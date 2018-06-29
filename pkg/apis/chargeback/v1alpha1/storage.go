package v1alpha1

import (
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const IsDefaultStorageLocationAnnotation = "storagelocation.chargeback.coreos.com/is-default"

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
	S3   *S3Bucket    `json:"s3,omitempty"`
	Hive *HiveStorage `json:"hiveStorage,omitempty"`
}

type S3Bucket struct {
	Bucket string `json:"bucket"`
	Pre***REMOVED***x string `json:"pre***REMOVED***x"`
}

type HiveStorage struct {
	TableProperties TableProperties `json:"tableProperties"`
}

type StorageLocationRef struct {
	StorageLocationName string               `json:"storageLocationName,omitempty"`
	StorageSpec         *StorageLocationSpec `json:"spec,omitempty"`
}
