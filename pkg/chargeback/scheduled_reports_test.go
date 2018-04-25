package chargeback

import (
	"testing"
	"time"

	"github.com/operator-framework/operator-metering/pkg/apis/chargeback/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetPreviousReportDay(t *testing.T) {
	tests := map[string]struct {
		period               v1alpha1.ScheduledReportPeriod
		nextTime             time.Time
		expectedPreviousTime time.Time
		expectError          bool
	}{
		"daily-on-beginning-of-month": {
			period:               v1alpha1.ScheduledReportPeriodDaily,
			nextTime:             time.Date(2018, time.July, 1, 0, 0, 0, 0, time.UTC),
			expectedPreviousTime: time.Date(2018, time.June, 30, 0, 0, 0, 0, time.UTC),
		},
		// Ensure this actually goes to the 28th rather than the 30/31st
		// when previous month is February
		"daily-on-beginning-of-month-march-february": {
			period:               v1alpha1.ScheduledReportPeriodDaily,
			nextTime:             time.Date(2018, time.March, 1, 0, 0, 0, 0, time.UTC),
			expectedPreviousTime: time.Date(2018, time.February, 28, 0, 0, 0, 0, time.UTC),
		},
		"weekly-on-beginning-of-month": {
			period:               v1alpha1.ScheduledReportPeriodWeekly,
			nextTime:             time.Date(2018, time.July, 1, 0, 0, 0, 0, time.UTC),
			expectedPreviousTime: time.Date(2018, time.June, 24, 0, 0, 0, 0, time.UTC),
		},
		"weekly-in-middle-of-month": {
			period:               v1alpha1.ScheduledReportPeriodWeekly,
			nextTime:             time.Date(2018, time.July, 15, 0, 0, 0, 0, time.UTC),
			expectedPreviousTime: time.Date(2018, time.July, 8, 0, 0, 0, 0, time.UTC),
		},
		"monthly-on-beginning-of-month": {
			period:               v1alpha1.ScheduledReportPeriodMonthly,
			nextTime:             time.Date(2018, time.July, 1, 0, 0, 0, 0, time.UTC),
			expectedPreviousTime: time.Date(2018, time.June, 1, 0, 0, 0, 0, time.UTC),
		},
		"monthly-on-end-of-month": {
			period:               v1alpha1.ScheduledReportPeriodMonthly,
			nextTime:             time.Date(2018, time.July, 31, 0, 0, 0, 0, time.UTC),
			expectedPreviousTime: time.Date(2018, time.July, 1, 0, 0, 0, 0, time.UTC),
		},
		"monthly-on-middle-of-month": {
			period:               v1alpha1.ScheduledReportPeriodMonthly,
			nextTime:             time.Date(2018, time.July, 15, 0, 0, 0, 0, time.UTC),
			expectedPreviousTime: time.Date(2018, time.June, 15, 0, 0, 0, 0, time.UTC),
		},
		"monthly-on-beginning-of-year": {
			period:               v1alpha1.ScheduledReportPeriodMonthly,
			nextTime:             time.Date(2018, time.January, 1, 0, 0, 0, 0, time.UTC),
			expectedPreviousTime: time.Date(2017, time.December, 1, 0, 0, 0, 0, time.UTC),
		},
		// The following 4 tests ensure all of the dates in a month are covered
		// with a series weekly report periods
		"weekly-covers-full-month-week-1": {
			period:               v1alpha1.ScheduledReportPeriodWeekly,
			nextTime:             time.Date(2018, time.July, 8, 0, 0, 0, 0, time.UTC),
			expectedPreviousTime: time.Date(2018, time.July, 1, 0, 0, 0, 0, time.UTC),
		},
		"weekly-covers-full-month-week-2": {
			period:               v1alpha1.ScheduledReportPeriodWeekly,
			nextTime:             time.Date(2018, time.July, 15, 0, 0, 0, 0, time.UTC),
			expectedPreviousTime: time.Date(2018, time.July, 8, 0, 0, 0, 0, time.UTC),
		},
		"weekly-covers-full-month-week-3": {
			period:               v1alpha1.ScheduledReportPeriodWeekly,
			nextTime:             time.Date(2018, time.July, 22, 0, 0, 0, 0, time.UTC),
			expectedPreviousTime: time.Date(2018, time.July, 15, 0, 0, 0, 0, time.UTC),
		},
		"weekly-covers-full-month-week-4": {
			period:               v1alpha1.ScheduledReportPeriodWeekly,
			nextTime:             time.Date(2018, time.July, 29, 0, 0, 0, 0, time.UTC),
			expectedPreviousTime: time.Date(2018, time.July, 22, 0, 0, 0, 0, time.UTC),
		},
		"weekly-covers-full-month-week-5": {
			period:               v1alpha1.ScheduledReportPeriodWeekly,
			nextTime:             time.Date(2018, time.August, 5, 0, 0, 0, 0, time.UTC),
			expectedPreviousTime: time.Date(2018, time.July, 29, 0, 0, 0, 0, time.UTC),
		},
		"hourly": {
			period:               v1alpha1.ScheduledReportPeriodHourly,
			nextTime:             time.Date(2018, time.July, 5, 0, 0, 0, 0, time.UTC),
			expectedPreviousTime: time.Date(2018, time.July, 4, 23, 0, 0, 0, time.UTC),
		},
		"invalid period": {
			period:      "yearly",
			expectError: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			previousTime, err := getPreviousReportDay(test.nextTime, test.period)
			if test.expectError {
				assert.Error(t, err)
			} ***REMOVED*** {
				assert.Nil(t, err)
			}
			assert.Equal(t, test.expectedPreviousTime, previousTime)
		})
	}
}

