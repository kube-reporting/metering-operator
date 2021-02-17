package e2e

import (
	"context"
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"

	"github.com/kube-reporting/metering-operator/test/deployframework"
	"github.com/kube-reporting/metering-operator/test/reportingframework"
	"github.com/kube-reporting/metering-operator/test/testhelpers"
)

var (
	df *deployframework.DeployFramework

	kubeConfig         string
	logLevel           string
	runTestsLocal      bool
	runDevSetup        bool
	runAllInstallTests bool

	meteringOperatorImageRepo  string
	meteringOperatorImageTag   string
	reportingOperatorImageRepo string
	reportingOperatorImageTag  string
	meteringOperatorImage      string
	reportingOperatorImage     string

	namespacePrefix                string
	testOutputPath                 string
	repoPath                       string
	repoVersion                    string
	registryImage                  string
	indexImage                     string
	subscriptionChannel            string
	upgradeFromSubscriptionChannel string
	catalogSourceName              string
	catalogSourceNamespace         string

	kubeNamespaceCharLimit          = 63
	namespacePrefixCharLimit        = 10
	packageName                     = "metering-ocp"
	preUpgradeTestDirName           = "pre-upgrade"
	postUpgradeTestDirName          = "post-upgrade"
	gatherTestArtifactsScript       = "gather-test-install-artifacts.sh"
	testMeteringConfigManifestsPath = "/test/e2e/manifests/meteringconfigs/"
	testNFSManifestPath             = "/test/e2e/manifests/nfs/"
)

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
	flag.BoolVar(&runAllInstallTests, "test-run-all-install-tests", true, "Controls whether or not the e2e suite installs all of the tests listed in the TestManualMeteringInstall function")
	flag.BoolVar(&runAWSBillingTests, "run-aws-billing-tests", runAWSBillingTests, "")

	flag.StringVar(&meteringOperatorImageRepo, "metering-operator-image-repo", meteringOperatorImageRepo, "")
	flag.StringVar(&meteringOperatorImageTag, "metering-operator-image-tag", meteringOperatorImageTag, "")
	flag.StringVar(&reportingOperatorImageRepo, "reporting-operator-image-repo", reportingOperatorImageRepo, "")
	flag.StringVar(&reportingOperatorImageTag, "reporting-operator-image-tag", reportingOperatorImageTag, "")

	flag.StringVar(&namespacePrefix, "namespace-prefix", "", "The namespace prefix to install the metering resources.")
	flag.StringVar(&repoPath, "repo-path", "../../", "The absolute path to the operator-metering directory.")
	flag.StringVar(&repoVersion, "repo-version", "", "The current version of the repository, e.g. 4.4, 4.5, etc.")
	flag.StringVar(&testOutputPath, "test-output-path", "", "The absolute/relative path that you want to store test logs within.")

	flag.StringVar(&registryImage, "registry-image", "registry.svc.ci.openshift.org/ocp/4.6:metering-ansible-operator-registry", "The name of an existing registry image containing a manifest bundle.")
	flag.StringVar(&indexImage, "index-image", "", "The name of the index image containing a metering bundle. Note: this flag take precedence over the --registry-image flag.")
	flag.StringVar(&subscriptionChannel, "subscription-channel", "4.6", "The name of an existing channel in the registry image you want to subscribe to.")
	flag.StringVar(&upgradeFromSubscriptionChannel, "upgrade-from-subscription-channel", "4.5", "The name of an existing channel in a catalog source that you want to upgrade from.")
	flag.Parse()

	logger := testhelpers.SetupLogger(logLevel)

	if len(namespacePrefix) > namespacePrefixCharLimit {
		logger.Fatalf("Error: the --namespace-prefix exceeds the limit of %d characters", namespacePrefixCharLimit)
	}

	var err error
	if df, err = deployframework.New(logger, runTestsLocal, runDevSetup, namespacePrefix, repoPath, repoVersion, kubeConfig); err != nil {
		logger.Fatalf("Failed to create a new deploy framework: %v", err)
	}

	if indexImage == "" {
		logger.Fatalf("You need to specify a non-empty --index-image flag value")
	}

	// TODO: determine whether it makes sense to have a toggle for creating
	// either a registry containing the old packagemanifest format vs.
	// always using an index image. For now, always use the index image.
	var registryProvisioned bool
	catalogSourceName, catalogSourceNamespace, err = df.CreateCatalogSourceFromIndex(indexImage)
	if err != nil {
		df.Logger.Fatalf("Failed to create the CatalogSource custom resource using the %s index image: %v", indexImage, err)
	}
	if !df.RunDevSetup {
		defer df.DeleteRegistryResources(registryProvisioned, catalogSourceName, catalogSourceNamespace)
	}

	err = df.WaitForPackageManifest(catalogSourceName, catalogSourceNamespace, subscriptionChannel)
	if err != nil {
		df.Logger.Fatalf("Failed to wait for the metering-ocp packagemanifest to become ready: %v", err)
	}

	return m.Run()
}

