package chargeback

import (
	"fmt"
	"time"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Query struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`

	Spec   QuerySpec   `json:"spec"`
	Status QueryStatus `json:"status"`
}

type QuerySpec struct {
	// ReportingStart is the beginning period of time that the report will be based on.
	ReportingStart time.Time `json:"reportingStart"`

	// ReportingEnd is the end period of time that the report will be based on.
	ReportingEnd time.Time `json:"reportingEnd"`

	// Chargeback is the bucket that stores chargeback metering data.
	Chargeback S3Bucket `json:"chargeback"`

	// AWS identifies the location of the a billing report, as configured in the AWS Console.
	AWS AWSUsage `json:"aws"`

	// Output is the S3 bucket where results are sent.
	Output S3Bucket `json:"output"`
}

type QueryStatus struct {
	Phase  QueryPhase `json:"phase"`
	Output string     `json:"output"`
}

type QueryPhase string

const (
	QueryPhaseFinished QueryPhase = "Finished"
	QueryPhaseWaiting  QueryPhase = "Waiting"
	QueryPhaseStarted  QueryPhase = "Started"
	QueryPhaseError    QueryPhase = "Error"
)

func (p *QueryPhase) UnmarshalText(text []byte) error {
	phase := QueryPhase(text)
	switch phase {
	case QueryPhaseFinished:
	case QueryPhaseWaiting:
	case QueryPhaseStarted:
	case QueryPhaseError:
	default:
		return fmt.Errorf("'%s' is not a QueryPhase", phase)
	}
	*p = phase
	return nil
}

type S3Bucket struct {
	Bucket string `json:"bucket"`
	Prefix string `json:"prefix"`
}

// AWSPodCostReport details the expense of running a Pod over a period of time on Amazon Web Services.
type AWSUsage struct {
	// ReportName as configured in AWS Console.
	ReportName string `json:"reportName"`

	// ReportPrefix as configured in AWS Console.
	ReportPrefix string `json:"reportPrefix"`

	// Bucket that the report is configured to store in. Setup in AWS Console.
	Bucket string `json:"bucket"`
}

type QueryList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`
	Items         []*Query `json:"items"`
}
