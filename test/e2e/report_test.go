package e2e

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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	reportTestTimeout         = 5 * time.Minute
	reportTestOutputDirectory string
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
}

func TestReportsProduceData(t *testing.T) {
	tests := []struct {
		// name is the name of the sub test but also the name of the report.
		name      string
		queryName string
		timeout   time.Duration
	}{
		{
			name:      "namespace-cpu-request",
			queryName: "namespace-cpu-request",
			timeout:   reportTestTimeout,
		},
		{
			name:      "namespace-cpu-usage",
			queryName: "namespace-cpu-usage",
			timeout:   reportTestTimeout,
		},
		{
			name:      "namespace-memory-request",
			queryName: "namespace-memory-request",
			timeout:   reportTestTimeout + time.Minute,
		},
		{
			name:      "namespace-memory-usage",
			queryName: "namespace-memory-usage",
			timeout:   reportTestTimeout + time.Minute,
		},
		{
			name:      "pod-cpu-request",
			queryName: "pod-cpu-request",
			timeout:   reportTestTimeout,
		},
		{
			name:      "pod-cpu-usage",
			queryName: "pod-cpu-usage",
			timeout:   reportTestTimeout,
		},
		{
			name:      "pod-memory-request",
			queryName: "pod-memory-request",
			timeout:   reportTestTimeout,
		},
		{
			name:      "pod-memory-usage",
			queryName: "pod-memory-usage",
			timeout:   reportTestTimeout,
		},
		{
			name:      "pod-memory-request-vs-node-memory-allocatable",
			queryName: "pod-memory-request-vs-node-memory-allocatable",
			timeout:   reportTestTimeout + time.Minute,
		},
		{
			name:      "node-cpu-utilization",
			queryName: "node-cpu-utilization",
			timeout:   reportTestTimeout,
		},
		{
			name:      "node-memory-utilization",
			queryName: "node-memory-utilization",
			timeout:   reportTestTimeout,
		},
		{
			name:      "pod-cpu-request-aws",
			queryName: "pod-cpu-request-aws",
			timeout:   reportTestTimeout,
		},
		{
			name:      "pod-memory-request-aws",
			queryName: "pod-memory-request-aws",
			timeout:   reportTestTimeout,
		},
		{
			name:      "aws-ec2-cluster-cost",
			queryName: "aws-ec2-cluster-cost",
			timeout:   reportTestTimeout,
		},
	}

	reportStart, reportEnd := collectMetricsOnce(t)
	t.Logf("reportStart: %s, reportEnd: %s", reportStart, reportEnd)

	for i, test := range tests {
		// Fix closure captures
		test := test
		i := i
		// The JVM has a warm up time and the first report always takes longer
		// than others, so give it a longer timeout
		if i == 0 {
			test.timeout += time.Minute
		}

		t.Run(test.name, func(t *testing.T) {
			if testing.Short() && i != 0 {
				t.Skipf("skipping test in short mode")
				return
			}

			report := testFramework.NewSimpleReport(test.name, test.queryName, reportStart, reportEnd)

			err := testFramework.ChargebackClient.Reports(testFramework.Namespace).Delete(report.Name, nil)
			assert.Condition(t, func() bool {
				return err == nil || errors.IsNotFound(err)
			}, "failed to ensure report doesn't exist before creating report")

			t.Logf("creating report %s", report.Name)
			err = testFramework.CreateChargebackReport(report)
			require.NoError(t, err, "creating report should succeed")

			defer func() {
				t.Logf("deleting report %s", report.Name)
				err := testFramework.ChargebackClient.Reports(testFramework.Namespace).Delete(report.Name, nil)
				assert.NoError(t, err, "expected delete report to succeed")
			}()

			query := map[string]string{
				"name":   test.name,
				"format": "json",
			}

			var reportResults []map[string]interface{}
			var reportData []byte
			err = wait.Poll(time.Second*5, test.timeout, func() (bool, error) {
				req := testFramework.NewChargebackSVCRequest("/api/v1/reports/get", query)
				result := req.Do()
				resp, err := result.Raw()
				if err != nil {
					return false, fmt.Errorf("error querying chargeback service got error: %v, body: %v", err, string(resp))
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

			fileName := path.Join(reportTestOutputDirectory, fmt.Sprintf("%s.json", test.name))
			err = ioutil.WriteFile(fileName, reportData, os.ModePerm)
			require.NoError(t, err, "expected writing report results to disk not to error")
		})
	}
}
