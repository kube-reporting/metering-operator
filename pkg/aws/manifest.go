package aws

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"

	cb "github.com/coreos-inc/kube-chargeback/pkg/chargeback"
)

const (
	// BillingDateFormat is the layout for parsing the AWS date format of 'yyyymmdd'.
	BillingDateFormat = "20060102"

	// ManifestSuffix is the extension of AWS Usage Data manifests.
	ManifestSuffix = ".json"
)

var (
	// S3Client allows mocks to be injected for testing.
	S3Client s3iface.S3API
)

// Manifest is a representation of the file AWS provides with metadata for current usage information.
type Manifest struct {
	AssemblyID string `json:"assemblyId"`
	Account    string `json:"account"`
	Columns    []struct {
		Category string `json:"category"`
		Name     string `json:"name"`
	} `json:"columns"`
	Charset       string `json:"charset"`
	Compression   string `json:"compression"`
	ContentType   string `json:"contentType"`
	ReportID      string `json:"reportId"`
	ReportName    string `json:"reportName"`
	BillingPeriod struct {
		Start string `json:"start"`
		End   string `json:"end"`
	} `json:"billingPeriod"`
	Bucket                 string   `json:"bucket"`
	ReportKeys             []string `json:"reportKeys"`
	AdditionalArtifactKeys []string `json:"additionalArtifactKeys"`
}

// Paths returns the directories containing usage data. The result will be free of duplicates.
func (m Manifest) Paths() (paths []string) {
	pathMap := map[string]bool{}
	for _, key := range m.ReportKeys {
		dirPath := filepath.Dir(key)
		pathMap[dirPath] = true
	}

	for path := range pathMap {
		paths = append(paths, path)
	}
	return
}

// RetrieveManifests downloads the billing manifest for the given bucket, prefix, and report name.
func RetrieveManifests(bucket, reportPrefix, reportName string, rng cb.Range) ([]Manifest, error) {
	client := getS3Client()

	// list all in <report-prefix>/<report-name>/ of bucket
	prefix := fmt.Sprintf("%s/%s/", reportPrefix, reportName)
	dateRngs, err := client.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		return nil, fmt.Errorf("could not list retrieve AWS billing report keys: %v", err)
	}

	// use UTC for all times
	rng.Start, rng.End = rng.Start.UTC(), rng.End.UTC()

	manifests := []Manifest{}
	for _, obj := range dateRngs.Contents {
		key := *obj.Key
		suffix := strings.TrimPrefix(key, prefix)
		kParts := strings.SplitN(suffix, "/", 3)
		rngStr, file := kParts[0], kParts[1]

		// only look for manifest files
		if !strings.HasSuffix(file, ManifestSuffix) {
			continue
		}

		if dirRng, err := rngFromDirName(rngStr); err != nil {
			return nil, fmt.Errorf("failed to determine range for '%s': %v", *obj.Key, err)
		} else if !dirRng.Within(rng.Start) && !dirRng.Within(rng.End) {
			// directory is not within range
			continue
		}

		manifest, err := retrieveManifest(client, bucket, key)
		if err != nil {
			return nil, fmt.Errorf("can't get manifest from bucket '%s' with key '%s': %v", bucket, key, err)
		}
		manifests = append(manifests, manifest)

	}
	return manifests, nil
}

// retrieveManifest retrieves a manifest from the given bucket and key.
func retrieveManifest(client s3iface.S3API, bucket, key string) (Manifest, error) {
	get := &s3.GetObjectInput{Bucket: aws.String(bucket), Key: aws.String(key)}
	obj, err := client.GetObject(get)
	if err != nil {
		return Manifest{}, err
	}

	data, err := ioutil.ReadAll(obj.Body)
	if err != nil {
		return Manifest{}, err
	}

	manifest := Manifest{}
	err = json.Unmarshal(data, &manifest)
	return manifest, err
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
func getS3Client() s3iface.S3API {
	if S3Client == nil {
		awsSession := session.Must(session.NewSession())
		S3Client = s3.New(awsSession)
	}
	return S3Client
}
