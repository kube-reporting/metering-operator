package util

import (
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1"
)

const (
	// Failure scheduledReport conditions:
	//
	// GenerateReportErrorReason is added to a ScheduledReport when an error
	// occurs while generating the report data.
	GenerateReportErrorReason = "GenerateReportError"

	// Running scheduledReport conditions:

	// ScheduledReason is added to a ScheduledReport when it's reached the next
	// reporting time in it's schedule.
	ScheduledReason = "Scheduled"
	// ValidatingScheduledReportReason is added to a ScheduledReport when the
	// report is having it's ReportGenerationQuery validated
	ValidatingScheduledReportReason = "ValidatingScheduledReport"
	// ReportPeriodWaitingReason is added to a ScheduledReport when the report
	// has to wait until the next scheduled reporting time.
	ReportPeriodWaitingReason = "ReportPeriodNotFinished"
	// ReportPeriodFinishedReason is added to a ScheduledReport when the report
	// has had it's report processed up until it's reportingEnd.
	ReportPeriodFinishedReason = "ReportPeriodFinished"
)

// NewScheduledReportCondition creates a new scheduledReport condition.
func NewScheduledReportCondition(condType v1alpha1.ScheduledReportConditionType, status v1.ConditionStatus, reason, message string) *v1alpha1.ScheduledReportCondition {
	return &v1alpha1.ScheduledReportCondition{
		Type:               condType,
		Status:             status,
		LastUpdateTime:     metav1.Now(),
		LastTransitionTime: metav1.Now(),
		Reason:             reason,
		Message:            message,
	}
}

// GetScheduledReportCondition returns the condition with the provided type.
func GetScheduledReportCondition(status v1alpha1.ScheduledReportStatus, condType v1alpha1.ScheduledReportConditionType) *v1alpha1.ScheduledReportCondition {
	for i := range status.Conditions {
		c := status.Conditions[i]
		if c.Type == condType {
			return &c
		}
	}
	return nil
}

// SetScheduledReportCondition updates the scheduledReport to include the provided condition. If the condition that
// we are about to add already exists and has the same status and reason then we are not going to update.
func SetScheduledReportCondition(status *v1alpha1.ScheduledReportStatus, condition v1alpha1.ScheduledReportCondition) {
	currentCond := GetScheduledReportCondition(*status, condition.Type)
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

// RemoveScheduledReportCondition removes the scheduledReport condition with the provided type.
func RemoveScheduledReportCondition(status *v1alpha1.ScheduledReportStatus, condType v1alpha1.ScheduledReportConditionType) {
	status.Conditions = filterOutCondition(status.Conditions, condType)
}

// filterOutCondition returns a new slice of scheduledReport conditions without conditions with the provided type.
func filterOutCondition(conditions []v1alpha1.ScheduledReportCondition, condType v1alpha1.ScheduledReportConditionType) []v1alpha1.ScheduledReportCondition {
	var newConditions []v1alpha1.ScheduledReportCondition
	for _, c := range conditions {
		if c.Type == condType {
			continue
		}
		newConditions = append(newConditions, c)
	}
	return newConditions
}
