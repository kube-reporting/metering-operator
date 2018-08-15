package v1alpha1

import (
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const IsDefaultStorageLocationAnnotation = "storagelocation.metering.openshift.io/is-default"

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type StorageLocationList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`
	Items         []*StorageLocation `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type StorageLocation struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`

	Spec StorageLocationSpec `json:"spec"`
}

type StorageLocationSpec struct {
	Hive *HiveStorage `json:"hive,omitempty"`
}

type HiveStorage struct {
	TableProperties TableProperties `json:"tableProperties"`
}

type StorageLocationRef struct {
	StorageLocationName string               `json:"storageLocationName,omitempty"`
	StorageSpec         *StorageLocationSpec `json:"spec,omitempty"`
}
