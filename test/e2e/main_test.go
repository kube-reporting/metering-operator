package e2e

import (
	"flag"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"

	meteringv1alpha1 "github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1"
	"github.com/operator-framework/operator-metering/test/framework"
)

var (
	testFramework *framework.Framework

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
	flag.Parse()

	var err error
	if testFramework, err = framework.New(*ns, *kubecon***REMOVED***g, *httpsAPI); err != nil {
		logrus.Fatalf("failed to setup framework: %v\n", err)
	}

	os.Exit(m.Run())
}

func TestReportingProducesData(t *testing.T) {
	// cron schedule to run every minute
	cronSchedule := &meteringv1alpha1.ReportSchedule{
		Period: meteringv1alpha1.ReportPeriodCron,
		Cron: &meteringv1alpha1.ReportScheduleCron{
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
		{queryName: "namespace-memory-usage"},
		{queryName: "persistentvolumeclaim-usage"},
		{queryName: "pod-cpu-request"},
		{queryName: "pod-cpu-usage"},
		{queryName: "pod-memory-request"},
		{queryName: "pod-memory-usage"},
		{queryName: "persistentvolumeclaim-request"},
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
			newReportFunc: testFramework.NewSimpleReport,
			skip:          query.skip,
			parallel:      !query.nonParallel,
		}
		reportRunOnceTestCase := reportProducesDataTestCase{
			name:          query.queryName + "-runonce",
			queryName:     query.queryName,
			schedule:      nil, // runOnce
			newReportFunc: testFramework.NewSimpleReport,
			skip:          query.skip,
			parallel:      !query.nonParallel,
		}

		reportsProduceDataTestCases = append(reportsProduceDataTestCases, reportcronTestCase, reportRunOnceTestCase)
	}

	// Align to the nearest minute
	periodEnd := time.Now().UTC().Truncate(time.Minute)
	// start the report 5 minutes ago
	periodStart := periodEnd.Add(-5 * time.Minute)

	testReportsProduceData(t, testFramework, periodStart, periodEnd, reportsProduceDataTestCases)
}
