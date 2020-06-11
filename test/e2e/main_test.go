package e2e

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	metering "github.com/kube-reporting/metering-operator/pkg/apis/metering/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"

	"github.com/kube-reporting/metering-operator/test/deployframework"
	"github.com/kube-reporting/metering-operator/test/reportingframework"
	"github.com/kube-reporting/metering-operator/test/testhelpers"
)

var (
	df *deployframework.DeployFramework

	kubeConfig    string
	logLevel      string
	runTestsLocal bool
	runDevSetup   bool

	meteringOperatorImageRepo  string
	meteringOperatorImageTag   string
	reportingOperatorImageRepo string
	reportingOperatorImageTag  string
	namespacePrefix            string
	testOutputPath             string
	repoPath                   string
	repoVersion                string

	kubeNamespaceCharLimit          = 63
	namespacePrefixCharLimit        = 10
	packageName                     = "metering-ocp"
	preUpgradeTestDirName           = "pre-upgrade"
	postUpgradeTestDirName          = "post-upgrade"
	gatherTestArtifactsScript       = "gather-test-install-artifacts.sh"
	testMeteringConfigManifestsPath = "/test/e2e/testdata/meteringconfigs/"
)

func init() {
	runAWSBillingTests = os.Getenv("ENABLE_AWS_BILLING_TESTS") == "true"

	meteringOperatorImageRepo = os.Getenv("METERING_OPERATOR_IMAGE_REPO")
	meteringOperatorImageTag = os.Getenv("METERING_OPERATOR_IMAGE_TAG")
	reportingOperatorImageRepo = os.Getenv("REPORTING_OPERATOR_IMAGE_REPO")
	reportingOperatorImageTag = os.Getenv("REPORTING_OPERATOR_IMAGE_TAG")
}

func TestMain(m *testing.M) {
	flag.StringVar(&kubeConfig, "kubeconfig", "", "kube config path, e.g. $HOME/.kube/config")
	flag.StringVar(&logLevel, "log-level", logrus.DebugLevel.String(), "The log level")
	flag.BoolVar(&runTestsLocal, "run-tests-local", false, "Controls whether the metering and reporting operators are run locally during tests")
	flag.BoolVar(&runDevSetup, "run-dev-setup", false, "Controls whether the e2e suite uses the dev-friendly configuration")
	flag.BoolVar(&runAWSBillingTests, "run-aws-billing-tests", runAWSBillingTests, "")

	flag.StringVar(&meteringOperatorImageRepo, "metering-operator-image-repo", meteringOperatorImageRepo, "")
	flag.StringVar(&meteringOperatorImageTag, "metering-operator-image-tag", meteringOperatorImageTag, "")
	flag.StringVar(&reportingOperatorImageRepo, "reporting-operator-image-repo", reportingOperatorImageRepo, "")
	flag.StringVar(&reportingOperatorImageTag, "reporting-operator-image-tag", reportingOperatorImageTag, "")

	flag.StringVar(&namespacePrefix, "namespace-prefix", "", "The namespace prefix to install the metering resources.")
	flag.StringVar(&repoPath, "repo-path", "../../", "The absolute path to the operator-metering directory.")
	flag.StringVar(&repoVersion, "repo-version", "", "The current version of the repository, e.g. 4.4, 4.5, etc.")
	flag.StringVar(&testOutputPath, "test-output-path", "", "The absolute/relative path that you want to store test logs within.")
	flag.Parse()

	logger := testhelpers.SetupLogger(logLevel)

	if len(namespacePrefix) > namespacePrefixCharLimit {
		logger.Fatalf("Error: the --namespace-prefix exceeds the limit of %d characters", namespacePrefixCharLimit)
	}

	var err error
	if df, err = deployframework.New(logger, runTestsLocal, runDevSetup, namespacePrefix, repoPath, repoVersion, kubeConfig); err != nil {
		logger.Fatalf("Failed to create a new deploy framework: %v", err)
	}

	os.Exit(m.Run())
}

type InstallTestCase struct {
	Name         string
	ExtraEnvVars []string
	TestFunc     func(t *testing.T, testReportingFramework *reportingframework.ReportingFramework)
}

