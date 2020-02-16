package e2e

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/types"

	"github.com/operator-framework/operator-metering/pkg/operator/prestostore"
	"github.com/operator-framework/operator-metering/test/reportingframework"
	"github.com/operator-framework/operator-metering/test/testhelpers"
)

var (
	reportTestTimeout                              = 5 * time.Minute
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
			name:      "namespace-persistentvolumeclaim-usage",
			queryName: "namespace-persistentvolumeclaim-usage",
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
			expectedReportOutputFileName: "testdata/reports/namespace-persistentvolumeclaim-usage.json",
			comparisonColumnNames:        []string{"persistentvolumeclaim_usage_bytes"},
			timeout:                      reportTestTimeout,
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
		{
			name:      "persistentvolumeclaim-capacity",
			queryName: "persistentvolumeclaim-capacity",
			dataSources: []testDatasource{
				{
					DatasourceName: "persistentvolumeclaim-capacity-bytes",
					FileName:       "testdata/datasources/persistentvolumeclaim-capacity-bytes.json",
				},
			},
			expectedReportOutputFileName: "testdata/reports/persistentvolumeclaim-capacity.json",
			comparisonColumnNames:        []string{"persistentvolumeclaim_capacity_bytes"},
			timeout:                      reportTestTimeout,
		},
	}
)

func testReportingProducesCorrectDataForInput(t *testing.T, testReportingFramework *reportingframework.ReportingFramework) {
	var queries []string
	t.Logf("Waiting for ReportDataSources tables to be created")
	_, err := testReportingFramework.WaitForAllMeteringReportDataSourceTables(t, time.Second*5, 5*time.Minute)
	require.NoError(t, err, "should not error when waiting for all ReportDataSource tables to be created")

	for _, test := range testReportsProduceCorrectDataForInputTestCases {
		queries = append(queries, test.queryName)
	}

	// validate all ReportQueries and ReportDataSources that are
	// used by the test cases are initialized
	t.Logf("Waiting for ReportQueries tables to become ready")
	testReportingFramework.RequireReportQueriesReady(t, queries, time.Second*5, 5*time.Minute)

	var reportStart, reportEnd time.Time
	dataSourcesSubmitted := make(map[string]struct{})

	// Inject all the dataSources we require for each test case
	t.Logf("Pushing fixture metrics required for tests into metering")
	for _, test := range testReportsProduceCorrectDataForInputTestCases {
		for _, dataSource := range test.dataSources {
			if _, alreadySubmitted := dataSourcesSubmitted[dataSource.DatasourceName]; !alreadySubmitted {
				// wait for the datasource table to exist
				_, err := testReportingFramework.WaitForMeteringReportDataSourceTable(t, dataSource.DatasourceName, time.Second*5, 2*time.Minute)
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
						err := testReportingFramework.StoreDataSourceData(dataSource.DatasourceName, metrics)
						require.NoError(t, err)
						metrics = nil
					}
				}
				// flush any metrics left over
				if len(metrics) != 0 {
					err = testReportingFramework.StoreDataSourceData(dataSource.DatasourceName, metrics)
					require.NoError(t, err)
				}

				reportEndStr := reportEnd.UTC().Format(time.RFC3339)
				reportStartStr := reportStart.UTC().Format(time.RFC3339)
				nowStr := time.Now().UTC().Format(time.RFC3339)
				jsonPatch := []byte(fmt.Sprintf(
					`[{ "op": "add", "path": "/status/prometheusMetricsImportStatus", "value": { "importDataStartTime": "%s", "importDataEndTime": "%s", "earliestImportedMetricTime": "%s", "newestImportedMetricTime": "%s", "lastImportTime": "%s" } } ]`,
					reportStartStr, reportEndStr, reportStartStr, reportEndStr, nowStr))
				_, err = testReportingFramework.MeteringClient.ReportDataSources(testReportingFramework.Namespace).Patch(dataSource.DatasourceName, types.JSONPatchType, jsonPatch)
				require.NoError(t, err)

				dataSourcesSubmitted[dataSource.DatasourceName] = struct{}{}
			}
		}
	}

	testReportsProduceCorrectDataForInput(t, testReportingFramework, reportStart, reportEnd, testReportsProduceCorrectDataForInputTestCases)
}

type testDatasource struct {
	DatasourceName string
	FileName       string
}

type reportsProduceCorrectDataForInputTestCase struct {
	name                         string
	queryName                    string
	dataSources                  []testDatasource
	expectedReportOutputFileName string
	comparisonColumnNames        []string
	timeout                      time.Duration
	parallel                     bool
}

func testReportsProduceCorrectDataForInput(
	t *testing.T,
	testReportingFramework *reportingframework.ReportingFramework,
	reportStart,
	reportEnd time.Time,
	testCases []reportsProduceCorrectDataForInputTestCase,
) {
	require.NotZero(t, reportStart, "reportStart should not be zero")
	require.NotZero(t, reportEnd, "reportEnd should not be zero")
	t.Logf("reportStart: %s, reportEnd: %s", reportStart, reportEnd)
	for _, test := range testCases {
		// Fix closure captures
		name := test.name
		test := test
		t.Run(name, func(t *testing.T) {
			if test.parallel {
				t.Parallel()
			}

			report := testReportingFramework.NewSimpleReport(test.name, test.queryName, nil, &reportStart, &reportEnd)

			reportRunTimeout := 10 * time.Minute
			t.Logf("creating report %s and waiting %s to finish", report.Name, reportRunTimeout)
			testReportingFramework.RequireReportSuccessfullyRuns(t, report, reportRunTimeout)

			resultTimeout := time.Minute
			t.Logf("waiting %s for report %s results", resultTimeout, report.Name)
			actualResults := testReportingFramework.GetReportResults(t, report, resultTimeout)

			// read expected results from a file
			expectedReportData, err := ioutil.ReadFile(test.expectedReportOutputFileName)
			require.NoError(t, err)
			// turn the expected results into a list of maps
			var expectedResults []map[string]interface{}
			err = json.Unmarshal(expectedReportData, &expectedResults)
			require.NoError(t, err)

			testhelpers.AssertReportResultsEqual(t, expectedResults, actualResults, test.comparisonColumnNames)
		})
	}
}
