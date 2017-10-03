package v1

import (
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	cbTypes "github.com/coreos-inc/kube-chargeback/pkg/chargeback/v1/types"
)

// Cron allows reports to be scheduled.
type Cron struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`

	Spec   CronSpec   `json:"spec"`
	Status CronStatus `json:"status"`
}

// CronSpec defines which report should be run and when.
type CronSpec struct {
	// Schedule is defined using a CRON expression. Details at https://en.wikipedia.org/wiki/Cron.
	Schedule string `json:"schedule"`

	// Suspend stops execution of the schedule.
	Suspend *bool `json:"suspend,omitempty"`

	// ReportTemplate dictates the report which is created at the given schedule.
	ReportTemplate cbTypes.ReportTemplateSpec `json:"reportTemplate"`
}

// CronStatus displays the state of a Cron schedule.
type CronStatus struct {
	// LastScheduleTime was the last successful run of a schedule.
	LastScheduleTime *meta.Time `json:"lastScheduleTime"`
}

// CronList is a collection of Cron schedules.
type CronList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`
	Items         []*Cron `json:"items"`
}