func TestMeteringUpgrades(t *testing.T) {
	tt := []struct {
		Name                           string
		MeteringOperatorImageRepo      string
		MeteringOperatorImageTag       string
		Skip                           bool
		PurgeReports                   bool
		PurgeReportDataSources         bool
		ExpectInstallErr               bool
		ExpectInstallErrMsg            []string
		InstallSubTest                 InstallTestCase
		MeteringConfigManifestFilename string
	}{
		{
			Name:                      "HDFS-OLM-Upgrade",
			MeteringOperatorImageRepo: meteringOperatorImageRepo,
			MeteringOperatorImageTag:  meteringOperatorImageTag,
			PurgeReports:              true,
			PurgeReportDataSources:    true,
			Skip:                      false,
			ExpectInstallErrMsg:       []string{},
			InstallSubTest: InstallTestCase{
				Name:     "testReportingProducesData",
				TestFunc: testReportingProducesData,
				ExtraEnvVars: []string{
					"REPORTING_OPERATOR_PROMETHEUS_DATASOURCE_MAX_IMPORT_BACKFILL_DURATION=15m",
					"REPORTING_OPERATOR_PROMETHEUS_METRICS_IMPORTER_INTERVAL=30s",
					"REPORTING_OPERATOR_PROMETHEUS_METRICS_IMPORTER_CHUNK_SIZE=5m",
					"REPORTING_OPERATOR_PROMETHEUS_METRICS_IMPORTER_INTERVAL=5m",
					"REPORTING_OPERATOR_PROMETHEUS_METRICS_IMPORTER_STEP_SIZE=60s",
				},
			},
			MeteringConfigManifestFilename: "prometheus-metrics-importer-enabled.yaml",
		},
	}

	for _, testCase := range tt {
		t := t
		testCase := testCase

		if testCase.Skip {
			continue
		}

		t.Run(testCase.Name, func(t *testing.T) {
			testManualOLMUpgradeInstall(
				t,
				testCase.Name,
				namespacePrefix,
				testCase.MeteringOperatorImageRepo,
				testCase.MeteringOperatorImageTag,
				testCase.MeteringConfigManifestFilename,
				testOutputPath,
				testCase.ExpectInstallErrMsg,
				testCase.ExpectInstallErr,
				testCase.PurgeReports,
				testCase.PurgeReportDataSources,
				testCase.InstallSubTest,
			)
		})
	}
}

