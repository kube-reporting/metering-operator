package v1alpha1

import (
	"k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ScheduledReportList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`
	Items         []*ScheduledReport `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ScheduledReport struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`

	Spec   ScheduledReportSpec   `json:"spec"`
	Status ScheduledReportStatus `json:"status"`
}

type ScheduledReportSpec struct {
	// GenerationQueryName speci***REMOVED***es the ReportGenerationQuery to execute when
	// the report runs.
	GenerationQueryName string `json:"generationQuery"`

	// Schedule con***REMOVED***gures when the report runs.
	Schedule ScheduledReportSchedule `json:"schedule"`

	// ReportingStart speci***REMOVED***es the time this ScheduledReport should start from
	// instead of the current time.
	// This is intended for allowing a ScheduledReport to start from the past
	// and report on data collected before the ScheduledReport was created.
	ReportingStart *meta.Time `json:"reportingStart,omitempty"`

	// ReportingEnd speci***REMOVED***es the time this ScheduledReport should stop
	// running. Once a ScheduledReport has reached ReportingEnd, no new results
	// will be generated.
	ReportingEnd *meta.Time `json:"reportingEnd,omitempty"`

	// GracePeriod controls how long after each period to wait until running
	// the report
	GracePeriod *meta.Duration `json:"gracePeriod,omitempty"`

	// OverwriteExistingData controls whether or not to delete any existing
	// data in the report table before the scheduled report runs. Useful for
	// having a report that is just a snapshot of the most recent data rather
	// than a log of all runs before it.
	OverwriteExistingData bool `json:"overwriteExistingData,omitempty"`

	// Inputs are the inputs to the ReportGenerationQuery
	Inputs ReportGenerationQueryInputValues `json:"inputs,omitempty"`

	// Output is the storage location where results are sent.
	Output *StorageLocationRef `json:"output,omitempty"`
}

type ScheduledReportPeriod string

const (
	ScheduledReportPeriodCron    ScheduledReportPeriod = "cron"
	ScheduledReportPeriodHourly  ScheduledReportPeriod = "hourly"
	ScheduledReportPeriodDaily   ScheduledReportPeriod = "daily"
	ScheduledReportPeriodWeekly  ScheduledReportPeriod = "weekly"
	ScheduledReportPeriodMonthly ScheduledReportPeriod = "monthly"
)

type ScheduledReportSchedule struct {
	Period ScheduledReportPeriod `json:"period"`

	Cron    *ScheduledReportScheduleCron    `json:"cron,omitempty"`
	Hourly  *ScheduledReportScheduleHourly  `json:"hourly,omitempty"`
	Daily   *ScheduledReportScheduleDaily   `json:"daily,omitempty"`
	Weekly  *ScheduledReportScheduleWeekly  `json:"weekly,omitempty"`
	Monthly *ScheduledReportScheduleMonthly `json:"monthly,omitempty"`
}

type ScheduledReportScheduleCron struct {
	Expression string `json:"expression,omitempty"`
}

type ScheduledReportScheduleHourly struct {
	Minute int64 `json:"minute,omitempty"`
	Second int64 `json:"second,omitempty"`
}

type ScheduledReportScheduleDaily struct {
	Hour   int64 `json:"hour,omitempty"`
	Minute int64 `json:"minute,omitempty"`
	Second int64 `json:"second,omitempty"`
}

type ScheduledReportScheduleWeekly struct {
	DayOfWeek *string `json:"dayOfWeek,omitempty"`
	Hour      int64   `json:"hour,omitempty"`
	Minute    int64   `json:"minute,omitempty"`
	Second    int64   `json:"second,omitempty"`
}

type ScheduledReportScheduleMonthly struct {
	DayOfMonth *int64 `json:"dayOfMonth,omitempty"`
	Hour       int64  `json:"hour,omitempty"`
	Minute     int64  `json:"minute,omitempty"`
	Second     int64  `json:"second,omitempty"`
}

type ScheduledReportStatus struct {
	Conditions     []ScheduledReportCondition `json:"conditions,omitempty"`
	LastReportTime *meta.Time                 `json:"lastReportTime,omitempty"`
	TableName      string                     `json:"table_name"`
}

type ScheduledReportCondition struct {
	// Type of ScheduledReport condition, Waiting, Active or Failed.
	Type ScheduledReportConditionType `json:"type"`
	// Status of the condition, one of True, False, Unknown.
	Status v1.ConditionStatus `json:"status"`
	// Last time the condition was checked.
	// +optional
	LastUpdateTime meta.Time `json:"lastUpdateTime,omitempty"`
	// Last time the condition transit from one status to another.
	// +optional
	LastTransitionTime meta.Time `json:"lastTransitionTime,omitempty"`
	// (brief) reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty"`
	// Human readable message indicating details about last transition.
	// +optional
	Message string `json:"message,omitempty"`
}

type ScheduledReportConditionType string

const (
	ScheduledReportRunning ScheduledReportConditionType = "Running"
	ScheduledReportFailure ScheduledReportConditionType = "Failure"
)
