package v1

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

type CronClientInterface interface {
	RESTClient() rest.Interface
	CronGetter
}

func NewForConfig(c *rest.Config) (*CronClient, error) {
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

	return &CronClient{client, dynamicClient}, nil
}

type CronClient struct {
	restClient    rest.Interface
	dynamicClient *dynamic.Client
}

func (c *CronClient) Crons(namespace string) CronInterface {
	return newCrons(c.restClient, c.dynamicClient, namespace)
}
