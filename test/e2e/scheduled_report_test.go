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

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/wait"

	meteringv1alpha1 "github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1"
	cbutil "github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1/util"
	"github.com/operator-framework/operator-metering/test/framework"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type scheduledReportProducesDataTestCase struct {
	name      string
	queryName string
	timeout   time.Duration
}

func testScheduledReportsProduceData(t *testing.T, testFramework *framework.Framework, periodStart, periodEnd time.Time, testCases []scheduledReportProducesDataTestCase) {
	t.Logf("periodStart: %s, periodEnd: %s", periodStart, periodEnd)
	for _, test := range testCases {
		name := test.name
		// Fix closure captures
		test := test
		t.Run(name, func(t *testing.T) {
			// all scheduled reports get skipped in short test mode
			if testing.Short() {
				t.Skipf("skipping test in short mode")
				return
			}

			// reportStart needs to be at least 1 hour and 5 minutes ago
			// because NewSimpleScheduledReport produces hourly reports, and it
			// won't run unless we set reportingStart in the past by the period
			// (1 hour) + the default gracePeriod (5 minutes).
			// to do this, we set reportStart to 1.5 hours back if it isn't
			// already
			reportStart := periodStart
			reportEnd := periodEnd
			if reportEnd.Sub(reportStart) < time.Hour {
				reportStart = reportEnd.Add(-time.Hour)
			}
			t.Logf("scheduledReport reportingStart: %s, reportingEnd: %s", reportStart, reportEnd)
			report := testFramework.NewSimpleScheduledReport(name, test.queryName, &reportStart, &reportEnd)

			err := testFramework.MeteringClient.ScheduledReports(testFramework.Namespace).Delete(report.Name, nil)
			assert.Condition(t, func() bool {
				return err == nil || errors.IsNotFound(err)
			}, "failed to ensure scheduled report doesn't exist before creating scheduled report")

			t.Logf("creating scheduled report %s", report.Name)
			err = testFramework.CreateMeteringScheduledReport(report)
			require.NoError(t, err, "creating scheduled report should succeed")

			defer func() {
				t.Logf("deleting scheduled report %s", report.Name)
				err := testFramework.MeteringClient.ScheduledReports(testFramework.Namespace).Delete(report.Name, nil)
				assert.NoError(t, err, "expected delete scheduled report to succeed")
			}()

			query := map[string]string{
				"name":   name,
				"format": "json",
			}

			err = wait.PollImmediate(time.Second*5, test.timeout, func() (bool, error) {
				// poll the status
				newReport, err := testFramework.GetMeteringScheduledReport(report.Name)
				if err != nil {
					return false, err
				}
				cond := cbutil.GetScheduledReportCondition(newReport.Status, meteringv1alpha1.ScheduledReportFailure)
				if cond != nil && cond.Status == v1.ConditionTrue {
					return false, fmt.Errorf("report is failed, reason: %s, message: %s", cond.Reason, cond.Message)
				}

				if newReport.Status.TableName == "" {
					t.Logf("ScheduledReport %s table isn't created yet, status: %+v", report.Name, newReport.Status)
					return false, nil
				}

				// If the last reportTime is updated, that means this report
				// has been collected at least once.
				if newReport.Status.LastReportTime == nil {
					t.Logf("report LastReportTime is unset")
					return false, nil
				}
				return true, nil
			})
			require.NoError(t, err, "expected getting ScheduledReport to not timeout")

			var reportResults []map[string]interface{}
			var reportData []byte
			err = wait.PollImmediate(time.Second*5, test.timeout, func() (bool, error) {
				req := testFramework.NewReportingOperatorSVCRequest("/api/v1/scheduledreports/get", query)
				result := req.Do()
				resp, err := result.Raw()
				require.NoError(t, err, "fetching ScheduledReport results should be successful")

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
			require.NoError(t, err, "expected getting ScheduledReport result to not timeout")
			assert.NotEmpty(t, reportResults, "reports should return at least 1 row")

			***REMOVED***leName := path.Join(reportTestOutputDirectory, fmt.Sprintf("%s-scheduled-report.json", name))
			err = ioutil.WriteFile(***REMOVED***leName, reportData, os.ModePerm)
			require.NoError(t, err, "expected writing report results to disk not to error")
		})
	}
}
