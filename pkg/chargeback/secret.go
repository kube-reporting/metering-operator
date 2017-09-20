package chargeback

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/credentials"
	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// BucketConfigMapName is the name of the ConfigMap resource containing the mapping between buckets and secrets.
	BucketConfigMapName = "buckets"

	// BucketSecret keys
	AWS_ID_STR            = "AWS_ACCESS_KEY_ID"
	AWS_KEY_STR           = "AWS_SECRET_ACCESS_KEY"
	AWS_SESSION_STR       = "AWS_SESSION_TOKEN"
	AWS_CRED_PROVIDER_STR = "CredentialsProvider"
)

// NewBucketSecret extracts credentials from a ConfigMap's data.
func NewBucketSecret(data map[string]string) (secret BucketSecret) {
	secret.AWSAccessKeyID, _ = data[AWS_ID_STR]
	secret.AWSSecretAccessKey, _ = data[AWS_KEY_STR]
	secret.AWSSessionToken, _ = data[AWS_SESSION_STR]
	secret.AWSCredentialsProvider, _ = data[AWS_CRED_PROVIDER_STR]
	return
}

// BucketSecret is a Secret containing credentials for Chargeback.
type BucketSecret struct {
	AWSAccessKeyID         string
	AWSSecretAccessKey     string
	AWSSessionToken        string
	AWSCredentialsProvider string
}

// AWSCreds returns credentials use to authenticate with AWS.
func (s BucketSecret) AWSCreds() *credentials.Credentials {
	return credentials.NewStaticCredentials(s.AWSAccessKeyID, s.AWSSecretAccessKey, s.AWSSessionToken)
}

func (c *Chargeback) getBucketSecret(bucket string) (BucketSecret, error) {
	bucketConfig, err := c.getBucketSecretConfig()
	if err != nil {
		return BucketSecret{}, err
	}

	secretName, ok := bucketConfig[bucket]
	if !ok {
		return BucketSecret{}, fmt.Errorf("did not find configuration for bucket '%s' in ConfigMap '%s",
			bucket, c.bucketConfigMapName())
	}

	secret, err := c.core.Secrets(c.namespace).Get(secretName, v1meta.GetOptions{})
	if err != nil {
		return BucketSecret{}, fmt.Errorf("couldn't retrieve Secret '%s' for bucket '%s", secretName, bucket)
	}

	data := stringValues(secret.Data)
	return NewBucketSecret(data), nil
}

func (c *Chargeback) getBucketSecretConfig() (map[string]string, error) {
	cfg, err := c.core.ConfigMaps(c.namespace).Get(BucketConfigMapName, v1meta.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("could not retrieve bucket to secret mapping from ConfigMap '%s': %v",
			c.bucketConfigMapName(), err)
	}

	return cfg.Data, nil
}

func (c Chargeback) bucketConfigMapName() string {
	return fmt.Sprintf("%s/%s", c.namespace, BucketConfigMapName)
}

func stringValues(in map[string][]byte) map[string]string {
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = string(v)
	}
	return out
}
