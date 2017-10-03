package v1

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

const (
	Group   = "chargeback.coreos.com"
	Version = "prealpha"
)

type ChargebackInterface interface {
	RESTClient() rest.Interface
	ReportGetter
}

func NewForCon***REMOVED***g(c *rest.Con***REMOVED***g) (*ChargebackClient, error) {
	scheme := runtime.NewScheme()
	newC := *c
	newC.GroupVersion = &schema.GroupVersion{
		Group:   Group,
		Version: Version,
	}
	newC.APIPath = "/apis"
	newC.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: serializer.NewCodecFactory(scheme)}

	client, err := rest.RESTClientFor(&newC)
	if err != nil {
		return nil, err
	}

	dynamicClient, err := dynamic.NewClient(&newC)
	if err != nil {
		return nil, err
	}

	return &ChargebackClient{client, dynamicClient}, nil
}

type ChargebackClient struct {
	restClient    rest.Interface
	dynamicClient *dynamic.Client
}

func (c *ChargebackClient) Reports(namespace string) ReportInterface {
	return newReports(c.restClient, c.dynamicClient, namespace)
}