type InstallTestCase struct {
	Name         string
	ExtraEnvVars []string
	TestFunc     func(t *testing.T, testReportingFramework *reportingframework.ReportingFramework)
}

type PreInstallFunc func(ctx *deployframework.DeployerCtx) error

func TestInvalidMeteringConfigs(t *testing.T) {
	namespace := fmt.Sprintf("%s-invalid-meteringconfigs", namespacePrefix)
	ns, err := createTestingNamespace(df.Client, namespace)
	require.NoError(t, err, "failed to successfully create the %s testing namespace", namespace)
	require.NotEmpty(t, ns.Name, "expected the testing namespace would not be nil")

	tt := []struct {
		Name                           string
		ExpectInstallErrMsg            []string
		MeteringConfigManifestFileName string
	}{
		{
			Name: "missing-storage-spec",
			ExpectInstallErrMsg: []string{
				"spec.storage in body is required|spec.storage: Required value",
			},
			MeteringConfigManifestFileName: "missing-storage.yaml",
		},
	}

	for _, tc := range tt {
		tc := tc
		t := t

		mc, err := testhelpers.DecodeMeteringConfigManifest(repoPath, testMeteringConfigManifestsPath, tc.MeteringConfigManifestFileName)
		require.NoError(t, err, "failed to successfully decode the YAML MeteringConfig manifest")

		_, err = df.MeteringClient.MeteringConfigs(ns.Name).Create(context.Background(), mc, metav1.CreateOptions{})
		testhelpers.AssertErrorContainsErrorMsgs(t, err, tc.ExpectInstallErrMsg)
	}

	// In the case that `make e2e-dev` has been specified, avoid
	// deleting the testing namespace used to validate the MC's.
	if runDevSetup {
		return
	}

	// May need to account for apierrors.IsNotFound(err) just to reduce
	// any potential e2e flakes and return w/o an error
	err = df.Client.CoreV1().Namespaces().Delete(context.Background(), ns.Name, metav1.DeleteOptions{})
	require.NoError(t, err, "failed to delete the %s testing namespace", ns.Name)
}

