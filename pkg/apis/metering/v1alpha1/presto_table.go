package v1alpha1

import (
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/operator-framework/operator-metering/pkg/hive"
	"github.com/operator-framework/operator-metering/pkg/presto"
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

	Status PrestoTableStatus `json:"status"`
}

type TableParameters hive.TableParameters

type TableProperties hive.TableProperties

type TablePartition presto.TablePartition

type PrestoTableStatus struct {
	Parameters TableParameters  `json:"parameters"`
	Properties TableProperties  `json:"properties"`
	Partitions []TablePartition `json:"partitions"`
}
