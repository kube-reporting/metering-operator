package v1

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var MeteringConfigGVK = SchemeGroupVersion.WithKind("MeteringConfig")

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type MeteringConfigList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`
	Items         []*MeteringConfig `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type MeteringConfig struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`

	Spec   MeteringConfigSpec   `json:"spec"`
	Status MeteringConfigStatus `json:"status"`
}

type MeteringConfigStatus struct {
	DisableOCPFeatures *bool `json:"disableOCPFeatures,omitempty"`
	LogHelmTemplate    *bool `json:"logHelmTemplate,omitempty"`

	Storage             *StorageConfig             `json:"storage,omitempty"`
	UnsupportedFeatures *UnsupportedFeaturesConfig `json:"unsupportedFeatures,omitempty"`
	TLS                 *MeteringConfigTLSConfig   `json:"tls,omitempty"`
	Monitoring          *MonitoringConfig          `json:"monitoring,omitempty"`
	Permissions         *MeteringPermissionConfig  `json:"permissions,omitempty"`
	OpenshiftReporting  *OpenshiftReportingConfig  `json:"openshift-reporting,omitempty"`

	Ghostunnel        *Ghostunnel        `json:"__ghostunnel,omitempty"`
	Hive              *Hive              `json:"hive,omitempty"`
	Hadoop            *Hadoop            `json:"hadoop,omitempty"`
	Presto            *Presto            `json:"presto,omitempty"`
	ReportingOperator *ReportingOperator `json:"reporting-operator,omitempty"`
}

type MeteringConfigSpec struct {
	DisableOCPFeatures *bool `json:"disableOCPFeatures,omitempty"`
	LogHelmTemplate    *bool `json:"logHelmTemplate,omitempty"`

	Storage             *StorageConfig             `json:"storage,omitempty"`
	UnsupportedFeatures *UnsupportedFeaturesConfig `json:"unsupportedFeatures,omitempty"`
	TLS                 *MeteringConfigTLSConfig   `json:"tls,omitempty"`
	Monitoring          *MonitoringConfig          `json:"monitoring,omitempty"`
	Permissions         *MeteringPermissionConfig  `json:"permissions,omitempty"`
	OpenshiftReporting  *OpenshiftReportingConfig  `json:"openshift-reporting,omitempty"`

	Ghostunnel        *Ghostunnel        `json:"__ghostunnel,omitempty"`
	Hive              *Hive              `json:"hive,omitempty"`
	Hadoop            *Hadoop            `json:"hadoop,omitempty"`
	Presto            *Presto            `json:"presto,omitempty"`
	ReportingOperator *ReportingOperator `json:"reporting-operator,omitempty"`
}

/*
Start of structures re-used throughout the top-level keys
*/

type ImageConfig struct {
	PullPolicy string `json:"pullPolicy,omitempty"`
	Repository string `json:"repository,omitempty"`
	Tag        string `json:"tag,omitempty"`
}
type TLSConfig struct {
	Enabled       *bool  `json:"enabled,omitempty"`
	CreateSecret  *bool  `json:"createSecret,omitempty"`
	Certificate   string `json:"certificate,omitempty"`
	Key           string `json:"key,omitempty"`
	CaCertificate string `json:"caCertificate,omitempty"`
	SecretName    string `json:"secretName,omitempty"`
}
type JVMConfig struct {
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

type PodDisruptionBudget struct {
	Enabled      *bool `json:"enabled,omitempty"`
	MinAvailable *bool `json:"minAvailable,omitempty"`
}

type ReportingOperator struct {
	Spec *ReportingOperatorSpec `json:"spec,omitempty"`
}
type ReportingOperatorSpec struct {
	Replicas            *int32                             `json:"replicas,omitempty"`
	Annotations         map[string]string                  `json:"annotations,omitempty"`
	Labels              map[string]string                  `json:"labels,omitempty"`
	NodeSelector        map[string]string                  `json:"nodeSelector,omitempty"`
	PodDisruptionBudget *PodDisruptionBudget               `json:"podDisruptionBudget,omitempty"`
	Affinity            *corev1.Affinity                   `json:"affinity,omitempty"`
	Resources           *corev1.ResourceRequirements       `json:"resources,omitempty"`
	UpdateStrategy      *appsv1.DeploymentStrategy         `json:"updateStrategy,omitempty"`
	Image               *ImageConfig                       `json:"image,omitempty"`
	Config              *ReportingOperatorConfig           `json:"config,omitempty"`
	APIService          *ReportingOperatorAPIServiceConfig `json:"apiService,omitempty"`
	Route               *ReportingOperatorRouteConfig      `json:"route,omitempty"`
	AuthProxy           *ReportingOperatorAuthProxyConfig  `json:"authProxy,omitempty"`
}
type ReportingOperatorConfig struct {
	AllNamespaces       *bool                              `json:"allNamespaces,omitempty"`
	EnableFinalizers    *bool                              `json:"enableFinalizers,omitempty"`
	LogDDLQueries       *bool                              `json:"logDDLQueries,omitempty"`
	LogDMLQueries       *bool                              `json:"logDMLQueries,omitempty"`
	LogReports          *bool                              `json:"logReports,omitempty"`
	LogLevel            string                             `json:"logLevel,omitempty"`
	LeaderLeaseDuration *meta.Duration                     `json:"leaderLeaseDuration,omitempty"`
	AWS                 *AWSConfig                         `json:"aws,omitempty"`
	Prometheus          *ReportingOperatorPrometheusConfig `json:"prometheus,omitempty"`
	Hive                *ReportingOperatorHiveConfig       `json:"hive,omitempty"`
	Presto              *ReportingOperatorPrestoConfig     `json:"presto,omitempty"`
	TLS                 *ReportingOperatorTLSConfig        `json:"tls,omitempty"`
}

type ReportingOperatorAuthProxyConfig struct {
	Enabled             *bool                                       `json:"enabled,omitempty"`
	Resources           *corev1.ResourceRequirements                `json:"resources,omitempty"`
	Image               *ImageConfig                                `json:"image,omitempty"`
	AuthenticatedEmails *ReportingOperatorAuthenticatedEmailConfig  `json:"authenticatedEmails,omitempty"`
	Cookie              *ReportingOperatorCookieConfig              `json:"cookie,omitempty"`
	DelegateURLs        *ReportingOperatorDelegateURLConfig         `json:"delegateURLs,omitempty"`
	SubjectAccessReview *ReportingOperatorSubjectAccessReviewConfig `json:"subjectAccessReview,omitempty"`
	Htpasswd            *ReportingOperatorHtpasswdConfig            `json:"htpasswd,omitempty"`
	Rbac                *ReportingOperatorRBACConfig                `json:"rbac,omitempty"`
}
type ReportingOperatorAuthenticatedEmailConfig struct {
	Enabled      *bool  `json:"enabled,omitempty"`
	CreateSecret *bool  `json:"createSecret,omitempty"`
	Data         string `json:"data,omitempty"`
	SecretName   string `json:"secretName,omitempty"`
}
type ReportingOperatorCookieConfig struct {
	CreateSecret *bool  `json:"createSecret,omitempty"`
	Seed         string `json:"seed,omitempty"`
	SecretName   string `json:"secretName,omitempty"`
}
type ReportingOperatorDelegateURLConfig struct {
	Enabled *bool  `json:"enabled,omitempty"`
	Policy  string `json:"policy,omitempty"`
}
type ReportingOperatorSubjectAccessReviewConfig struct {
	Enabled *bool  `json:"enabled,omitempty"`
	Policy  string `json:"policy,omitempty"`
}
type ReportingOperatorHtpasswdConfig struct {
	CreateSecret *bool  `json:"createSecret,omitempty"`
	Data         string `json:"data,omitempty"`
	SecretName   string `json:"secretName,omitempty"`
}

type ReportingOperatorRBACConfig struct {
	CreateAuthProxyClusterRole *bool `json:"createAuthProxyClusterRole,omitempty"`
}

type ReportingOperatorPrestoConfig struct {
	MaxQueryLength int                                `json:"maxQueryLength,omitempty"`
	Host           string                             `json:"host,omitempty"`
	TLS            *ReportingOperatorConfigTLSConfig  `json:"tls,omitempty"`
	Auth           *ReportingOperatorConfigAuthConfig `json:"auth,omitempty"`
}
type ReportingOperatorHiveConfig struct {
	Host string                             `json:"host,omitempty"`
	TLS  *ReportingOperatorConfigTLSConfig  `json:"tls,omitempty"`
	Auth *ReportingOperatorConfigAuthConfig `json:"auth,omitempty"`
}

// ReportingOperatorConfigTLSConfig contains TLS-related fields for Presto/Hive
type ReportingOperatorConfigTLSConfig struct {
	CaCertificate string `json:"caCertificate,omitempty"`
	CreateSecret  *bool  `json:"createSecret,omitempty"`
	Enabled       *bool  `json:"enabled,omitempty"`
	SecretName    string `json:"secretName,omitempty"`
}

//ReportingOperatorConfigAuthConfig contains auth-related fields for Presto/Hive
type ReportingOperatorConfigAuthConfig struct {
	Certificate  string `json:"certificate,omitempty"`
	Key          string `json:"key,omitempty"`
	CreateSecret *bool  `json:"createSecret,omitempty"`
	Enabled      *bool  `json:"enabled,omitempty"`
	SecretName   string `json:"secretName,omitempty"`
}

type ReportingOperatorPrometheusConfig struct {
	URL                  string                                                 `json:"url,omitempty"`
	CertificateAuthority *ReportingOperatorPrometheusCertificateAuthorityConfig `json:"certificateAuthority,omitempty"`
	MetricsImporter      *ReportingOperatorPrometheusMetricsImporterConfig      `json:"metricsImporter,omitempty"`
}
type ReportingOperatorPrometheusMetricsImporterConfig struct {
	Enabled *bool                                                 `json:"enabled,omitempty"`
	Config  *ReportingOperatorPrometheusMetricsImporterConfigSpec `json:"config,omitempty"`
	Auth    *ReportingOperatorPrometheusAuthConfig                `json:"auth,omitempty"`
}
type ReportingOperatorPrometheusMetricsImporterConfigSpec struct {
	ChunkSize                 *meta.Duration `json:"chunkSize,omitempty"`
	PollInterval              *meta.Duration `json:"pollInterval,omitempty"`
	StepSize                  *meta.Duration `json:"stepSize,omitempty"`
	ImportFrom                *meta.Time     `json:"importFrom,omitempty"`
	MaxImportBackfillDuration *meta.Duration `json:"maxImportBackfillDuration,omitempty"`
	MaxQueryRangeDuration     *meta.Duration `json:"maxQueryRangeDuration,omitempty"`
}
type ReportingOperatorPrometheusConfigMapConfig struct {
	Create   *bool  `json:"create,omitempty"`
	Enabled  *bool  `json:"enabled,omitempty"`
	Filename string `json:"filename,omitempty"`
	Name     string `json:"name,omitempty"`
	Value    string `json:"value,omitempty"`
}
type ReportingOperatorPrometheusCertificateAuthorityConfig struct {
	UseServiceAccountCA *bool                                       `json:"useServiceAccountCA,omitempty"`
	ConfigMap           *ReportingOperatorPrometheusConfigMapConfig `json:"configMap,omitempty"`
}
type ReportingOperatorPrometheusTokenSecretConfig struct {
	Create  *bool  `json:"create,omitempty"`
	Enabled *bool  `json:"enabled,omitempty"`
	Name    string `json:"name,omitempty"`
	Value   string `json:"value,omitempty"`
}
type ReportingOperatorPrometheusAuthConfig struct {
	UseServiceAccountToken *bool                                         `json:"useServiceAccountToken,omitempty"`
	TokenSecret            *ReportingOperatorPrometheusTokenSecretConfig `json:"tokenSecret,omitempty"`
}
type ReportingOperatorTLSConfig struct {
	API *TLSConfig `json:"api,omitempty"`
}
type ReportingOperatorAPIServiceConfig struct {
	Annotations map[string]string `json:"annotations,omitempty"`
	NodePort    string            `json:"nodePort,omitempty"`
	Type        string            `json:"type,omitempty"`
}
type ReportingOperatorRouteConfig struct {
	Enabled *bool  `json:"enabled,omitempty"`
	Name    string `json:"name,omitempty"`
}

/*
End of Reporting Operator section
*/

type MonitoringConfig struct {
	CreateRBAC *bool  `json:"createRBAC,omitempty"`
	Enabled    *bool  `json:"enabled,omitempty"`
	Namespace  string `json:"namespace,omitempty"`
}

type MeteringPermissionConfig struct {
	MeteringAdmins  []MeteringPermissionConfigSpec `json:"meteringAdmins,omitempty"`
	MeteringViewers []MeteringPermissionConfigSpec `json:"meteringViewers,omitempty"`
	ReportExporters []MeteringPermissionConfigSpec `json:"reportExporters,omitempty"`
	ReportAdmins    []MeteringPermissionConfigSpec `json:"reportingAdmins,omitempty"`
	ReportViewers   []MeteringPermissionConfigSpec `json:"reportingViewers,omitempty"`
}
type MeteringPermissionConfigSpec struct {
	Kind      string `json:"kind,omitempty"`
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	APIGroup  string `json:"apiGroup,omitempty"`
}

type OpenshiftReportingConfig struct {
	Spec *OpenshiftReportingConfigSpec `json:"spec,omitempty"`
}
type OpenshiftReportingConfigSpec struct {
	OpenshiftReportingDefaultStorageLocation     *OpenshiftReportingDefaultStorageLocationConfig     `json:"defaultStorageLocation,omitempty"`
	OpenshiftReportingAWSBillingReportDataSource *OpenshiftReportingAWSBillingReportDataSourceConfig `json:"awsBillingReportDataSource,omitempty"`
	OpenshiftReportingDefaultReportDataSources   *OpenshiftReportingDefaultReportDataSourcesConfig   `json:"defaultReportDataSources,omitempty"`
}
type OpenshiftReportingDefaultStorageLocationConfig struct {
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
type OpenshiftReportingAWSBillingReportDataSourceConfig struct {
	Enabled *bool  `json:"enabled,omitempty"`
	Bucket  string `json:"bucket,omitempty"`
	Prefix  string `json:"prefix,omitempty"`
	Region  string `json:"region,omitempty"`
}
type OpenshiftReportingDefaultReportDataSourcesConfig struct {
	Base            *OpenshiftReportingDefaultReportDataSourcesBaseConfig `json:"base,omitempty"`
	PostKubeVersion *OpenshiftReportingPostKubeVersionConfig              `json:"postKube_1_14,omitempty"`
}
type OpenshiftReportingPostKubeVersionConfig struct {
	Enabled *bool `json:"enabled,omitempty"`
}
type OpenshiftReportingDefaultReportDataSourcesBaseConfig struct {
	Enabled *bool                                 `json:"enabled,omitempty"`
	Items   []OpenshiftReportingReportQueryConfig `json:"items,omitempty"`
}
type OpenshiftReportingReportQueryConfig struct {
	Name string                                   `json:"name,omitempty"`
	Spec *OpenshiftReportingReportQueryConfigSpec `json:"spec,omitempty"`
}
type OpenshiftReportingReportQueryConfigSpec struct {
	OpenshiftReportingReportQueryView *OpenshiftReportingReportQueryView `json:"reportQueryView,omitempty"`
}
type OpenshiftReportingReportQueryView struct {
	QueryName string `json:"queryName,omitempty"`
}

type MeteringConfigTLSConfig struct {
	Enabled     *bool  `json:"enabled,omitempty"`
	Certificate string `json:"certificate,omitempty"`
	Key         string `json:"key,omitempty"`
	SecretName  string `json:"secretName,omitempty"`
}

type UnsupportedFeaturesConfig struct {
	EnableHDFS *bool `json:"enableHDFS,omitempty"`
}

type Ghostunnel struct {
	Image *ImageConfig `json:"image,omitempty"`
}

type StorageConfig struct {
	Type string             `json:"type,omitempty"`
	Hive *HiveStorageConfig `json:"hive,omitempty"`
}
type HiveStorageConfig struct {
	Type         string              `json:"type,omitempty"`
	Azure        *AzureConfig        `json:"azure,omitempty"`
	Gcs          *GCSConfig          `json:"gcs,omitempty"`
	Hdfs         *HiveHDFSConfig     `json:"hdfs,omitempty"`
	S3           *S3Config           `json:"s3,omitempty"`
	S3Compatible *S3CompatibleConfig `json:"s3Compatible,omitempty"`
	SharedPVC    *SharedPVCConfig    `json:"sharedPVC,omitempty"`
}
type AzureConfig struct {
	CreateSecret       *bool  `json:"createSecret,omitempty"`
	Container          string `json:"container,omitempty"`
	RootDirectory      string `json:"rootDirectory,omitempty"`
	SecretAccessKey    string `json:"secretAccessKey,omitempty"`
	SecretName         string `json:"secretName,omitempty"`
	StorageAccountName string `json:"storageAccountName,omitempty"`
}
type GCSConfig struct {
	CreateSecret          *bool  `json:"createSecret,omitempty"`
	Bucket                string `json:"bucket,omitempty"`
	SecretName            string `json:"secretName,omitempty"`
	ServiceAccountKeyJSON string `json:"serviceAccountKeyJSON,omitempty"`
}
type HiveHDFSConfig struct {
	Namenode string `json:"namenode,omitempty"`
}
type S3Config struct {
	CreateBucket *bool  `json:"createBucket,omitempty"`
	Bucket       string `json:"bucket,omitempty"`
	Region       string `json:"region,omitempty"`
	SecretName   string `json:"secretName,omitempty"`
}
type AWSConfig struct {
	CreateSecret    *bool  `json:"createSecret,omitempty"`
	AccessKeyID     string `json:"accessKeyID,omitempty"`
	SecretAccessKey string `json:"secretAccessKey,omitempty"`
	SecretName      string `json:"secretName,omitempty"`
}
type S3CompatibleConfig struct {
	CreateSecret    *bool  `json:"createSecret,omitempty"`
	AccessKeyID     string `json:"accessKeyID,omitempty"`
	Bucket          string `json:"bucket,omitempty"`
	Endpoint        string `json:"endpoint,omitempty"`
	SecretAccessKey string `json:"secretAccessKey,omitempty"`
	SecretName      string `json:"secretName,omitempty"`
}
type SharedPVCConfig struct {
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
	Image           *ImageConfig            `json:"image,omitempty"`
	Config          *PrestoConfig           `json:"config,omitempty"`
	Coordinator     *PrestoCoordinatorSpec  `json:"coordinator,omitempty"`
	Worker          *PrestoWorkerSpec       `json:"worker,omitempty"`
}
type PrestoConfig struct {
	NodeSchedulerIncludeCoordinator *bool                  `json:"nodeSchedulerIncludeCoordinator,omitempty"`
	Environment                     string                 `json:"environment,omitempty"`
	MaxQueryLength                  string                 `json:"maxQueryLength,omitempty"`
	AWS                             *AWSConfig             `json:"aws,omitempty"`
	Azure                           *AzureConfig           `json:"azure,omitempty"`
	Gcs                             *GCSConfig             `json:"gcs,omitempty"`
	S3Compatible                    *S3CompatibleConfig    `json:"s3Compatible,omitempty"`
	TLS                             *TLSConfig             `json:"tls,omitempty"`
	Auth                            *TLSConfig             `json:"auth,omitempty"`
	Connectors                      *PrestoConnectorConfig `json:"connectors,omitempty"`
}
type PrestoCoordinatorSpec struct {
	TerminationGracePeriodSeconds *int64                       `json:"terminationGracePeriodSeconds,omitempty"`
	NodeSelector                  map[string]string            `json:"nodeSelector,omitempty"`
	PodDisruptionBudget           *PodDisruptionBudget         `json:"podDisruptionBudget,omitempty"`
	Affinity                      *corev1.Affinity             `json:"affinity,omitempty"`
	Resources                     *corev1.ResourceRequirements `json:"resources,omitempty"`
	Tolerations                   []corev1.Toleration          `json:"tolerations,omitempty"`
	Config                        *PrestoServerConfig          `json:"config,omitempty"`
}
type PrestoWorkerSpec struct {
	Replicas                      *int32                       `json:"replicas,omitempty"`
	TerminationGracePeriodSeconds *int64                       `json:"terminationGracePeriodSeconds,omitempty"`
	NodeSelector                  map[string]string            `json:"nodeSelector,omitempty"`
	PodDisruptionBudget           *PodDisruptionBudget         `json:"podDisruptionBudget,omitempty"`
	Affinity                      *corev1.Affinity             `json:"affinity,omitempty"`
	Resources                     *corev1.ResourceRequirements `json:"resources,omitempty"`
	Tolerations                   []corev1.Toleration          `json:"tolerations,omitempty"`
	Config                        *PrestoServerConfig          `json:"config,omitempty"`
}
type PrestoConnectorConfig struct {
	Hive              *PrestoConnectorHiveConfig       `json:"hive,omitempty"`
	Prometheus        *PrestoConnectorPrometheusConfig `json:"prometheus,omitempty"`
	ConnectorFileList *PrestoConnectorFileList         `json:"extraConnectorFiles,omitempty"`
}
type PrestoConnectorHiveConfig struct {
	UseHadoopConfig        *bool      `json:"useHadoopConfig,omitempty"`
	HadoopConfigSecretName string     `json:"hadoopConfigSecretName,omitempty"`
	MetastoreURI           string     `json:"metastoreURI,omitempty"`
	MetastoreTimeout       string     `json:"metastoreTimeout,omitempty"`
	TLS                    *TLSConfig `json:"tls,omitempty"`
}
type PrestoConnectorPrometheusConfig struct {
	Enabled *bool                                  `json:"enabled,omitempty"`
	Config  *PrestoConnectorPrometheusConfigConfig `json:"config,omitempty"`
	Auth    *PrestoConnectorPrometheusConfigAuth   `json:"auth,omitempty"`
}
type PrestoConnectorPrometheusConfigConfig struct {
	URI                   string         `json:"uri,omitempty"`
	ChunkSize             *meta.Duration `json:"chunkSizeDuration,omitempty"`
	MaxQueryRangeDuration *meta.Duration `json:"maxQueryRangeDuration,omitempty"`
	CacheDuration         *meta.Duration `json:"cacheDuration,omitempty"`
}
type PrestoConnectorPrometheusConfigAuth struct {
	BearerTokenFile        string `json:"bearerTokenFile,omitempty"`
	UseServiceAccountToken *bool  `json:"useServiceAccountToken,omitempty"`
}
type PrestoConnectorFileList struct {
	Name    string `json:"name,omitempty"`
	Content string `json:"content,omitempty"`
}

// PrestoServerConfig handles the configuration of the Presto coordinator/worker
type PrestoServerConfig struct {
	TaskMaxWorkerThreads int        `json:"taskMaxWorkerThreads,omitempty"`
	TaskMinDrivers       int        `json:"taskMinDrivers,omitempty"`
	LogLevel             string     `json:"logLevel,omitempty"`
	Jvm                  *JVMConfig `json:"jvm,omitempty"`
}

/*
End of Presto section
*/

type Hive struct {
	Spec *HiveSpec `json:"spec,omitempty"`
}
type HiveSpec struct {
	TerminationGracePeriodSeconds *int64                  `json:"terminationGracePeriodSeconds,omitempty"`
	Labels                        map[string]string       `json:"labels,omitempty"`
	Annotations                   map[string]string       `json:"annotations,omitempty"`
	SecurityContext               *corev1.SecurityContext `json:"securityContext,omitempty"`
	Image                         *ImageConfig            `json:"image,omitempty"`
	Config                        *HiveSpecConfig         `json:"config,omitempty"`
	Metastore                     *HiveMetastoreSpec      `json:"metastore,omitempty"`
	Server                        *HiveServerSpec         `json:"server,omitempty"`
}
type HiveSpecConfig struct {
	UseHadoopConfig              *bool                   `json:"useHadoopConfig,omitempty"`
	MetastoreClientSocketTimeout string                  `json:"metastoreClientSocketTimeout,omitempty"`
	MetastoreWarehouseDir        string                  `json:"metastoreWarehouseDir,omitempty"`
	DefaultCompression           string                  `json:"defaultCompression,omitempty"`
	DefaultFileFormat            string                  `json:"defaultFileFormat,omitempty"`
	HadoopConfigSecretName       string                  `json:"hadoopConfigSecretName,omitempty"`
	AWS                          *AWSConfig              `json:"aws,omitempty"`
	Azure                        *AzureConfig            `json:"azure,omitempty"`
	Gcs                          *GCSConfig              `json:"gcs,omitempty"`
	S3Compatible                 *S3CompatibleConfig     `json:"s3Compatible,omitempty"`
	DB                           *HiveDBConfig           `json:"db,omitempty"`
	SharedVolume                 *HiveSharedVolumeConfig `json:"sharedVolume,omitempty"`
}
type HiveDBConfig struct {
	AutoCreateMetastoreSchema         *bool  `json:"autoCreateMetastoreSchema,omitempty"`
	EnableMetastoreSchemaVerification *bool  `json:"enableMetastoreSchemaVerification,omitempty"`
	Driver                            string `json:"driver,omitempty"`
	Password                          string `json:"password,omitempty"`
	URL                               string `json:"url,omitempty"`
	Username                          string `json:"username,omitempty"`
	SecretName                        string `json:"secretName,omitempty"`
}
type HiveSharedVolumeConfig struct {
	CreatePVC    *bool  `json:"createPVC,omitempty"`
	Enabled      *bool  `json:"enabled,omitempty"`
	ClaimName    string `json:"claimName,omitempty"`
	MountPath    string `json:"mountPath,omitempty"`
	Size         string `json:"size,omitempty"`
	StorageClass string `json:"storageClass,omitempty"`
}
type HiveMetastoreSpec struct {
	NodeSelector        map[string]string            `json:"nodeSelector,omitempty"`
	PodDisruptionBudget *PodDisruptionBudget         `json:"podDisruptionBudget,omitempty"`
	Affinity            *corev1.Affinity             `json:"affinity,omitempty"`
	LivenessProbe       *corev1.Probe                `json:"livenessProbe,omitempty"`
	ReadinessProbe      *corev1.Probe                `json:"readinessProbe,omitempty"`
	Resources           *corev1.ResourceRequirements `json:"resources,omitempty"`
	Tolerations         []corev1.Toleration          `json:"tolerations,omitempty"`
	Config              *HiveMetastoreSpecConfig     `json:"config,omitempty"`
	Storage             *HiveMetastoreStorageConfig  `json:"storage,omitempty"`
}
type HiveMetastoreSpecConfig struct {
	LogLevel string                  `json:"logLevel,omitempty"`
	Jvm      *JVMConfig              `json:"jvm,omitempty"`
	TLS      *TLSConfig              `json:"tls,omitempty"`
	Auth     *HiveResourceAuthConfig `json:"auth,omitempty"`
}
type HiveResourceAuthConfig struct {
	Enabled *bool `json:"enabled,omitempty"`
}
type HiveMetastoreStorageConfig struct {
	Create *bool  `json:"create,omitempty"`
	Class  string `json:"class,omitempty"`
	Size   string `json:"size,omitempty"`
}
type HiveServerSpec struct {
	NodeSelector        map[string]string            `json:"nodeSelector,omitempty"`
	PodDisruptionBudget *PodDisruptionBudget         `json:"podDisruptionBudget,omitempty"`
	Affinity            *corev1.Affinity             `json:"affinity,omitempty"`
	LivenessProbe       *corev1.Probe                `json:"livenessProbe,omitempty"`
	ReadinessProbe      *corev1.Probe                `json:"readinessProbe,omitempty"`
	Resources           *corev1.ResourceRequirements `json:"resources,omitempty"`
	Tolerations         []corev1.Toleration          `json:"tolerations,omitempty"`
	Config              *HiveServerSpecConfig        `json:"config,omitempty"`
}
type HiveServerSpecConfig struct {
	LogLevel     string                  `json:"logLevel,omitempty"`
	Jvm          *JVMConfig              `json:"jvm,omitempty"`
	TLS          *TLSConfig              `json:"tls,omitempty"`
	MetastoreTLS *TLSConfig              `json:"metastoreTLS,omitempty"`
	Auth         *HiveResourceAuthConfig `json:"auth,omitempty"`
}

/*
End of Hive section
*/

type Hadoop struct {
	Spec *HadoopSpec `json:"spec,omitempty"`
}
type HadoopSpec struct {
	ConfigSecretName string            `json:"configSecretName,omitempty"`
	Image            *ImageConfig      `json:"image,omitempty"`
	HDFS             *HadoopHDFS       `json:"hdfs,omitempty"`
	Config           *HadoopSpecConfig `json:"config,omitempty"`
}
type HadoopHDFS struct {
	Enabled         *bool                   `json:"enabled,omitempty"`
	SecurityContext *corev1.SecurityContext `json:"securityContext,omitempty"`
	Config          *HadoopHDFSConfig       `json:"config,omitempty"`
	Datanode        *HadoopHDFSDatanodeSpec `json:"datanode,omitempty"`
	Namenode        *HadoopHDFSNamenodeSpec `json:"namenode,omitempty"`
}
type HadoopSpecConfig struct {
	DefaultFS    string              `json:"defaultFS,omitempty"`
	AWS          *AWSConfig          `json:"aws,omitempty"`
	Azure        *AzureConfig        `json:"azure,omitempty"`
	Gcs          *GCSConfig          `json:"gcs,omitempty"`
	S3Compatible *S3CompatibleConfig `json:"s3Compatible,omitempty"`
}
type HadoopHDFSConfig struct {
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
	Affinity                      *corev1.Affinity             `json:"affinity,omitempty"`
	Resources                     *corev1.ResourceRequirements `json:"resources,omitempty"`
	Tolerations                   []corev1.Toleration          `json:"tolerations,omitempty"`
	Config                        *HadoopHDFSNodeConfig        `json:"config,omitempty"`
	Storage                       *HadoopHDFSStorageConfig     `json:"storage,omitempty"`
}
type HadoopHDFSStorageConfig struct {
	Class string `json:"class,omitempty"`
	Size  string `json:"size,omitempty"`
}
type HadoopHDFSNodeConfig struct {
	Jvm *JVMConfig `json:"jvm,omitempty"`
}
type HadoopHDFSNamenodeSpec struct {
	Replicas                      *int32                       `json:"replicas,omitempty"`
	TerminationGracePeriodSeconds *int64                       `json:"terminationGracePeriodSeconds,omitempty"`
	Annotations                   map[string]string            `json:"annotations,omitempty"`
	Labels                        map[string]string            `json:"labels,omitempty"`
	NodeSelector                  map[string]string            `json:"nodeSelector,omitempty"`
	Affinity                      *corev1.Affinity             `json:"affinity,omitempty"`
	Resources                     *corev1.ResourceRequirements `json:"resources,omitempty"`
	Tolerations                   []corev1.Toleration          `json:"tolerations,omitempty"`
	Config                        *HadoopHDFSNodeConfig        `json:"config,omitempty"`
	Storage                       *HadoopHDFSStorageConfig     `json:"storage,omitempty"`
}

/*
End of Hadoop Section
*/
