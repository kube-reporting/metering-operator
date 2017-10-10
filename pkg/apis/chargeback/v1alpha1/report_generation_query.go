package v1alpha1

import (
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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

	Spec ReportGenerationQuerySpec `json:"spec"`
}

type ReportGenerationQuerySpec struct {
	DataStoreName string           `json:"reportDataStore"`
	Query         string           `json:"query"`
	Columns       []GenQueryColumn `json:"columns"`
}

type GenQueryColumn struct {
	Name string `json:"name"`
	Type string `json:"type"`
}
