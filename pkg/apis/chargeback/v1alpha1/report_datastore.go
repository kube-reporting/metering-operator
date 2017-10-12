package v1alpha1

import (
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ReportDataStoreList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`
	Items         []*ReportDataStore `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ReportDataStore struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`

	Spec      ReportDataStoreSpec `json:"spec"`
	TableName string              `json:"tableName"`
}

type ReportDataStoreSpec struct {
	DataStoreSource `json:",inline"`
}

type DataStoreSource struct {
	// Prommsum represents a datastore which holds Prometheus metrics
	Promsum *PromsumDataSource `json:"promsum"`
	// AWSBilling represents a datastore which points to a pre-existing S3
	// bucket.
	AWSBilling *AWSBillingDataSource `json:"awsBilling"`
}

type AWSBillingDataSource struct {
	Source *S3Bucket `json:"source"`
}

type PromsumDataSource struct {
	Queries []string          `json:"queries"`
	Storage *DataStoreStorage `json:"storage"`
}

type DataStoreStorage struct {
	S3    *S3Bucket     `json:"s3"`
	Local *LocalStorage `json:"local"`
}

type LocalStorage struct {
}
