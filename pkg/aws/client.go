package aws

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	log "github.com/sirupsen/logrus"
)

const (
	// BillingDateFormat is the layout for parsing the AWS date format of 'yyyymmdd'.
	BillingDateFormat = "20060102"

	// ManifestSuffix is the extension of AWS Usage Data manifests.
	ManifestSuffix = ".json"

	// defaultS3Region is used to make the API call used to determine a bucket's region.
	defaultS3Region = "us-east-1"

	// maxS3Keys is the maximum amount of keys to be returned by a single S3
	// list objects API response
	maxS3Keys = 200
)

type ManifestRetriever interface {
	RetrieveManifests() ([]*Manifest, error)
}

type manifestRetriever struct {
	logger log.FieldLogger
	s3API  s3iface.S3API
	bucket string
	prefix string
}

func NewManifestRetriever(logger log.FieldLogger, region, bucket, prefix, caBundlePath string) (ManifestRetriever, error) {
	var (
		proxy                 string
		useProxyConfiguration bool
		transport             http.Transport
	)
	// check whether the @caBundlePath is non-empty, and a valid
	// path to a CA bundle. If non-empty, we need to load that
	// certificate bundle into the http.Transport object that
	// we are building up. The current implementation expects
	// that the $HTTPS_PROXY environment variable is non-empty
	// when the @caBundlePath has been provided. In future, we
	// most likely just want to pass something like a *ProxyConfig
	// parameter and reference the fields from that object to
	// do the heavy-lifting. If the @caBundlePath is empty, then
	// check if the $HTTP_PROXY environment variable has a value,
	// which we will use when registering the custom http client.
	if caBundlePath != "" {
		if _, err := os.Stat(caBundlePath); err != nil {
			return nil, fmt.Errorf("failed to stat the %s path: %v", caBundlePath, err)
		}
		caBundle, err := ioutil.ReadFile(caBundlePath)
		if err != nil {
			return nil, fmt.Errorf("failed to load the trusted CA bundle: %v", err)
		}
		caRoot := x509.NewCertPool()
		caRoot.AppendCertsFromPEM(caBundle)
		transport.TLSClientConfig = &tls.Config{
			RootCAs: caRoot,
		}

		proxy = os.Getenv("HTTPS_PROXY")
		if proxy == "" {
			return nil, fmt.Errorf("expected the $HTTPS_PROXY to be non-empty when a trustbundle has been provided")
		}
		useProxyConfiguration = true
	} else {
		proxy = os.Getenv("HTTP_PROXY")
		if proxy != "" {
			useProxyConfiguration = true
		}
	}

	// start building up a barebones AWS configuration object
	config := &aws.Config{
		Region: aws.String(region),
	}

	// in the case where we grab a value from either of the HTTP*_PROXY
	// environment variables, we need to register a custom http client
	// that uses this proxy URL. We can also reasonably assume that the
	// HTTP*_PROXY environment variables are valid as the metering-operator,
	// which sets these values, reads directly from this cluster-scoped
	// proxy/cluster object, and the values stored in the object are
	// validated by the Cluster Networking Operator.
	if useProxyConfiguration {
		proxyURL, err := url.Parse(proxy)
		if err != nil {
			return nil, fmt.Errorf("failed to parse the proxy url '%s': %v", proxy, err)
		}
		transport.Proxy = http.ProxyURL(proxyURL)

		httpClient := &http.Client{
			Transport: &transport,
			Timeout:   time.Second * 60,
		}
		config.HTTPClient = httpClient
		logger.Debugf("registering a custom HTTP client for the AWS configuration")
	}

	session, err := session.NewSession(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create a new aws session: %v", err)
	}
	client := s3.New(session)

	return &manifestRetriever{
		logger: logger,
		s3API:  client,
		bucket: bucket,
		prefix: prefix,
	}, nil
}

