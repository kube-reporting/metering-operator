package v1alpha1

import (
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

	Spec   ScheduledReportSpec `json:"spec"`
	Status ReportStatus        `json:"status"`
}

type ScheduledReportStorageLocation struct {
	StorageLocationName string               `json:"storageLocationName,omitempty"`
	StorageSpec         *StorageLocationSpec `json:"spec,omitempty"`
}

type ScheduledReportSpec struct {
	GenerationQueryName string `json:"generationQuery"`

	Schedule ScheduledReportSchedule `json:"schedule"`

	// GracePeriod controls how long after each period to wait until running
	// the report
	GracePeriod *meta.Duration `json:"gracePeriod,omitempty"`

	// Output is the storage location where results are sent.
	Output *ReportStorageLocation `json:"output,omitempty"`
}

type ScheduledReportPeriod string

const (
	ScheduledReportPeriodHourly  ScheduledReportPeriod = "hourly"
	ScheduledReportPeriodDaily   ScheduledReportPeriod = "daily"
	ScheduledReportPeriodWeekly  ScheduledReportPeriod = "weekly"
	ScheduledReportPeriodMonthly ScheduledReportPeriod = "monthly"
)

type ScheduledReportSchedule struct {
	Period ScheduledReportPeriod `json:"period"`

	Hourly  *ScheduledReportScheduleHourly  `json:"hourly,omitempty"`
	Daily   *ScheduledReportScheduleDaily   `json:"daily,omitempty"`
	Weekly  *ScheduledReportScheduleWeekly  `json:"weekly,omitempty"`
	Monthly *ScheduledReportScheduleMonthly `json:"monthly,omitempty"`
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
	DayOfWeek string `json:"dayOfWeek,omitempty"`
	Hour      int64  `json:"hour,omitempty"`
	Minute    int64  `json:"minute,omitempty"`
	Second    int64  `json:"second,omitempty"`
}

type ScheduledReportScheduleMonthly struct {
	DayOfMonth int64 `json:"dayOfMonth,omitempty"`
	Hour       int64 `json:"hour,omitempty"`
	Minute     int64 `json:"minute,omitempty"`
	Second     int64 `json:"second,omitempty"`
}
