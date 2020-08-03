package operator

import (
	"context"
	"time"

	metering "github.com/kube-reporting/metering-operator/pkg/apis/metering/v1"
	"github.com/kube-reporting/metering-operator/pkg/operator/reporting"
)

type ReportingOperator interface {
	Run(ctx context.Context) error
}

// DependencyResolver analyzes report dependencies for reports, report queries, and data sources.
type DependencyResolver interface {
	// ResolveDependencies determines, for the given namespace and report query inputs, any report, report
	// query, and data source dependencies.
	ResolveDependencies(namespace string, inputDefs []metering.ReportQueryInputDefinition, inputVals []metering.ReportQueryInputValue) (*reporting.DependencyResolutionResult, error)
}

// TLSConfig allows configuration of using TLS (on/off) as well as the cert and key.
type TLSConfig struct {
	// UseTLS if true TLS is requested and cert/key must be provided for valid config.
	// TODO: we should just key on the existence of this, not a bool, unless we want to
	// configure TLS and not use it.
	UseTLS bool
	// TLSCert is the certificate used to serve TLS.
	TLSCert string
	// TLSKey is the key used to serve TLS.
	TLSKey string
}

// PrometheusConfig provides the configuration options to set up a Prometheus connections from a URL.
type PrometheusConfig struct {
	// Address is the URL to reach Prometheus.
	Address string
	// SkipTLSVerify should not be used in a production environment.  This is used to configure
	// the transport to not verify the sever it is connecting too.  For testing only.
	SkipTLSVerify bool
	// BearerToken is the bearer token for authentication.
	BearerToken string
	// BearerTokenFile is a path to a file that contains the bearer token.  If configured the
	// the contents are periodically read and the last successfully read value takes precedence over
	// BearerToken.
	BearerTokenFile string
	// CAFile is the path of the PEM-encoded server trusted root certificates.
	CAFile string
}

