package e2e

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/wait"

	meteringv1alpha1 "github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	reportsProduceDataTestCases = []struct {
		// name is the name of the sub test but also the name of the report.
		name      string
		queryName string
		timeout   time.Duration
		skip      bool
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
			skip:      !runAWSBillingTests,
		},
		{
			name:      "pod-memory-request-aws",
			queryName: "pod-memory-request-aws",
			timeout:   reportTestTimeout,
			skip:      !runAWSBillingTests,
		},
		{
			name:      "aws-ec2-cluster-cost",
			queryName: "aws-ec2-cluster-cost",
			timeout:   reportTestTimeout,
			skip:      !runAWSBillingTests,
		},
	}
)

func testReportsProduceData(t *testing.T) {
	t.Logf("reportStart: %s, reportEnd: %s", periodStart, periodEnd)
	for i, test := range reportsProduceDataTestCases {
		// Fix closure captures
		test := test
		i := i
		// The JVM has a warm up time and the ***REMOVED***rst report always takes longer
		// than others, so give it a longer timeout
		if i == 0 {
			test.timeout += time.Minute
		}

		t.Run(test.name, func(t *testing.T) {
			if testing.Short() && i != 0 {
				t.Skip("skipping test in short mode")
				return
			}

			if test.skip {
				t.Skip("test con***REMOVED***gured to be skipped")
				return
			}

			report := testFramework.NewSimpleReport(test.name, test.queryName, periodStart, periodEnd)

			err := testFramework.MeteringClient.Reports(testFramework.Namespace).Delete(report.Name, nil)
			require.Condition(t, func() bool {
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
				"name":   test.name,
				"format": "json",
			}

			err = wait.PollImmediate(time.Second*5, test.timeout, func() (bool, error) {
				// poll the status
				newReport, err := testFramework.GetMeteringReport(report.Name)
				if err != nil {
					return false, err
				}
				if newReport.Status.Phase == meteringv1alpha1.ReportPhaseError {
					return false, fmt.Errorf("report is failed, message: %s", newReport.Status.Output)
				}

				if newReport.Status.TableName == "" {
					t.Logf("Report %s table isn't created yet, status: %+v", report.Name, newReport.Status)
					return false, nil
				}
				return true, nil
			})
			require.NoError(t, err, "expected getting Report to not timeout")

			var reportResults []map[string]interface{}
			var reportData []byte
			err = wait.PollImmediate(time.Second*5, test.timeout, func() (bool, error) {
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

			***REMOVED***leName := path.Join(reportTestOutputDirectory, fmt.Sprintf("%s-report.json", test.name))
			err = ioutil.WriteFile(***REMOVED***leName, reportData, os.ModePerm)
			require.NoError(t, err, "expected writing report results to disk not to error")
		})
	}
}
