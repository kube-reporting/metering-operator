package chargeback

import (
	"testing"
	"time"

	"github.com/operator-framework/operator-metering/pkg/apis/chargeback/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetNextReportPeriod(t *testing.T) {
	baseTime := time.Date(2018, time.July, 1, 0, 0, 0, 0, time.UTC)
	tests := map[string]struct {
		period              v1alpha1.ScheduledReportPeriod
		expectError         bool
		expectReportPeriods []reportPeriod
	}{
		"hourly": {
			period: v1alpha1.ScheduledReportPeriodHourly,
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
			period: v1alpha1.ScheduledReportPeriodDaily,
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
			period: v1alpha1.ScheduledReportPeriodWeekly,
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
			period: v1alpha1.ScheduledReportPeriodMonthly,
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
			apiSched := v1alpha1.ScheduledReportSchedule{
				Period: test.period,
				// Normally only one is set, but we simply use a zero value
				// for each to make it easier in tests.
				Hourly:  &v1alpha1.ScheduledReportScheduleHourly{},
				Daily:   &v1alpha1.ScheduledReportScheduleDaily{},
				Weekly:  &v1alpha1.ScheduledReportScheduleWeekly{},
				Monthly: &v1alpha1.ScheduledReportScheduleMonthly{},
			}

			schedule, err := getSchedule(apiSched)
			require.NoError(t, err)

			lastScheduled := baseTime

			for _, expectedReportPeriod := range test.expectReportPeriods {
				reportPeriod := getNextReportPeriod(schedule, test.period, lastScheduled)
				assert.Equal(t, expectedReportPeriod, reportPeriod)
				lastScheduled = expectedReportPeriod.periodEnd
			}

		})
	}
}
