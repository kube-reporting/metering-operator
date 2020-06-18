package v1

import (
	presto "github.com/kube-reporting/metering-operator/pkg/presto"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var PrestoTableGVK = SchemeGroupVersion.WithKind("PrestoTable")

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type PrestoTableList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`
	Items         []*PrestoTable `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type PrestoTable struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`

	Spec   PrestoTableSpec   `json:"spec"`
	Status PrestoTableStatus `json:"status"`
}

type PrestoTableSpec struct {
	// Unmanaged indicates that this table is not to be actively managed by the operator.
	Unmanaged bool `json:"unmanaged"`

	Catalog   string          `json:"catalog"`
	Schema    string          `json:"schema"`
	TableName string          `json:"tableName"`
	Columns   []presto.Column `json:"columns"`

	Properties map[string]string `json:"properties,omitempty"`
	Comment    string            `json:"comment,string"`

	// If true, uses "query" to create a view instead of a table.
	View bool `json:"view,omitempty"`
	// If true, uses "query" to create a table using CREATE TABLE AS.
	CreateTableAs bool   `json:"createTableAs,omitempty"`
	Query         string `json:"query,omitempty"`
}

type PrestoTableStatus struct {
	Catalog   string          `json:"catalog"`
	Schema    string          `json:"schema"`
	TableName string          `json:"tableName"`
	Columns   []presto.Column `json:"columns,omitempty"`

	Properties map[string]string `json:"properties,omitempty"`
	Comment    string            `json:"comment,string"`

	// If true, uses "query" to create a view instead of a table.
	View bool `json:"view,omitempty"`
	// If true, uses "query" to create a table using CREATE TABLE AS.
	CreateTableAs bool   `json:"createTableAs,omitempty"`
	Query         string `json:"query,omitempty"`
}
