package v1alpha1

import (
	v1alpha1 "github.com/coreos-inc/kube-chargeback/pkg/apis/chargeback/v1alpha1"
	"github.com/coreos-inc/kube-chargeback/pkg/generated/clientset/versioned/scheme"
	serializer "k8s.io/apimachinery/pkg/runtime/serializer"
	rest "k8s.io/client-go/rest"
)

type ChargebackV1alpha1Interface interface {
	RESTClient() rest.Interface
	ReportsGetter
	ReportDataStoresGetter
	ReportGenerationQueriesGetter
	ReportPrometheusQueriesGetter
}

// ChargebackV1alpha1Client is used to interact with features provided by the chargeback.coreos.com group.
type ChargebackV1alpha1Client struct {
	restClient rest.Interface
}

func (c *ChargebackV1alpha1Client) Reports(namespace string) ReportInterface {
	return newReports(c, namespace)
}

func (c *ChargebackV1alpha1Client) ReportDataStores(namespace string) ReportDataStoreInterface {
	return newReportDataStores(c, namespace)
}

func (c *ChargebackV1alpha1Client) ReportGenerationQueries(namespace string) ReportGenerationQueryInterface {
	return newReportGenerationQueries(c, namespace)
}

func (c *ChargebackV1alpha1Client) ReportPrometheusQueries(namespace string) ReportPrometheusQueryInterface {
	return newReportPrometheusQueries(c, namespace)
}

// NewForCon***REMOVED***g creates a new ChargebackV1alpha1Client for the given con***REMOVED***g.
func NewForCon***REMOVED***g(c *rest.Con***REMOVED***g) (*ChargebackV1alpha1Client, error) {
	con***REMOVED***g := *c
	if err := setCon***REMOVED***gDefaults(&con***REMOVED***g); err != nil {
		return nil, err
	}
	client, err := rest.RESTClientFor(&con***REMOVED***g)
	if err != nil {
		return nil, err
	}
	return &ChargebackV1alpha1Client{client}, nil
}

// NewForCon***REMOVED***gOrDie creates a new ChargebackV1alpha1Client for the given con***REMOVED***g and
// panics if there is an error in the con***REMOVED***g.
func NewForCon***REMOVED***gOrDie(c *rest.Con***REMOVED***g) *ChargebackV1alpha1Client {
	client, err := NewForCon***REMOVED***g(c)
	if err != nil {
		panic(err)
	}
	return client
}

// New creates a new ChargebackV1alpha1Client for the given RESTClient.
func New(c rest.Interface) *ChargebackV1alpha1Client {
	return &ChargebackV1alpha1Client{c}
}

func setCon***REMOVED***gDefaults(con***REMOVED***g *rest.Con***REMOVED***g) error {
	gv := v1alpha1.SchemeGroupVersion
	con***REMOVED***g.GroupVersion = &gv
	con***REMOVED***g.APIPath = "/apis"
	con***REMOVED***g.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: scheme.Codecs}

	if con***REMOVED***g.UserAgent == "" {
		con***REMOVED***g.UserAgent = rest.DefaultKubernetesUserAgent()
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
