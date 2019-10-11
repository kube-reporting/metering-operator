package e2e

import (
	"flag"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"

	meteringv1 "github.com/operator-framework/operator-metering/pkg/apis/metering/v1"
	"github.com/operator-framework/operator-metering/pkg/operator/deploy"
	"github.com/operator-framework/operator-metering/test/deployframework"
	"github.com/operator-framework/operator-metering/test/reportingframework"
	"github.com/operator-framework/operator-metering/test/testhelpers"
)

var (
	df *deployframework.DeployFramework

	testOutputDirectory string
	runAWSBillingTests  bool
)

/*
TODO:
- make sure that the metering and reporting operator images are overrided
- update this PR so that the CI is passing w/o knowledge of local flags (https-api, use-kube-proxy...)
  * these flags are more contingent on the meteringconfig spec
  * https-api: spec.tls.enabled (default true)
  * use-route: always true
  * use-https: always true
  * kube proxy: always false
  * reporting API URL: setting this to empty

Current Problems:
- the reporting framework flags are hardcoded
- need a way to be able to dump the actual test logs (and log to stdout)
- need to support overriding the metering and reporting operator images
*/

func init() {
	runAWSBillingTests = os.Getenv("ENABLE_AWS_BILLING_TESTS") == "true"
}

func TestMain(m *testing.M) {
	kubeConfigFlag := flag.String("kubeconfig", "", "kube config path, e.g. $HOME/.kube/config")
	nsPrefixFlag := flag.String("namespace-prefix", "", "The namespace prefix to install the metering resources.")
	manifestDirFlag := flag.String("deploy-manifests-dir", "../../manifests/deploy", "The absolute/relative path to the metering manifest directory.")
	cleanupScriptPathFlag := flag.String("cleanup-script-path", "../../hack/run-test-cleanup.sh", "The absolute/relative path to the testing cleanup hack script.")
	testOutputPathFlag := flag.String("test-output-path", "", "The absolute/relative path that you want to store test logs within.")
	logLevelFlag := flag.String("log-level", logrus.DebugLevel.String(), "The log level")
	flag.Parse()

	logger := testhelpers.SetupLogger(*logLevelFlag)

	var testOutputBaseDir string
	// check if the a test output path has been provided, else create tmpdir
	// TODO stat directory before executing os.MkdirAll (support passing an existing dir)
	if *testOutputPathFlag != "" {
		err := os.MkdirAll(*testOutputPathFlag, 0777)
		if err != nil {
			logger.Fatalf("Failed to create the test output directory '%s': %v", testOutputDirectory, err)
		}
		testOutputBaseDir = *testOutputPathFlag
	} else {
		testOutputBaseDir, err := ioutil.TempDir("", *nsPrefixFlag)
		if err != nil {
			logger.Fatalf("Failed to create the temporary directory '%s': %v", testOutputBaseDir, err)
		}
	}

	logger.Infof("Logging resource and container logs to '%s'", testOutputBaseDir)

	var err error
	if df, err = deployframework.New(logger, *nsPrefixFlag, *manifestDirFlag, *kubeConfigFlag, *cleanupScriptPathFlag, testOutputBaseDir); err != nil {
		logger.Fatalf("Failed to create a new deploy framework: %v", err)
	}

	os.Exit(m.Run())
}

func TestMultipleInstalls(t *testing.T) {
	defaultTargetPods := 7
	defaultPlatform := "openshift"

	testInstallConfigs := []struct {
		TargetPods int
		Name       string
		Config     deploy.Config
	}{
		{
			Name:       "HDFSInstall",
			TargetPods: defaultTargetPods,
			Config: deploy.Config{
				Platform:        defaultPlatform,
				DeleteNamespace: true,
				MeteringConfig: &meteringv1.MeteringConfig{
					Spec: testhelpers.NewMeteringConfigSpec(),
				},
			},
		},
	}

	for _, testCase := range testInstallConfigs {
		t := t
		testCase := testCase

		t.Run(testCase.Name, func(t *testing.T) {
			testInstall(t, testCase.Config, testCase.Name, testCase.TargetPods)
		})
	}
}

