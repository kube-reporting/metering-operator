package util

import (
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1"
)

const (
	// Failure report conditions:
	//
	// GenerateReportErrorReason is added to a Report when an error
	// occurs while generating the report data.
	GenerateReportErrorReason = "GenerateReportError"

	// FailedValidationReason is added to a Report when the
	// Report is invalid or it's ReportGenerationQuery is invalid or
	// not ready
	FailedValidationReason = "FailedValidation"

	// Running report conditions:

	// ScheduledReason is added to a Report when it's reached the next
	// reporting time in it's schedule.
	ScheduledReason = "Scheduled"
	// ValidatingReportReason is added to a Report when the
	// report is being validated
	ValidatingReportReason = "ValidatingReport"
	// ReportPeriodWaitingReason is added to a Report when the report
	// has to wait until the next scheduled reporting time.
	ReportPeriodWaitingReason = "ReportPeriodNotFinished"
	// ReportPeriodFinishedReason is added to a Report when the report
	// has had it's report processed up until it's reportingEnd.
	ReportPeriodFinishedReason = "ReportPeriodFinished"
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
	newConditions := ***REMOVED***lterOutCondition(status.Conditions, condition.Type)
	status.Conditions = append(newConditions, condition)
}

// RemoveReportCondition removes the report condition with the provided type.
func RemoveReportCondition(status *v1alpha1.ReportStatus, condType v1alpha1.ReportConditionType) {
	status.Conditions = ***REMOVED***lterOutCondition(status.Conditions, condType)
}

// ***REMOVED***lterOutCondition returns a new slice of report conditions without conditions with the provided type.
func ***REMOVED***lterOutCondition(conditions []v1alpha1.ReportCondition, condType v1alpha1.ReportConditionType) []v1alpha1.ReportCondition {
	var newConditions []v1alpha1.ReportCondition
	for _, c := range conditions {
		if c.Type == condType {
			continue
		}
		newConditions = append(newConditions, c)
	}
	return newConditions
}
