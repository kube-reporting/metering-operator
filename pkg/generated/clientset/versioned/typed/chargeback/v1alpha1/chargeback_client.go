package v1alpha1

import (
	v1alpha1 "github.com/coreos-inc/kube-chargeback/pkg/apis/chargeback/v1alpha1"
	"github.com/coreos-inc/kube-chargeback/pkg/generated/clientset/versioned/scheme"
	serializer "k8s.io/apimachinery/pkg/runtime/serializer"
	rest "k8s.io/client-go/rest"
)

type ChargebackV1alpha1Interface interface {
	RESTClient() rest.Interface
	PrestoTablesGetter
	ReportsGetter
	ReportDataSourcesGetter
	ReportGenerationQueriesGetter
	ReportPrometheusQueriesGetter
	ScheduledReportsGetter
	StorageLocationsGetter
}

// ChargebackV1alpha1Client is used to interact with features provided by the chargeback.coreos.com group.
type ChargebackV1alpha1Client struct {
	restClient rest.Interface
}

func (c *ChargebackV1alpha1Client) PrestoTables(namespace string) PrestoTableInterface {
	return newPrestoTables(c, namespace)
}

func (c *ChargebackV1alpha1Client) Reports(namespace string) ReportInterface {
	return newReports(c, namespace)
}

func (c *ChargebackV1alpha1Client) ReportDataSources(namespace string) ReportDataSourceInterface {
	return newReportDataSources(c, namespace)
}

func (c *ChargebackV1alpha1Client) ReportGenerationQueries(namespace string) ReportGenerationQueryInterface {
	return newReportGenerationQueries(c, namespace)
}

func (c *ChargebackV1alpha1Client) ReportPrometheusQueries(namespace string) ReportPrometheusQueryInterface {
	return newReportPrometheusQueries(c, namespace)
}

func (c *ChargebackV1alpha1Client) ScheduledReports(namespace string) ScheduledReportInterface {
	return newScheduledReports(c, namespace)
}

func (c *ChargebackV1alpha1Client) StorageLocations(namespace string) StorageLocationInterface {
	return newStorageLocations(c, namespace)
}

// NewForConfig creates a new ChargebackV1alpha1Client for the given config.
func NewForConfig(c *rest.Config) (*ChargebackV1alpha1Client, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	return &ChargebackV1alpha1Client{client}, nil
}

// NewForConfigOrDie creates a new ChargebackV1alpha1Client for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *ChargebackV1alpha1Client {
	client, err := NewForConfig(c)
	if err != nil {
		panic(err)
	}
	return client
}

// New creates a new ChargebackV1alpha1Client for the given RESTClient.
func New(c rest.Interface) *ChargebackV1alpha1Client {
	return &ChargebackV1alpha1Client{c}
}

func setConfigDefaults(config *rest.Config) error {
	gv := v1alpha1.SchemeGroupVersion
	config.GroupVersion = &gv
	config.APIPath = "/apis"
	config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: scheme.Codecs}

	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	return nil
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *ChargebackV1alpha1Client) RESTClient() rest.Interface {
	if c == nil {
		return nil
	}
	return c.restClient
}
