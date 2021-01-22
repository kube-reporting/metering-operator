package e2e

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	meteringv1 "github.com/kube-reporting/metering-operator/pkg/apis/metering/v1"
	"github.com/kube-reporting/metering-operator/test/reportingframework"
	"github.com/kube-reporting/metering-operator/test/testhelpers"
)

func testReportIsDeletedWhenNoDeps(t *testing.T, testReportingFramework *reportingframework.ReportingFramework) {
	reportName := "report-should-delete"
	// cron schedule to run every minute
	cronSchedule := &meteringv1.ReportSchedule{
		Period: meteringv1.ReportPeriodCron,
		Cron: &meteringv1.ReportScheduleCron{
			Expression: "*/1 * * * *",
		},
	}
	var foundOnce bool
	expiration := &metav1.Duration{Duration: 30.0 * time.Second}
	// If a report is in dependency unmet state it won't be deleted and it can take awhile for import to happen.
	// 4 minutes seems to be enough time for that setup, and the delete, to reliably happen.
	waitTimeWhileCheckDeletion := 4 * time.Minute

	report := testReportingFramework.NewSimpleReport(
		reportName,
		"namespace-memory-request",
		cronSchedule,
		nil,
		nil,
		expiration,
	)
	t.Logf("creating report %s and waiting %s to finish", report.Name, waitTimeWhileCheckDeletion)
	createErr := testReportingFramework.CreateMeteringReport(report)
	assert.NoError(t, createErr, "creating report should produce no error")

	err := wait.Poll(time.Second*5, waitTimeWhileCheckDeletion, func() (bool, error) {
		_, err := testReportingFramework.GetMeteringReport(reportName)
		// `foundOnce` provides a latch that allows us to wait for report creation then exit on deletion
		if err == nil {
			foundOnce = true
		}
		if err != nil {
			if foundOnce && apierrors.IsNotFound(err) {
				return true, errors.New("report was deleted after being created")
			}
		}
		// if we're exiting due to timeout, it's a fail
		return false, nil
	})
	if err != nil {
		assert.Contains(t, err.Error(), "was deleted", "report should have been deleted by retention period being reached")
	}
}

func testReportIsNotDeletedWhenReportDependsOnIt(t *testing.T, testReportingFramework *reportingframework.ReportingFramework) {
	testQueryName := "namespace-cpu-usage"
	subReportName := "subreport-should-not-delete"
	// cron schedule to run every minute
	cronSchedule := &meteringv1.ReportSchedule{
		Period: meteringv1.ReportPeriodCron,
		Cron: &meteringv1.ReportScheduleCron{
			Expression: "*/1 * * * *",
		},
	}
	var foundOnce bool
	expiration := &metav1.Duration{Duration: 30.0 * time.Second}
	// If a subReport is in dependency unmet state it won't be deleted and it can take awhile for import to happen.
	// 4 minutes seems to be enough time for that setup, and the delete, to reliably happen.
	waitTimeWhileCheckDeletion := 4 * time.Minute

	subReport := testReportingFramework.NewSimpleReport(
		subReportName,
		testQueryName,
		cronSchedule,
		nil,
		nil,
		expiration,
	)
	t.Logf("creating subReport %s and waiting %s to finish", subReport.Name, waitTimeWhileCheckDeletion)

	createErr := testReportingFramework.CreateMeteringReport(subReport)
	assert.NoError(t, createErr, "creating subreport should produce no error")

	newDefault := func(s string) *json.RawMessage {
		v := json.RawMessage(s)
		return &v
	}

	report := testhelpers.NewReport(
		"report-depends-on-subreport",
		testReportingFramework.Namespace,
		testQueryName,
		[]meteringv1.ReportQueryInputValue{
			meteringv1.ReportQueryInputValue{
				Name:  "Report",
				Value: newDefault(`"` + subReportName + `"`),
			},
		},
		nil,
		nil,
		meteringv1.ReportStatus{},
		cronSchedule,
		false,
		nil,
	)

	createErr = testReportingFramework.CreateMeteringReport(report)
	assert.NoError(t, createErr, "creating report should produce no error")

	err := wait.Poll(time.Second*5, waitTimeWhileCheckDeletion, func() (bool, error) {
		_, err := testReportingFramework.GetMeteringReport(subReportName)
		// `foundOnce` provides a latch that allows us to wait for subReport creation then exit on deletion
		if err == nil {
			foundOnce = true
		}
		if err != nil {
			if foundOnce && apierrors.IsNotFound(err) {
				return true, errors.New("subReport was deleted after being created")
			}
		}
		// if we're exiting due to timeout, it's correct, because the subReport should not be deleted
		return false, nil
	})
	assert.Contains(t,
		err.Error(),
		"timed out waiting",
		"should have timed out waiting on subReport expiration as subReport should not have been deleted",
	)
	listOptions := metav1.ListOptions{}
	events, err := testReportingFramework.KubeClient.CoreV1().Events(testReportingFramework.Namespace).List(
		context.Background(),
		listOptions,
	)
	require.NoError(t, err, "failed to list the events in the test namespace")
	var found bool
	for _, event := range events.Items {
		if strings.Contains(event.InvolvedObject.Name, "subreport-should-not-delete") {
			found = true
		}
	}

	assert.True(t, found, "expected to find an event related to Report not deleted because of dependencies")
}
