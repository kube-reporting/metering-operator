package e2e

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	corev1 "k8s.io/api/core/v1"
	storagev1beta1 "k8s.io/api/storage/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"

	"github.com/kube-reporting/metering-operator/pkg/deploy"
	"github.com/kube-reporting/metering-operator/test/deployframework"
	"github.com/kube-reporting/metering-operator/test/testhelpers"
)

const (
	nonHDFSTargetPodCount = 5

	secretName            = "aws-creds"
	testingNamespaceLabel = "metering-testing-ns"
	nfsTestingNamespace   = "nfs"
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

			t.Run(installFunc.Name, func(t *testing.T) {
				installFunc.TestFunc(t, rf)
			})

			deployerCtx.Logger.Infof("The %s test has finished running", installFunc.Name)
		}
	}

	err = deployerCtx.Teardown(deployerCtx.Deployer.UninstallOLM)
	assert.NoError(t, err, "capturing logs and uninstalling metering should produce no error")
}

// createTestingNamespace is a helper function that is responsible
// for creating a namespace with the @namespace metadata.name and
// contains the `name: "<namespace_prefix>-metering-testing-ns` label.
// During the teardown function of the hack/e2e.sh script, we search for any
// namespaces that match that label. Note: manually running the e2e suite
// specifying a list of go test flags does not ensure proper cleanup.
func createTestingNamespace(client kubernetes.Interface, namespace string) (*corev1.Namespace, error) {
	ns, err := client.CoreV1().Namespaces().Get(context.Background(), namespace, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return nil, err
	}
	if apierrors.IsNotFound(err) {
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: namespace,
				Labels: map[string]string{
					"name": namespacePrefix + "-" + testingNamespaceLabel,
				},
			},
		}
		_, err = client.CoreV1().Namespaces().Create(context.Background(), ns, metav1.CreateOptions{})
		if err != nil {
			return ns, nil
		}
	}
	return ns, nil
}

func createNFSProvisioner(ctx *deployframework.DeployerCtx) error {
	_, err := createTestingNamespace(ctx.Client, nfsTestingNamespace)
	if err != nil {
		return err
	}

	ctx.TargetPodsCount = nonHDFSTargetPodCount
	files := map[string]string{
		"pod":          "server.yaml",
		"service":      "service.yaml",
		"storageClass": "storageclass.yaml",
	}
	server := &corev1.Pod{}
	service := &corev1.Service{}
	storageClass := &storagev1beta1.StorageClass{}

	for name, file := range files {
		absFile := filepath.Join(repoPath, testNFSManifestPath, file)
		switch name {
		case "pod":
			if err := deploy.DecodeYAMLManifestToObject(absFile, server); err != nil {
				return err
			}
			ctx.Client.CoreV1().Pods(nfsTestingNamespace).Create(context.TODO(), server, metav1.CreateOptions{})
			if err != nil && !apierrors.IsAlreadyExists(err) {
				return err
			}
		case "service":
			if err := deploy.DecodeYAMLManifestToObject(absFile, service); err != nil {
				return err
			}
			ctx.Client.CoreV1().Services(nfsTestingNamespace).Create(context.TODO(), service, metav1.CreateOptions{})
			if err != nil && !apierrors.IsAlreadyExists(err) {
				return err
			}
		case "storageClass":
			if err := deploy.DecodeYAMLManifestToObject(absFile, storageClass); err != nil {
				return err
			}
			ctx.Client.StorageV1beta1().StorageClasses().Create(context.TODO(), storageClass, metav1.CreateOptions{})
			if err != nil && !apierrors.IsAlreadyExists(err) {
				return err
			}
		}
	}

	svc, err := ctx.Client.CoreV1().Services(nfsTestingNamespace).Get(context.TODO(), "nfs-service", metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		return err
	}
	if err != nil || svc.Spec.ClusterIP == "" {
		err = wait.Poll(3*time.Second, 5*time.Minute, func() (bool, error) {
			svc, err = ctx.Client.CoreV1().Services(ctx.Namespace).Get(context.TODO(), "nfs-service", metav1.GetOptions{})
			if apierrors.IsNotFound(err) {
				return false, err
			}
			if err != nil {
				return false, nil
			}
			return svc.Spec.ClusterIP != "", nil
		})
		if err != nil {
			return err
		}
	}

	nfsFile := "persistentvolume.yaml"
	nfsPv := &corev1.PersistentVolume{}
	absFile := filepath.Join(repoPath, testNFSManifestPath, nfsFile)
	if err := deploy.DecodeYAMLManifestToObject(absFile, nfsPv); err != nil {
		return err
	}

	nfsPv.Spec.NFS.Server = svc.Spec.ClusterIP
	_, err = ctx.Client.CoreV1().PersistentVolumes().Create(context.TODO(), nfsPv, metav1.CreateOptions{})
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}
	return nil
}

