package framework

import (
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
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	meteringv1alpha1 "github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1"
	meteringutil "github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1/util"
)

func (f *Framework) CreateMeteringReport(report *meteringv1alpha1.Report) error {
	_, err := f.MeteringClient.Reports(f.Namespace).Create(report)
	return err
}

func (f *Framework) GetMeteringReport(name string) (*meteringv1alpha1.Report, error) {
	return f.MeteringClient.Reports(f.Namespace).Get(name, meta.GetOptions{})
}

func (f *Framework) NewSimpleReport(name, queryName string, schedule *meteringv1alpha1.ReportSchedule, reportingStart, reportingEnd *time.Time) *meteringv1alpha1.Report {
	var start, end *meta.Time
	if reportingStart != nil {
		start = &meta.Time{*reportingStart}
	}
	if reportingEnd != nil {
		end = &meta.Time{*reportingEnd}
	}
	return &meteringv1alpha1.Report{
		ObjectMeta: meta.ObjectMeta{
			Name:      name,
			Namespace: f.Namespace,
		},
		Spec: meteringv1alpha1.ReportSpec{
			GenerationQueryName: queryName,
			Schedule:            schedule,
			ReportingStart:      start,
			ReportingEnd:        end,
		},
	}
}

func (f *Framework) RequireReportSuccessfullyRuns(t *testing.T, report *meteringv1alpha1.Report, waitTimeout time.Duration) {
	err := f.MeteringClient.Reports(f.Namespace).Delete(report.Name, nil)
	assert.Condition(t, func() bool {
		return err == nil || errors.IsNotFound(err)
	}, "failed to ensure Report doesn't exist before creating")

	t.Logf("creating Report %s", report.Name)
	err = f.CreateMeteringReport(report)
	require.NoError(t, err, "creating Report should succeed")

	prevLogMsg := ""
	err = wait.Poll(time.Second*5, waitTimeout, func() (bool, error) {
		// poll the status
		report, err := f.GetMeteringReport(report.Name)
		if err != nil {
			return false, err
		}
		cond := meteringutil.GetReportCondition(report.Status, meteringv1alpha1.ReportRunning)
		if cond != nil && cond.Status == v1.ConditionFalse && cond.Reason == meteringutil.ReportFinishedReason {
			return true, nil
		}

		newLogMsg := fmt.Sprintf("Report %s not ***REMOVED***nished, status: %s", report.Name, spew.Sprintf("%v", report.Status))
		if newLogMsg != prevLogMsg {
			t.Log(newLogMsg)
		}
		prevLogMsg = newLogMsg
		return false, nil
	})
	require.NoErrorf(t, err, "expected Report to ***REMOVED***nished within %s timeout", waitTimeout)
}

func (f *Framework) GetReportResults(t *testing.T, report *meteringv1alpha1.Report, waitTimeout time.Duration) []map[string]interface{} {
	var reportResults []map[string]interface{}
	var reportData []byte

	queryParams := map[string]string{
		"name":      report.Name,
		"namespace": report.Namespace,
		"format":    "json",
	}
	err := wait.Poll(time.Second*5, waitTimeout, func() (bool, error) {
		respBody, respCode, err := f.ReportingOperatorRequest("/api/v1/reports/get", queryParams)
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

	***REMOVED***leName := path.Join(f.ReportOutputDirectory, fmt.Sprintf("%s.json", report.Name))
	err = ioutil.WriteFile(***REMOVED***leName, reportData, os.ModePerm)
	require.NoError(t, err, "expected writing report results to disk not to error")
	return reportResults
}
