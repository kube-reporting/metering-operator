package v1

import (
	v1 "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var ReportDataSourceGVK = SchemeGroupVersion.WithKind("ReportDataSource")

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

	Spec   ReportDataSourceSpec   `json:"spec"`
	Status ReportDataSourceStatus `json:"status"`
}

type ReportDataSourceSpec struct {
	// Prometheus represents a datasource which holds Prometheus metrics
	PrometheusMetricsImporter *PrometheusMetricsImporterDataSource `json:"prometheusMetricsImporter,omitempty"`
	// AWSBilling represents a datasource which points to a pre-existing S3
	// bucket.
	AWSBilling *AWSBillingDataSource `json:"awsBilling,omitempty"`
	// PrestoTable represents a datasource which points to an existing
	// PrestoTable CR.
	PrestoTable *PrestoTableDataSource `json:"prestoTable,omitempty"`

	// ReportQueryView  represents a datasource which creates a Presto
	// view from a ReportQuery
	ReportQueryView *ReportQueryViewDataSource `json:"reportQueryView,omitempty"`
}

type AWSBillingDataSource struct {
	Source       *S3Bucket `json:"source"`
	DatabaseName string    `json:"databaseName,omitempty"`
}

type S3Bucket struct {
	Region string `json:"region"`
	Bucket string `json:"bucket"`
	Pre***REMOVED***x string `json:"pre***REMOVED***x"`
}

type PrometheusQueryCon***REMOVED***g struct {
	QueryInterval *meta.Duration `json:"queryInterval,omitempty"`
	StepSize      *meta.Duration `json:"stepSize,omitempty"`
	ChunkSize     *meta.Duration `json:"chunkSize,omitempty"`
}

type PrometheusConnectionCon***REMOVED***g struct {
	URL string `json:"url,omitempty"`
}

type PrometheusMetricsImporterDataSource struct {
	Query            string                      `json:"query"`
	QueryCon***REMOVED***g      *PrometheusQueryCon***REMOVED***g      `json:"queryCon***REMOVED***g,omitempty"`
	Storage          *StorageLocationRef         `json:"storage,omitempty"`
	PrometheusCon***REMOVED***g *PrometheusConnectionCon***REMOVED***g `json:"prometheusCon***REMOVED***g,omitempty"`
}

type PrestoTableDataSource struct {
	TableRef v1.LocalObjectReference `json:"tableRef"`
}

type ReportQueryViewDataSource struct {
	// QueryName speci***REMOVED***es the ReportQuery to execute when the report
	// runs.
	QueryName string `json:"queryName"`
	// Inputs are the inputs to the ReportQuery
	Inputs  ReportQueryInputValues `json:"inputs,omitempty"`
	Storage *StorageLocationRef    `json:"storage,omitempty"`
}

type ReportDataSourceStatus struct {
	TableRef                      v1.LocalObjectReference        `json:"tableRef"`
	PrometheusMetricsImportStatus *PrometheusMetricsImportStatus `json:"prometheusMetricsImportStatus,omitempty"`
}

type PrometheusMetricsImportStatus struct {
	// LastImportTime is the time the import last import was ran.
	LastImportTime *meta.Time `json:"lastImportTime,omitempty"`

	// ImportDataStartTime is the start of the time ***REMOVED***rst time range queried.
	ImportDataStartTime *meta.Time `json:"importDataStartTime,omitempty"`
	// ImportDataEndTime is the end of the time last time range queried.
	ImportDataEndTime *meta.Time `json:"importDataEndTime,omitempty"`

	// EarliestImportedMetricTime is the timestamp for the earliest metric
	// imported for this ReportDataSource.
	EarliestImportedMetricTime *meta.Time `json:"earliestImportedMetricTime,omitempty"`
	// NewestImportedMetricTime is the timestamp for the newest metric
	// imported for this ReportDataSource.
	NewestImportedMetricTime *meta.Time `json:"newestImportedMetricTime,omitempty"`
}