func TestManualMeteringInstall(t *testing.T) {
	testInstallConfigs := []struct {
		Name                           string
		MeteringOperatorImageRepo      string
		MeteringOperatorImageTag       string
		Skip                           bool
		InstallSubTests                []InstallTestCase
		PreInstallFunc                 PreInstallFunc
		MeteringConfigManifestFilename string
	}{
		{
			Name:                      "HDFS-ValidNodeSelector",
			MeteringOperatorImageRepo: meteringOperatorImageRepo,
			MeteringOperatorImageTag:  meteringOperatorImageTag,
			// TODO: transistion this to a periodic test and
			// update the `Skip` condition to !runAllInstallTests
			// TODO: disabling this test for the time being as
			// we're labeling nodes and firing off this metering
			// installation before the machineautoscaler has provisioned
			// any new machine. The result is the new machines that get
			// provisioned, don't have this custom node label we added in
			// the preInstallFunc closure.
			Skip:           !testing.Short(),
			PreInstallFunc: customNodeSelectorFunc,
			InstallSubTests: []InstallTestCase{
				{
					Name:     "testNodeSelectorConfigurationWorks",
					TestFunc: testNodeSelectorConfigurationWorks,
				},
				{
					Name:     "testReportingProducesCorrectDataForInput",
					TestFunc: testReportingProducesCorrectDataForInput,
					ExtraEnvVars: []string{
						"REPORTING_OPERATOR_DISABLE_PROMETHEUS_METRICS_IMPORTER=true",
					},
				},
				{
					Name:     "testMeteringAnsibleOperatorMetricsWork",
					TestFunc: testMeteringAnsibleOperatorMetricsWork,
				},
				{
					Name:     "testReportingOperatorServiceCABundleExists",
					TestFunc: testReportingOperatorServiceCABundleExists,
				},
				{
					Name:     "testFailedPrometheusQueryEvents",
					TestFunc: testFailedPrometheusQueryEvents,
				},
			},
			MeteringConfigManifestFilename: "node-selector-prometheus-importer-disabled.yaml",
		},
		{
			Name:                      "HDFS-ReportDynamicInputData",
			MeteringOperatorImageRepo: meteringOperatorImageRepo,
			MeteringOperatorImageTag:  meteringOperatorImageTag,
			Skip:                      false,
			InstallSubTests: []InstallTestCase{
				{
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
				{
					Name:     "testFailedPrometheusQueryEvents",
					TestFunc: testFailedPrometheusQueryEvents,
				},
			},
			MeteringConfigManifestFilename: "prometheus-metrics-importer-enabled.yaml",
		},
		{
			Name:                      "HDFS-ReportStaticInputData",
			MeteringOperatorImageRepo: meteringOperatorImageRepo,
			MeteringOperatorImageTag:  meteringOperatorImageTag,
			Skip:                      false,
			InstallSubTests: []InstallTestCase{
				{
					Name:     "testReportingProducesCorrectDataForInput",
					TestFunc: testReportingProducesCorrectDataForInput,
					ExtraEnvVars: []string{
						"REPORTING_OPERATOR_DISABLE_PROMETHEUS_METRICS_IMPORTER=true",
					},
				},
				{
					Name:     "testPrometheusConnectorWorks",
					TestFunc: testPrometheusConnectorWorks,
				},
				{
					Name:     "testReportingOperatorServiceCABundleExists",
					TestFunc: testReportingOperatorServiceCABundleExists,
				},
				{
					Name:     "testLeaderElectionEventIsCreated",
					TestFunc: testLeaderElectionEventIsCreated,
				},
				{
					Name:     "testFailedPrometheusQueryEvents",
					TestFunc: testFailedPrometheusQueryEvents,
				},
				{
					Name:         "testReportIsDeletedWhenNoDeps",
					TestFunc:     testReportIsDeletedWhenNoDeps,
					ExtraEnvVars: []string{},
				},
				{
					Name:         "testReportIsNotDeletedWhenReportDependsOnIt",
					TestFunc:     testReportIsNotDeletedWhenReportDependsOnIt,
					ExtraEnvVars: []string{},
				},
			},
			MeteringConfigManifestFilename: "prometheus-metrics-importer-disabled.yaml",
		},
		{
			Name:                      "HDFS-MySQLDatabase",
			MeteringOperatorImageRepo: meteringOperatorImageRepo,
			MeteringOperatorImageTag:  meteringOperatorImageTag,
			PreInstallFunc:            createMySQLDatabase,
			InstallSubTests: []InstallTestCase{
				{
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
				{
					Name:     "testFailedPrometheusQueryEvents",
					TestFunc: testFailedPrometheusQueryEvents,
				},
			},
			MeteringConfigManifestFilename: "mysql.yaml",
		},
		{
			Name:                      "S3-ReportDynamicInputData",
			MeteringOperatorImageRepo: meteringOperatorImageRepo,
			MeteringOperatorImageTag:  meteringOperatorImageTag,
			// TODO: disabling this for now as we work towards
			// migrating a subset of all tests to periodic jobs.
			// We should check-in with DPTP to make sure they're
			// aware we're creating a s3 bucket in their CI account
			// so their pruner is aware of this bucket location.
			Skip:           true,
			PreInstallFunc: s3InstallFunc,
			InstallSubTests: []InstallTestCase{
				{
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
				{
					Name:         "testEnsureS3BucketIsDeleted",
					TestFunc:     testEnsureS3BucketIsDeleted,
					ExtraEnvVars: []string{},
				},
				{
					Name:     "testFailedPrometheusQueryEvents",
					TestFunc: testFailedPrometheusQueryEvents,
				},
			},
			MeteringConfigManifestFilename: "s3.yaml",
		},
		{
			Name:                      "NFS-ReportDynamicInputData",
			MeteringOperatorImageRepo: meteringOperatorImageRepo,
			MeteringOperatorImageTag:  meteringOperatorImageTag,
			Skip:                      !runAllInstallTests,
			PreInstallFunc:            createNFSProvisioner,
			InstallSubTests: []InstallTestCase{
				{
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
			},
			MeteringConfigManifestFilename: "nfs.yaml",
		},
	}

	for _, testCase := range testInstallConfigs {
		// capture the range variables
		testCase := testCase
		t := t

		if testCase.Skip {
			continue
		}

		t.Run(testCase.Name, func(t *testing.T) {
			// If we call t.Parallel() here, the top-level test will
			// blocked from returning until all of the goroutines that
			// t.Run spawns have completed.
			t.Parallel()

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
				testCase.PreInstallFunc,
				testCase.InstallSubTests,
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
				upgradeFromSubscriptionChannel,
				subscriptionChannel,
				testOutputPath,
				testCase.PurgeReports,
				testCase.PurgeReportDataSources,
				testCase.InstallSubTest,
			)
		})
	}
}
