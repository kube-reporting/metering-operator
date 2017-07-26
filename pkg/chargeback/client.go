package chargeback

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/pkg/api"
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

func NewForConfig(c *rest.Config) (*ChargebackClient, error) {
	newC := *c
	newC.GroupVersion = &schema.GroupVersion{
		Group:   Group,
		Version: Version,
	}
	newC.APIPath = "/apis"
	newC.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: api.Codecs}

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
