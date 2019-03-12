package e2e

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	meteringv1alpha1 "github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1"
	"github.com/operator-framework/operator-metering/test/framework"
)

type reportProducesDataTestCase struct {
	name          string
	queryName     string
	schedule      *meteringv1alpha1.ReportSchedule
	newReportFunc func(name, queryName string, schedule *meteringv1alpha1.ReportSchedule, start, end *time.Time) *meteringv1alpha1.Report
	skip          bool
	parallel      bool
}

func testReportsProduceData(t *testing.T, testFramework *framework.Framework, reportStart, reportEnd time.Time, testCases []reportProducesDataTestCase) {
	t.Logf("reportStart: %s, reportEnd: %s", reportStart, reportEnd)
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

			report := test.newReportFunc(test.name, test.queryName, test.schedule, &reportStart, &reportEnd)
			reportRunTimeout := 10 * time.Minute
			t.Logf("creating report %s and waiting %s to finish", report.Name, reportRunTimeout)
			testFramework.RequireReportSuccessfullyRuns(t, report, reportRunTimeout)

			resultTimeout := time.Minute
			t.Logf("waiting %s for report %s results", resultTimeout, report.Name)
			reportResults := testFramework.GetReportResults(t, report, resultTimeout)
			assert.NotEmpty(t, reportResults, "reports should return at least 1 row")
		})
	}
}
