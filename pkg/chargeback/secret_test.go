package chargeback

import (
	"os"
	"testing"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/coreos-inc/kube-chargeback/pkg/aws"
	cb "github.com/coreos-inc/kube-chargeback/pkg/chargeback/v1"
)

const (
	// S3BucketVar is the environment variable holding the S3 bucket used for testing.
	S3BucketVar = "TEST_S3_BUCKET"

	// KubeconfigVar is the environment variable holding the Kubeconfig bucket used for testing.
	KubeconfigVar = "TEST_KUBECONFIG"

	// NamespaceVar is the environment variable holding the cluster Namespace used for testing.
	NamespaceVar = "TEST_NAMESPACE"
)

func TestBucketSecret(t *testing.T) {
	kubeClientCfg, bucket, namespace := setupBucketSecretTest(t)
	op := operator(t, kubeClientCfg, namespace)

	bucketSecret, err := op.getBucketSecret(bucket)
	if err != nil {
		t.Fatalf("failed to retrieve bucket secret: %v", err)
	}

	if len(bucketSecret.AWSAccessKeyID) == 0 {
		t.Error("returned secret did not include access key")
	} else {
		t.Logf("Found AWS Access Key: %s", bucketSecret.AWSAccessKeyID)
	}
}

func TestRetrieveManifestWithCredential(t *testing.T) {
	kubeClientCfg, bucket, namespace := setupBucketSecretTest(t)
	op := operator(t, kubeClientCfg, namespace)

	bucketSecret, err := op.getBucketSecret(bucket)
	if err != nil {
		t.Fatalf("failed to retrieve bucket secret: %v", err)
	}

	_, err = aws.RetrieveManifests(bucket, "NoSuchKey", cb.Range{}, bucketSecret.AWSCreds())
	if err != nil {
		t.Errorf("encountered error using created credentials: %v", err)
	}
}

func operator(t *testing.T, kubeClientCfg *rest.Config, namespace string) *Chargeback {
	cfg := Config{
		ClientCfg: kubeClientCfg,
		Namespace: namespace,
	}
	op, err := New(cfg)
	if err != nil {
		t.Fatalf("unable to configure Chargeback operator: %v", err)
	}
	return op
}

func setupBucketSecretTest(t *testing.T) (kubeCfg *rest.Config, s3Bucket, namespace string) {
	var kubeconfigPath string
	kubeconfigPath, s3Bucket, namespace = testEnvVars(t)

	var err error
	if kubeCfg, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath); err != nil {
		t.Fatalf("Could not configure Kubernetes client: %v", err)
	}
	return
}

func testEnvVars(t *testing.T) (kubeconfigPath, s3Bucket, namespace string) {
	skipMissingEnv := func(what, env string) {
		t.Skipf("To test Hive, set the environment variable '%s' to the %s to be used.", env, what)
		t.SkipNow()
	}

	var ok bool
	if kubeconfigPath, ok = os.LookupEnv(KubeconfigVar); !ok {
		skipMissingEnv("kubeconfig path", KubeconfigVar)
	} else if s3Bucket, ok = os.LookupEnv(S3BucketVar); !ok {
		skipMissingEnv("S3 bucket", S3BucketVar)
	} else if namespace, ok = os.LookupEnv(NamespaceVar); !ok {
		skipMissingEnv("Kubernetes namespace", NamespaceVar)
	}
	return
}
