package v1alpha1

import meta "k8s.io/apimachinery/pkg/apis/meta/v1"

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

type ReportStorageLocation struct {
	StorageLocationName string               `json:"storageLocationName,omitempty"`
	StorageSpec         *StorageLocationSpec `json:"spec,omitempty"`
}

type StorageLocationSpec struct {
	S3    *S3Bucket     `json:"s3,omitempty"`
	Local *LocalStorage `json:"local,omitempty"`
}

type S3Bucket struct {
	Bucket string `json:"bucket"`
	Pre***REMOVED***x string `json:"pre***REMOVED***x"`
}

type LocalStorage struct{}
