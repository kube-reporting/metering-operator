package v1alpha1

import (
	"fmt"

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
	// ReportingStart (deprecated: use inputs) is the beginning period of time
	// that the report will be based on.
	ReportingStart *meta.Time `json:"reportingStart"`

	// ReportingEnd (deprecated: use inputs)  is the end period of time that
	// the report will be based on.
	ReportingEnd *meta.Time `json:"reportingEnd"`

	// Inputs are the inputs to the ReportGenerationQuery
	Inputs ReportGenerationQueryInputValues `json:"inputs,omitempty"`

	// GenerationQueryName is the name of the ReportGenerationQuery that this
	// report should run
	GenerationQueryName string `json:"generationQuery"`

	// RunImmediately will run the report immediately, ignoring ReportingEnd and
	// GracePeriod.
	RunImmediately bool `json:"runImmediately,omitempty"`

	// GracePeriod controls how long after `ReportingEnd` to wait until running
	// the report
	GracePeriod *meta.Duration `json:"gracePeriod,omitempty"`

	// Output is the storage location where results are sent.
	Output *StorageLocationRef `json:"output,omitempty"`

	// ReportingEndInputName allows overriding the default expected input name that maps to the ReportPeriodEnd
	ReportingEndInputName string `json:"reportingEndInputName,omitempty"`
}

type ReportStatus struct {
	Phase     ReportPhase `json:"phase,omitempty"`
	Output    string      `json:"output,omitempty"`
	TableName string      `json:"tableName"`
}

type ReportPhase string

const (
	ReportPhaseFinished ReportPhase = "Finished"
	ReportPhaseWaiting  ReportPhase = "Waiting"
	ReportPhaseStarted  ReportPhase = "Started"
	ReportPhaseError    ReportPhase = "Error"
)

func (p *ReportPhase) UnmarshalText(text []byte) error {
	phase := ReportPhase(text)
	switch phase {
	case ReportPhaseFinished:
	case ReportPhaseWaiting:
	case ReportPhaseStarted:
	case ReportPhaseError:
	case ReportPhase(""): // default to waiting
		phase = ReportPhaseWaiting
	default:
		return fmt.Errorf("'%s' is not a ReportPhase", phase)
	}
	*p = phase
	return nil
}