func testInstall(
	t *testing.T,
	deployerConfig deploy.Config,
	testName string,
	targetPods int,
) {
	testOutputDir := filepath.Join(df.LoggingPath, testName)
	err := os.Mkdir(testOutputDir, 0777)
	assert.NoError(t, err, "creating the test case output directory should produce no error")

	cfg, err := df.Setup(deployerConfig, testOutputDir, targetPods)
	assert.NoError(t, err, "deploying metering should produce no error")
	assert.NotNil(t, cfg, "the ReportingFrameworkConfig object returned from df.Setup should not be nil")

	defer func() {
		err := df.Teardown(testOutputDir)
		assert.NoError(t, err, "capturing logs and uninstalling metering should produce no error")
	}()

	if cfg != nil {
		testReportingFramework, err := reportingframework.New(
			cfg.Namespace,
			cfg.KubeConfigPath,
			cfg.HTTPSAPI,
			cfg.UseKubeProxyForReportingAPI,
			cfg.UseRouteForReportingAPI,
			cfg.RouteBearerToken,
			cfg.ReportingAPIURL,
			cfg.ReportResultsOutputPath,
		)
		assert.NoError(t, err, "initializing the reporting test framework should produce no error")

		testReportingProducesData(t, testReportingFramework)
	}
}

func testReportingProducesData(t *testing.T, testReportingFramework *reportingframework.ReportingFramework) {
	// cron schedule to run every minute
	cronSchedule := &meteringv1.ReportSchedule{
		Period: meteringv1.ReportPeriodCron,
		Cron: &meteringv1.ReportScheduleCron{
			Expression: fmt.Sprintf("*/1 * * * *"),
		},
	}

	queries := []struct {
		queryName   string
		skip        bool
		nonParallel bool
	}{
		{queryName: "namespace-cpu-request"},
		{queryName: "namespace-cpu-usage"},
		{queryName: "namespace-memory-request"},
		{queryName: "namespace-persistentvolumeclaim-request"},
		{queryName: "namespace-persistentvolumeclaim-usage"},
		{queryName: "namespace-memory-usage"},
		{queryName: "persistentvolumeclaim-usage"},
		{queryName: "persistentvolumeclaim-capacity"},
		{queryName: "persistentvolumeclaim-request"},
		{queryName: "pod-cpu-request"},
		{queryName: "pod-cpu-usage"},
		{queryName: "pod-memory-request"},
		{queryName: "pod-memory-usage"},
		{queryName: "node-cpu-utilization"},
		{queryName: "node-memory-utilization"},
		{queryName: "cluster-persistentvolumeclaim-request"},
		{queryName: "cluster-cpu-capacity"},
		{queryName: "cluster-memory-capacity"},
		{queryName: "cluster-cpu-usage"},
		{queryName: "cluster-memory-usage"},
		{queryName: "cluster-cpu-utilization"},
		{queryName: "cluster-memory-utilization"},
		{queryName: "namespace-memory-utilization"},
		{queryName: "namespace-cpu-utilization"},
		{queryName: "pod-cpu-request-aws", skip: !runAWSBillingTests, nonParallel: true},
		{queryName: "pod-memory-request-aws", skip: !runAWSBillingTests, nonParallel: true},
		{queryName: "aws-ec2-cluster-cost", skip: !runAWSBillingTests, nonParallel: true},
	}

	var reportsProduceDataTestCases []reportProducesDataTestCase

	for _, query := range queries {
		reportcronTestCase := reportProducesDataTestCase{
			name:          query.queryName + "-cron",
			queryName:     query.queryName,
			schedule:      cronSchedule,
			newReportFunc: testReportingFramework.NewSimpleReport,
			skip:          query.skip,
			parallel:      !query.nonParallel,
		}
		reportRunOnceTestCase := reportProducesDataTestCase{
			name:          query.queryName + "-runonce",
			queryName:     query.queryName,
			schedule:      nil, // runOnce
			newReportFunc: testReportingFramework.NewSimpleReport,
			skip:          query.skip,
			parallel:      !query.nonParallel,
		}

		reportsProduceDataTestCases = append(reportsProduceDataTestCases, reportcronTestCase, reportRunOnceTestCase)
	}

	testReportsProduceData(t, testReportingFramework, reportsProduceDataTestCases)
}
