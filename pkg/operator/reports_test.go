package operator

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/operator-framework/operator-metering/pkg/operator/reporting"

	metering "github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1"
	meteringUtil "github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1/util"
	"github.com/operator-framework/operator-metering/test/testhelpers"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetNextReportPeriod(t *testing.T) {
	baseTime := time.Date(2018, time.July, 1, 0, 0, 0, 0, time.UTC)
	tests := map[string]struct {
		period              metering.ReportPeriod
		expectError         bool
		expectReportPeriods []reportPeriod
	}{
		"hourly": {
			period: metering.ReportPeriodHourly,
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
			period: metering.ReportPeriodDaily,
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
			period: metering.ReportPeriodWeekly,
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
			period: metering.ReportPeriodMonthly,
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
			apiSched := &metering.ReportSchedule{
				Period: test.period,
				// Normally only one is set, but we simply use a zero value
				// for each to make it easier in tests.
				Hourly:  &metering.ReportScheduleHourly{},
				Daily:   &metering.ReportScheduleDaily{},
				Weekly:  &metering.ReportScheduleWeekly{},
				Monthly: &metering.ReportScheduleMonthly{},
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

func TestIsReportFinished(t *testing.T) {
	const (
		testNamespace     = "default"
		testReportName    = "test-report"
		testQueryName     = "test-query"
		testReportMessage = "test-message"
	)

	schedule := &metering.ReportSchedule{
		Period: metering.ReportPeriodCron,
		Cron:   &metering.ReportScheduleCron{Expression: "5 4 * * *"},
	}

	reportStart := &time.Time{}
	reportEndTmp := reportStart.AddDate(0, 1, 0)
	reportEnd := &reportEndTmp

	testTable := []struct {
		name           string
		report         *metering.Report
		expectFinished bool
	}{
		{
			name:           "new report returns false",
			report:         testhelpers.NewReport(testReportName, testNamespace, testQueryName, reportStart, reportEnd, metering.ReportStatus{}, nil, false),
			expectFinished: false,
		},
		{
			name: "***REMOVED***nished status on run-once report returns true",
			report: testhelpers.NewReport(testReportName, testNamespace, testQueryName, reportStart, reportEnd, metering.ReportStatus{
				Conditions: []metering.ReportCondition{
					*meteringUtil.NewReportCondition(metering.ReportRunning, v1.ConditionFalse, meteringUtil.ReportFinishedReason, testReportMessage),
				},
			}, nil, false),
			expectFinished: true,
		},
		{
			name: "unset reportingEnd returns false",
			report: testhelpers.NewReport(testReportName, testNamespace, testQueryName, reportStart, nil, metering.ReportStatus{
				Conditions: []metering.ReportCondition{
					*meteringUtil.NewReportCondition(metering.ReportRunning, v1.ConditionFalse, meteringUtil.ReportFinishedReason, testReportMessage),
				},
			}, schedule, false),
			expectFinished: false,
		},
		{
			name: "reportingEnd > lastReportTime returns false",
			report: testhelpers.NewReport(testReportName, testNamespace, testQueryName, reportStart, reportEnd, metering.ReportStatus{
				Conditions: []metering.ReportCondition{
					*meteringUtil.NewReportCondition(metering.ReportRunning, v1.ConditionFalse, meteringUtil.ReportFinishedReason, testReportMessage),
				},
				LastReportTime: &metav1.Time{Time: reportStart.AddDate(0, 0, 0)},
			}, schedule, false),
			expectFinished: false,
		},
		{
			name: "reportingEnd < lastReportTime returns true",
			report: testhelpers.NewReport(testReportName, testNamespace, testQueryName, reportStart, reportEnd, metering.ReportStatus{
				Conditions: []metering.ReportCondition{
					*meteringUtil.NewReportCondition(metering.ReportRunning, v1.ConditionFalse, meteringUtil.ReportFinishedReason, testReportMessage),
				},
				LastReportTime: &metav1.Time{Time: reportStart.AddDate(0, 2, 0)},
			}, schedule, false),
			expectFinished: true,
		},
		{
			name: "when status running is false and reason is Scheduled return false",
			report: testhelpers.NewReport(testReportName, testNamespace, testQueryName, reportStart, reportEnd, metering.ReportStatus{
				Conditions: []metering.ReportCondition{
					*meteringUtil.NewReportCondition(metering.ReportRunning, v1.ConditionFalse, meteringUtil.ScheduledReason, testReportMessage),
				},
			}, schedule, false),
			expectFinished: false,
		},
		{
			name: "when status running is true and reason is Scheduled return false",
			report: testhelpers.NewReport(testReportName, testNamespace, testQueryName, reportStart, reportEnd, metering.ReportStatus{
				Conditions: []metering.ReportCondition{
					*meteringUtil.NewReportCondition(metering.ReportRunning, v1.ConditionTrue, meteringUtil.ScheduledReason, testReportMessage),
				},
			}, schedule, false),
			expectFinished: false,
		},
		{
			name: "when status running is false and reason is InvalidReport return false",
			report: testhelpers.NewReport(testReportName, testNamespace, testQueryName, reportStart, reportEnd, metering.ReportStatus{
				Conditions: []metering.ReportCondition{
					*meteringUtil.NewReportCondition(metering.ReportRunning, v1.ConditionFalse, meteringUtil.InvalidReportReason, testReportMessage),
				},
			}, schedule, false),
			expectFinished: false,
		},
		{
			name: "when status running is true and reason is InvalidReport return false",
			report: testhelpers.NewReport(testReportName, testNamespace, testQueryName, reportStart, reportEnd, metering.ReportStatus{
				Conditions: []metering.ReportCondition{
					*meteringUtil.NewReportCondition(metering.ReportRunning, v1.ConditionTrue, meteringUtil.InvalidReportReason, testReportMessage),
				},
			}, schedule, false),
			expectFinished: false,
		},
		{
			name: "when status running is false and reason is RunImmediately return false",
			report: testhelpers.NewReport(testReportName, testNamespace, testQueryName, reportStart, reportEnd, metering.ReportStatus{
				Conditions: []metering.ReportCondition{
					*meteringUtil.NewReportCondition(metering.ReportRunning, v1.ConditionFalse, meteringUtil.RunImmediatelyReason, testReportMessage),
				},
			}, schedule, false),
			expectFinished: false,
		},
		{
			name: "when status running is true and reason is RunImmediately return false",
			report: testhelpers.NewReport(testReportName, testNamespace, testQueryName, reportStart, reportEnd, metering.ReportStatus{
				Conditions: []metering.ReportCondition{
					*meteringUtil.NewReportCondition(metering.ReportRunning, v1.ConditionTrue, meteringUtil.RunImmediatelyReason, testReportMessage),
				},
			}, schedule, false),
			expectFinished: false,
		},
	}

	for _, testCase := range testTable {
		var mockLogger = logrus.New()
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			runningCond := isReportFinished(mockLogger, testCase.report)
			assert.Equalf(t, runningCond, testCase.expectFinished, "expected the report would return '%t', but got '%t'", testCase.expectFinished, runningCond)
		})
	}
}

func TestValidateReport(t *testing.T) {
	const (
		testNamespace            = "default"
		testReportName           = "test-report"
		testQueryName            = "test-query"
		testInvalidQueryName     = "invalid-query"
		testInvalidQueryName2    = "invalid-query2"
		testNonExistentQueryName = "does-not-exist"
	)

	ds1 := testhelpers.NewReportDataSource("datasource1", testNamespace)
	ds1.Status.TableRef = v1.LocalObjectReference{Name: "initialized-datasource"}
	// ds2 is uninitialized
	ds2 := testhelpers.NewReportDataSource("datasource2", testNamespace)

	newDefault := func(s string) *json.RawMessage {
		v := json.RawMessage(s)
		return &v
	}

	testValidQuery := &metering.ReportQuery{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testQueryName,
			Namespace: testNamespace,
		},
		Spec: metering.ReportQuerySpec{
			Inputs: []metering.ReportQueryInputDe***REMOVED***nition{
				{
					Name:     "ds",
					Type:     "ReportDataSource",
					Required: true,
					Default:  newDefault((`"` + ds1.Name + `"`)),
				},
			},
		},
	}

	testInvalidQuery := &metering.ReportQuery{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testInvalidQueryName,
			Namespace: testNamespace,
		},
		Spec: metering.ReportQuerySpec{
			Inputs: []metering.ReportQueryInputDe***REMOVED***nition{
				{
					Name:     "ds",
					Type:     "ReportDataSource",
					Required: true,
					Default:  newDefault(`"this-does-not-exist"`),
				},
			},
		},
	}

	testInvalidQuery2 := &metering.ReportQuery{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testInvalidQueryName2,
			Namespace: testNamespace,
		},
		Spec: metering.ReportQuerySpec{
			Inputs: []metering.ReportQueryInputDe***REMOVED***nition{
				{
					Name:     "ds",
					Type:     "ReportDataSource",
					Required: true,
					Default:  newDefault((`"` + ds2.Name + `"`)),
				},
			},
		},
	}

	dataSourceGetter := testhelpers.NewReportDataSourceStore([]*metering.ReportDataSource{ds1, ds2})
	queryGetter := testhelpers.NewReportQueryStore([]*metering.ReportQuery{testValidQuery, testInvalidQuery, testInvalidQuery2})
	reportGetter := testhelpers.NewReportStore(nil)
	dependencyResolver := reporting.NewDependencyResolver(queryGetter, dataSourceGetter, reportGetter)

	reportStart := &time.Time{}
	reportEndTmp := reportStart.AddDate(0, 1, 0)
	reportEnd := &reportEndTmp

	testTable := []struct {
		name         string
		report       *metering.Report
		expectErr    bool
		expectErrMsg string
	}{
		{
			name:         "empty spec.query returns err",
			report:       testhelpers.NewReport(testReportName, testNamespace, "", reportStart, reportEnd, metering.ReportStatus{}, nil, false),
			expectErr:    true,
			expectErrMsg: "must set spec.query",
		},
		{
			name:         "spec.ReportingStart > spec.ReportingEnd returns err",
			report:       testhelpers.NewReport(testReportName, testNamespace, testNonExistentQueryName, reportEnd, reportStart, metering.ReportStatus{}, nil, false),
			expectErr:    true,
			expectErrMsg: fmt.Sprintf("spec.reportingEnd (%s) must be after spec.reportingStart (%s)", reportStart.String(), reportEnd.String()),
		},
		{
			name:         "spec.ReportingEnd is unset and spec.RunImmediately is set returns err",
			report:       testhelpers.NewReport(testReportName, testNamespace, testNonExistentQueryName, reportStart, nil, metering.ReportStatus{}, nil, true),
			expectErr:    true,
			expectErrMsg: "spec.reportingEnd must be set if report.spec.runImmediately is true",
		},
		{
			name:         "spec.QueryName does not exist returns err",
			report:       testhelpers.NewReport(testReportName, testNamespace, testNonExistentQueryName, reportStart, reportEnd, metering.ReportStatus{}, nil, false),
			expectErr:    true,
			expectErrMsg: fmt.Sprintf("ReportQuery (%s) does not exist", testNonExistentQueryName),
		},
		{
			name:         "valid report with missing DataSource returns error",
			report:       testhelpers.NewReport(testReportName, testNamespace, testInvalidQueryName, reportStart, reportEnd, metering.ReportStatus{}, nil, true),
			expectErr:    true,
			expectErrMsg: fmt.Sprintf("failed to resolve ReportQuery dependencies %s: %s", testInvalidQueryName, "ReportDataSource.metering.openshift.io \"this-does-not-exist\" not found"),
		},
		{
			name:         "valid report with uninitalized DataSource returns error",
			report:       testhelpers.NewReport(testReportName, testNamespace, testInvalidQueryName2, reportStart, reportEnd, metering.ReportStatus{}, nil, true),
			expectErr:    true,
			expectErrMsg: fmt.Sprintf("failed to validate ReportQuery dependencies %s: ReportQueryDependencyValidationError: uninitialized ReportDataSource dependencies: %s", testInvalidQueryName2, ds2.Name),
		},
		{
			name:         "valid report with valid DataSource returns nil",
			report:       testhelpers.NewReport(testReportName, testNamespace, testQueryName, reportStart, reportEnd, metering.ReportStatus{}, nil, true),
			expectErr:    false,
			expectErrMsg: "",
		},
	}

	for _, testCase := range testTable {
		testCase := testCase

		noopHandler := &reporting.UninitialiedDependendenciesHandler{HandleUninitializedReportDataSource: func(ds *metering.ReportDataSource) {}}
		t.Run(testCase.name, func(t *testing.T) {
			_, _, err := validateReport(testCase.report, queryGetter, dependencyResolver, noopHandler)

			if testCase.expectErr {
				assert.EqualErrorf(t, err, testCase.expectErrMsg, "expected that validateReport would return the correct error message")
			} ***REMOVED*** {
				assert.NoErrorf(t, err, "expected the report would return no error, but got '%v'", err)
			}
		})
	}
}

