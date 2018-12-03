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

	"github.com/operator-framework/operator-metering/test/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const reportComparisionEpsilon = 0.0001

var (
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
		},
	}
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

type reportsProduceCorrectDataForInputTestCase struct {
	name                         string
	queryName                    string
	dataSources                  []testDatasource
	expectedReportOutputFileName string
	comparisonColumnNames        []string
	timeout                      time.Duration
}

type testDatasource struct {
	DatasourceName string
	FileName       string
}

func testReportsProduceCorrectDataForInput(t *testing.T, reportStart, reportEnd time.Time, testCases []reportsProduceCorrectDataForInputTestCase) {
	if reportTestOutputDirectory == "" {
		t.Fatalf("$TEST_RESULT_REPORT_OUTPUT_DIRECTORY must be set")
	}

	require.NotZero(t, reportStart, "reportStart should not be zero")
	require.NotZero(t, reportEnd, "reportEnd should not be zero")

	for _, test := range testCases {
		// Fix closure captures
		name := test.name
		test := test
		t.Run(name, func(t *testing.T) {
			if testing.Short() {
				t.Skipf("skipping test in short mode")
				return
			}

			t.Logf("reportStart: %s, reportEnd: %s", reportStart, reportEnd)

			report := testFramework.NewSimpleReport(test.name, test.queryName, &reportStart, &reportEnd)

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
			var expectedResults []map[string]interface{}
			err = json.Unmarshal(expectedReportData, &expectedResults)
			require.NoError(t, err)

			// turn the actual results into a list of maps
			var actualResults []map[string]interface{}
			err = json.Unmarshal(reportData, &actualResults)
			require.NoError(t, err)

			testhelpers.AssertReportResultsEqual(t, expectedResults, actualResults, test.comparisonColumnNames)
		})
	}
}
