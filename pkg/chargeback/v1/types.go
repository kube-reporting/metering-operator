package v1

import (
	"fmt"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen=true
type Report struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`

	Spec   ReportSpec   `json:"spec"`
	Status ReportStatus `json:"status"`
}

// +k8s:deepcopy-gen=true
type ReportSpec struct {
	// ReportingStart is the beginning period of time that the report will be based on.
	ReportingStart meta.Time `json:"reportingStart"`

	// ReportingEnd is the end period of time that the report will be based on.
	ReportingEnd meta.Time `json:"reportingEnd"`

	// Chargeback is the bucket that stores chargeback metering data.
	Chargeback S3Bucket `json:"chargeback"`

	// AWS identi***REMOVED***es the location of the a billing report, as con***REMOVED***gured in the AWS Console.
	AWS AWSUsage `json:"aws"`

	// Output is the S3 bucket where results are sent.
	Output S3Bucket `json:"output"`
}

type ReportTemplateSpec struct {
	meta.ObjectMeta `json:"metadata,omitempty"`

	Spec ReportSpec `json:"spec"`
}

// +k8s:deepcopy-gen=true
type ReportStatus struct {
	Phase  ReportPhase `json:"phase"`
	Output string      `json:"output"`
}

// +k8s:deepcopy-gen=true
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
	default:
		return fmt.Errorf("'%s' is not a ReportPhase", phase)
	}
	*p = phase
	return nil
}

// +k8s:deepcopy-gen=true
type S3Bucket struct {
	Bucket string `json:"bucket"`
	Pre***REMOVED***x string `json:"pre***REMOVED***x"`
}

// +k8s:deepcopy-gen=true
// AWSPodCostReport details the expense of running a Pod over a period of time on Amazon Web Services.
type AWSUsage struct {
	// ReportName as con***REMOVED***gured in AWS Console.
	ReportName string `json:"reportName"`

	// ReportPre***REMOVED***x as con***REMOVED***gured in AWS Console.
	ReportPre***REMOVED***x string `json:"reportPre***REMOVED***x"`

	// Bucket that the report is con***REMOVED***gured to store in. Setup in AWS Console.
	Bucket string `json:"bucket"`
}

// +k8s:deepcopy-gen=true
type ReportList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`
	Items         []*Report `json:"items"`
}