func s3InstallFunc(ctx *deployframework.DeployerCtx) error {
	// The default ctx.TargetPodsCount value assumes that HDFS
	// is being used a storage backend, so we need to decrement
	// that value to ensure we're not going to poll forever waiting
	// for Pods that will not be created by the metering-ansible-operator
	ctx.TargetPodsCount = nonHDFSTargetPodCount

	_, err := createTestingNamespace(ctx.Client, ctx.Namespace)
	if err != nil {
		return err
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

// customNodeSelectorFunc is a test helper function that adds a
// custom testing label to all of the worker nodes in a particular
// cluster. This is intended to be run prior to firing off any
// Metering installations, or running any post-install tests,
// that require testing this kind of configuration.
func customNodeSelectorFunc(ctx *deployframework.DeployerCtx) error {
	nodes, err := ctx.Client.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{
		LabelSelector: "node-role.kubernetes.io/worker=",
	})
	if err != nil || len(nodes.Items) == 0 {
		return fmt.Errorf("Failed to list the worker nodes in the cluster: %v", err)
	}

	nodeTestingLabelKey := "metering-node-testing-label"
	for _, node := range nodes.Items {
		// we can make the safe assumption that the nodes were iterating
		// over already have an initialized labels dictionary, so this only
		// requires indexing into this dictionary and adding the node testing
		// label key and value.
		node.Labels[nodeTestingLabelKey] = "true"

		_, err = ctx.Client.CoreV1().Nodes().Update(context.Background(), &node, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("Failed to apply the %s testing label to the %s node: %v", nodeTestingLabelKey, node.Name, err)
		}
		ctx.Logger.Infof("Labeled the %s node with the %s node label", node.Name, nodeTestingLabelKey)
	}

	return nil
}

// createMySQLDatabase is a test helper function that is
// responsible for initializing an ephemeral mysql database
// that will be used as the underlying Hive metastore database
// for an individual Metering test installation.
func createMySQLDatabase(ctx *deployframework.DeployerCtx) error {
	const (
		mysqlNamespace     = "mysql"
		mysqlLabelSelector = "db=mysql"
	)
	_, err := createTestingNamespace(ctx.Client, mysqlNamespace)
	if err != nil {
		return err
	}

	cmd := exec.Command(
		"oc",
		"-n", mysqlNamespace,
		"new-app",
		"--image-stream", "mysql:8.0",
		"MYSQL_USER=testuser",
		"MYSQL_PASSWORD=testpass",
		"MYSQL_DATABASE=metastore",
		"-l", mysqlLabelSelector,
	)
	cmd.Stdout = ctx.LoggerOutFile
	cmd.Stderr = ctx.LoggerOutFile
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("Failed to run the cmd: %v", err)
	}

	err = wait.Poll(5*time.Second, 5*time.Minute, func() (done bool, err error) {
		ctx.Logger.Debugf("Waiting for the db=mysql Pod to report a Ready status...")

		pods, err := ctx.Client.CoreV1().Pods(mysqlNamespace).List(context.Background(), metav1.ListOptions{
			LabelSelector: mysqlLabelSelector,
		})
		if err != nil || len(pods.Items) == 0 {
			return false, err
		}

		for _, pod := range pods.Items {
			if pod.Status.Phase != corev1.PodRunning {
				return false, nil
			}
		}

		return true, nil
	})
	if err != nil {
		return fmt.Errorf("failed to wait for the MySQL database instance to be created: %v", err)
	}
	ctx.Logger.Infof("The %s MySQL database instance is ready in the %s namespace", mysqlLabelSelector, mysqlNamespace)

	return nil
}
