package e2e

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kube-reporting/metering-operator/test/deployframework"
	"github.com/kube-reporting/metering-operator/test/testhelpers"
)

const (
	upgradeFromCatalogSource          = "redhat-operators"
	upgradeFromCatalogSourceNamespace = "openshift-marketplace"
)

type InstallConfig struct {
	CatalogSourceName      string
	CatalogSourceNamespace string
	PackageName            string
	SubscriptionChannel    string
}

func NewInstallConfig(name, namespace, packageName, channel string) *InstallConfig {
	if name == "" {
		name = "redhat-operators"
	}
	if namespace == "" {
		namespace = "openshift-marketplace"
	}
	if packageName == "" {
		packageName = "metering-ocp"
	}
	if channel == "" {
		channel = "4.8"
	}

	return &InstallConfig{
		CatalogSourceName:      name,
		CatalogSourceNamespace: namespace,
		PackageName:            packageName,
		SubscriptionChannel:    channel,
	}
}

func testManualOLMUpgradeInstall(
	t *testing.T,
	testCaseName,
	namespacePrefix,
	meteringOperatorImageRepo,
	meteringOperatorImageTag,
	manifestFilename,
	catalogSourceName,
	catalogSourceNamespace,
	upgradeFromSubscriptionChannel,
	subscriptionChannel,
	testOutputPath string,
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
	require.LessOrEqual(t, len(testFuncNamespace), kubeNamespaceCharLimit)

	mc, err := testhelpers.DecodeMeteringConfigManifest(repoPath, testMeteringConfigManifestsPath, manifestFilename)
	require.NoError(t, err, "failed to successfully decode the YAML MeteringConfig manifest")

	deployerCtx, err := df.NewDeployerCtx(
		testFuncNamespace,
		meteringOperatorImageRepo,
		meteringOperatorImageTag,
		reportingOperatorImageRepo,
		reportingOperatorImageTag,
		upgradeFromCatalogSource,
		upgradeFromCatalogSourceNamespace,
		upgradeFromSubscriptionChannel,
		preUpgradeTestOutputDir,
		testInstallFunction.ExtraEnvVars,
		deployframework.DefaultDeleteNamespace,
		deployframework.DefaultDeleteCRD,
		deployframework.DefaultDeleteCRB,
		deployframework.DefaultDeletePVC,
		mc.Spec,
	)
	require.NoError(t, err, "creating a new deployer context should produce no error")
	deployerCtx.Logger.Infof("DeployerCtx: %+v", deployerCtx)

	rf, err := deployerCtx.Setup(deployerCtx.Deployer.InstallOLM)
	require.NoError(t, err, "installing metering should produce no error")

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
	rf, err = deployerCtx.Upgrade(catalogSourceName, catalogSourceNamespace, subscriptionChannel, purgeReports, purgeReportDataSources)
	if err != nil {
		gatherErr := deployerCtx.MustGatherMeteringResources(gatherTestArtifactsScript)
		assert.NoError(t, gatherErr, "gathering metering resources should produce no error")
		require.NoError(t, err, "upgrading metering should produce no error")
	}

	// run tests against the upgraded installation
	postUpgradeTestName := fmt.Sprintf("post-upgrade-%s", testInstallFunction.Name)
	t.Run(postUpgradeTestName, func(t *testing.T) {
		testInstallFunction.TestFunc(t, rf)
	})

	err = deployerCtx.Teardown(deployerCtx.Deployer.UninstallOLM)
	require.NoError(t, err, "capturing logs and uninstalling metering should produce no error")
}
