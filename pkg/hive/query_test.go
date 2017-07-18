package hive

import (
	"os"
	"testing"
)

const (
	// HiveHostVar is environment variable holding the Hive host used for testing.
	HiveHostVar = "TEST_HIVE_HOST"

	// S3BucketVar is environment variable holding the S3 bucket used for Hive testing.
	S3BucketVar = "TEST_S3_BUCKET"

	// S3LocationVar is environment variable holding the S3 location used for Hive testing.
	S3LocationVar = "TEST_S3_LOCATION"
)

func setupHiveTest(t *testing.T) (hiveHost, s3Bucket, s3Pre***REMOVED***x string) {
	var ok bool
	if hiveHost, ok = os.LookupEnv(HiveHostVar); !ok {
		skipMissingEnv(t, "hive host", HiveHostVar)
	} ***REMOVED*** if s3Bucket, ok = os.LookupEnv(S3BucketVar); !ok {
		skipMissingEnv(t, "S3 bucket", S3BucketVar)
	} ***REMOVED*** if s3Pre***REMOVED***x, ok = os.LookupEnv(S3LocationVar); !ok {
		skipMissingEnv(t, "S3 location", S3LocationVar)
	}
	return
}

func skipMissingEnv(t *testing.T, what, env string) {
	t.Skipf("To test Hive, set the environment variable '%s' to the %s to be used.", env, what)
	t.SkipNow()
}
