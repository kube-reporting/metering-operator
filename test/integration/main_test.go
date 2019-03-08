package integration

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	types "k8s.io/apimachinery/pkg/types"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"

	"github.com/operator-framework/operator-metering/pkg/operator/prestostore"
	"github.com/operator-framework/operator-metering/test/framework"
)

var (
	testFramework *framework.Framework

	reportTestTimeout         = 5 * time.Minute
	reportTestOutputDirectory string

	testReportsProduceCorrectDataForInputTestCases = []reportsProduceCorrectDataForInputTestCase{
		{
			name:      "namespace-cpu-request",
			queryName: "namespace-cpu-request",
			dataSources: []testDatasource{
				{
					DatasourceName: "pod-request-cpu-cores",
					FileName:       "testdata/datasources/pod-request-cpu-cores.json",
				},
			},
			expectedReportOutputFileName: "testdata/reports/namespace-cpu-request.json",
			comparisonColumnNames:        []string{"pod_request_cpu_core_seconds"},
			timeout:                      reportTestTimeout,
			parallel:                     true,
		},
		{
			name:      "namespace-cpu-usage",
			queryName: "namespace-cpu-usage",
			dataSources: []testDatasource{
				{
					DatasourceName: "pod-usage-cpu-cores",
					FileName:       "testdata/datasources/pod-usage-cpu-cores.json",
				},
			},
			expectedReportOutputFileName: "testdata/reports/namespace-cpu-usage.json",
			comparisonColumnNames:        []string{"pod_usage_cpu_core_seconds"},
			timeout:                      reportTestTimeout,
			parallel:                     true,
		},
		{
			name:      "namespace-memory-request",
			queryName: "namespace-memory-request",
			dataSources: []testDatasource{
				{
					DatasourceName: "pod-request-memory-bytes",
					FileName:       "testdata/datasources/pod-request-memory-bytes.json",
				},
			},
			expectedReportOutputFileName: "testdata/reports/namespace-memory-request.json",
			comparisonColumnNames:        []string{"pod_request_memory_byte_seconds"},
			timeout:                      reportTestTimeout,
			parallel:                     true,
		},
		{
			name:      "namespace-memory-usage",
			queryName: "namespace-memory-usage",
			dataSources: []testDatasource{
				{
					DatasourceName: "pod-usage-memory-bytes",
					FileName:       "testdata/datasources/pod-usage-memory-bytes.json",
				},
			},
			expectedReportOutputFileName: "testdata/reports/namespace-memory-usage.json",
			comparisonColumnNames:        []string{"pod_usage_memory_core_seconds"},
			timeout:                      reportTestTimeout,
			parallel:                     true,
		},
		{
			name:      "pod-cpu-request",
			queryName: "pod-cpu-request",
			dataSources: []testDatasource{
				{
					DatasourceName: "pod-request-cpu-cores",
					FileName:       "testdata/datasources/pod-request-cpu-cores.json",
				},
			},
			expectedReportOutputFileName: "testdata/reports/pod-cpu-request.json",
			comparisonColumnNames:        []string{"pod_request_cpu_core_seconds"},
			timeout:                      reportTestTimeout,
			parallel:                     true,
		},
		{
			name:      "pod-cpu-usage",
			queryName: "pod-cpu-usage",
			dataSources: []testDatasource{
				{
					DatasourceName: "pod-usage-cpu-cores",
					FileName:       "testdata/datasources/pod-usage-cpu-cores.json",
				},
			},
			expectedReportOutputFileName: "testdata/reports/pod-cpu-usage.json",
			comparisonColumnNames:        []string{"pod_usage_cpu_core_seconds"},
			timeout:                      reportTestTimeout,
			parallel:                     true,
		},
		{
			name:      "pod-memory-request",
			queryName: "pod-memory-request",
			dataSources: []testDatasource{
				{
					DatasourceName: "pod-request-memory-bytes",
					FileName:       "testdata/datasources/pod-request-memory-bytes.json",
				},
			},
			expectedReportOutputFileName: "testdata/reports/pod-memory-request.json",
			comparisonColumnNames:        []string{"pod_request_memory_byte_seconds"},
			timeout:                      reportTestTimeout,
			parallel:                     true,
		},
		{
			name:      "pod-memory-usage",
			queryName: "pod-memory-usage",
			dataSources: []testDatasource{
				{
					DatasourceName: "pod-usage-memory-bytes",
					FileName:       "testdata/datasources/pod-usage-memory-bytes.json",
				},
			},
			expectedReportOutputFileName: "testdata/reports/pod-memory-usage.json",
			comparisonColumnNames:        []string{"pod_usage_memory_byte_seconds"},
			timeout:                      reportTestTimeout,
			parallel:                     true,
		},
		{
			name:      "node-cpu-utilization",
			queryName: "node-cpu-utilization",
			dataSources: []testDatasource{
				{
					DatasourceName: "node-allocatable-cpu-cores",
					FileName:       "testdata/datasources/node-allocatable-cpu-cores.json",
				},
				{
					DatasourceName: "pod-request-cpu-cores",
					FileName:       "testdata/datasources/pod-request-cpu-cores.json",
				},
			},
			expectedReportOutputFileName: "testdata/reports/node-cpu-utilization.json",
			comparisonColumnNames:        []string{"node_allocatable_cpu_core_seconds", "pod_request_cpu_core_seconds", "cpu_used_percent", "cpu_unused_percent"},
			timeout:                      reportTestTimeout,
			parallel:                     true,
		},
		{
			name:      "node-memory-utilization",
			queryName: "node-memory-utilization",
			dataSources: []testDatasource{
				{
					DatasourceName: "node-allocatable-memory-bytes",
					FileName:       "testdata/datasources/node-allocatable-memory-bytes.json",
				},
				{
					DatasourceName: "pod-request-memory-bytes",
					FileName:       "testdata/datasources/pod-request-memory-bytes.json",
				},
			},
			expectedReportOutputFileName: "testdata/reports/node-memory-utilization.json",
			comparisonColumnNames:        []string{"node_allocatable_memory_byte_seconds", "pod_request_memory_byte_seconds", "memory_used_percent", "memory_unused_percent"},
			timeout:                      reportTestTimeout,
			parallel:                     true,
		},
		{
			name:      "persistentvolumeclaim-usage",
			queryName: "persistentvolumeclaim-usage",
			dataSources: []testDatasource{
				{
					DatasourceName: "persistentvolumeclaim-phase",
					FileName:       "testdata/datasources/persistentvolumeclaim-phase.json",
				},
				{
					DatasourceName: "persistentvolumeclaim-usage-bytes",
					FileName:       "testdata/datasources/persistentvolumeclaim-usage-bytes.json",
				},
			},
			expectedReportOutputFileName: "testdata/reports/persistentvolumeclaim-usage.json",
			comparisonColumnNames:        []string{"persistentvolumeclaim_usage_bytes"},
			timeout:                      reportTestTimeout,
		},
	}
)

