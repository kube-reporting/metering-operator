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

	"github.com/kube-reporting/metering-operator/test/testhelpers"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	metering "github.com/kube-reporting/metering-operator/pkg/apis/metering/v1"
	"github.com/kube-reporting/metering-operator/pkg/operator/reporting"
	"github.com/kube-reporting/metering-operator/test/reportingframework"
)

var (
	runAWSBillingTests bool
)

func testReportingProducesData(t *testing.T, testReportingFramework *reportingframework.ReportingFramework) {
	// cron schedule to run every minute
	cronSchedule := &metering.ReportSchedule{
		Period: metering.ReportPeriodCron,
		Cron: &metering.ReportScheduleCron{
			Expression: "*/1 * * * *",
		},
	}

	queries := []struct {
		queryName   string
		skip        bool
		nonParallel bool
	}{
		{queryName: "namespace-cpu-request"},
		{queryName: "namespace-cpu-usage"},
		{queryName: "namespace-memory-request"},
		{queryName: "namespace-persistentvolumeclaim-request"},
		{queryName: "namespace-persistentvolumeclaim-usage"},
		{queryName: "namespace-memory-usage"},
		{queryName: "persistentvolumeclaim-usage"},
		{queryName: "persistentvolumeclaim-capacity"},
		{queryName: "persistentvolumeclaim-request"},
		{queryName: "pod-cpu-request"},
		{queryName: "pod-cpu-usage"},
		{queryName: "pod-memory-request"},
		{queryName: "pod-memory-usage"},
		{queryName: "node-cpu-utilization"},
		{queryName: "node-memory-utilization"},
		{queryName: "cluster-persistentvolumeclaim-request"},
		{queryName: "cluster-cpu-capacity"},
		{queryName: "cluster-memory-capacity"},
		{queryName: "cluster-cpu-usage"},
		{queryName: "cluster-memory-usage"},
		{queryName: "cluster-cpu-utilization"},
		{queryName: "cluster-memory-utilization"},
		{queryName: "namespace-memory-utilization"},
		{queryName: "namespace-cpu-utilization"},
		{queryName: "pod-cpu-request-aws", skip: !runAWSBillingTests, nonParallel: true},
		{queryName: "pod-memory-request-aws", skip: !runAWSBillingTests, nonParallel: true},
		{queryName: "aws-ec2-cluster-cost", skip: !runAWSBillingTests, nonParallel: true},
	}

	var reportsProduceDataTestCases []reportProducesDataTestCase

	for _, query := range queries {
		reportcronTestCase := reportProducesDataTestCase{
			name:          query.queryName + "-cron",
			queryName:     query.queryName,
			schedule:      cronSchedule,
			newReportFunc: testReportingFramework.NewSimpleReport,
			skip:          query.skip,
			parallel:      !query.nonParallel,
		}
		reportRunOnceTestCase := reportProducesDataTestCase{
			name:          query.queryName + "-runonce",
			queryName:     query.queryName,
			schedule:      nil, // runOnce
			newReportFunc: testReportingFramework.NewSimpleReport,
			skip:          query.skip,
			parallel:      !query.nonParallel,
		}

		reportsProduceDataTestCases = append(reportsProduceDataTestCases, reportcronTestCase, reportRunOnceTestCase)
	}

	testReportsProduceData(t, testReportingFramework, reportsProduceDataTestCases)
}

type reportProducesDataTestCase struct {
	name          string
	queryName     string
	schedule      *metering.ReportSchedule
	newReportFunc func(name, queryName string, schedule *metering.ReportSchedule, start, end *time.Time, expiration *metav1.Duration) *metering.Report
	skip          bool
	parallel      bool
}

