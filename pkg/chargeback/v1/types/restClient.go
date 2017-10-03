package types

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"
)

const (
	Group   = "chargeback.coreos.com"
	Version = "prealpha"
)

func GetRestClient() (rest.Interface, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	scheme := runtime.NewScheme()
	newC := *config
	newC.GroupVersion = &schema.GroupVersion{
		Group:   Group,
		Version: Version,
	}
	newC.APIPath = "/apis"
	newC.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: serializer.NewCodecFactory(scheme)}

	return rest.RESTClientFor(&newC)
}
