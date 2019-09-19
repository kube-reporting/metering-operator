package v1

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var MeteringCon***REMOVED***gGVK = SchemeGroupVersion.WithKind("MeteringCon***REMOVED***g")

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type MeteringCon***REMOVED***gList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`
	Items         []*MeteringCon***REMOVED***g `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type MeteringCon***REMOVED***g struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`

	Spec   MeteringCon***REMOVED***gSpec   `json:"spec"`
	Status MeteringCon***REMOVED***gStatus `json:"status"`
}

type MeteringCon***REMOVED***gStatus struct {
	DisableOCPFeatures *bool `json:"disableOCPFeatures,omitempty"`
	LogHelmTemplate    *bool `json:"logHelmTemplate,omitempty"`

	Storage             *StorageCon***REMOVED***g             `json:"storage,omitempty"`
	UnsupportedFeatures *UnsupportedFeaturesCon***REMOVED***g `json:"unsupportedFeatures,omitempty"`
	TLS                 *MeteringCon***REMOVED***gTLSCon***REMOVED***g   `json:"tls,omitempty"`
	Monitoring          *MonitoringCon***REMOVED***g          `json:"monitoring,omitempty"`
	Permissions         *MeteringPermissionCon***REMOVED***g  `json:"permissions,omitempty"`
	OpenshiftReporting  *OpenshiftReportingCon***REMOVED***g  `json:"openshift-reporting,omitempty"`

	Ghostunnel        *Ghostunnel        `json:"__ghostunnel,omitempty"`
	Hive              *Hive              `json:"hive,omitempty"`
	Hadoop            *Hadoop            `json:"hadoop,omitempty"`
	Presto            *Presto            `json:"presto,omitempty"`
	ReportingOperator *ReportingOperator `json:"reporting-operator,omitempty"`
}

type MeteringCon***REMOVED***gSpec struct {
	DisableOCPFeatures *bool `json:"disableOCPFeatures,omitempty"`
	LogHelmTemplate    *bool `json:"logHelmTemplate,omitempty"`

	Storage             *StorageCon***REMOVED***g             `json:"storage,omitempty"`
	UnsupportedFeatures *UnsupportedFeaturesCon***REMOVED***g `json:"unsupportedFeatures,omitempty"`
	TLS                 *MeteringCon***REMOVED***gTLSCon***REMOVED***g   `json:"tls,omitempty"`
	Monitoring          *MonitoringCon***REMOVED***g          `json:"monitoring,omitempty"`
	Permissions         *MeteringPermissionCon***REMOVED***g  `json:"permissions,omitempty"`
	OpenshiftReporting  *OpenshiftReportingCon***REMOVED***g  `json:"openshift-reporting,omitempty"`

	Ghostunnel        *Ghostunnel        `json:"__ghostunnel,omitempty"`
	Hive              *Hive              `json:"hive,omitempty"`
	Hadoop            *Hadoop            `json:"hadoop,omitempty"`
	Presto            *Presto            `json:"presto,omitempty"`
	ReportingOperator *ReportingOperator `json:"reporting-operator,omitempty"`
}

/*
Start of structures re-used throughout the top-level keys
*/

type ImageCon***REMOVED***g struct {
	PullPolicy string `json:"pullPolicy,omitempty"`
	Repository string `json:"repository,omitempty"`
	Tag        string `json:"tag,omitempty"`
}
type TLSCon***REMOVED***g struct {
	Enabled       *bool  `json:"enabled,omitempty"`
	CreateSecret  *bool  `json:"createSecret,omitempty"`
	Certi***REMOVED***cate   string `json:"certi***REMOVED***cate,omitempty"`
	Key           string `json:"key,omitempty"`
	CaCerti***REMOVED***cate string `json:"caCerti***REMOVED***cate,omitempty"`
	SecretName    string `json:"secretName,omitempty"`
}
type JVMCon***REMOVED***g struct {
	G1HeapRegionSize               int    `json:"G1HeapRegionSize,omitempty"`
	ConcGCThreads                  int    `json:"concGCThreads,omitempty"`
	InitiatingHeapOccupancyPercent int    `json:"initiatingHeapOccupancyPercent,omitempty"`
	MaxGcPauseMillis               int    `json:"maxGcPauseMillis,omitempty"`
	ParallelGCThreads              int    `json:"parallelGCThreads,omitempty"`
	InitialRAMPercentage           int    `json:"initialRAMPercentage,omitempty"`
	MaxRAMPercentage               int    `json:"maxRAMPercentage,omitempty"`
	MinRAMPercentage               int    `json:"minRAMPercentage,omitempty"`
	MaxCachedBufferSize            int    `json:"maxCachedBufferSize,omitempty"`
	MaxDirectMemorySize            int    `json:"maxDirectMemorySize,omitempty"`
	PermSize                       string `json:"permSize,omitempty"`
	ReservedCodeCacheSize          string `json:"reservedCodeCacheSize,omitempty"`
}

/*
End of structures re-used throughout the top-level keys
*/

type ReportingOperator struct {
	Spec *ReportingOperatorSpec `json:"spec,omitempty"`
}
type ReportingOperatorSpec struct {
	Replicas       *int32                             `json:"replicas,omitempty"`
	Annotations    map[string]string                  `json:"annotations,omitempty"`
	Labels         map[string]string                  `json:"labels,omitempty"`
	NodeSelector   map[string]string                  `json:"nodeSelector,omitempty"`
	Af***REMOVED***nity       *corev1.Af***REMOVED***nity                   `json:"af***REMOVED***nity,omitempty"`
	Resources      *corev1.ResourceRequirements       `json:"resources,omitempty"`
	UpdateStrategy *appsv1.DeploymentStrategy         `json:"updateStrategy,omitempty"`
	Image          *ImageCon***REMOVED***g                       `json:"image,omitempty"`
	Con***REMOVED***g         *ReportingOperatorCon***REMOVED***g           `json:"con***REMOVED***g,omitempty"`
	APIService     *ReportingOperatorAPIServiceCon***REMOVED***g `json:"apiService,omitempty"`
	Route          *ReportingOperatorRouteCon***REMOVED***g      `json:"route,omitempty"`
	AuthProxy      *ReportingOperatorAuthProxyCon***REMOVED***g  `json:"authProxy,omitempty"`
}
type ReportingOperatorCon***REMOVED***g struct {
	AllNamespaces       *bool                              `json:"allNamespaces,omitempty"`
	EnableFinalizers    *bool                              `json:"enableFinalizers,omitempty"`
	LogDDLQueries       *bool                              `json:"logDDLQueries,omitempty"`
	LogDMLQueries       *bool                              `json:"logDMLQueries,omitempty"`
	LogReports          *bool                              `json:"logReports,omitempty"`
	LogLevel            string                             `json:"logLevel,omitempty"`
	LeaderLeaseDuration *meta.Duration                     `json:"leaderLeaseDuration,omitempty"`
	AWS                 *AWSCon***REMOVED***g                         `json:"aws,omitempty"`
	Prometheus          *ReportingOperatorPrometheusCon***REMOVED***g `json:"prometheus,omitempty"`
	Hive                *ReportingOperatorHiveCon***REMOVED***g       `json:"hive,omitempty"`
	Presto              *ReportingOperatorPrestoCon***REMOVED***g     `json:"presto,omitempty"`
	TLS                 *ReportingOperatorTLSCon***REMOVED***g        `json:"tls,omitempty"`
}

type ReportingOperatorAuthProxyCon***REMOVED***g struct {
	Enabled             *bool                                       `json:"enabled,omitempty"`
	Resources           *corev1.ResourceRequirements                `json:"resources,omitempty"`
	Image               *ImageCon***REMOVED***g                                `json:"image,omitempty"`
	AuthenticatedEmails *ReportingOperatorAuthenticatedEmailCon***REMOVED***g  `json:"authenticatedEmails,omitempty"`
	Cookie              *ReportingOperatorCookieCon***REMOVED***g              `json:"cookie,omitempty"`
	DelegateURLs        *ReportingOperatorDelegateURLCon***REMOVED***g         `json:"delegateURLs,omitempty"`
	SubjectAccessReview *ReportingOperatorSubjectAccessReviewCon***REMOVED***g `json:"subjectAccessReview,omitempty"`
	Htpasswd            *ReportingOperatorHtpasswdCon***REMOVED***g            `json:"htpasswd,omitempty"`
	Rbac                *ReportingOperatorRBACCon***REMOVED***g                `json:"rbac,omitempty"`
}
type ReportingOperatorAuthenticatedEmailCon***REMOVED***g struct {
	Enabled      *bool  `json:"enabled,omitempty"`
	CreateSecret *bool  `json:"createSecret,omitempty"`
	Data         string `json:"data,omitempty"`
	SecretName   string `json:"secretName,omitempty"`
}
type ReportingOperatorCookieCon***REMOVED***g struct {
	CreateSecret *bool  `json:"createSecret,omitempty"`
	Seed         string `json:"seed,omitempty"`
	SecretName   string `json:"secretName,omitempty"`
}
type ReportingOperatorDelegateURLCon***REMOVED***g struct {
	Enabled *bool  `json:"enabled,omitempty"`
	Policy  string `json:"policy,omitempty"`
}
type ReportingOperatorSubjectAccessReviewCon***REMOVED***g struct {
	Enabled *bool  `json:"enabled,omitempty"`
	Policy  string `json:"policy,omitempty"`
}
type ReportingOperatorHtpasswdCon***REMOVED***g struct {
	CreateSecret *bool  `json:"createSecret,omitempty"`
	Data         string `json:"data,omitempty"`
	SecretName   string `json:"secretName,omitempty"`
}

type ReportingOperatorRBACCon***REMOVED***g struct {
	CreateAuthProxyClusterRole *bool `json:"createAuthProxyClusterRole,omitempty"`
}

type ReportingOperatorPrestoCon***REMOVED***g struct {
	MaxQueryLength int                                `json:"maxQueryLength,omitempty"`
	Host           string                             `json:"host,omitempty"`
	TLS            *ReportingOperatorCon***REMOVED***gTLSCon***REMOVED***g  `json:"tls,omitempty"`
	Auth           *ReportingOperatorCon***REMOVED***gAuthCon***REMOVED***g `json:"auth,omitempty"`
}
type ReportingOperatorHiveCon***REMOVED***g struct {
	Host string                             `json:"host,omitempty"`
	TLS  *ReportingOperatorCon***REMOVED***gTLSCon***REMOVED***g  `json:"tls,omitempty"`
	Auth *ReportingOperatorCon***REMOVED***gAuthCon***REMOVED***g `json:"auth,omitempty"`
}

// ReportingOperatorCon***REMOVED***gTLSCon***REMOVED***g contains TLS-related ***REMOVED***elds for Presto/Hive
type ReportingOperatorCon***REMOVED***gTLSCon***REMOVED***g struct {
	CaCerti***REMOVED***cate string `json:"caCerti***REMOVED***cate,omitempty"`
	CreateSecret  *bool  `json:"createSecret,omitempty"`
	Enabled       *bool  `json:"enabled,omitempty"`
	SecretName    string `json:"secretName,omitempty"`
}

//ReportingOperatorCon***REMOVED***gAuthCon***REMOVED***g contains auth-related ***REMOVED***elds for Presto/Hive
type ReportingOperatorCon***REMOVED***gAuthCon***REMOVED***g struct {
	Certi***REMOVED***cate  string `json:"certi***REMOVED***cate,omitempty"`
	Key          string `json:"key,omitempty"`
	CreateSecret *bool  `json:"createSecret,omitempty"`
	Enabled      *bool  `json:"enabled,omitempty"`
	SecretName   string `json:"secretName,omitempty"`
}

type ReportingOperatorPrometheusCon***REMOVED***g struct {
	URL                  string                                                 `json:"url,omitempty"`
	Certi***REMOVED***cateAuthority *ReportingOperatorPrometheusCerti***REMOVED***cateAuthorityCon***REMOVED***g `json:"certi***REMOVED***cateAuthority,omitempty"`
	MetricsImporter      *ReportingOperatorPrometheusMetricsImporterCon***REMOVED***g      `json:"metricsImporter,omitempty"`
}
type ReportingOperatorPrometheusMetricsImporterCon***REMOVED***g struct {
	Enabled *bool                                                 `json:"enabled,omitempty"`
	Con***REMOVED***g  *ReportingOperatorPrometheusMetricsImporterCon***REMOVED***gSpec `json:"con***REMOVED***g,omitempty"`
	Auth    *ReportingOperatorPrometheusAuthCon***REMOVED***g                `json:"auth,omitempty"`
}
type ReportingOperatorPrometheusMetricsImporterCon***REMOVED***gSpec struct {
	ChunkSize                 *meta.Duration `json:"chunkSize,omitempty"`
	PollInterval              *meta.Duration `json:"pollInterval,omitempty"`
	StepSize                  *meta.Duration `json:"stepSize,omitempty"`
	ImportFrom                *meta.Time     `json:"importFrom,omitempty"`
	MaxImportBack***REMOVED***llDuration *meta.Duration `json:"maxImportBack***REMOVED***llDuration,omitempty"`
	MaxQueryRangeDuration     string         `json:"maxQueryRangeDuration,omitempty"`
}
type ReportingOperatorPrometheusCon***REMOVED***gMapCon***REMOVED***g struct {
	Create   *bool  `json:"create,omitempty"`
	Enabled  *bool  `json:"enabled,omitempty"`
	Filename string `json:"***REMOVED***lename,omitempty"`
	Name     string `json:"name,omitempty"`
	Value    string `json:"value,omitempty"`
}
type ReportingOperatorPrometheusCerti***REMOVED***cateAuthorityCon***REMOVED***g struct {
	UseServiceAccountCA *bool                                       `json:"useServiceAccountCA,omitempty"`
	Con***REMOVED***gMap           *ReportingOperatorPrometheusCon***REMOVED***gMapCon***REMOVED***g `json:"con***REMOVED***gMap,omitempty"`
}
type ReportingOperatorPrometheusTokenSecretCon***REMOVED***g struct {
	Create  *bool  `json:"create,omitempty"`
	Enabled *bool  `json:"enabled,omitempty"`
	Name    string `json:"name,omitempty"`
	Value   string `json:"value,omitempty"`
}
type ReportingOperatorPrometheusAuthCon***REMOVED***g struct {
	UseServiceAccountToken *bool                                         `json:"useServiceAccountToken,omitempty"`
	TokenSecret            *ReportingOperatorPrometheusTokenSecretCon***REMOVED***g `json:"tokenSecret,omitempty"`
}
type ReportingOperatorTLSCon***REMOVED***g struct {
	API *TLSCon***REMOVED***g `json:"api,omitempty"`
}
type ReportingOperatorAPIServiceCon***REMOVED***g struct {
	Annotations map[string]string `json:"annotations,omitempty"`
	NodePort    string            `json:"nodePort,omitempty"`
	Type        string            `json:"type,omitempty"`
}
type ReportingOperatorRouteCon***REMOVED***g struct {
	Enabled *bool  `json:"enabled,omitempty"`
	Name    string `json:"name,omitempty"`
}

/*
End of Reporting Operator section
*/

type MonitoringCon***REMOVED***g struct {
	CreateRBAC *bool  `json:"createRBAC,omitempty"`
	Enabled    *bool  `json:"enabled,omitempty"`
	Namespace  string `json:"namespace,omitempty"`
}

type MeteringPermissionCon***REMOVED***g struct {
	MeteringAdmins  []MeteringPermissionCon***REMOVED***gSpec `json:"meteringAdmins,omitempty"`
	MeteringViewers []MeteringPermissionCon***REMOVED***gSpec `json:"meteringViewers,omitempty"`
	ReportExporters []MeteringPermissionCon***REMOVED***gSpec `json:"reportExporters,omitempty"`
	ReportAdmins    []MeteringPermissionCon***REMOVED***gSpec `json:"reportingAdmins,omitempty"`
	ReportViewers   []MeteringPermissionCon***REMOVED***gSpec `json:"reportingViewers,omitempty"`
}
type MeteringPermissionCon***REMOVED***gSpec struct {
	Kind      string `json:"kind,omitempty"`
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	APIGroup  string `json:"apiGroup,omitempty"`
}

type OpenshiftReportingCon***REMOVED***g struct {
	Spec *OpenshiftReportingCon***REMOVED***gSpec `json:"spec,omitempty"`
}
type OpenshiftReportingCon***REMOVED***gSpec struct {
	OpenshiftReportingDefaultStorageLocation     *OpenshiftReportingDefaultStorageLocationCon***REMOVED***g     `json:"defaultStorageLocation,omitempty"`
	OpenshiftReportingAWSBillingReportDataSource *OpenshiftReportingAWSBillingReportDataSourceCon***REMOVED***g `json:"awsBillingReportDataSource,omitempty"`
	OpenshiftReportingDefaultReportDataSources   *OpenshiftReportingDefaultReportDataSourcesCon***REMOVED***g   `json:"defaultReportDataSources,omitempty"`
}
type OpenshiftReportingDefaultStorageLocationCon***REMOVED***g struct {
	Enabled   *bool                                  `json:"enabled,omitempty"`
	IsDefault *bool                                  `json:"isDefault,omitempty"`
	Name      string                                 `json:"name,omitempty"`
	Type      string                                 `json:"type,omitempty"`
	Hive      *OpenshiftReportingHiveStorageLocation `json:"hive,omitempty"`
}
type OpenshiftReportingHiveStorageLocation struct {
	UnmanagedDatabase *bool  `json:"unmanagedDatabase,omitempty"`
	DatabaseName      string `json:"databaseName,omitempty"`
	Location          string `json:"location,omitempty"`
}
type OpenshiftReportingAWSBillingReportDataSourceCon***REMOVED***g struct {
	Enabled *bool  `json:"enabled,omitempty"`
	Bucket  string `json:"bucket,omitempty"`
	Pre***REMOVED***x  string `json:"pre***REMOVED***x,omitempty"`
	Region  string `json:"region,omitempty"`
}
type OpenshiftReportingDefaultReportDataSourcesCon***REMOVED***g struct {
	Base            *OpenshiftReportingDefaultReportDataSourcesBaseCon***REMOVED***g `json:"base,omitempty"`
	PostKubeVersion *OpenshiftReportingPostKubeVersionCon***REMOVED***g              `json:"postKube_1_14,omitempty"`
}
type OpenshiftReportingPostKubeVersionCon***REMOVED***g struct {
	Enabled *bool `json:"enabled,omitempty"`
}
type OpenshiftReportingDefaultReportDataSourcesBaseCon***REMOVED***g struct {
	Enabled *bool                                 `json:"enabled,omitempty"`
	Items   []OpenshiftReportingReportQueryCon***REMOVED***g `json:"items,omitempty"`
}
type OpenshiftReportingReportQueryCon***REMOVED***g struct {
	Name string                                   `json:"name,omitempty"`
	Spec *OpenshiftReportingReportQueryCon***REMOVED***gSpec `json:"spec,omitempty"`
}
type OpenshiftReportingReportQueryCon***REMOVED***gSpec struct {
	OpenshiftReportingReportQueryView *OpenshiftReportingReportQueryView `json:"reportQueryView,omitempty"`
}
type OpenshiftReportingReportQueryView struct {
	QueryName string `json:"queryName,omitempty"`
}

type MeteringCon***REMOVED***gTLSCon***REMOVED***g struct {
	Enabled     *bool  `json:"enabled,omitempty"`
	Certi***REMOVED***cate string `json:"certi***REMOVED***cate,omitempty"`
	Key         string `json:"key,omitempty"`
	SecretName  string `json:"secretName,omitempty"`
}

type UnsupportedFeaturesCon***REMOVED***g struct {
	EnableHDFS *bool `json:"enableHDFS,omitempty"`
}

type Ghostunnel struct {
	Image *ImageCon***REMOVED***g `json:"image,omitempty"`
}

type StorageCon***REMOVED***g struct {
	Type string             `json:"type,omitempty"`
	Hive *HiveStorageCon***REMOVED***g `json:"hive,omitempty"`
}
type HiveStorageCon***REMOVED***g struct {
	Type         string              `json:"type,omitempty"`
	Azure        *AzureCon***REMOVED***g        `json:"azure,omitempty"`
	Gcs          *GCSCon***REMOVED***g          `json:"gcs,omitempty"`
	Hdfs         *HiveHDFSCon***REMOVED***g     `json:"hdfs,omitempty"`
	S3           *S3Con***REMOVED***g           `json:"s3,omitempty"`
	S3Compatible *S3CompatibleCon***REMOVED***g `json:"s3Compatible,omitempty"`
	SharedPVC    *SharedPVCCon***REMOVED***g    `json:"sharedPVC,omitempty"`
}
type AzureCon***REMOVED***g struct {
	CreateSecret       *bool  `json:"createSecret,omitempty"`
	Container          string `json:"container,omitempty"`
	RootDirectory      string `json:"rootDirectory,omitempty"`
	SecretAccessKey    string `json:"secretAccessKey,omitempty"`
	SecretName         string `json:"secretName,omitempty"`
	StorageAccountName string `json:"storageAccountName,omitempty"`
}
type GCSCon***REMOVED***g struct {
	CreateSecret          *bool  `json:"createSecret,omitempty"`
	Bucket                string `json:"bucket,omitempty"`
	SecretName            string `json:"secretName,omitempty"`
	ServiceAccountKeyJSON string `json:"serviceAccountKeyJSON,omitempty"`
}
type HiveHDFSCon***REMOVED***g struct {
	Namenode string `json:"namenode,omitempty"`
}
type S3Con***REMOVED***g struct {
	CreateBucket *bool  `json:"createBucket,omitempty"`
	Bucket       string `json:"bucket,omitempty"`
	Region       string `json:"region,omitempty"`
	SecretName   string `json:"secretName,omitempty"`
}
type AWSCon***REMOVED***g struct {
	CreateSecret    *bool  `json:"createSecret,omitempty"`
	AccessKeyID     string `json:"accessKeyID,omitempty"`
	SecretAccessKey string `json:"secretAccessKey,omitempty"`
	SecretName      string `json:"secretName,omitempty"`
}
type S3CompatibleCon***REMOVED***g struct {
	CreateSecret    *bool  `json:"createSecret,omitempty"`
	AccessKeyID     string `json:"accessKeyID,omitempty"`
	Bucket          string `json:"bucket,omitempty"`
	Endpoint        string `json:"endpoint,omitempty"`
	SecretAccessKey string `json:"secretAccessKey,omitempty"`
	SecretName      string `json:"secretName,omitempty"`
}
type SharedPVCCon***REMOVED***g struct {
	CreatePVC    *bool  `json:"createPVC,omitempty"`
	ClaimName    string `json:"claimName,omitempty"`
	MountPath    string `json:"mountPath,omitempty"`
	Size         string `json:"size,omitempty"`
	StorageClass string `json:"storageClass,omitempty"`
}

/*
End of storage section
*/

type Presto struct {
	Spec *PrestoSpec `json:"spec,omitempty"`
}
type PrestoSpec struct {
	Labels          map[string]string       `json:"labels,omitempty"`
	SecurityContext *corev1.SecurityContext `json:"securityContext,omitempty"`
	Image           *ImageCon***REMOVED***g            `json:"image,omitempty"`
	Con***REMOVED***g          *PrestoCon***REMOVED***g           `json:"con***REMOVED***g,omitempty"`
	Coordinator     *PrestoCoordinatorSpec  `json:"coordinator,omitempty"`
	Worker          *PrestoWorkerSpec       `json:"worker,omitempty"`
}
type PrestoCon***REMOVED***g struct {
	NodeSchedulerIncludeCoordinator *bool                  `json:"nodeSchedulerIncludeCoordinator,omitempty"`
	Environment                     string                 `json:"environment,omitempty"`
	MaxQueryLength                  string                 `json:"maxQueryLength,omitempty"`
	AWS                             *AWSCon***REMOVED***g             `json:"aws,omitempty"`
	Azure                           *AzureCon***REMOVED***g           `json:"azure,omitempty"`
	Gcs                             *GCSCon***REMOVED***g             `json:"gcs,omitempty"`
	S3Compatible                    *S3CompatibleCon***REMOVED***g    `json:"s3Compatible,omitempty"`
	TLS                             *TLSCon***REMOVED***g             `json:"tls,omitempty"`
	Auth                            *TLSCon***REMOVED***g             `json:"auth,omitempty"`
	Connectors                      *PrestoConnectorCon***REMOVED***g `json:"connectors,omitempty"`
}
type PrestoCoordinatorSpec struct {
	TerminationGracePeriodSeconds *int64                       `json:"terminationGracePeriodSeconds,omitempty"`
	NodeSelector                  map[string]string            `json:"nodeSelector,omitempty"`
	Af***REMOVED***nity                      *corev1.Af***REMOVED***nity             `json:"af***REMOVED***nity,omitempty"`
	Resources                     *corev1.ResourceRequirements `json:"resources,omitempty"`
	Tolerations                   []corev1.Toleration          `json:"tolerations,omitempty"`
	Con***REMOVED***g                        *PrestoServerCon***REMOVED***g          `json:"con***REMOVED***g,omitempty"`
}
type PrestoWorkerSpec struct {
	Replicas                      *int32                       `json:"replicas,omitempty"`
	TerminationGracePeriodSeconds *int64                       `json:"terminationGracePeriodSeconds,omitempty"`
	NodeSelector                  map[string]string            `json:"nodeSelector,omitempty"`
	Af***REMOVED***nity                      *corev1.Af***REMOVED***nity             `json:"af***REMOVED***nity,omitempty"`
	Resources                     *corev1.ResourceRequirements `json:"resources,omitempty"`
	Tolerations                   []corev1.Toleration          `json:"tolerations,omitempty"`
	Con***REMOVED***g                        *PrestoServerCon***REMOVED***g          `json:"con***REMOVED***g,omitempty"`
}
type PrestoConnectorCon***REMOVED***g struct {
	Hive              *PrestoConnectorHiveCon***REMOVED***g `json:"hive,omitempty"`
	ConnectorFileList *PrestoConnectorFileList   `json:"extraConnectorFiles,omitempty"`
}
type PrestoConnectorHiveCon***REMOVED***g struct {
	UseHadoopCon***REMOVED***g        *bool      `json:"useHadoopCon***REMOVED***g,omitempty"`
	HadoopCon***REMOVED***gSecretName string     `json:"hadoopCon***REMOVED***gSecretName,omitempty"`
	MetastoreURI           string     `json:"metastoreURI,omitempty"`
	MetastoreTimeout       string     `json:"metastoreTimeout,omitempty"`
	TLS                    *TLSCon***REMOVED***g `json:"tls,omitempty"`
}
type PrestoConnectorFileList struct {
	Name    string `json:"name,omitempty"`
	Content string `json:"content,omitempty"`
}

// PrestoServerCon***REMOVED***g handles the con***REMOVED***guration of the Presto coordinator/worker
type PrestoServerCon***REMOVED***g struct {
	TaskMaxWorkerThreads int        `json:"taskMaxWorkerThreads,omitempty"`
	TaskMinDrivers       int        `json:"taskMinDrivers,omitempty"`
	LogLevel             string     `json:"logLevel,omitempty"`
	Jvm                  *JVMCon***REMOVED***g `json:"jvm,omitempty"`
}

/*
End of Presto section
*/

type Hive struct {
	Spec HiveSpec `json:"spec,omitempty"`
}
type HiveSpec struct {
	TerminationGracePeriodSeconds *int64                  `json:"terminationGracePeriodSeconds,omitempty"`
	Labels                        map[string]string       `json:"labels,omitempty"`
	Annotations                   map[string]string       `json:"annotations,omitempty"`
	SecurityContext               *corev1.SecurityContext `json:"securityContext,omitempty"`
	Image                         *ImageCon***REMOVED***g            `json:"image,omitempty"`
	Con***REMOVED***g                        *HiveSpecCon***REMOVED***g         `json:"con***REMOVED***g,omitempty"`
	Metastore                     *HiveMetastoreSpec      `json:"metastore,omitempty"`
	Server                        *HiveServerSpec         `json:"server,omitempty"`
}
type HiveSpecCon***REMOVED***g struct {
	UseHadoopCon***REMOVED***g              *bool                   `json:"useHadoopCon***REMOVED***g,omitempty"`
	MetastoreClientSocketTimeout string                  `json:"metastoreClientSocketTimeout,omitempty"`
	MetastoreWarehouseDir        string                  `json:"metastoreWarehouseDir,omitempty"`
	DefaultCompression           string                  `json:"defaultCompression,omitempty"`
	DefaultFileFormat            string                  `json:"defaultFileFormat,omitempty"`
	HadoopCon***REMOVED***gSecretName       string                  `json:"hadoopCon***REMOVED***gSecretName,omitempty"`
	AWS                          *AWSCon***REMOVED***g              `json:"aws,omitempty"`
	Azure                        *AzureCon***REMOVED***g            `json:"azure,omitempty"`
	Gcs                          *GCSCon***REMOVED***g              `json:"gcs,omitempty"`
	S3Compatible                 *S3CompatibleCon***REMOVED***g     `json:"s3Compatible,omitempty"`
	DB                           *HiveDBCon***REMOVED***g           `json:"db,omitempty"`
	SharedVolume                 *HiveSharedVolumeCon***REMOVED***g `json:"sharedVolume,omitempty"`
}
type HiveDBCon***REMOVED***g struct {
	AutoCreateMetastoreSchema         *bool  `json:"autoCreateMetastoreSchema,omitempty"`
	EnableMetastoreSchemaVeri***REMOVED***cation *bool  `json:"enableMetastoreSchemaVeri***REMOVED***cation,omitempty"`
	Driver                            string `json:"driver,omitempty"`
	Password                          string `json:"password,omitempty"`
	URL                               string `json:"url,omitempty"`
	Username                          string `json:"username,omitempty"`
}
type HiveSharedVolumeCon***REMOVED***g struct {
	CreatePVC    *bool  `json:"createPVC,omitempty"`
	Enabled      *bool  `json:"enabled,omitempty"`
	ClaimName    string `json:"claimName,omitempty"`
	MountPath    string `json:"mountPath,omitempty"`
	Size         string `json:"size,omitempty"`
	StorageClass string `json:"storageClass,omitempty"`
}
type HiveMetastoreSpec struct {
	NodeSelector   map[string]string            `json:"nodeSelector,omitempty"`
	Af***REMOVED***nity       *corev1.Af***REMOVED***nity             `json:"af***REMOVED***nity,omitempty"`
	LivenessProbe  *corev1.Probe                `json:"livenessProbe,omitempty"`
	ReadinessProbe *corev1.Probe                `json:"readinessProbe,omitempty"`
	Resources      *corev1.ResourceRequirements `json:"resources,omitempty"`
	Tolerations    []corev1.Toleration          `json:"tolerations,omitempty"`
	Con***REMOVED***g         *HiveMetastoreSpecCon***REMOVED***g     `json:"con***REMOVED***g,omitempty"`
	Storage        *HiveMetastoreStorageCon***REMOVED***g  `json:"storage,omitempty"`
}
type HiveMetastoreSpecCon***REMOVED***g struct {
	LogLevel string                  `json:"logLevel,omitempty"`
	Jvm      *JVMCon***REMOVED***g              `json:"jvm,omitempty"`
	TLS      *TLSCon***REMOVED***g              `json:"tls,omitempty"`
	Auth     *HiveResourceAuthCon***REMOVED***g `json:"auth,omitempty"`
}
type HiveResourceAuthCon***REMOVED***g struct {
	Enabled *bool `json:"enabled,omitempty"`
}
type HiveMetastoreStorageCon***REMOVED***g struct {
	Create *bool  `json:"create,omitempty"`
	Class  string `json:"class,omitempty"`
	Size   string `json:"size,omitempty"`
}
type HiveServerSpec struct {
	NodeSelector   map[string]string            `json:"nodeSelector,omitempty"`
	Af***REMOVED***nity       *corev1.Af***REMOVED***nity             `json:"af***REMOVED***nity,omitempty"`
	LivenessProbe  *corev1.Probe                `json:"livenessProbe,omitempty"`
	ReadinessProbe *corev1.Probe                `json:"readinessProbe,omitempty"`
	Resources      *corev1.ResourceRequirements `json:"resources,omitempty"`
	Tolerations    []corev1.Toleration          `json:"tolerations,omitempty"`
	Con***REMOVED***g         *HiveServerSpecCon***REMOVED***g        `json:"con***REMOVED***g,omitempty"`
}
type HiveServerSpecCon***REMOVED***g struct {
	LogLevel     string                  `json:"logLevel,omitempty"`
	Jvm          *JVMCon***REMOVED***g              `json:"jvm,omitempty"`
	TLS          *TLSCon***REMOVED***g              `json:"tls,omitempty"`
	MetastoreTLS *TLSCon***REMOVED***g              `json:"metastoreTLS,omitempty"`
	Auth         *HiveResourceAuthCon***REMOVED***g `json:"auth,omitempty"`
}

/*
End of Hive section
*/

type Hadoop struct {
	Spec *HadoopSpec `json:"spec,omitempty"`
}
type HadoopSpec struct {
	Con***REMOVED***gSecretName string            `json:"con***REMOVED***gSecretName,omitempty"`
	Image            *ImageCon***REMOVED***g      `json:"image,omitempty"`
	HDFS             *HadoopHDFS       `json:"hdfs,omitempty"`
	Con***REMOVED***g           *HadoopSpecCon***REMOVED***g `json:"con***REMOVED***g,omitempty"`
}
type HadoopHDFS struct {
	Enabled         *bool                   `json:"enabled,omitempty"`
	SecurityContext *corev1.SecurityContext `json:"securityContext,omitempty"`
	Con***REMOVED***g          *HadoopHDFSCon***REMOVED***g       `json:"con***REMOVED***g,omitempty"`
	Datanode        *HadoopHDFSDatanodeSpec `json:"datanode,omitempty"`
	Namenode        *HadoopHDFSNamenodeSpec `json:"namenode,omitempty"`
}
type HadoopSpecCon***REMOVED***g struct {
	DefaultFS    string              `json:"defaultFS,omitempty"`
	AWS          *AWSCon***REMOVED***g          `json:"aws,omitempty"`
	Azure        *AzureCon***REMOVED***g        `json:"azure,omitempty"`
	Gcs          *GCSCon***REMOVED***g          `json:"gcs,omitempty"`
	S3Compatible *S3CompatibleCon***REMOVED***g `json:"s3Compatible,omitempty"`
}
type HadoopHDFSCon***REMOVED***g struct {
	ReplicationFactor    *int32 `json:"replicationFactor,omitempty"`
	DatanodeDataDirPerms string `json:"datanodeDataDirPerms,omitempty"`
	LogLevel             string `json:"logLevel,omitempty"`
}
type HadoopHDFSDatanodeSpec struct {
	Replicas                      *int32                       `json:"replicas,omitempty"`
	TerminationGracePeriodSeconds *int64                       `json:"terminationGracePeriodSeconds,omitempty"`
	Annotations                   map[string]string            `json:"annotations,omitempty"`
	Labels                        map[string]string            `json:"labels,omitempty"`
	NodeSelector                  map[string]string            `json:"nodeSelector,omitempty"`
	Af***REMOVED***nity                      *corev1.Af***REMOVED***nity             `json:"af***REMOVED***nity,omitempty"`
	Resources                     *corev1.ResourceRequirements `json:"resources,omitempty"`
	Tolerations                   []corev1.Toleration          `json:"tolerations,omitempty"`
	Con***REMOVED***g                        *HadoopHDFSNodeCon***REMOVED***g        `json:"con***REMOVED***g,omitempty"`
	Storage                       *HadoopHDFSStorageCon***REMOVED***g     `json:"storage,omitempty"`
}
type HadoopHDFSStorageCon***REMOVED***g struct {
	Class string `json:"class,omitempty"`
	Size  string `json:"size,omitempty"`
}
type HadoopHDFSNodeCon***REMOVED***g struct {
	Jvm *JVMCon***REMOVED***g `json:"jvm,omitempty"`
}
type HadoopHDFSNamenodeSpec struct {
	Replicas                      *int32                       `json:"replicas,omitempty"`
	TerminationGracePeriodSeconds *int64                       `json:"terminationGracePeriodSeconds,omitempty"`
	Annotations                   map[string]string            `json:"annotations,omitempty"`
	Labels                        map[string]string            `json:"labels,omitempty"`
	NodeSelector                  map[string]string            `json:"nodeSelector,omitempty"`
	Af***REMOVED***nity                      *corev1.Af***REMOVED***nity             `json:"af***REMOVED***nity,omitempty"`
	Resources                     *corev1.ResourceRequirements `json:"resources,omitempty"`
	Tolerations                   []corev1.Toleration          `json:"tolerations,omitempty"`
	Con***REMOVED***g                        *HadoopHDFSNodeCon***REMOVED***g        `json:"con***REMOVED***g,omitempty"`
	Storage                       *HadoopHDFSStorageCon***REMOVED***g     `json:"storage,omitempty"`
}

/*
End of Hadoop Section
*/