// RetrieveManifests downloads the billing manifest for the given bucket and
// prefix. It includes only the top level manifest files, and ignores manifest
// files that are within assemblyId subdirectories, as the top level manifest
// points to the directory containing the most up to date billing report data.
func (r *manifestRetriever) RetrieveManifests() ([]*Manifest, error) {
	// ensure that there is a slash at end of location
	prefix := r.prefix
	if len(prefix) == 0 {
		prefix = "/"
	} else if prefix[len(prefix)-1] != '/' {
		prefix += "/"
	}
	logger := r.logger.WithFields(log.Fields{
		"bucket": r.bucket,
		"prefix": prefix,
	})

	var (
		manifests   []*Manifest
		manifestErr error
		page        int
	)
	pageFn := func(out *s3.ListObjectsV2Output, lastPage bool) bool {
		page++
		filteredKeys := r.filterObjects(prefix, out.Contents)
		if len(filteredKeys) == 0 {
			logger.Debugf("page %d had no manifests", page)
			return true
		}

		for _, key := range filteredKeys {
			logger.WithField("key", key).Debugf("retrieving manifest")
			manifest, err := retrieveManifest(r.s3API, r.bucket, key)
			if err != nil {
				manifestErr = fmt.Errorf("failed to get the manifest from the bucket '%s' with the key '%s': %v", r.bucket, key, err)
				return false
			}
			manifests = append(manifests, manifest)
		}

		return true
	}

	// list all in <report-prefix>/<report-name>/ of bucket
	err := r.s3API.ListObjectsV2Pages(&s3.ListObjectsV2Input{
		Bucket:  aws.String(r.bucket),
		Prefix:  aws.String(prefix),
		MaxKeys: aws.Int64(maxS3Keys),
	}, pageFn)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve the AWS billing report keys: %v", err)
	}
	if manifestErr != nil {
		return nil, manifestErr
	}

	return manifests, nil
}

func (r *manifestRetriever) filterObjects(prefix string, objects []*s3.Object) []string {
	var keys []string
	for _, obj := range objects {
		key := *obj.Key

		// only look for manifest files
		if !strings.HasSuffix(key, ManifestSuffix) {
			continue
		}

		// We're looking for the top-level manifest for a given time range.
		// These manifests are copies of manifests within the assemblyId
		// directories, and are updated everytime a report for the given time
		// period is run. We use these to determine the "most up to date" set
		// of data.
		// We're looking for manifests in the following format:
		// <report-prefix>/<report-name>/YYYYMMDD-YYYYMMDD/<report-name>-Manifest.json
		// We ignore the following manifests:
		// <report-prefix>/<report-name>/YYYYMMDD-YYYYMMDD/<assemblyId>/<report-name>-Manifest.json

		// Strip off <report-prefix>/<report-name>
		trimmedPath := strings.TrimPrefix(key, prefix)
		// manifestDir will be <YYYYMMDD-YYYYMMDD>/<assemblyId> or <YYYYMMDD-YYYYMMDD>
		// The latter is what we're looking for (without the assemblyId subdir)
		manifestDir := path.Dir(trimmedPath)
		// assemblyDir will be empty if manifestDir is without an assemblyId
		// subdirectory: <YYYYMMDD-YYYYMMDD>
		assemblyDir, _ := path.Split(manifestDir)
		// If there's another directory, it isn't the top-level manifest.
		if assemblyDir != "" {
			continue
		}
		// If we've gotten this far, then "key" is a top-level manifest that we
		// care about.
		keys = append(keys, key)
	}
	return keys
}

// retrieveManifest retrieves a manifest from the given bucket and key.
func retrieveManifest(client s3iface.S3API, bucket, key string) (*Manifest, error) {
	obj, err := client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	defer obj.Body.Close()

	decoder := json.NewDecoder(obj.Body)

	var manifest Manifest
	err = decoder.Decode(&manifest)
	if err != nil {
		return nil, err
	}
	return &manifest, err
}
