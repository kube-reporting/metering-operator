package e2e

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kube-reporting/metering-operator/test/reportingframework"
)

func testEnsureS3BucketIsDeleted(t *testing.T, rf *reportingframework.ReportingFramework) {
	mc, err := rf.MeteringClient.MeteringConfigs(rf.Namespace).Get(context.Background(), "operator-metering", metav1.GetOptions{})
	require.False(t, apierrors.IsNotFound(err), "expected querying for the operator-metering MeteringConfig custom resource in the %s namespace would produce no error", rf.Namespace)

	s, err := rf.KubeClient.CoreV1().Secrets(rf.Namespace).Get(context.Background(), "aws-creds", metav1.GetOptions{})
	require.NoErrorf(t, err, "expected querying for the aws-creds secret in the %s namespace would produce no error", rf.Namespace)

	region := mc.Spec.Storage.Hive.S3.Region
	session, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(string(s.Data["aws-access-key-id"]), string(s.Data["aws-secret-access-key"]), ""),
	})
	require.NoError(t, err, "failed to create the s3 service clientset")
	client := s3.New(session)

	bucket := mc.Spec.Storage.Hive.S3.Bucket
	iter := s3manager.NewDeleteListIterator(client, &s3.ListObjectsInput{
		Bucket: aws.String(bucket),
	})
	err = s3manager.NewBatchDeleteWithClient(client).Delete(aws.BackgroundContext(), iter)
	require.NoError(t, err, "expected deleting the objects in an s3 bucket would produce no error")

	_, err = client.DeleteBucket(&s3.DeleteBucketInput{Bucket: aws.String(bucket)})
	require.NoErrorf(t, err, "failed to delete the %s bucket", bucket)
	t.Logf("Deleted the %s bucket in the %s region", bucket, region)
}
