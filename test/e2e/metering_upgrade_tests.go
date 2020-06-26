package e2e

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/yaml"

	metering "github.com/kube-reporting/metering-operator/pkg/apis/metering/v1"
	"github.com/kube-reporting/metering-operator/test/reportingframework"
	"github.com/kube-reporting/metering-operator/test/testhelpers"
)

func testManualOLMUpgradeInstall(
	t *testing.T,
	testCaseName,
	namespacePrefix,
	meteringOperatorImageRepo,
	meteringOperatorImageTag,
	manifestFilename,
	catalogSourceName,
	catalogSourceNamespace,
	subscriptionName,
	testOutputPath string,
	expectInstallErrMsg []string,
	expectInstallErr,
	purgeReports,
	purgeReportDataSources bool,
	testInstallFunction InstallTestCase,
) {
	// create a directory used to store the @testCaseName container and resource logs
	testCaseOutputBaseDir := filepath.Join(testOutputPath, testCaseName)
	err := os.Mkdir(testCaseOutputBaseDir, 0777)
	require.NoError(t, err, "creating the test case output directory should produce no error")

	// create a pre-upgrade test case directory
	preUpgradeTestOutputDir := filepath.Join(testCaseOutputBaseDir, preUpgradeTestDirName)
	err = os.Mkdir(preUpgradeTestOutputDir, 0777)
	require.NoError(t, err, "creating the test case output directory should produce no error")

	testFuncNamespace := fmt.Sprintf("%s-%s", namespacePrefix, strings.ToLower(testCaseName))
	if len(testFuncNamespace) > kubeNamespaceCharLimit {
		require.Fail(t, "The length of the test function namespace exceeded the kube namespace limit of %d characters", kubeNamespaceCharLimit)
	}

	manifestFullPath := filepath.Join(repoPath, testMeteringConfigManifestsPath, manifestFilename)
	file, err := os.Open(manifestFullPath)
	require.NoError(t, err, "failed to open manifest file")

	mc := &metering.MeteringConfig{}
	err = yaml.NewYAMLOrJSONDecoder(file, 100).Decode(&mc)
	require.NoError(t, err, "failed to decode the yaml meteringconfig manifest")
	require.NotNil(t, mc, "the decoded meteringconfig object is nil")

	deployerCtx, err := df.NewDeployerCtx(
		testFuncNamespace,
		meteringOperatorImageRepo,
		meteringOperatorImageTag,
		reportingOperatorImageRepo,
		reportingOperatorImageTag,
		catalogSourceName,
		catalogSourceNamespace,
		subscriptionName,
		preUpgradeTestOutputDir,
		expectInstallErrMsg,
		mc.Spec,
	)
	require.NoError(t, err, "creating a new deployer context should produce no error")
	deployerCtx.Logger.Infof("DeployerCtx: %+v", deployerCtx)

	var (
		canSafelyRunTest bool
		rf               *reportingframework.ReportingFramework
	)
	rf, err = deployerCtx.Setup(deployerCtx.Deployer.InstallOLM, expectInstallErr)
	if canSafelyRunTest = testhelpers.AssertCanSafelyRunReportingTests(t, err, expectInstallErr, expectInstallErrMsg); !canSafelyRunTest {
		// if we encounter an unexpected Setup error, fail this test case
		// early and gather the metering and OLM resource logs we care about.
		err = deployerCtx.MustGatherMeteringResources(gatherTestArtifactsScript)
		assert.NoError(t, err, "gathering metering resources should produce no error")
		t.Fatal("Exiting test case early as the pre-upgrade tests failed")
	}

	preUpgradeTestName := fmt.Sprintf("pre-upgrade-%s", testInstallFunction.Name)
	t.Run(preUpgradeTestName, func(t *testing.T) {
		testInstallFunction.TestFunc(t, rf)
	})

	err = deployerCtx.MustGatherMeteringResources(gatherTestArtifactsScript)
	assert.NoError(t, err, "gathering metering resources should produce no error")

	// create a post-upgrade test case directory
	postUpgradeTestOutputDir := filepath.Join(testCaseOutputBaseDir, postUpgradeTestDirName)
	err = os.Mkdir(postUpgradeTestOutputDir, 0777)
	assert.NoError(t, err, "creating the test case output directory should produce no error")

	deployerCtx.TestCaseOutputPath = postUpgradeTestOutputDir
	rf, err = deployerCtx.Upgrade(packageName, df.RepoVersion, purgeReports, purgeReportDataSources)
	if canSafelyRunTest = testhelpers.AssertCanSafelyRunReportingTests(t, err, expectInstallErr, expectInstallErrMsg); !canSafelyRunTest {
		err = deployerCtx.MustGatherMeteringResources(gatherTestArtifactsScript)
		assert.NoError(t, err, "gathering metering resources should produce no error")
	}

	if canSafelyRunTest {
		// run tests against the upgraded installation
		postUpgradeTestName := fmt.Sprintf("post-upgrade-%s", testInstallFunction.Name)
		t.Run(postUpgradeTestName, func(t *testing.T) {
			testInstallFunction.TestFunc(t, rf)
		})
	}

	err = deployerCtx.Teardown(deployerCtx.Deployer.UninstallOLM)
	require.NoError(t, err, "capturing logs and uninstalling metering should produce no error")
}
