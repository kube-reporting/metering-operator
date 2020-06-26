package reportingframework

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	metering "github.com/kube-reporting/metering-operator/pkg/apis/metering/v1"
	meteringUtil "github.com/kube-reporting/metering-operator/pkg/apis/metering/v1/util"
)

func (rf *ReportingFramework) CreateMeteringReport(report *metering.Report) error {
	_, err := rf.MeteringClient.Reports(rf.Namespace).Create(context.Background(), report, metav1.CreateOptions{})
	return err
}

func (rf *ReportingFramework) GetMeteringReport(name string) (*metering.Report, error) {
	return rf.MeteringClient.Reports(rf.Namespace).Get(context.Background(), name, metav1.GetOptions{})
}

func (rf *ReportingFramework) NewSimpleReport(name, queryName string, schedule *metering.ReportSchedule, reportingStart, reportingEnd *time.Time) *metering.Report {
	var start, end *metav1.Time
	if reportingStart != nil {
		start = &metav1.Time{Time: *reportingStart}
	}
	if reportingEnd != nil {
		end = &metav1.Time{Time: *reportingEnd}
	}
	return &metering.Report{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: rf.Namespace,
		},
		Spec: metering.ReportSpec{
			QueryName:      queryName,
			Schedule:       schedule,
			ReportingStart: start,
			ReportingEnd:   end,
		},
	}
}

func (rf *ReportingFramework) RequireReportSuccessfullyRuns(t *testing.T, report *metering.Report, waitTimeout time.Duration) {
	t.Helper()
	err := rf.MeteringClient.Reports(rf.Namespace).Delete(context.Background(), report.Name, metav1.DeleteOptions{})
	assert.Condition(t, func() bool {
		return err == nil || errors.IsNotFound(err)
	}, "failed to ensure Report doesn't exist before creating")

	t.Logf("creating Report %s", report.Name)
	err = rf.CreateMeteringReport(report)
	require.NoError(t, err, "creating Report should succeed")

	prevLogMsg := ""
	err = wait.Poll(time.Second*5, waitTimeout, func() (bool, error) {
		// poll the status
		report, err := rf.GetMeteringReport(report.Name)
		if err != nil {
			return false, err
		}
		cond := meteringUtil.GetReportCondition(report.Status, metering.ReportRunning)
		if cond != nil && cond.Status == v1.ConditionFalse && cond.Reason == meteringUtil.ReportFinishedReason {
			return true, nil
		}

		newLogMsg := fmt.Sprintf("Report %s not finished, status: %s", report.Name, spew.Sprintf("%v", report.Status))
		if newLogMsg != prevLogMsg {
			t.Log(newLogMsg)
		}
		prevLogMsg = newLogMsg
		return false, nil
	})
	require.NoErrorf(t, err, "expected Report to finished within %s timeout", waitTimeout)
}

func (rf *ReportingFramework) GetReportResults(t *testing.T, report *metering.Report, waitTimeout time.Duration) []map[string]interface{} {
	t.Helper()
	var reportResults []map[string]interface{}
	var reportData []byte

	queryParams := map[string]string{
		"name":      report.Name,
		"namespace": report.Namespace,
		"format":    "json",
	}
	err := wait.Poll(time.Second*5, waitTimeout, func() (bool, error) {
		respBody, respCode, err := rf.ReportingOperatorRequest("/api/v1/reports/get", queryParams)
		require.NoError(t, err, "fetching Report results should be successful")

		if respCode == http.StatusAccepted {
			t.Logf("Report %s is still running", report.Name)
			return false, nil
		}

		require.Equal(t, http.StatusOK, respCode, "http response status code should be ok")
		err = json.Unmarshal(respBody, &reportResults)
		require.NoError(t, err, "expected to unmarshal response")
		reportData = respBody
		return true, nil
	})
	require.NoError(t, err, "expected Report to have 1 row of results before timing out")
	assert.NotEmpty(t, reportResults, "reports should return at least 1 row")

	fileName := path.Join(rf.ReportOutputDirectory, fmt.Sprintf("%s.json", report.Name))
	err = ioutil.WriteFile(fileName, reportData, os.ModePerm)
	require.NoError(t, err, "expected writing report results to disk not to error")
	return reportResults
}
