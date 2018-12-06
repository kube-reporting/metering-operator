package v1alpha1

import (
	"k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ReportList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`
	Items         []*Report `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Report struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`

	Spec   ReportSpec   `json:"spec"`
	Status ReportStatus `json:"status"`
}

type ReportSpec struct {
	// GenerationQueryName speci***REMOVED***es the ReportGenerationQuery to execute when
	// the report runs.
	GenerationQueryName string `json:"generationQuery"`

	// Schedule con***REMOVED***gures when the report runs.
	Schedule *ReportSchedule `json:"schedule,omitempty"`

	// ReportingStart speci***REMOVED***es the time this Report should start from
	// instead of the current time.
	// This is intended for allowing a Report to start from the past
	// and report on data collected before the Report was created.
	ReportingStart *meta.Time `json:"reportingStart,omitempty"`

	// ReportingEnd speci***REMOVED***es the time this Report should stop
	// running. Once a Report has reached ReportingEnd, no new results
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

	// RunImmediately will run the report immediately, ignoring ReportingStart,
	// ReportingEnd and GracePeriod.
	RunImmediately bool `json:"runImmediately,omitempty"`

	// Inputs are the inputs to the ReportGenerationQuery
	Inputs ReportGenerationQueryInputValues `json:"inputs,omitempty"`

	// Output is the storage location where results are sent.
	Output *StorageLocationRef `json:"output,omitempty"`
}

type ReportPeriod string

const (
	ReportPeriodCron    ReportPeriod = "cron"
	ReportPeriodHourly  ReportPeriod = "hourly"
	ReportPeriodDaily   ReportPeriod = "daily"
	ReportPeriodWeekly  ReportPeriod = "weekly"
	ReportPeriodMonthly ReportPeriod = "monthly"
)

type ReportSchedule struct {
	Period ReportPeriod `json:"period"`

	Cron    *ReportScheduleCron    `json:"cron,omitempty"`
	Hourly  *ReportScheduleHourly  `json:"hourly,omitempty"`
	Daily   *ReportScheduleDaily   `json:"daily,omitempty"`
	Weekly  *ReportScheduleWeekly  `json:"weekly,omitempty"`
	Monthly *ReportScheduleMonthly `json:"monthly,omitempty"`
}

type ReportScheduleCron struct {
	Expression string `json:"expression,omitempty"`
}

type ReportScheduleHourly struct {
	Minute int64 `json:"minute,omitempty"`
	Second int64 `json:"second,omitempty"`
}

type ReportScheduleDaily struct {
	Hour   int64 `json:"hour,omitempty"`
	Minute int64 `json:"minute,omitempty"`
	Second int64 `json:"second,omitempty"`
}

type ReportScheduleWeekly struct {
	DayOfWeek *string `json:"dayOfWeek,omitempty"`
	Hour      int64   `json:"hour,omitempty"`
	Minute    int64   `json:"minute,omitempty"`
	Second    int64   `json:"second,omitempty"`
}

type ReportScheduleMonthly struct {
	DayOfMonth *int64 `json:"dayOfMonth,omitempty"`
	Hour       int64  `json:"hour,omitempty"`
	Minute     int64  `json:"minute,omitempty"`
	Second     int64  `json:"second,omitempty"`
}

type ReportStatus struct {
	Conditions     []ReportCondition `json:"conditions,omitempty"`
	LastReportTime *meta.Time                 `json:"lastReportTime,omitempty"`
	TableName      string                     `json:"tableName"`
}

type ReportCondition struct {
	// Type of Report condition, Waiting, Active or Failed.
	Type ReportConditionType `json:"type"`
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

type ReportConditionType string

const (
	ReportRunning ReportConditionType = "Running"
	ReportFailure ReportConditionType = "Failure"
)
