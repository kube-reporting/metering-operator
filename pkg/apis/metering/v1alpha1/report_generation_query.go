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

	Spec   ReportGenerationQuerySpec   `json:"spec"`
	Status ReportGenerationQueryStatus `json:"status"`
}

type ReportGenerationQuerySpec struct {
	Columns              []ReportGenerationQueryColumn          `json:"columns"`
	Query                string                                 `json:"query"`
	View                 GenQueryView                           `json:"view"`
	ReportQueries        []string                               `json:"reportQueries,omitempty"`
	DynamicReportQueries []string                               `json:"dynamicReportQueries,omitempty"`
	DataSources          []string                               `json:"reportDataSources,omitempty"`
	Reports              []string                               `json:"reports,omitempty"`
	Inputs               []ReportGenerationQueryInputDefinition `json:"inputs,omitempty"`
}

type ReportGenerationQueryColumn struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	TableHidden bool   `json:"tableHidden"`
	Unit        string `json:"unit,omitempty"`
}

type GenQueryView struct {
	// Disabled controls whether or not to create a view in presto for this
	// ReportGenerationQuery
	Disabled bool `json:"disabled"`
}

type ReportGenerationQueryInputDefinition struct {
	Name     string `json:"name"`
	Required bool   `json:"required"`
}

type ReportGenerationQueryInputValue struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type ReportGenerationQueryInputValues []ReportGenerationQueryInputValue

type ReportGenerationQueryStatus struct {
	// ViewName is the name of the view in Presto for this query, if the view
	// has been created. If it is empty, the view does not exist.
	ViewName string `json:"viewName,omitempty"`
}
