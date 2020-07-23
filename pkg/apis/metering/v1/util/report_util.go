package util

import (
	"errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	metering "github.com/kube-reporting/metering-operator/pkg/apis/metering/v1"
)

const (
	// Running true

	// ScheduledReason is set when the report is running due to it's schedule
	// and the current time is beyond the next reporting period end.
	ScheduledReason = "Scheduled"

	// RunImmediatelyReason is set when the report is running because it's
	// spec.runImmediately is true.
	RunImmediatelyReason = "RunImmediately"

	// Running false

	// ReportingPeriodWaitingReason is set when a report is not running because it is
	// waiting for the next period of time in it's schedule to elapse.
	ReportingPeriodWaitingReason = "ReportingPeriodWaiting"

	// ReportFinishedReason is set in the report has generated report results
	// for all periods between reportingStart and reportingEnd, according to
	// it's configured schedule. Run-once reports are finished after their
	// execution.
	ReportFinishedReason = "Finished"

	// InvalidReportReason is added to a Report when the
	// Report is invalid or it's ReportQuery is invalid.
	InvalidReportReason = "InvalidReport"

	// ReportingPeriodUnmetDependenciesReason is set when a Report cannot run
	// because it's dependencies (ReportQueries dependencies on
	// ReportDataSources, ReportQueries, and other Reports) do not
	// have data available for the reporting period currently being processed.
	ReportingPeriodUnmetDependenciesReason = "ReportingPeriodUnmetDependencies"

	// GenerateReportFailedReason is set when a Report is not running because
	// it previously failed when generating results previously.
	GenerateReportFailedReason = "GenerateReportFailed"
)

// NewReportCondition creates a new report condition.
func NewReportCondition(condType metering.ReportConditionType, status v1.ConditionStatus, reason, message string) *metering.ReportCondition {
	return &metering.ReportCondition{
		Type:               condType,
		Status:             status,
		LastUpdateTime:     metav1.Now(),
		LastTransitionTime: metav1.Now(),
		Reason:             reason,
		Message:            message,
	}
}

// GetReportCondition returns the condition with the provided type.
func GetReportCondition(status metering.ReportStatus, condType metering.ReportConditionType) *metering.ReportCondition {
	for i := range status.Conditions {
		c := status.Conditions[i]
		if c.Type == condType {
			return &c
		}
	}
	return nil
}

// SetReportCondition updates the report to include the provided condition. If the condition that
// we are about to add already exists and has the same status and reason then we are not going to update.
func SetReportCondition(status *metering.ReportStatus, condition metering.ReportCondition) error {
	if status == nil {
		return errors.New("cannot add condition to nil status")
	}
	currentCond := GetReportCondition(*status, condition.Type)
	if currentCond != nil && currentCond.Status == condition.Status && currentCond.Reason == condition.Reason {
		return nil
	}
	// Do not update lastTransitionTime if the status of the condition doesn't change.
	if currentCond != nil && currentCond.Status == condition.Status {
		condition.LastTransitionTime = currentCond.LastTransitionTime
	}
	newConditions := filterOutCondition(status.Conditions, condition.Type)
	status.Conditions = append(newConditions, condition)

	return nil
}

// RemoveReportCondition removes the report condition with the provided type.
func RemoveReportCondition(status *metering.ReportStatus, condType metering.ReportConditionType) error {
	if status == nil {
		return errors.New("cannot remove condition from nil status")
	}
	status.Conditions = filterOutCondition(status.Conditions, condType)
	return nil
}

// filterOutCondition returns a new slice of report conditions without conditions with the provided type.
func filterOutCondition(conditions []metering.ReportCondition, condType metering.ReportConditionType) []metering.ReportCondition {
	var newConditions []metering.ReportCondition
	for _, c := range conditions {
		if c.Type == condType {
			continue
		}
		newConditions = append(newConditions, c)
	}
	return newConditions
}
