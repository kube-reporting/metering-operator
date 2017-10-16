package aws

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"

	cb "github.com/coreos-inc/kube-chargeback/pkg/chargeback/v1"
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

// RetrieveManifests downloads the billing manifest for the given bucket, prefix, and report name.
func (r *manifestRetriever) RetrieveManifests() ([]*Manifest, error) {
	// ensure that there is a slash at end of location
	prefix := r.prefix
	if prefix[len(prefix)-1] != '/' {
		prefix += "/"
	}

	// list all in <report-prefix>/<report-name>/ of bucket
	dateRngs, err := r.s3API.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(r.bucket),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		return nil, fmt.Errorf("could not list retrieve AWS billing report keys: %v", err)
	}

	var manifests []*Manifest
	for _, obj := range dateRngs.Contents {
		key := *obj.Key

		// only look for manifest files
		if !strings.HasSuffix(key, ManifestSuffix) {
			continue
		}

		manifest, err := retrieveManifest(r.s3API, r.bucket, key)
		if err != nil {
			return nil, fmt.Errorf("can't get manifest from bucket '%s' with key '%s': %v", r.bucket, key, err)
		}
		manifests = append(manifests, manifest)

	}
	return manifests, nil
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

// rangeFromDirName returns the start and end times encoded in an AWS usage record directory name.
func rngFromDirName(dir string) (rng cb.Range, err error) {
	rngParts := strings.Split(dir, "-")
	if len(rngParts) != 2 {
		err = errors.New("expected only 1 instance of '-'")
	} else if rng.Start, err = parseBillingDate(rngParts[0]); err != nil {
		err = fmt.Errorf("can't determine start: %v", err)
	} else if rng.End, err = parseBillingDate(rngParts[1]); err != nil {
		err = fmt.Errorf("can't determine end: %v", err)
	}

	if err != nil {
		err = fmt.Errorf("expected format for billing dates is 'yyyymmdd-yyyymmdd', given '%s': %v", dir, err)
	}
	return
}

// parseBillingDate returns a Time based on a date string formatted in the pattern 'yyyymmdd'. Times will be UTC.
func parseBillingDate(dateStr string) (t time.Time, err error) {
	if t, err = time.Parse(BillingDateFormat, dateStr); err != nil {
		err = fmt.Errorf("failed to parse date from '%s': %v", dateStr, err)
	}
	return
}

// getS3Client returns the singleton client.
func getS3Client(bucket string, creds *credentials.Credentials) (s3iface.S3API, error) {
	awsSession := session.Must(session.NewSession())
	if creds != nil {
		awsSession.Config.Credentials = creds
	}

	var err error
	tmpClient := s3.New(awsSession, aws.NewConfig().WithRegion(defaultS3Region))
	if awsSession.Config.Region, err = retrieveRegion(tmpClient, bucket); err != nil {
		return nil, err
	}

	return s3.New(awsSession), nil
}

// retrieveRegion performs a request to determine the region the bucket has been created in.
func retrieveRegion(client s3iface.S3API, bucket string) (*string, error) {
	bucketResp, err := client.GetBucketLocation(&s3.GetBucketLocationInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve bucket region: %v", err)
	}

	if bucketResp == nil || bucketResp.LocationConstraint == nil {
		return nil, errors.New("bucket response or bucket name was nil")
	}
	return bucketResp.LocationConstraint, nil
}
