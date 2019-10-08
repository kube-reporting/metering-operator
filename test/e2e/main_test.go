package e2e

import (
	"flag"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"

	metering "github.com/operator-framework/operator-metering/pkg/apis/metering/v1"
	"github.com/operator-framework/operator-metering/test/reportingframework"
)

var (
	testReportingFramework *reportingframework.ReportingFramework

	reportTestOutputDirectory string
	runAWSBillingTests        bool
)

func init() {
	reportTestOutputDirectory = os.Getenv("TEST_RESULT_REPORT_OUTPUT_DIRECTORY")
	if reportTestOutputDirectory == "" {
		log.Fatalf("$TEST_RESULT_REPORT_OUTPUT_DIRECTORY must be set")
	}

	err := os.MkdirAll(reportTestOutputDirectory, 0777)
	if err != nil {
		log.Fatalf("error making directory %s, err: %s", reportTestOutputDirectory, err)
	}

	runAWSBillingTests = os.Getenv("ENABLE_AWS_BILLING_TESTS") == "true"
}

func TestMain(m *testing.M) {
	kubecon***REMOVED***g := flag.String("kubecon***REMOVED***g", "", "kube con***REMOVED***g path, e.g. $HOME/.kube/con***REMOVED***g")
	ns := flag.String("namespace", "metering-ci", "test namespace")
	httpsAPI := flag.Bool("https-api", false, "If true, use https to talk to Metering API")
	useKubeProxyForReportingAPI := flag.Bool("use-kube-proxy-for-reporting-api", false, "If true, uses kubernetes API proxy to access reportingAPI")
	useRouteForReportingAPI := flag.Bool("use-route-for-reporting-api", true, "If true, uses a route to access reportingAPI")
	reportingAPIURL := flag.String("reporting-api-url", "", "reporting-operator URL if useKubeProxyForReportingAPI is false")
	routeBearerToken := flag.String("route-bearer-token", "", "The bearer token to the reporting-operator serviceaccount if use-route-for-reporting-api is set to true")
	flag.Parse()

	var err error
	if testReportingFramework, err = reportingframework.New(*ns, *kubecon***REMOVED***g, *httpsAPI, *useKubeProxyForReportingAPI, *useRouteForReportingAPI, *routeBearerToken, *reportingAPIURL, reportTestOutputDirectory); err != nil {
		logrus.Fatalf("failed to setup framework: %v\n", err)
	}

	os.Exit(m.Run())
}

func TestReportingProducesData(t *testing.T) {
	// cron schedule to run every minute
	cronSchedule := &metering.ReportSchedule{
		Period: metering.ReportPeriodCron,
		Cron: &metering.ReportScheduleCron{
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
