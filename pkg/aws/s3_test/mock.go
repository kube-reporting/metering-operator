package s3_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
)

func NewMockS3() *MockS3 {
	return &MockS3{
		buckets: map[string]map[string][]byte{},
	}
}

// MockS3 mimics an S3 blob store for testing.
type MockS3 struct {
	sync.RWMutex
	buckets map[string]map[string][]byte
	s3iface.S3API
}

func (m *MockS3) NewBucket(name string) {
	m.buckets[name] = map[string][]byte{}
}

func (m *MockS3) PutObject(in *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
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

func (m *MockS3) GetObject(in *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
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

func (m *MockS3) ListObjectsV2(in *s3.ListObjectsV2Input) (*s3.ListObjectsV2Output, error) {
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