func TestGetReportPeriod(t *testing.T) {
	const (
		testNamespace  = "default"
		testReportName = "test-report"
		testQueryName  = "test-query"
	)

	invalidSchedule := &metering.ReportSchedule{
		Period: metering.ReportPeriodCron,
		Cron:   nil,
	}

	validSchedule := &metering.ReportSchedule{
		Period: metering.ReportPeriodCron,
		Cron:   &metering.ReportScheduleCron{Expression: "5 4 * * *"},
	}

	reportStart := &time.Time{}
	reportEndTmp := reportStart.AddDate(0, 1, 0)
	reportEnd := &reportEndTmp
	lastReportTime := &metav1.Time{Time: reportStart.AddDate(0, 0, 0)}
	nextReportTime := &metav1.Time{Time: reportStart.AddDate(0, 1, 0)}

	testTable := []struct {
		name        string
		report      *metering.Report
		expectErr   bool
		expectPanic bool
	}{
		{
			name:      "invalid report with an unset spec.Schedule ***REMOVED***eld returns an error",
			report:    testhelpers.NewReport(testReportName, testNamespace, testQueryName, nil, nil, metering.ReportStatus{}, nil, false),
			expectErr: true,
		},
		{
			name:      "valid report with an unset spec.Schedule ***REMOVED***eld returns nil",
			report:    testhelpers.NewReport(testReportName, testNamespace, testQueryName, reportStart, reportEnd, metering.ReportStatus{}, nil, false),
			expectErr: false,
		},
		{
			name:      "invalid schedule with a set spec.Schedule ***REMOVED***eld returns error",
			report:    testhelpers.NewReport(testReportName, testNamespace, testQueryName, reportStart, reportEnd, metering.ReportStatus{}, invalidSchedule, false),
			expectErr: true,
		},
		{
			name:      "valid schedule with a set spec.Schedule ***REMOVED***eld and an unset Spec.Status.LastReportTime returns nil",
			report:    testhelpers.NewReport(testReportName, testNamespace, testQueryName, reportStart, reportEnd, metering.ReportStatus{}, validSchedule, false),
			expectErr: false,
		},
		{
			name:      "valid schedule with a set spec.Schedule ***REMOVED***eld and a set Spec.Status.LastReportTime returns nil",
			report:    testhelpers.NewReport(testReportName, testNamespace, testQueryName, reportStart, reportEnd, metering.ReportStatus{LastReportTime: lastReportTime}, validSchedule, false),
			expectErr: false,
		},
		{
			name:      "valid schedule with a set spec.Schedule ***REMOVED***eld and an unset Spec.Status.LastReportTime and a set Spec.ReportingStart returns nil",
			report:    testhelpers.NewReport(testReportName, testNamespace, testQueryName, reportStart, reportEnd, metering.ReportStatus{}, validSchedule, false),
			expectErr: false,
		},
		{
			name:      "valid schedule with a set spec.Schedule ***REMOVED***eld and an unset Spec.Status.LastReportTime and an unset Spec.ReportingStart returns nil",
			report:    testhelpers.NewReport(testReportName, testNamespace, testQueryName, nil, reportEnd, metering.ReportStatus{}, validSchedule, false),
			expectErr: false,
		},
		{
			name:      "valid schedule with a set spec.Schedule ***REMOVED***eld and an unset Spec.Status.LastReportTime and a set Spec.NextReportTime returns nil",
			report:    testhelpers.NewReport(testReportName, testNamespace, testQueryName, nil, reportEnd, metering.ReportStatus{NextReportTime: nextReportTime}, validSchedule, false),
			expectErr: false,
		},
		{
			name:        "unset Spec.Schedule with reportPeriod.periodStart > reportPeriod.periodEnd returns panic",
			report:      testhelpers.NewReport(testReportName, testNamespace, testQueryName, reportEnd, reportStart, metering.ReportStatus{NextReportTime: nextReportTime}, nil, false),
			expectErr:   false,
			expectPanic: true,
		},
	}

	for _, testCase := range testTable {
		var mockLogger = logrus.New()
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			if testCase.expectPanic {
				assert.Panics(t, func() { getReportPeriod(time.Now(), mockLogger, testCase.report) }, "expected the test case would panic, but it did not")
			} ***REMOVED*** {
				_, err := getReportPeriod(time.Now(), mockLogger, testCase.report)
				if testCase.expectErr {
					assert.Error(t, err, "expected that getting the report period  would return a non-nil error")
				} ***REMOVED*** {
					assert.Nil(t, err, "expected that getting the report period would return a nil error")
				}
			}
		})
	}
}
