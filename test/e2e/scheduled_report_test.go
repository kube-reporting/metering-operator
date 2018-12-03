package e2e

import (
	"testing"
	"time"

	meteringv1alpha1 "github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1"
	"github.com/operator-framework/operator-metering/test/framework"
	"github.com/stretchr/testify/assert"
)

type scheduledReportProducesDataTestCase struct {
	name          string
	queryName     string
	schedule      *meteringv1alpha1.ScheduledReportSchedule
	newReportFunc func(name, queryName string, schedule *meteringv1alpha1.ScheduledReportSchedule, start, end *time.Time) *meteringv1alpha1.ScheduledReport
	timeout       time.Duration
	skip          bool
}

func testScheduledReportsProduceData(t *testing.T, testFramework *framework.Framework, periodStart, periodEnd time.Time, testCases []scheduledReportProducesDataTestCase) {
	t.Logf("periodStart: %s, periodEnd: %s", periodStart, periodEnd)
	for _, test := range testCases {
		name := test.name
		// Fix closure captures
		test := test
		t.Run(name, func(t *testing.T) {
			if test.skip {
				t.Skip("test configured to be skipped")
				return
			}

			// set reportStart to the nearest hour since the hourly
			// scheduledReport will align to the hour
			reportStart := periodStart.Truncate(time.Hour)
			reportEnd := periodEnd.Truncate(time.Hour)

			// if truncation causes them to be the same, set reportStart to 1
			// hour before reportEnd
			if reportEnd.Equal(reportStart) {
				reportStart.Add(-time.Hour)
			}

			t.Logf("scheduledReport reportingStart: %s, reportingEnd: %s", reportStart, reportEnd)

			report := test.newReportFunc(test.name, test.queryName, test.schedule, &reportStart, &reportEnd)
			defer func() {
				t.Logf("deleting scheduled report %s", report.Name)
				err := testFramework.MeteringClient.ScheduledReports(testFramework.Namespace).Delete(report.Name, nil)
				assert.NoError(t, err, "expected delete scheduled report to succeed")
			}()

			testFramework.RequireScheduledReportSuccessfullyRuns(t, report, time.Minute)
			reportResults := testFramework.GetScheduledReportResults(t, report, time.Minute)

			assert.NotEmpty(t, reportResults, "reports should return at least 1 row")
		})
	}
}