// Config holds the user-facing configuration options for running the reporting operator.
type Config struct {
	// Hostname is used as the identity of the resource lock for leader election as well as the event
	// source when recording events.
	Hostname string

	// OwnNamespace is the namespace the operator is running in.  It is also used as the informer namespace
	// if AllNamespaces or TargetNamespaces are not defined.  OwnNamespace is also the target sink for event
	// recording.
	OwnNamespace string
	// AllNamespaces should be set to true if the operator should watch all namespaces for metering.openshift.io
	// resources.  This should be set to true if more than one namespace is passed in TargetNamespaces.
	AllNamespaces bool
	// TargetNamespaces are the the namespaces for reporting-operator to watch for metering.openshift.io resources.
	TargetNamespaces []string

	// Kubeconfig is the path to the kubeconfig file.  If unset default config loading rules will be used.
	Kubeconfig string

	// APIListen configures the ip:port to listen on for the reporting API.
	APIListen string
	// MetricsListen configures the ip:port to listen on for Prometheus metrics.
	MetricsListen string
	// PprofListen configures the ip:port to listen on for the pprof debug info.
	PprofListen string

	// HiveHost configures the hostname:port for connecting to Hive.
	HiveHost string
	// HiveUseTLS, when set to true, enables TLS when connecting to Hive.  When set, HiveCAFile should also be set.
	HiveUseTLS bool
	// HiveCAFile configures the path to the certificate authority to use to connect to Hive. If empty, defaults to
	// system CAs.  Should be set when HiveUseTLS is set to true.
	HiveCAFile string
	// HiveTLSInsecureSkipVerify is not for production use.  Setting to true disables TLS verification when connecting
	// to Hive.  For testing only.
	HiveTLSInsecureSkipVerify bool

	// HiveUseClientCertAuth enables TLS client certificate authentication when HiveUseTLS is also enabled.
	HiveUseClientCertAuth bool
	// HiveClientCertFile configures the path to the client certificate to use to connect to Hive.
	HiveClientCertFile string
	// HiveClientKeyFile configures the path to the client key to use to connect to Hive.
	HiveClientKeyFile string

	// PrestoHost configures the hostname:port for connecting to Presto.
	PrestoHost string
	// PrestoUseTLS enables TLS when connecting to Presto.
	PrestoUseTLS bool
	// PrestoCAFile configures path to the certificate authority to use to connect to Presto.
	PrestoCAFile string
	// PrestoTLSInsecureSkipVerify is not for production use.  Setting to true disables TLS verification when connecting
	// to Presto.  For testing only.
	PrestoTLSInsecureSkipVerify bool

	// PrestoUseClientCertAuth enables TLS client certificate authentication when PrestoUseTLS is also enabled.
	PrestoUseClientCertAuth bool
	// PrestoClientCertFile configures the path to the client certificate to use to connect to Presto.
	PrestoClientCertFile string
	// PrestoClientKeyFile configures the path to the client key to use to connect to Presto.
	PrestoClientKeyFile string

	// PrestoMaxQueryLength configures the capacity of the buffer pool for the Presto store.
	// TODO: why would someone set this?
	PrestoMaxQueryLength int

	// DisablePrometheusMetricsImporter disables collecting Prometheus metrics periodically.
	DisablePrometheusMetricsImporter bool
	// EnableFinalizers, if enabled, then finalizers will be set on some resources to ensure the reporting-operator
	// is able to perform cleanup before the resource is deleted from the API.
	EnableFinalizers bool

	// LogDMLQueries controls if we log data manipulation queries made via Presto (SELECT, INSERT, etc).
	LogDMLQueries bool
	// LogDDLQueries controls if we log data definition language queries made via Hive (CREATE TABLE, DROP TABLE, etc).
	LogDDLQueries bool

	// PrometheusQueryConfig holds the PrometheusQueryConfig api configuration.
	PrometheusQueryConfig metering.PrometheusQueryConfig
	// PrometheusDataSourceMaxQueryRangeDuration if non-zero specifies the maximum duration of time to query from
	// Prometheus. When back filling, this value is used for the chunkSize when querying Prometheus.
	PrometheusDataSourceMaxQueryRangeDuration time.Duration
	// PrometheusDataSourceMaxBackfillImportDuration if non-zero specifies the maximum duration of time before the
	// current to look back for data when back filling. Only one of PrometheusDataSourceMaxBackfillImportDuration and
	// PrometheusDataSourceGlobalImportFromTime should be set.
	PrometheusDataSourceMaxBackfillImportDuration time.Duration
	// PrometheusDataSourceGlobalImportFromTime, if non-empty, indicates when Prometheus ReportDataSource data should
	// be back filled from.
	PrometheusDataSourceGlobalImportFromTime *time.Time

	// ProxyTrustedCABundle configures the path to the certificate authority bundle used to connect to the cluster-wide
	// https proxy.
	ProxyTrustedCABundle string

	// LeaderLeaseDuration is the duration that non-leader candidates will wait to force acquire leadership.  This
	// value, halved, will be used as the renewal deadline duration for the acting master.
	LeaderLeaseDuration time.Duration

	// APITLSConfig configures TLS options for API traffic.
	APITLSConfig TLSConfig
	// MetricsTLSConfig configures TLS options for Prometheus metrics endpoint traffic.
	MetricsTLSConfig TLSConfig
	// PrometheusConfig configures connectivity options for Prometheus.
	PrometheusConfig PrometheusConfig
}

const (
	// defaultResyncPeriod is the default informer sync period.
	defaultResyncPeriod = time.Minute * 15
	// prestoUsername is the default presto user name for building the presto query endpoint.
	prestoUsername = "reporting-operator"

	// DefaultPrometheusQueryInterval - Query Prometheus every 5 minutes
	DefaultPrometheusQueryInterval = time.Minute * 5
	// DefaultPrometheusQueryStepSize - Query data from Prometheus at a 60 second resolution
	// (one data point per minute max)
	DefaultPrometheusQueryStepSize = time.Minute
	// DefaultPrometheusQueryChunkSize the default value for how much data we will insert into Presto per Prometheus query.
	DefaultPrometheusQueryChunkSize = 5 * time.Minute
	// DefaultPrometheusDataSourceMaxQueryRangeDuration is how much data we will query from Prometheus at once
	DefaultPrometheusDataSourceMaxQueryRangeDuration = 10 * time.Minute
	// DefaultPrometheusDataSourceMaxBackfillImportDuration how far we will query for backlogged data.
	DefaultPrometheusDataSourceMaxBackfillImportDuration = 2 * time.Hour
)
