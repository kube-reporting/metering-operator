package v1alpha1

import (
	"encoding/json"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var ReportGenerationQueryGVK = SchemeGroupVersion.WithKind("ReportGenerationQuery")

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ReportGenerationQueryList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`
	Items         []*ReportGenerationQuery `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ReportGenerationQuery struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`

	Spec   ReportGenerationQuerySpec   `json:"spec"`
	Status ReportGenerationQueryStatus `json:"status"`
}

type ReportGenerationQuerySpec struct {
	Columns              []ReportGenerationQueryColumn          `json:"columns"`
	Query                string                                 `json:"query"`
	DynamicReportQueries []string                               `json:"dynamicReportQueries,omitempty"`
	DataSources          []string                               `json:"reportDataSources,omitempty"`
	Reports              []string                               `json:"reports,omitempty"`
	Inputs               []ReportGenerationQueryInputDe***REMOVED***nition `json:"inputs,omitempty"`
}

type ReportGenerationQueryColumn struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	TableHidden bool   `json:"tableHidden"`
	Unit        string `json:"unit,omitempty"`
}

type ReportGenerationQueryInputDe***REMOVED***nition struct {
	Name     string `json:"name"`
	Required bool   `json:"required"`
	Type     string `json:"type,omitempty"`
}

type ReportGenerationQueryInputValue struct {
	Name  string           `json:"name"`
	Value *json.RawMessage `json:"value,omitempty"`
}

type ReportGenerationQueryInputValues []ReportGenerationQueryInputValue

type ReportGenerationQueryStatus struct {
}
