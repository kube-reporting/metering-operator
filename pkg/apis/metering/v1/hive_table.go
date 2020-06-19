package v1

import (
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kube-reporting/metering-operator/pkg/hive"
)

var HiveTableGVK = SchemeGroupVersion.WithKind("HiveTable")

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type HiveTableList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`
	Items         []*HiveTable `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type HiveTable struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`

	Spec   HiveTableSpec   `json:"spec"`
	Status HiveTableStatus `json:"status"`
}

type HiveTablePartition hive.TablePartition

type HiveColumn hive.Column

type SortColumn struct {
	Name      string `json:"name"`
	Decending *bool  `json:"decending,omitempty"`
}

type HiveTableSpec struct {
	DatabaseName  string        `json:"databaseName,omitempty"`
	TableName     string        `json:"tableName"`
	Columns       []hive.Column `json:"columns"`
	PartitionedBy []hive.Column `json:"partitionedBy,omitempty"`
	ClusteredBy   []string      `json:"clusteredBy,omitempty"`
	SortedBy      []SortColumn  `json:"sortedBy,omitempty"`
	NumBuckets    int           `json:"numBuckets,omitempty"`

	Location        string            `json:"location,omitempty"`
	RowFormat       string            `json:"rowFormat,omitempty"`
	FileFormat      string            `json:"fileFormat,omitempty"`
	TableProperties map[string]string `json:"tableProperties,omitempty"`
	External        bool              `json:"external,omitempty"`

	ManagePartitions bool                 `json:"managePartitions"`
	Partitions       []HiveTablePartition `json:"partitions,omitempty"`
}

type HiveTableStatus struct {
	DatabaseName  string        `json:"databaseName,omitempty"`
	TableName     string        `json:"tableName,omitempty"`
	Columns       []hive.Column `json:"columns,omitempty"`
	PartitionedBy []hive.Column `json:"partitionedBy,omitempty"`
	ClusteredBy   []string      `json:"clusteredBy,omitempty"`
	SortedBy      []SortColumn  `json:"sortedBy,omitempty"`
	NumBuckets    int           `json:"numBuckets,omitempty"`

	Location        string            `json:"location,omitempty"`
	RowFormat       string            `json:"rowFormat,omitempty"`
	FileFormat      string            `json:"fileFormat,omitempty"`
	TableProperties map[string]string `json:"tableProperties,omitempty"`
	External        bool              `json:"external,omitempty"`

	Partitions []HiveTablePartition `json:"partitions,omitempty"`
}
