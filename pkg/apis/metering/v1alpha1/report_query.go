package v1alpha1

import (
	"encoding/json"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var ReportQueryGVK = SchemeGroupVersion.WithKind("ReportQuery")

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ReportQueryList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`
	Items         []*ReportQuery `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ReportQuery struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`

	Spec   ReportQuerySpec   `json:"spec"`
	Status ReportQueryStatus `json:"status"`
}

type ReportQuerySpec struct {
	Columns []ReportQueryColumn          `json:"columns"`
	Query   string                       `json:"query"`
	Inputs  []ReportQueryInputDe***REMOVED***nition `json:"inputs,omitempty"`
}

type ReportQueryColumn struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	TableHidden bool   `json:"tableHidden"`
	Unit        string `json:"unit,omitempty"`
}

type ReportQueryInputDe***REMOVED***nition struct {
	Name     string           `json:"name"`
	Required bool             `json:"required"`
	Type     string           `json:"type,omitempty"`
	Default  *json.RawMessage `json:"default,omitempty"`
}

type ReportQueryInputValue struct {
	Name  string           `json:"name"`
	Value *json.RawMessage `json:"value,omitempty"`
}

type ReportQueryInputValues []ReportQueryInputValue

type ReportQueryStatus struct {
}
