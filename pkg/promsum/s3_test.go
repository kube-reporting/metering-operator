package promsum

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
)

const (
	// TestActualS3EnvVar is the name of the environment variable that if set will test against real S3 not a mock.
	TestActualS3EnvVar = "TEST_ACTUAL_S3"
)

func TestMockS3StoreReadWrite(t *testing.T) {
	bucket, path := "promsum-test", "test-path"
	store := NewS3Store(bucket, path)

	mockS3 := newMockS3()
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

func newMockS3() *mockS3 {
	return &mockS3{
		buckets: map[string]map[string][]byte{},
	}
}

// mockS3 mimics an S3 blob store for testing.
type mockS3 struct {
	sync.RWMutex
	buckets map[string]map[string][]byte
	s3iface.S3API
}

func (m *mockS3) NewBucket(name string) {
	m.buckets[name] = map[string][]byte{}
}

func (m *mockS3) PutObject(in *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
	data, err := ioutil.ReadAll(in.Body)
	if err != nil {
		return nil, err
	}

	m.Lock()
	defer m.Unlock()

	bucket, ok := m.buckets[*in.Bucket]
	if !ok {
		bucket = map[string][]byte{}
		m.buckets[*in.Bucket] = bucket
	}

	bucket[*in.Key] = data
	return &s3.PutObjectOutput{}, nil
}

func (m *mockS3) GetObject(in *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
	m.RLock()
	defer m.RUnlock()

	bucket, ok := m.buckets[*in.Bucket]
	if !ok {
		return nil, fmt.Errorf("bucket '%s' does not exist", *in.Bucket)
	}

	data, ok := bucket[*in.Key]
	if !ok {
		return nil, fmt.Errorf("key '%s' does not exist in bucket '%s'", *in.Key, *in.Bucket)
	}

	return &s3.GetObjectOutput{
		Body: ioutil.NopCloser(bytes.NewBuffer(data)),
	}, nil
}

func (m *mockS3) ListObjectsV2(in *s3.ListObjectsV2Input) (*s3.ListObjectsV2Output, error) {
	bucket, ok := m.buckets[*in.Bucket]
	if !ok {
		return nil, fmt.Errorf("bucket '%s' does not exist", *in.Bucket)
	}

	var objects []*s3.Object
	for key := range bucket {
		if strings.HasPrefix(key, *in.Prefix) {
			objKey := key
			obj := &s3.Object{Key: &objKey}
			objects = append(objects, obj)
		}
	}
	out := new(s3.ListObjectsV2Output)
	out.SetContents(objects)

	return out, nil
}