func testManualOLMUpgradeInstall(
	t *testing.T,
	testCaseName,
	namespacePrefix,
	meteringOperatorImageRepo,
	meteringOperatorImageTag,
	manifestFilename,
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

func TestManualMeteringInstall(t *testing.T) {
	testInstallConfigs := []struct {
		Name                           string
		MeteringOperatorImageRepo      string
		MeteringOperatorImageTag       string
		Skip                           bool
		ExpectInstallErr               bool
		ExpectInstallErrMsg            []string
		InstallSubTest                 InstallTestCase
		MeteringConfigManifestFilename string
	}{
		{
			Name:                      "InvalidHDFS-MissingStorageSpec",
			MeteringOperatorImageRepo: meteringOperatorImageRepo,
			MeteringOperatorImageTag:  meteringOperatorImageTag,
			Skip:                      false,
			ExpectInstallErr:          true,
			ExpectInstallErrMsg: []string{
				"failed to install metering",
				"failed to create the MeteringConfig resource",
				"spec.storage in body is required|spec.storage: Required value",
			},
			InstallSubTest: InstallTestCase{
				Name:     "testInvalidMeteringConfigMissingStorageSpec",
				TestFunc: testInvalidMeteringConfigMissingStorageSpec,
			},
			MeteringConfigManifestFilename: "missing-storage.yaml",
		},
		{
			Name:                      "PrometheusConnectorWorks",
			MeteringOperatorImageRepo: meteringOperatorImageRepo,
			MeteringOperatorImageTag:  meteringOperatorImageTag,
			Skip:                      false,
			InstallSubTest: InstallTestCase{
				Name:     "testPrometheusConnectorWorks",
				TestFunc: testPrometheusConnectorWorks,
			},
			MeteringConfigManifestFilename: "prometheus-metrics-importer-disabled.yaml",
		},
		{
			Name:                      "ValidHDFS-ReportDynamicInputData",
			MeteringOperatorImageRepo: meteringOperatorImageRepo,
			MeteringOperatorImageTag:  meteringOperatorImageTag,
			Skip:                      false,
			InstallSubTest: InstallTestCase{
				Name:     "testReportingProducesData",
				TestFunc: testReportingProducesData,
				ExtraEnvVars: []string{
					"REPORTING_OPERATOR_PROMETHEUS_DATASOURCE_MAX_IMPORT_BACKFILL_DURATION=15m",
					"REPORTING_OPERATOR_PROMETHEUS_METRICS_IMPORTER_INTERVAL=30s",
					"REPORTING_OPERATOR_PROMETHEUS_METRICS_IMPORTER_CHUNK_SIZE=5m",
					"REPORTING_OPERATOR_PROMETHEUS_METRICS_IMPORTER_INTERVAL=5m",
					"REPORTING_OPERATOR_PROMETHEUS_METRICS_IMPORTER_STEP_SIZE=60s",
				},
			},
			MeteringConfigManifestFilename: "prometheus-metrics-importer-enabled.yaml",
		},
		{
			Name:                      "ValidHDFS-ReportStaticInputData",
			MeteringOperatorImageRepo: meteringOperatorImageRepo,
			MeteringOperatorImageTag:  meteringOperatorImageTag,
			Skip:                      false,
			InstallSubTest: InstallTestCase{
				Name:     "testReportingProducesCorrectDataForInput",
				TestFunc: testReportingProducesCorrectDataForInput,
				ExtraEnvVars: []string{
					"REPORTING_OPERATOR_DISABLE_PROMETHEUS_METRICS_IMPORTER=true",
				},
			},
			MeteringConfigManifestFilename: "prometheus-metrics-importer-disabled.yaml",
		},
	}

	for _, testCase := range testInstallConfigs {
		t := t
		testCase := testCase

		if testCase.Skip {
			continue
		}

		t.Run(testCase.Name, func(t *testing.T) {
			testManualMeteringInstall(
				t,
				testCase.Name,
				namespacePrefix,
				testCase.MeteringOperatorImageRepo,
				testCase.MeteringOperatorImageTag,
				testCase.MeteringConfigManifestFilename,
				testOutputPath,
				testCase.ExpectInstallErrMsg,
				testCase.ExpectInstallErr,
				testCase.InstallSubTest,
			)
		})
	}
}

func testManualMeteringInstall(
	t *testing.T,
	testCaseName,
	namespacePrefix,
	meteringOperatorImageRepo,
	meteringOperatorImageTag,
	manifestFilename,
	testOutputPath string,
	expectInstallErrMsg []string,
	expectInstallErr bool,
	testInstallFunction InstallTestCase,
) {
	// create a directory used to store the @testCaseName container and resource logs
	testCaseOutputBaseDir := filepath.Join(testOutputPath, testCaseName)
	err := os.Mkdir(testCaseOutputBaseDir, 0777)
	assert.NoError(t, err, "creating the test case output directory should produce no error")

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
		testCaseOutputBaseDir,
		testInstallFunction.ExtraEnvVars,
		mc.Spec,
	)
	require.NoError(t, err, "creating a new deployer context should produce no error")

	deployerCtx.Logger.Infof("DeployerCtx: %+v", deployerCtx)
	rf, err := deployerCtx.Setup(deployerCtx.Deployer.Install, expectInstallErr)

	canSafelyRunTest := testhelpers.AssertCanSafelyRunReportingTests(t, err, expectInstallErr, expectInstallErrMsg)
	if canSafelyRunTest {
		t.Run(testInstallFunction.Name, func(t *testing.T) {
			testInstallFunction.TestFunc(t, rf)
		})

		deployerCtx.Logger.Infof("The %s test has finished running", testInstallFunction.Name)
	}

	err = deployerCtx.Teardown(deployerCtx.Deployer.Uninstall)
	assert.NoError(t, err, "capturing logs and uninstalling metering should produce no error")
}