func testReportsProduceData(t *testing.T, testReportingFramework *reportingframework.ReportingFramework, testCases []reportProducesDataTestCase) {
	for _, test := range testCases {
		name := test.name
		// Fix closure captures
		test := test
		t.Run(name, func(t *testing.T) {
			if test.skip {
				t.Skip("test configured to be skipped")
				return
			}
			if test.parallel {
				t.Parallel()
			}

			query, err := testReportingFramework.WaitForMeteringReportQuery(t, test.queryName, 5*time.Second, 5*time.Minute)
			require.NoError(t, err, "report query for report should exist")

			dsGetter := reporting.NewReportDataSourceClientGetter(testReportingFramework.MeteringClient)
			queryGetter := reporting.NewReportQueryClientGetter(testReportingFramework.MeteringClient)
			reportGetter := reporting.NewReportClientGetter(testReportingFramework.MeteringClient)

			// get all the datasources for the query used in our report
			dependencies, err := reporting.GetQueryDependencies(queryGetter, dsGetter, reportGetter, query, nil)
			require.NoError(t, err, "datasources for query should exist")
			require.NotEqual(t, 0, len(dependencies.ReportDataSources), "Report should have at least 1 datasource dependency")

			var reportStart time.Time
			// for each datasource, wait until it's EarliestImportedMetricTime is set
			for _, ds := range dependencies.ReportDataSources {
				_, err := testReportingFramework.WaitForMeteringReportDataSource(t, ds.Name, 5*time.Second, 5*time.Minute, func(dataSource *metering.ReportDataSource) (bool, error) {
					if dataSource.Spec.PrometheusMetricsImporter == nil {
						return true, nil
					}
					if dataSource.Status.PrometheusMetricsImportStatus != nil && dataSource.Status.PrometheusMetricsImportStatus.EarliestImportedMetricTime != nil {
						// keep the EarliestImportedMetricTime that is the
						// least far back, so that we ensure the reportStart is
						// a time that all datasources have metrics for.
						if reportStart.IsZero() || dataSource.Status.PrometheusMetricsImportStatus.EarliestImportedMetricTime.Time.After(reportStart) {
							reportStart = dataSource.Status.PrometheusMetricsImportStatus.EarliestImportedMetricTime.Time
						}
						return true, nil
					}
					return false, nil
				})
				require.NoError(t, err, "expected ReportDataSource %s to have an earliestImportedMetricTime", ds.Name)
			}

			if reportStart.IsZero() {
				t.Errorf("reportStart is zero")
			}

			reportStart = reportStart.UTC()
			// The report spans 5 minutes
			reportEnd := reportStart.Add(5 * time.Minute).UTC()

			t.Logf("reportStart: %s, reportEnd: %s", reportStart, reportEnd)

			report := test.newReportFunc(test.name, test.queryName, test.schedule, &reportStart, &reportEnd, nil)
			reportRunTimeout := 10 * time.Minute
			t.Logf("creating report %s and waiting %s to finish", report.Name, reportRunTimeout)
			testReportingFramework.RequireReportSuccessfullyRuns(t, report, reportRunTimeout)

			resultTimeout := time.Minute
			t.Logf("waiting %s for report %s results", resultTimeout, report.Name)
			reportResults := testReportingFramework.GetReportResults(t, report, resultTimeout)
			assert.NotEmpty(t, reportResults, "reports should return at least 1 row")
		})
	}
}

func testReportIsDeletedWhenNoDeps(t *testing.T, testReportingFramework *reportingframework.ReportingFramework) {
	reportName := "report-should-delete"
	// cron schedule to run every minute
	cronSchedule := &metering.ReportSchedule{
		Period: metering.ReportPeriodCron,
		Cron: &metering.ReportScheduleCron{
			Expression: "*/1 * * * *",
		},
	}
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

	var foundOnce bool
	err := wait.Poll(time.Second*5, waitTimeWhileCheckDeletion, func() (bool, error) {
		_, err := testReportingFramework.GetMeteringReport(reportName)
		// `foundOnce` provides a latch that allows us to wait for report creation then exit on deletion
		if apierrors.IsAlreadyExists(err) {
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
	cronSchedule := &metering.ReportSchedule{
		Period: metering.ReportPeriodCron,
		Cron: &metering.ReportScheduleCron{
			Expression: "*/1 * * * *",
		},
	}
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
		[]metering.ReportQueryInputValue{
			{
				Name:  "Report",
				Value: newDefault(`"` + subReportName + `"`),
			},
		},
		nil,
		nil,
		metering.ReportStatus{},
		cronSchedule,
		false,
		nil,
	)

	createErr = testReportingFramework.CreateMeteringReport(report)
	assert.NoError(t, createErr, "creating report should produce no error")

	var foundOnce bool
	err := wait.Poll(time.Second*5, waitTimeWhileCheckDeletion, func() (bool, error) {
		_, err := testReportingFramework.GetMeteringReport(subReportName)
		// `foundOnce` provides a latch that allows us to wait for subReport creation then exit on deletion
		if apierrors.IsAlreadyExists(err) {
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
	events, err := testReportingFramework.KubeClient.CoreV1().Events(testReportingFramework.Namespace).List(
		context.Background(),
		metav1.ListOptions{},
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
