package types

import (
	"encoding/json"
	"fmt"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

const (
	ReportGenerationQueryPlural = "reportgenerationqueries"
)

// +k8s:deepcopy-gen=true
type ReportGenerationQuery struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`

	Spec ReportGenerationQuerySpec `json:"spec"`
}

// +k8s:deepcopy-gen=true
type ReportGenerationQuerySpec struct {
	DataStoreName string           `json:"reportDataStore"`
	Query         string           `json:"query"`
	Columns       []GenQueryColumn `json:"columns"`
}

// +k8s:deepcopy-gen=true
type GenQueryColumn struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func GetReportGenerationQuery(r rest.Interface, namespace, name string) (ReportGenerationQuery, error) {
	req := r.Get().
		Namespace(namespace).
		Resource(ReportGenerationQueryPlural).
		Name(name).
		FieldsSelectorParam(nil)

	b, err := req.DoRaw()
	if err != nil {
		fmt.Printf("get op for namespace %q got err: %v\n", namespace, err)
		return ReportGenerationQuery{}, err
	}
	var reportGenerationQuery ReportGenerationQuery
	return reportGenerationQuery, json.Unmarshal(b, &reportGenerationQuery)
}
