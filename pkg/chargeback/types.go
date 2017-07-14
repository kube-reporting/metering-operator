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
	// Range of time to be queried.
	Range `json:"range"`

	S3 S3Output `json:"s3"`
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

type S3Output struct {
	Bucket string `json:"bucket"`
	Pre***REMOVED***x string `json:"pre***REMOVED***x"`
	Secret string `json:"secret"`
}
