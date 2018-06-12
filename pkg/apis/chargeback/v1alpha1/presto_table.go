package v1alpha1

import (
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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

	State PrestoTableState `json:"state"`
}

type PrestoTableState struct {
	// CreationParameters holds all arguments used in the call to
	// pkg/hive/query.go#createTable
	CreationParameters PrestoTableCreationParameters `json:"creationParameters"`

	// Partitions holds all currently con***REMOVED***gured partitions for a given table.
	// Currently only relevant to tables backed by AWS billing reports.
	Partitions []PrestoTablePartition `json:"partitions"`
}

type PrestoTableCreationParameters struct {
	TableName    string              `json:"tableName"`
	Location     string              `json:"location,omitempty"`
	SerdeFmt     string              `json:"serdeFmt,omitempty"`
	Format       string              `json:"format,omitempty"`
	SerdeProps   map[string]string   `json:"serdeProps,omitempty"`
	Columns      []PrestoTableColumn `json:"columns"`
	Partitions   []PrestoTableColumn `json:"partitions,omitempty"`
	External     bool                `json:"external,omitempty"`
	IgnoreExists bool                `json:"ignoreExists,omitempty"`
}

type PrestoTableColumn struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type PrestoTablePartition struct {
	Start    string `json:"start"`
	End      string `json:"end"`
	Location string `json:"location"`
}
