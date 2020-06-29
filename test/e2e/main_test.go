package e2e

import (
	"flag"
	"os"
	"testing"

	"github.com/sirupsen/logrus"

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
	meteringOperatorImage      string
	reportingOperatorImage     string

	namespacePrefix        string
	testOutputPath         string
	repoPath               string
	repoVersion            string
	registryImage          string
	subscriptionChannel    string
	catalogSourceName      string
	catalogSourceNamespace string

	kubeNamespaceCharLimit          = 63
	namespacePrefixCharLimit        = 10
	packageName                     = "metering-ocp"
	preUpgradeTestDirName           = "pre-upgrade"
	postUpgradeTestDirName          = "post-upgrade"
	gatherTestArtifactsScript       = "gather-test-install-artifacts.sh"
	testMeteringConfigManifestsPath = "/test/e2e/testdata/meteringconfigs/"
)

type InstallTestCase struct {
	Name         string
	ExtraEnvVars []string
	TestFunc     func(t *testing.T, testReportingFramework *reportingframework.ReportingFramework)
}

func init() {
	runAWSBillingTests = os.Getenv("ENABLE_AWS_BILLING_TESTS") == "true"

	meteringOperatorImageRepo = os.Getenv("METERING_OPERATOR_IMAGE_REPO")
	meteringOperatorImageTag = os.Getenv("METERING_OPERATOR_IMAGE_TAG")
	reportingOperatorImageRepo = os.Getenv("REPORTING_OPERATOR_IMAGE_REPO")
	reportingOperatorImageTag = os.Getenv("REPORTING_OPERATOR_IMAGE_TAG")
}

func TestMain(m *testing.M) {
	os.Exit(testMainWrapper(m))
}

// testMainWrapper is a wrapper function around the
// top-level TestMain function and this pattern is
// needed as os.Exit() doesn't respect any defer calls
// that may occur during the TestMain workflow. If we
// we instead doing the heavy-lifting in this function,
// and then return an integer code that os.Exit can correctly
// interpret, then the defer call will work.
//
// See the following references for more information:
// - https://golang.org/pkg/os/#Exit
// - http://blog.englund.nu/golang,/testing/2017/03/12/using-defer-in-testmain.html
func testMainWrapper(m *testing.M) int {
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

	flag.StringVar(&registryImage, "registry-image", "", "The name of an existing registry image containing a manifest bundle.")
	flag.StringVar(&subscriptionChannel, "subscription-channel", "", "The name of an existing channel in the registry image you want to subscribe to.")
	flag.Parse()

	logger := testhelpers.SetupLogger(logLevel)

	if len(namespacePrefix) > namespacePrefixCharLimit {
		logger.Fatalf("Error: the --namespace-prefix exceeds the limit of %d characters", namespacePrefixCharLimit)
	}
	if registryImage == "" {
		logger.Fatalf("Error: the --registry-image field is empty and that field is required")
	}

	var err error
	if df, err = deployframework.New(logger, runTestsLocal, runDevSetup, namespacePrefix, repoPath, repoVersion, kubeConfig); err != nil {
		logger.Fatalf("Failed to create a new deploy framework: %v", err)
	}

	// Check if any of the operator image variables are non-empty.
	// In the case where one or both are non-empty, we need to pass any
	// of the overriden operator images so that we can manipulate
	// the default images in the metering CSV. In the case where
	// some version of the `make e2e-dev` has been provided, skip
	// the deletion the local registry resources and CatalogSource CR.
	if meteringOperatorImageRepo != "" {
		meteringOperatorImage = meteringOperatorImageRepo + ":" + meteringOperatorImageTag
	}
	if reportingOperatorImageRepo != "" {
		reportingOperatorImage = reportingOperatorImageRepo + ":" + reportingOperatorImageTag
	}
	catalogSourceName, catalogSourceNamespace, err = df.CreateRegistryResources(registryImage, meteringOperatorImage, reportingOperatorImage)
	if err != nil {
		df.Logger.Fatalf("Failed to create the CatalogSource custom resource using the %s registry image: %v", registryImage, err)
	}
	if !df.RunDevSetup {
		defer df.DeleteRegistryResources(catalogSourceName, catalogSourceNamespace)
	}

	err = df.WaitForPackageManifest(catalogSourceName, catalogSourceNamespace, subscriptionChannel)
	if err != nil {
		df.Logger.Fatalf("Failed to wait for the metering-ocp packagemanifest to become ready: %v", err)
	}

	return m.Run()
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
				catalogSourceName,
				catalogSourceNamespace,
				subscriptionChannel,
				testOutputPath,
				testCase.ExpectInstallErrMsg,
				testCase.ExpectInstallErr,
				testCase.InstallSubTest,
			)
		})
	}
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
				catalogSourceName,
				catalogSourceNamespace,
				subscriptionChannel,
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
