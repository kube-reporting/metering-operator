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

	Spec   StorageLocationSpec   `json:"spec"`
	Status StorageLocationStatus `json:"status"`
}

type StorageLocationRef struct {
	StorageLocationName string `json:"storageLocationName,omitempty"`
}

type StorageLocationSpec struct {
	Hive *HiveStorage `json:"hive,omitempty"`
}

type HiveStorage struct {
	UnmanagedDatabase      bool                               `json:"unmanagedDatabase"`
	DatabaseName           string                             `json:"databaseName"`
	Location               string                             `json:"location,omitempty"`
	DefaultTableProperties *HiveStorageDefaultTableProperties `json:"defaultTableProperties,omitempty"`
}

type HiveStorageDefaultTableProperties struct {
	SerdeFormat        string            `json:"serdeFormat,omitempty"`
	FileFormat         string            `json:"***REMOVED***leFormat,omitempty"`
	SerdeRowProperties map[string]string `json:"serdeRowProperties,omitempty"`
}

type StorageLocationStatus struct {
	Hive HiveStorageStatus `json:"hive,omitempty"`
}

type HiveStorageStatus struct {
	DatabaseName string `json:"databaseName"`
	Location     string `json:"location,omitempty"`
}
