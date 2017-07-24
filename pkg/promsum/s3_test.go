package promsum

import (
	"os"
	"testing"

	s3 "github.com/coreos-inc/kube-chargeback/pkg/aws/s3_test"
)

const (
	// TestActualS3EnvVar is the name of the environment variable that if set will test against real S3 not a mock.
	TestActualS3EnvVar = "TEST_ACTUAL_S3"
)

func TestMockS3StoreReadWrite(t *testing.T) {
	bucket, path := "promsum-test", "test-path"
	store := NewS3Store(bucket, path)

	mockS3 := s3.NewMockS3()
	mockS3.NewBucket(bucket)
	store.s3 = mockS3

	testStoreReadWrite(t, store)
}

func TestActualS3StoreReadWrite(t *testing.T) {
	bucket, path := "promsum-test", "test-path"
	store := NewS3Store(bucket, path)

	// replace S3 client with mock if missing envvar
	if _, set := os.LookupEnv(TestActualS3EnvVar); !set {
		t.Skipf("To test against an actual S3 endpoint set the environment variable '%s'.", TestActualS3EnvVar)
		return
	}

	testStoreReadWrite(t, store)
}
