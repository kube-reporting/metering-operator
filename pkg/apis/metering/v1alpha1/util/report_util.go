package util

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1"
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
	// Report is invalid or it's ReportGenerationQuery is invalid.
	InvalidReportReason = "InvalidReport"

	// ReportingPeriodUnmetDependenciesReason is set when a Report cannot run
	// because it's dependencies (ReportGenerationQueries dependencies on
	// ReportDataSources, ReportGenerationQueries, and other Reports) do not
	// have data available for the reporting period currently being processed.
	ReportingPeriodUnmetDependenciesReason = "ReportingPeriodUnmetDependencies"

	// GenerateReportFailedReason is set when a Report is not running because
	// it previously failed when generating results previously.
	GenerateReportFailedReason = "GenerateReportFailed"
)

// NewReportCondition creates a new report condition.
func NewReportCondition(condType v1alpha1.ReportConditionType, status v1.ConditionStatus, reason, message string) *v1alpha1.ReportCondition {
	return &v1alpha1.ReportCondition{
		Type:               condType,
		Status:             status,
		LastUpdateTime:     metav1.Now(),
		LastTransitionTime: metav1.Now(),
		Reason:             reason,
		Message:            message,
	}
}

// GetReportCondition returns the condition with the provided type.
func GetReportCondition(status v1alpha1.ReportStatus, condType v1alpha1.ReportConditionType) *v1alpha1.ReportCondition {
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
func SetReportCondition(status *v1alpha1.ReportStatus, condition v1alpha1.ReportCondition) {
	currentCond := GetReportCondition(*status, condition.Type)
	if currentCond != nil && currentCond.Status == condition.Status && currentCond.Reason == condition.Reason {
		return
	}
	// Do not update lastTransitionTime if the status of the condition doesn't change.
	if currentCond != nil && currentCond.Status == condition.Status {
		condition.LastTransitionTime = currentCond.LastTransitionTime
	}
	newConditions := filterOutCondition(status.Conditions, condition.Type)
	status.Conditions = append(newConditions, condition)
}

// RemoveReportCondition removes the report condition with the provided type.
func RemoveReportCondition(status *v1alpha1.ReportStatus, condType v1alpha1.ReportConditionType) {
	status.Conditions = filterOutCondition(status.Conditions, condType)
}

// filterOutCondition returns a new slice of report conditions without conditions with the provided type.
func filterOutCondition(conditions []v1alpha1.ReportCondition, condType v1alpha1.ReportConditionType) []v1alpha1.ReportCondition {
	var newConditions []v1alpha1.ReportCondition
	for _, c := range conditions {
		if c.Type == condType {
			continue
		}
		newConditions = append(newConditions, c)
	}
	return newConditions
}
