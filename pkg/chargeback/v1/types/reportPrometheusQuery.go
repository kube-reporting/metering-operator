package types

import (
	"encoding/json"
	"fmt"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

const (
	ReportPrometheusQueryPlural = "reportprometheusqueries"
)

// +k8s:deepcopy-gen=true
type ReportPrometheusQuery struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`

	Spec ReportPrometheusQuerySpec `json:"spec"`
}

// +k8s:deepcopy-gen=true
type ReportPrometheusQuerySpec struct {
	Query string `json:"query"`
}

func GetReportPrometheusQuery(r rest.Interface, namespace, name string) (ReportPrometheusQuery, error) {
	req := r.Get().
		Namespace(namespace).
		Resource(ReportPrometheusQueryPlural).
		Name(name).
		FieldsSelectorParam(nil)

	b, err := req.DoRaw()
	if err != nil {
		fmt.Printf("get op for namespace %q got err: %v\n", namespace, err)
		return ReportPrometheusQuery{}, err
	}
	var prometheusQuery ReportPrometheusQuery
	return prometheusQuery, json.Unmarshal(b, &prometheusQuery)
}
