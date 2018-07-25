package aws

import (
	"encoding/json"
	"fmt"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
)

const (
	// BillingDateFormat is the layout for parsing the AWS date format of 'yyyymmdd'.
	BillingDateFormat = "20060102"

	// ManifestSuf***REMOVED***x is the extension of AWS Usage Data manifests.
	ManifestSuf***REMOVED***x = ".json"

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
	s3API          s3iface.S3API
	bucket, pre***REMOVED***x string
}

func NewManifestRetriever(region, bucket, pre***REMOVED***x string) ManifestRetriever {
	awsSession := session.Must(session.NewSession())
	client := s3.New(awsSession, aws.NewCon***REMOVED***g().WithRegion(region))
	return &manifestRetriever{
		s3API:  client,
		bucket: bucket,
		pre***REMOVED***x: pre***REMOVED***x,
	}
}

// RetrieveManifests downloads the billing manifest for the given bucket and
// pre***REMOVED***x. It includes only the top level manifest ***REMOVED***les, and ignores manifest
// ***REMOVED***les that are within assemblyId subdirectories, as the top level manifest
// points to the directory containing the most up to date billing report data.
func (r *manifestRetriever) RetrieveManifests() ([]*Manifest, error) {
	// ensure that there is a slash at end of location
	pre***REMOVED***x := r.pre***REMOVED***x
	if len(pre***REMOVED***x) == 0 {
		pre***REMOVED***x = "/"
	} ***REMOVED*** if pre***REMOVED***x[len(pre***REMOVED***x)-1] != '/' {
		pre***REMOVED***x += "/"
	}

	var manifests []*Manifest
	var manifestErr error
	pageFn := func(out *s3.ListObjectsV2Output, lastPage bool) bool {
		keys := r.***REMOVED***lterObjects(pre***REMOVED***x, out.Contents)

		for _, key := range keys {
			manifest, err := retrieveManifest(r.s3API, r.bucket, key)
			if err != nil {
				manifestErr = fmt.Errorf("can't get manifest from bucket '%s' with key '%s': %v", r.bucket, key, err)
				return false
			}
			manifests = append(manifests, manifest)
		}

		return true
	}

	// list all in <report-pre***REMOVED***x>/<report-name>/ of bucket
	err := r.s3API.ListObjectsV2Pages(&s3.ListObjectsV2Input{
		Bucket:  aws.String(r.bucket),
		Pre***REMOVED***x:  aws.String(pre***REMOVED***x),
		MaxKeys: aws.Int64(maxS3Keys),
	}, pageFn)
	if err != nil {
		return nil, fmt.Errorf("could not list retrieve AWS billing report keys: %v", err)
	}
	if manifestErr != nil {
		return nil, manifestErr
	}

	return manifests, nil
}

func (r *manifestRetriever) ***REMOVED***lterObjects(pre***REMOVED***x string, objects []*s3.Object) []string {
	var keys []string
	for _, obj := range objects {
		key := *obj.Key

		// only look for manifest ***REMOVED***les
		if !strings.HasSuf***REMOVED***x(key, ManifestSuf***REMOVED***x) {
			continue
		}

		// We're looking for the top-level manifest for a given time range.
		// These manifests are copies of manifests within the assemblyId
		// directories, and are updated everytime a report for the given time
		// period is run. We use these to determine the "most up to date" set
		// of data.
		// We're looking for manifests in the following format:
		// <report-pre***REMOVED***x>/<report-name>/YYYYMMDD-YYYYMMDD/<report-name>-Manifest.json
		// We ignore the following manifests:
		// <report-pre***REMOVED***x>/<report-name>/YYYYMMDD-YYYYMMDD/<assemblyId>/<report-name>-Manifest.json

		// Strip off <report-pre***REMOVED***x>/<report-name>
		trimmedPath := strings.TrimPre***REMOVED***x(key, pre***REMOVED***x)
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
