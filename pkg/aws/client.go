package aws

import (
	"encoding/json"
	"fmt"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
)

const (
	// BillingDateFormat is the layout for parsing the AWS date format of 'yyyymmdd'.
	BillingDateFormat = "20060102"

	// ManifestSuffix is the extension of AWS Usage Data manifests.
	ManifestSuffix = ".json"

	// defaultS3Region is used to make the API call used to determine a bucket's region.
	defaultS3Region = "us-east-1"
)

type ManifestRetriever interface {
	RetrieveManifests() ([]*Manifest, error)
}

type manifestRetriever struct {
	s3API          s3iface.S3API
	bucket, prefix string
}

func NewManifestRetriever(bucket, prefix string) (ManifestRetriever, error) {
	client, err := getS3Client(bucket, nil)
	if err != nil {
		return nil, err
	}
	return &manifestRetriever{
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

	var keys []string
	pageFn := func(out *s3.ListObjectsV2Output, lastPage bool) bool {
		keys = append(keys, r.filterObjects(prefix, out.Contents)...)
		return true
	}

	// list all in <report-prefix>/<report-name>/ of bucket
	err := r.s3API.ListObjectsV2Pages(&s3.ListObjectsV2Input{
		Bucket: aws.String(r.bucket),
		Prefix: aws.String(prefix),
	}, pageFn)
	if err != nil {
		return nil, fmt.Errorf("could not list retrieve AWS billing report keys: %v", err)
	}

	var manifests []*Manifest

	for _, key := range keys {
		manifest, err := retrieveManifest(r.s3API, r.bucket, key)
		if err != nil {
			return nil, fmt.Errorf("can't get manifest from bucket '%s' with key '%s': %v", r.bucket, key, err)
		}
		manifests = append(manifests, manifest)
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

// getS3Client returns the singleton client.
func getS3Client(bucket string, creds *credentials.Credentials) (s3iface.S3API, error) {
	awsSession := session.Must(session.NewSession())
	if creds != nil {
		awsSession.Config.Credentials = creds
	}

	tmpClient := s3.New(awsSession, aws.NewConfig().WithRegion(defaultS3Region))
	region, err := retrieveRegion(tmpClient, bucket)
	if err != nil {
		return nil, err
	}
	awsSession.Config.Region = &region

	return s3.New(awsSession), nil
}

// retrieveRegion performs a request to determine the region the bucket has been created in.
func retrieveRegion(client s3iface.S3API, bucket string) (string, error) {
	bucketResp, err := client.GetBucketLocation(&s3.GetBucketLocationInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		return "", fmt.Errorf("failed to retrieve bucket region: %v", err)
	}

	if bucketResp.LocationConstraint == nil || *bucketResp.LocationConstraint == "" {
		return "us-east-1", nil
	}
	return *bucketResp.LocationConstraint, nil
}
