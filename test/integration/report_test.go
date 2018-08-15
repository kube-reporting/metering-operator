package integration

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/operator-framework/operator-metering/pkg/operator/prestostore"
	"github.com/operator-framework/operator-metering/pkg/util/orderedmap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const reportComparisionEpsilon = 0.0001

var (
	reportTestTimeout         = 5 * time.Minute
	reportTestOutputDirectory string
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
}

type testDatasource struct {
	DatasourceName string
	FileName       string
}

func TestReportsProduceCorrectDataForInput(t *testing.T) {
	if reportTestOutputDirectory == "" {
		t.Fatalf("$TEST_RESULT_REPORT_OUTPUT_DIRECTORY must be set")
	}
	tests := map[string]struct {
		// name is the name of the sub test but also the name of the report.
		queryName                    string
		dataSources                  []testDatasource
		expectedReportOutputFileName string
		comparisonColumnNames        []string
		timeout                      time.Duration
	}{
		"namespace-cpu-request": {
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
		},
		"namespace-cpu-usage": {
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
		},
		"namespace-memory-request": {
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
		},
		"namespace-memory-usage": {
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
		},
		"pod-cpu-request": {
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
		},
		"pod-cpu-usage": {
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
		},
		"pod-memory-request": {
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
		},
		"pod-memory-usage": {
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
		},
		"node-cpu-utilization": {
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
		},
		"node-memory-utilization": {
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
		},
	}

	// For each datasource file, keep track
	dataSourcesSubmitted := make(map[string]struct {
		metricsStart, metricsEnd time.Time
	})

	for name, test := range tests {
		// Fix closure captures
		test := test
		t.Run(name, func(t *testing.T) {
			if testing.Short() {
				t.Skipf("skipping test in short mode")
				return
			}

			var reportStart, reportEnd time.Time

			for _, dataSource := range test.dataSources {
				if dataSourceTimes, alreadySubmitted := dataSourcesSubmitted[dataSource.DatasourceName]; !alreadySubmitted {
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

					dataSourcesSubmitted[dataSource.DatasourceName] = struct{ metricsStart, metricsEnd time.Time }{
						metricsStart: reportStart,
						metricsEnd:   reportEnd,
					}
				} else {
					reportStart = dataSourceTimes.metricsStart
					reportEnd = dataSourceTimes.metricsEnd
				}
			}

			require.NotZero(t, reportStart, "reportStart should not be zero")
			require.NotZero(t, reportEnd, "reportEnd should not be zero")

			t.Logf("reportStart: %s, reportEnd: %s", reportStart, reportEnd)

			report := testFramework.NewSimpleReport(name, test.queryName, reportStart, reportEnd)

			err := testFramework.MeteringClient.Reports(testFramework.Namespace).Delete(report.Name, nil)
			assert.Condition(t, func() bool {
				return err == nil || errors.IsNotFound(err)
			}, "failed to ensure report doesn't exist before creating report")

			t.Logf("creating report %s", report.Name)
			err = testFramework.CreateMeteringReport(report)
			require.NoError(t, err, "creating report should succeed")

			defer func() {
				t.Logf("deleting report %s", report.Name)
				err := testFramework.MeteringClient.Reports(testFramework.Namespace).Delete(report.Name, nil)
				assert.NoError(t, err, "expected delete report to succeed")
			}()

			query := map[string]string{
				"name":   name,
				"format": "json",
			}

			var reportResults []map[string]interface{}
			var reportData []byte
			err = wait.Poll(time.Second*5, test.timeout, func() (bool, error) {
				req := testFramework.NewReportingOperatorSVCRequest("/api/v1/reports/get", query)
				result := req.Do()
				resp, err := result.Raw()
				if err != nil {
					return false, fmt.Errorf("error querying metering service got error: %v, body: %v", err, string(resp))
				}

				var statusCode int
				result.StatusCode(&statusCode)

				if statusCode == http.StatusAccepted {
					t.Logf("report is still running")
					return false, nil
				}

				require.Equal(t, http.StatusOK, statusCode, "http response status code should be ok")

				err = json.Unmarshal(resp, &reportResults)
				require.NoError(t, err, "expected to unmarshal response")
				reportData = resp
				return true, nil
			})
			require.NoError(t, err, "expected getting report result to not timeout")
			assert.NotEmpty(t, reportResults, "reports should return at least 1 row")

			reportFileName := path.Join(reportTestOutputDirectory, fmt.Sprintf("%s.json", name))
			err = ioutil.WriteFile(reportFileName, reportData, os.ModePerm)
			require.NoError(t, err, "expected writing report results to disk not to error")

			expectedReportData, err := ioutil.ReadFile(test.expectedReportOutputFileName)
			require.NoError(t, err)

			// turn the expected results into a list of maps
			var tmpExpectedResults []map[string]interface{}
			err = json.Unmarshal(expectedReportData, &tmpExpectedResults)
			require.NoError(t, err)

			// turn the actual results into a list of maps
			var tmpResults []map[string]interface{}
			err = json.Unmarshal(reportData, &tmpResults)
			require.NoError(t, err)

			// turn the list of expected results maps into a list of ordered maps
			expectedResults := make([]*orderedmap.OrderedMap, len(tmpExpectedResults))
			for i, item := range tmpExpectedResults {
				expectedResults[i], err = orderedmap.NewFromMap(item)
				require.NoError(t, err)
			}

			// turn the list of actual results maps into a list of ordered maps
			results := make([]*orderedmap.OrderedMap, len(tmpResults))
			for i, item := range tmpResults {
				results[i], err = orderedmap.NewFromMap(item)
				require.NoError(t, err)
			}

			require.Len(t, results, len(expectedResults), "new should have same number of rows as existing report")

			// now that we have a slice of ordered maps, we should be able to
			// iterate over each row, and for each row, iterate over all
			// columns/keys in the row ensuring they match.
			// if the column is the comparison  column, then we allow a small
			// error, due to floating point precision
			// in summary, this does an deep equal comparison with a few tweaks
			// to allow for small error in the calculations.
			for i, row := range results {
				expectedRow := expectedResults[i]
				columns := row.Keys()
				expectedColumns := expectedRow.Keys()
				assert.Equal(t, columns, expectedColumns, "expecting key iteration between actual and expected to be the same")
				for _, column := range columns {

					actualValue, actualExists := row.Get(column)
					if !actualExists {
						t.Errorf("")
					}
					expectedValue, expectedExists := row.Get(column)
					if !expectedExists {
						t.Errorf("")
					}
					isCompareColumn := false
					for _, comparisionColumn := range test.comparisonColumnNames {
						if comparisionColumn == column {
							isCompareColumn = true
							break
						}
					}
					if isCompareColumn {
						assert.InEpsilonf(t, actualValue, expectedValue, reportComparisionEpsilon, "expected column %q value to be within delta of expected row", column)
					} else {
						assert.Equal(t, actualValue, expectedValue, "expected column values between actual and expected rows to be the same")
					}
				}
			}
		})
	}
}