type testDatasource struct {
	DatasourceName string
	FileName       string
}

func init() {
	reportTestOutputDirectory = os.Getenv("TEST_RESULT_REPORT_OUTPUT_DIRECTORY")
	if reportTestOutputDirectory == "" {
		log.Fatalf("$TEST_RESULT_REPORT_OUTPUT_DIRECTORY must be set")
	}

	err := os.MkdirAll(reportTestOutputDirectory, 0777)
	if err != nil {
		log.Fatalf("error making directory %s, err: %s", reportTestOutputDirectory, err)
	}
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

func TestReportingProducesCorrectDataForInput(t *testing.T) {
	var queries []string
	waitTimeout := time.Minute

	t.Logf("Waiting for ReportDataSources tables to be created")
	_, err := testFramework.WaitForAllMeteringReportDataSourceTables(t, time.Second*5, waitTimeout)
	require.NoError(t, err, "should not error when waiting for all ReportDataSource tables to be created")

	for _, test := range testReportsProduceCorrectDataForInputTestCases {
		queries = append(queries, test.queryName)
	}

	// validate all ReportGenerationQueries and ReportDataSources that are
	// used by the test cases are initialized
	t.Logf("Waiting for ReportGenerationQueries tables to become ready")
	testFramework.RequireReportGenerationQueriesReady(t, queries, time.Second*5, waitTimeout)

	var reportStart, reportEnd time.Time
	dataSourcesSubmitted := make(map[string]struct{})

	// Inject all the dataSources we require for each test case
	t.Logf("Pushing fixture metrics required for tests into metering")
	for _, test := range testReportsProduceCorrectDataForInputTestCases {
		for _, dataSource := range test.dataSources {
			if _, alreadySubmitted := dataSourcesSubmitted[dataSource.DatasourceName]; !alreadySubmitted {
				// wait for the datasource table to exist
				_, err := testFramework.WaitForMeteringReportDataSourceTable(t, dataSource.DatasourceName, time.Second*5, test.timeout)
				require.NoError(t, err, "ReportDataSource table should exist before storing data into it")

				metricsFile, err := os.Open(dataSource.FileName)
				require.NoError(t, err)

				decoder := json.NewDecoder(metricsFile)

				_, err = decoder.Token()
				require.NoError(t, err)

				var metrics []*prestostore.PrometheusMetric
				for decoder.More() {
					var metric prestostore.PrometheusMetric
					err = decoder.Decode(&metric)
					require.NoError(t, err)
					if reportStart.IsZero() || metric.Timestamp.Before(reportStart) {
						reportStart = metric.Timestamp
					}
					if metric.Timestamp.After(reportEnd) {
						reportEnd = metric.Timestamp
					}
					metrics = append(metrics, &metric)
					// batch store metrics in amounts of 100
					if len(metrics) >= 100 {
						err := testFramework.StoreDataSourceData(dataSource.DatasourceName, metrics)
						require.NoError(t, err)
						metrics = nil
					}
				}
				// flush any metrics left over
				if len(metrics) != 0 {
					err = testFramework.StoreDataSourceData(dataSource.DatasourceName, metrics)
					require.NoError(t, err)
				}

				reportEndStr := reportEnd.UTC().Format(time.RFC3339)
				reportStartStr := reportStart.UTC().Format(time.RFC3339)
				nowStr := time.Now().UTC().Format(time.RFC3339)
				jsonPatch := []byte(fmt.Sprintf(
					`[{ "op": "add", "path": "/status/prometheusMetricImportStatus", "value": { "importDataEndTime": "%s", "earliestImportedMetricTime": "%s", "newestImportedMetricTime": "%s", "lastImportTime": "%s" } } ]`,
					reportEndStr, reportStartStr, reportEndStr, nowStr))
				_, err = testFramework.MeteringClient.ReportDataSources(testFramework.Namespace).Patch(dataSource.DatasourceName, types.JSONPatchType, jsonPatch)
				require.NoError(t, err)

				dataSourcesSubmitted[dataSource.DatasourceName] = struct{}{}
			}
		}
	}

	testReportsProduceCorrectDataForInput(t, reportStart, reportEnd, testReportsProduceCorrectDataForInputTestCases)
}
