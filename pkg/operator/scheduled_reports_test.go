package operator

import (
	"testing"
	"time"

	"github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetNextReportPeriod(t *testing.T) {
	baseTime := time.Date(2018, time.July, 1, 0, 0, 0, 0, time.UTC)
	tests := map[string]struct {
		period              v1alpha1.ReportPeriod
		expectError         bool
		expectReportPeriods []reportPeriod
	}{
		"hourly": {
			period: v1alpha1.ReportPeriodHourly,
			expectReportPeriods: []reportPeriod{
				{
					periodStart: baseTime,
					periodEnd:   time.Date(2018, time.July, 1, 1, 0, 0, 0, time.UTC),
				},
				{
					periodStart: time.Date(2018, time.July, 1, 1, 0, 0, 0, time.UTC),
					periodEnd:   time.Date(2018, time.July, 1, 2, 0, 0, 0, time.UTC),
				},
			},
		},
		"daily": {
			period: v1alpha1.ReportPeriodDaily,
			expectReportPeriods: []reportPeriod{
				{
					periodStart: baseTime,
					periodEnd:   time.Date(2018, time.July, 2, 0, 0, 0, 0, time.UTC),
				},
				{
					periodStart: time.Date(2018, time.July, 2, 0, 0, 0, 0, time.UTC),
					periodEnd:   time.Date(2018, time.July, 3, 0, 0, 0, 0, time.UTC),
				},
			},
		},
		"weekly": {
			period: v1alpha1.ReportPeriodWeekly,
			expectReportPeriods: []reportPeriod{
				{
					periodStart: baseTime,
					periodEnd:   time.Date(2018, time.July, 8, 0, 0, 0, 0, time.UTC),
				},
				{
					periodStart: time.Date(2018, time.July, 8, 0, 0, 0, 0, time.UTC),
					periodEnd:   time.Date(2018, time.July, 15, 0, 0, 0, 0, time.UTC),
				},
			},
		},
		"monthly": {
			period: v1alpha1.ReportPeriodMonthly,
			expectReportPeriods: []reportPeriod{
				{
					periodStart: baseTime,
					periodEnd:   time.Date(2018, time.August, 1, 0, 0, 0, 0, time.UTC),
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			apiSched := &v1alpha1.ReportSchedule{
				Period: test.period,
				// Normally only one is set, but we simply use a zero value
				// for each to make it easier in tests.
				Hourly:  &v1alpha1.ReportScheduleHourly{},
				Daily:   &v1alpha1.ReportScheduleDaily{},
				Weekly:  &v1alpha1.ReportScheduleWeekly{},
				Monthly: &v1alpha1.ReportScheduleMonthly{},
			}

			schedule, err := getSchedule(apiSched)
			require.NoError(t, err)

			lastScheduled := baseTime

			for _, expectedReportPeriod := range test.expectReportPeriods {
				reportPeriod := getNextReportPeriod(schedule, test.period, lastScheduled)
				assert.Equal(t, &expectedReportPeriod, reportPeriod)
				lastScheduled = expectedReportPeriod.periodEnd
			}

		})
	}
}
