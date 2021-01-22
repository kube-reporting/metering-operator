package e2e

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"

	meteringv1 "github.com/kube-reporting/metering-operator/pkg/apis/metering/v1"
	"github.com/kube-reporting/metering-operator/test/reportingframework"
	"github.com/kube-reporting/metering-operator/test/testhelpers"
)

var (
	// If a report is in dependency unmet state it won't be deleted and
	// it can take awhile for import to happen. 4 minutes seems to be enough
	// time for that setup, and the delete, to reliably happen.
	waitTimeWhileCheckDeletion = 4 * time.Minute
	// Create a 30s time duration as input for a Report expiration period.
	expiration = &metav1.Duration{Duration: 30.0 * time.Second}
	// Use a one minute cron schedule as input for schedule Reports.
	cronSchedule = &meteringv1.ReportSchedule{
		Period: meteringv1.ReportPeriodCron,
		Cron: &meteringv1.ReportScheduleCron{
			Expression: "*/1 * * * *",
		},
	}
	// Use the `namespace-cpu-usage` ReportQuery for the Report
	// expiration tests.
	reportQueryName = "namespace-cpu-usage"

	emptyReportStatus   = meteringv1.ReportStatus{}
	emptyReportSchedule = meteringv1.ReportSchedule{}
)

// newDefault is a helper function that translates
// a string into the JSON representation of that string.
func newDefault(s string) *json.RawMessage {
	v := json.RawMessage(s)
	return &v
}

// ensureEventExists is a helper function that lists all the events
// in a namespace and checks whether an event has been created
func ensureEventExists(client kubernetes.Interface, name, namespace string) error {
	events, err := client.CoreV1().Events(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	var found bool
	for _, event := range events.Items {
		if strings.Contains(event.InvolvedObject.Name, name) {
			found = true
		}
	}
	if !found {
		return fmt.Errorf("failed to find the %s involved object in the %s namespace list of events", name, namespace)
	}

	return nil
}

// testReportIsDeletedWhenNoDeps ensures that creating a scheduled
// Report with a configured expiration and no other Report/ReportQuery
// dependencies will be deleting by the reporting-operator before the
// poll timeout period.
func testReportIsDeletedWhenNoDeps(t *testing.T, rf *reportingframework.ReportingFramework) {
	reportName := "e2e-report-with-expiration-should-delete"

	t.Logf("Creating the %s run-once Report and waiting %v to finish", reportName, waitTimeWhileCheckDeletion)
	report := rf.NewSimpleReport(reportName, reportQueryName, cronSchedule, nil, nil, expiration)
	err := rf.CreateMeteringReport(report)
	require.NoError(t, err, "creating a report that has a configured expiration should produce no error")

	var foundOnce bool
	err = wait.Poll(5*time.Second, waitTimeWhileCheckDeletion, func() (bool, error) {
		_, err := rf.GetMeteringReport(report.Name)
		if apierrors.IsNotFound(err) {
			if !foundOnce {
				return false, nil
			}
			t.Logf("The %s Report has been deleted", report.Name)
			return true, errors.New("report was deleted after being created")
		}
		if err != nil {
			return false, err
		}

		foundOnce = true
		return false, nil
	})
	if err != nil {
		assert.Contains(t, err.Error(), "was deleted", "report should have been deleted by retention period being reached")
	}
}

// testReportIsNotDeletedWhenReportDependsOnIt ensures that a Report
// that has been configured with a retention period will not be deleted
// when another Report references it as an input, i.e. depends on it.
func testReportIsNotDeletedWhenReportDependsOnIt(t *testing.T, rf *reportingframework.ReportingFramework) {
	now := time.Now()
	t1 := now.UTC().AddDate(0, 0, -1)
	t2 := now.UTC()
	testSubReportName := "subreport-cpu-usage-run-immediately-expiration-report"

	t.Logf("Creating the %s sub-Report and waiting %v to finish", testSubReportName, waitTimeWhileCheckDeletion)
	subReport := testhelpers.NewReport(testSubReportName, rf.Namespace, reportQueryName, nil, &t1, &t2, emptyReportStatus, nil, true, expiration)
	err := rf.CreateMeteringReport(subReport)
	require.NoError(t, err, "creating a cron schedule report with a 30s expiration should produce no error")

	// TODO: should verify the status as a valid report
	testReportQueryName := fmt.Sprintf("custom-%s-%s-rq", namespacePrefix, reportQueryName)
	inputReportQueryName := strings.ReplaceAll(subReport.Name, "-", "")
	rq := testhelpers.NewReportQuery(testReportQueryName, rf.Namespace, []meteringv1.ReportQueryColumn{{Name: "foo", Type: "double"}})
	rq.Spec.Inputs = []meteringv1.ReportQueryInputDefinition{
		{
			Name:     inputReportQueryName,
			Type:     "Report",
			Required: true,
			Default:  newDefault(`"` + subReport.Name + `"`),
		},
	}
	rq.Spec.Query = "SELECT 1"
	t.Logf("Creating the %s aggregate ReportQuery from the %s Report", testReportQueryName, subReport.Name)
	_, err = rf.MeteringClient.ReportQueries(rf.Namespace).Create(context.Background(), rq, metav1.CreateOptions{})
	require.NoError(t, err, "creating a custom ReportQuery from a sub-report should produce no error")

	rollupReportName := "report-depends-on-subreport"
	rollupReportInputs := []meteringv1.ReportQueryInputValue{
		{
			Name:  strings.ReplaceAll(subReport.Name, "-", ""),
			Value: newDefault(`"` + subReport.Name + `"`),
		},
	}
	report := testhelpers.NewReport(rollupReportName, rf.Namespace, testReportQueryName, rollupReportInputs, nil, nil, emptyReportStatus, cronSchedule, false, nil)

	t.Logf("Creating the %s rollup Report from the %s aggregate ReportQuery", rollupReportName, testReportQueryName)
	err = rf.CreateMeteringReport(report)
	assert.NoError(t, err, "creating a rollup report should produce no error", rollupReportName)

	var foundOnce bool
	err = wait.Poll(5*time.Second, waitTimeWhileCheckDeletion, func() (bool, error) {
		_, err := rf.GetMeteringReport(subReport.Name)
		if apierrors.IsAlreadyExists(err) {
			t.Logf("The %s Report has been created", subReport.Name)
			foundOnce = true
		}
		if foundOnce && apierrors.IsNotFound(err) {
			return true, errors.New("subReport was deleted after being created")
		}

		// if we're exiting due to timeout, it's correct, because the subReport should not be deleted
		return false, nil
	})
	assert.Contains(t,
		err.Error(),
		"timed out waiting",
		"should have timed out waiting on subReport expiration as subReport should not have been deleted",
	)

	err = ensureEventExists(rf.KubeClient, subReport.Name, rf.Namespace)
	require.NoError(t, err, "expected to find an event related to Report not deleted because of dependencies")
}
