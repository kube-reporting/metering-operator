package v1

import (
	"fmt"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	cb "github.com/coreos-inc/kube-chargeback/pkg/chargeback/v1"
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
	// Frequency specifies how often a report is run and determines report period if range isn't set in ReportSpec.
	Frequency CronFrequency `json:"frequency"`

	// Suspend stops execution of the schedule.
	Suspend *bool `json:"suspend,omitempty"`

	// ReportTemplate dictates the report which is created at the given schedule.
	ReportTemplate cb.ReportTemplateSpec `json:"reportTemplate"`
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

// CronFrequency offers fixed options for recurring reports.
type CronFrequency string

const (
	CronFrequencyHourly CronFrequency = "Hourly"
	CronFrequencyDaily  CronFrequency = "Daily"
	CronFrequencyWeekly CronFrequency = "Weekly"
)

func (f *CronFrequency) UnmarshalText(text []byte) error {
	freq := CronFrequency(text)
	switch freq {
	case CronFrequencyHourly:
	case CronFrequencyDaily:
	case CronFrequencyWeekly:
	default:
		return fmt.Errorf("'%s' is not a CronFrequency", freq)
	}
	*f = freq
	return nil
}
