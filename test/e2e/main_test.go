package e2e

import (
	"flag"
	"log"
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"

	meteringv1alpha1 "github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1"
	"github.com/operator-framework/operator-metering/pkg/operator"
	"github.com/operator-framework/operator-metering/test/framework"
)

var (
	testFramework *framework.Framework

	reportTestTimeout         = 5 * time.Minute
	reportTestOutputDirectory string
	runAWSBillingTests        bool
)

func init() {
	if reportTestTimeoutStr := os.Getenv("REPORT_TEST_TIMEOUT"); reportTestTimeoutStr != "" {
		var err error
		reportTestTimeout, err = time.ParseDuration(reportTestTimeoutStr)
		if err != nil {
			log.Fatalf("Invalid REPORT_TEST_TIMEOUT: %v", err)
		}
	}
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
	kubeconfig := flag.String("kubeconfig", "", "kube config path, e.g. $HOME/.kube/config")
	ns := flag.String("namespace", "metering-ci", "test namespace")
	httpsAPI := flag.Bool("https-api", false, "If true, use https to talk to Metering API")
	flag.Parse()

	var err error
	if testFramework, err = framework.New(*ns, *kubeconfig, *httpsAPI); err != nil {
		logrus.Fatalf("failed to setup framework: %v\n", err)
	}

	os.Exit(m.Run())
}

func TestReportingE2E(t *testing.T) {
	hourlySchedule := &meteringv1alpha1.ScheduledReportSchedule{
		Period: meteringv1alpha1.ScheduledReportPeriodHourly,
	}

	queries := []struct {
		queryName string
		skip      bool
	}{
		{queryName: "namespace-cpu-request"},
		{queryName: "namespace-cpu-usage"},
		{queryName: "namespace-memory-request"},
		{queryName: "namespace-persistentvolumeclaim-request"},
		{queryName: "namespace-memory-usage"},
		{queryName: "pod-cpu-request"},
		{queryName: "pod-cpu-usage"},
		{queryName: "pod-memory-request"},
		{queryName: "pod-memory-usage"},
		{queryName: "pod-memory-request-vs-node-memory-allocatable"},
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
		{queryName: "pod-cpu-request-aws", skip: !runAWSBillingTests},
		{queryName: "pod-memory-request-aws", skip: !runAWSBillingTests},
		{queryName: "aws-ec2-cluster-cost", skip: !runAWSBillingTests},
	}

	var scheduledReportsProduceDataTestCases []scheduledReportProducesDataTestCase

	for _, query := range queries {
		scheduledReportHourlyTestCase := scheduledReportProducesDataTestCase{
			name:          query.queryName + "-hourly",
			queryName:     query.queryName,
			schedule:      hourlySchedule,
			newReportFunc: testFramework.NewSimpleScheduledReport,
			timeout:       reportTestTimeout,
			skip:          query.skip,
		}
		scheduledReportRunOnceTestCase := scheduledReportProducesDataTestCase{
			name:          query.queryName + "-runonce",
			queryName:     query.queryName,
			schedule:      nil, // runOnce
			newReportFunc: testFramework.NewSimpleScheduledReport,
			timeout:       reportTestTimeout,
			skip:          query.skip,
		}

		scheduledReportsProduceDataTestCases = append(scheduledReportsProduceDataTestCases, scheduledReportHourlyTestCase)
		scheduledReportsProduceDataTestCases = append(scheduledReportsProduceDataTestCases, scheduledReportRunOnceTestCase)
	}

	t.Run("TestReportingProducesResults", func(t *testing.T) {
		// validate all the ReportDataSources for our tests exist before running
		// collect
		var queries []string
		waitTimeout := time.Minute

		// We wait for all ReportDataSources before anything else since even if
		// we don't use them, the collect endpoint will attempt to collect data
		// for all ReportDataSources
		_, err := testFramework.WaitForAllMeteringReportDataSourceTables(t, time.Second*5, waitTimeout)
		require.NoError(t, err, "should not error when waiting for all ReportDataSource tables to be created")

		seenQuery := make(map[string]struct{})
		for _, test := range scheduledReportsProduceDataTestCases {
			if test.skip {
				continue
			}
			if _, ok := seenQuery[test.queryName]; ok {
				continue
			}
			seenQuery[test.queryName] = struct{}{}
			queries = append(queries, test.queryName)
		}

		// validate all ReportGenerationQueries and ReportDataSources that are
		// used by our test cases are initialized
		testFramework.RequireReportGenerationQueriesReady(t, queries, time.Second*5, waitTimeout)

		var periodStart, periodEnd time.Time
		var collectResp operator.CollectPromsumDataResponse
		periodStart, periodEnd, collectResp = testFramework.CollectMetricsOnce(t)
		testFramework.RequireReportDataSourcesForQueryHaveData(t, queries, collectResp)

		t.Run("TestScheduledReportsProduceData", func(t *testing.T) {
			testScheduledReportsProduceData(t, testFramework, periodStart, periodEnd, scheduledReportsProduceDataTestCases)
		})
	})
}