func TestGetNextReprtPeriod(t *testing.T) {
	tests := map[string]struct {
		period             v1alpha1.ScheduledReportPeriod
		lastScheduled      time.Time
		expectError        bool
		expectReportPeriod reportPeriod
	}{
		"hourly": {
			period:        v1alpha1.ScheduledReportPeriodHourly,
			lastScheduled: time.Date(2018, time.July, 1, 0, 0, 0, 0, time.UTC),
			expectReportPeriod: reportPeriod{
				periodStart: time.Date(2018, time.July, 1, 0, 0, 0, 0, time.UTC),
				periodEnd:   time.Date(2018, time.July, 1, 1, 0, 0, 0, time.UTC),
			},
		},
		"daily": {
			period:        v1alpha1.ScheduledReportPeriodDaily,
			lastScheduled: time.Date(2018, time.July, 1, 0, 0, 0, 0, time.UTC),
			expectReportPeriod: reportPeriod{
				periodStart: time.Date(2018, time.July, 1, 0, 0, 0, 0, time.UTC),
				periodEnd:   time.Date(2018, time.July, 2, 0, 0, 0, 0, time.UTC),
			},
		},
		"weekly": {
			period:        v1alpha1.ScheduledReportPeriodWeekly,
			lastScheduled: time.Date(2018, time.July, 1, 0, 0, 0, 0, time.UTC),
			expectReportPeriod: reportPeriod{
				periodStart: time.Date(2018, time.July, 1, 0, 0, 0, 0, time.UTC),
				periodEnd:   time.Date(2018, time.July, 8, 0, 0, 0, 0, time.UTC),
			},
		},
		"monthly": {
			period:        v1alpha1.ScheduledReportPeriodMonthly,
			lastScheduled: time.Date(2018, time.July, 1, 0, 0, 0, 0, time.UTC),
			expectReportPeriod: reportPeriod{
				periodStart: time.Date(2018, time.July, 1, 0, 0, 0, 0, time.UTC),
				periodEnd:   time.Date(2018, time.August, 1, 0, 0, 0, 0, time.UTC),
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
			if test.expectError {
				require.Error(t, err)
			} ***REMOVED*** {
				require.NoError(t, err)
			}

			reportPeriod, err := getNextReportPeriod(schedule, test.period, test.lastScheduled)
			if test.expectError {
				assert.Error(t, err)
			} ***REMOVED*** {
				assert.NoError(t, err)
			}
			assert.Equal(t, test.expectReportPeriod, reportPeriod)
		})
	}
}
