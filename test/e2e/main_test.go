package e2e

import (
	"flag"
	"log"
	"os"
	"testing"
	"time"

	"github.com/operator-framework/operator-metering/pkg/operator/reporting"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"

	"github.com/operator-framework/operator-metering/test/framework"
)

var (
	testFramework *framework.Framework

	reportTestTimeout         = 5 * time.Minute
	reportTestOutputDirectory string
	runAWSBillingTests        bool

	periodStart, periodEnd time.Time
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
	t.Run("TestReportingProducesResults", func(t *testing.T) {
		// validate all the ReportDataSources for our tests exist before running
		// collect
		var queries []string
		readyReportDataSources := make(map[string]struct{})
		readyReportGenQueries := make(map[string]struct{})
		waitTimeout := time.Minute

		// We wait for all ReportDataSources before anything else since even if
		// we don't use them, the collect endpoint will attempt to collect data
		// for all ReportDataSources
		_, err := testFramework.WaitForAllMeteringReportDataSourceTables(t, time.Second*5, waitTimeout)
		require.NoError(t, err, "should not error when waiting for all ReportDataSource tables to be created")

		// below we validate all ReportGenerationQueries and ReportDataSources
		// that are used by our test cases are initialized
		queryGetter := reporting.NewReportGenerationQueryClientGetter(testFramework.MeteringClient)
		dataSourceGetter := reporting.NewReportDataSourceClientGetter(testFramework.MeteringClient)
		for _, test := range reportsProduceDataTestCases {
			if test.skip {
				continue
			}
			queries = append(queries, test.queryName)
		}
		for _, test := range scheduledReportsProduceDataTestCases {
			queries = append(queries, test.queryName)
		}

		for _, queryName := range queries {
			if _, exists := readyReportGenQueries[queryName]; exists {
				continue
			}

			t.Logf("waiting for ReportGenerationQuery %s to exist", queryName)
			reportGenQuery, err := testFramework.WaitForMeteringReportGenerationQuery(t, queryName, time.Second*5, waitTimeout)
			require.NoError(t, err, "ReportGenerationQuery should exist before creating report using it")

			depStatus, err := reporting.GetGenerationQueryDependenciesStatus(queryGetter, dataSourceGetter, reportGenQuery)
			require.NoError(t, err, "should not have errors getting dependent ReportGenerationQueries")

			var uninitializedReportGenerationQueries, uninitializedReportDataSources []string

			for _, q := range depStatus.UninitializedReportGenerationQueries {
				uninitializedReportGenerationQueries = append(uninitializedReportGenerationQueries, q.Name)
			}

			for _, ds := range depStatus.UninitializedReportDataSources {
				uninitializedReportDataSources = append(uninitializedReportDataSources, ds.Name)
			}

			t.Logf("waiting for ReportGenerationQuery %s UninitializedReportGenerationQueries: %v", queryName, uninitializedReportGenerationQueries)
			t.Logf("waiting for ReportGenerationQuery %s UninitializedReportDataSources: %v", queryName, uninitializedReportDataSources)

			for _, q := range depStatus.UninitializedReportGenerationQueries {
				if _, exists := readyReportGenQueries[q.Name]; exists {
					continue
				}

				t.Logf("waiting for ReportGenerationQuery %s to exist", q.Name)
				_, err := testFramework.WaitForMeteringReportGenerationQuery(t, q.Name, time.Second*5, waitTimeout)
				require.NoError(t, err, "ReportGenerationQuery should exist before creating report using it")

				readyReportGenQueries[queryName] = struct{}{}
			}

			for _, ds := range depStatus.UninitializedReportDataSources {
				if _, exists := readyReportDataSources[ds.Name]; exists {
					continue
				}
				t.Logf("waiting for ReportDataSource %s to exist", ds.Name)
				_, err := testFramework.WaitForMeteringReportDataSourceTable(t, ds.Name, time.Second*5, waitTimeout)
				require.NoError(t, err, "ReportDataSource %s table for ReportGenerationQuery %s should exist before running reports against it", ds.Name, queryName)
				readyReportDataSources[ds.Name] = struct{}{}
			}

			readyReportGenQueries[queryName] = struct{}{}
		}

		periodStart, periodEnd = testFramework.CollectMetricsOnce(t)

		t.Run("TestReportsProduceData", testReportsProduceData)
		t.Run("TestScheduledReportsProduceData", testScheduledReportsProduceData)
	})
}
