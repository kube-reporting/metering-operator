package v1alpha1

import (
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ReportDataSourceList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`
	Items         []*ReportDataSource `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ReportDataSource struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`

	Spec      ReportDataSourceSpec `json:"spec"`
	TableName string               `json:"tableName"`
}

type ReportDataSourceSpec struct {
	// Prommsum represents a datasource which holds Prometheus metrics
	Promsum *PromsumDataSource `json:"promsum"`
	// AWSBilling represents a datasource which points to a pre-existing S3
	// bucket.
	AWSBilling *AWSBillingDataSource `json:"awsBilling"`
}

type AWSBillingDataSource struct {
	Source *S3Bucket `json:"source"`
}

type PromsumDataSource struct {
	Query   string                            `json:"query"`
	Storage *PromsumDataSourceStorageLocation `json:"storage"`
}

type PromsumDataSourceStorageLocation struct {
	StorageLocationName string               `json:"storageLocationName,omitempty"`
	StorageSpec         *StorageLocationSpec `json:"spec,omitempty"`
}
