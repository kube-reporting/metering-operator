package e2e

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kube-reporting/metering-operator/test/deployframework"
	"github.com/kube-reporting/metering-operator/test/testhelpers"
)

const (
	nonHDFSTargetPodCount = 5

	secretName            = "aws-creds"
	testingNamespaceLabel = "metering-testing-ns"
)

func testManualMeteringInstall(
	t *testing.T,
	testCaseName,
	namespacePrefix,
	meteringOperatorImageRepo,
	meteringOperatorImageTag,
	manifestFilename,
	catalogSourceName,
	catalogSourceNamespace,
	subscriptionChannel,
	testOutputPath string,
	expectInstallErrMsg []string,
	expectInstallErr bool,
	preInstallFunc PreInstallFunc,
	testInstallFunctions []InstallTestCase,
) {
	// create a directory used to store the @testCaseName container and resource logs
	testCaseOutputBaseDir := filepath.Join(testOutputPath, testCaseName)
	err := os.Mkdir(testCaseOutputBaseDir, 0777)
	assert.NoError(t, err, "creating the test case output directory should produce no error")

	testFuncNamespace := fmt.Sprintf("%s-%s", namespacePrefix, strings.ToLower(testCaseName))
	if len(testFuncNamespace) > kubeNamespaceCharLimit {
		require.Fail(t, "The length of the test function namespace exceeded the kube namespace limit of %d characters", kubeNamespaceCharLimit)
	}

	mc, err := testhelpers.DecodeMeteringConfigManifest(repoPath, testMeteringConfigManifestsPath, manifestFilename)
	require.NoError(t, err, "failed to successfully decode the YAML MeteringConfig manifest")

	var envVars []string
	for _, installFunc := range testInstallFunctions {
		if len(installFunc.ExtraEnvVars) != 0 {
			envVars = append(envVars, installFunc.ExtraEnvVars...)
		}
	}

	deployerCtx, err := df.NewDeployerCtx(
		testFuncNamespace,
		meteringOperatorImageRepo,
		meteringOperatorImageTag,
		reportingOperatorImageRepo,
		reportingOperatorImageTag,
		catalogSourceName,
		catalogSourceNamespace,
		subscriptionChannel,
		testCaseOutputBaseDir,
		envVars,
		mc.Spec,
	)
	require.NoError(t, err, "creating a new deployer context should produce no error")
	defer deployerCtx.LoggerOutFile.Close()

	if preInstallFunc != nil {
		err = preInstallFunc(deployerCtx)
		require.NoError(t, err, "expected no error while running any pre-install functions")
	}

	rf, err := deployerCtx.Setup(deployerCtx.Deployer.InstallOLM, expectInstallErr)
	canSafelyRunTest := testhelpers.AssertCanSafelyRunReportingTests(t, err, expectInstallErr, expectInstallErrMsg)

	if canSafelyRunTest {
		for _, installFunc := range testInstallFunctions {
			installFunc := installFunc
			t := t

			// namespace got deleted early when running t.Parallel()
			// so re-running without that specified
			t.Run(installFunc.Name, func(t *testing.T) {
				installFunc.TestFunc(t, rf)
			})

			deployerCtx.Logger.Infof("The %s test has finished running", installFunc.Name)
		}
	}

	err = deployerCtx.Teardown(deployerCtx.Deployer.UninstallOLM)
	assert.NoError(t, err, "capturing logs and uninstalling metering should produce no error")
}

func s3InstallFunc(ctx *deployframework.DeployerCtx) error {
	// The default ctx.TargetPodsCount value assumes that HDFS
	// is being used a storage backend, so we need to decrement
	// that value to ensure we're not going to poll forever waiting
	// for Pods that will not be created by the metering-ansible-operator
	ctx.TargetPodsCount = nonHDFSTargetPodCount

	// Before we can create the AWS credentials secret in the ctx.Namespace, we need to ensure
	// that namespace has been created before attempting to query for a resource that does not exist.
	_, err := ctx.Client.CoreV1().Namespaces().Get(context.Background(), ctx.Namespace, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return nil
	}
	if apierrors.IsNotFound(err) {
		n := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: ctx.Namespace,
				Labels: map[string]string{
					"name": ctx.Namespace + "-" + testingNamespaceLabel,
				},
			},
		}
		_, err = ctx.Client.CoreV1().Namespaces().Create(context.Background(), n, metav1.CreateOptions{})
		if err != nil {
			return err
		}
		ctx.Logger.Debugf("Created the %s namespace", ctx.Namespace)
	}

	// Attempt to mirror the AWS credentials used to provision the IPI-based AWS cluster
	// from the kube-system namespace, to the ctx.Namespace that the test context is configured
	// to use. In the case where this secret containing the base64-encrypted access key id and
	// secret access key does not exist, replace all instances of the `_` delimiter with `_`,
	// and create that Secret resource in the ctx.Namespace. In the case where the secret already
	// exists, exit early.
	_, err = ctx.Client.CoreV1().Secrets(ctx.Namespace).Get(context.Background(), secretName, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return nil
	}
	if apierrors.IsNotFound(err) {
		s, err := ctx.Client.CoreV1().Secrets("kube-system").Get(context.Background(), secretName, metav1.GetOptions{})
		if err != nil {
			return err
		}

		key, ok := s.Data["aws_access_key_id"]
		if !ok {
			return fmt.Errorf("failed to retrieve the AWS access key ID from the secret data")
		}
		access, ok := s.Data["aws_secret_access_key"]
		if !ok {
			return fmt.Errorf("failed to retrieve the AWS secret access key from the secret data")
		}

		// Note: we need to do some light translation of data keys as the helm charts
		// expect the data keys to use the `-` delimiter, e.g. `data["aws_access_key_id"]`
		// becomes `data["aws-access-key-id"]`.
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: ctx.Namespace,
			},
			Data: map[string][]byte{
				"aws-access-key-id":     key,
				"aws-secret-access-key": access,
			},
		}

		_, err = ctx.Client.CoreV1().Secrets(ctx.Namespace).Create(context.Background(), secret, metav1.CreateOptions{})
		if err != nil {
			return err
		}
		ctx.Logger.Debugf("Created the %s AWS credentials secret in the %s namespace", secretName, ctx.Namespace)
	}

	return nil
}
