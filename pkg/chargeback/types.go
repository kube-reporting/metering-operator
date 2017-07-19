package chargeback

import (
	"fmt"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Query struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`

	Spec   QuerySpec   `json:"spec"`
	Status QueryStatus `json:"status"`
}

type QuerySpec struct {
	// ReportingRange is the period of time that the report will be based on.
	ReportingRange Range `json:"reportingRange"`

	// Chargeback is the bucket that stores chargeback metering data.
	Chargeback S3Bucket `json:"chargeback"`

	// AWS identi***REMOVED***es the location of the a billing report, as con***REMOVED***gured in the AWS Console.
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
	Pre***REMOVED***x string `json:"pre***REMOVED***x"`
}

// AWSPodCostReport details the expense of running a Pod over a period of time on Amazon Web Services.
type AWSUsage struct {
	// ReportName as con***REMOVED***gured in AWS Console.
	ReportName string `json:"reportName"`

	// ReportPre***REMOVED***x as con***REMOVED***gured in AWS Console.
	ReportPre***REMOVED***x string `json:"reportPre***REMOVED***x"`

	// Bucket that the report is con***REMOVED***gured to store in. Setup in AWS Console.
	Bucket string `json:"bucket"`
}
