package e2e

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"

	meteringv1 "github.com/operator-framework/operator-metering/pkg/apis/metering/v1"
	"github.com/operator-framework/operator-metering/pkg/operator/deploy"
	"github.com/operator-framework/operator-metering/test/deployframework"
	"github.com/operator-framework/operator-metering/test/reportingframework"
	"github.com/operator-framework/operator-metering/test/testhelpers"
)

var (
	df *deployframework.DeployFramework

	reportTestOutputDirectory string
	testOutputDirectory       string
	runAWSBillingTests        bool
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
*/

func init() {
	testOutputDirectory = os.Getenv("TEST_OUTPUT_PATH")
	runAWSBillingTests = os.Getenv("ENABLE_AWS_BILLING_TESTS") == "true"
}

func TestMain(m *testing.M) {
	var err error

	kubeConfigFlag := flag.String("kubeconfig", "", "kube config path, e.g. $HOME/.kube/config")
	nsPrefix := flag.String("namespace-prefix", "", "The namespace prefix to install the metering resources.")
	manifestDir := flag.String("deploy-manifests-dir", "../../manifests/deploy", "The absolute/relative path to the metering manifest directory.")
	cleanupScriptPath := flag.String("cleanup-script-path", "../../hack/run-test-cleanup.sh", "The absolute/relative path to the testing cleanup hack script.")
	testOutputPath := flag.String("test-output-path", "", "The absolute/relative path that you want to store test logs within.")
	logLevel := flag.String("log-level", logrus.DebugLevel.String(), "The log level")
	flag.Parse()

	logger := testhelpers.SetupLogger(*logLevel)

	if testOutputDirectory == "" {
		if *testOutputPath == "" {
			logger.Fatalf("You must specify the $TEST_OUTPUT_PATH or --test-output-path.")
		}
		testOutputDirectory = *testOutputPath
	}

	err = os.MkdirAll(testOutputDirectory, 077)
	if err != nil {
		logger.Fatalf("Failed to create the directory '%s' to log test output: %v", testOutputDirectory, err)
	}

	//loggingPath, err := ioutil.TempDir(testOutputDirectory, *nsPrefix)
	//if err != nil {
	//	logger.Fatalf("Failed to create the directory '%s' to log test output: %v", testOutputDirectory, err)
	//}

	logger.Infof("Logging resource and container logs to '%s'", testOutputDirectory)

	if df, err = deployframework.New(logger, *nsPrefix, *manifestDir, *kubeConfigFlag, *cleanupScriptPath, testOutputDirectory); err != nil {
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
	if err != nil {
		df.Logger.Fatalf("Failed to make the directory '%s': %v", testOutputDir, err)
	}

	cfg, err := df.Setup(deployerConfig, testOutputDir, targetPods)
	require.NoError(t, err, "Initializing the deploy framework should produce no error")

	defer func() {
		err := df.Teardown(testOutputDir)
		if err != nil {
			df.Logger.Warnf("Failed to teardown the metering deployment in the %s namespace: %v", cfg.Namespace, err)
		}
	}()

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
	require.NoError(t, err, "Initializing the test framework should produce no error")

	testReportingProducesData(t, testReportingFramework)
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
