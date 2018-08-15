package framework

import (
	"k8s.io/apimachinery/pkg/util/net"
	restclient "k8s.io/client-go/rest"
)

const (
	meteringServiceName     = "metering"
	meteringServicePortName = "http"
)

func (f *Framework) NewMeteringSVCRequest(endpoint string, query map[string]string) *restclient.Request {
	wrapper := f.KubeClient.CoreV1().Services(f.Namespace).ProxyGet(f.protocol, meteringServiceName, meteringServicePortName, endpoint, query)
	return wrapper.(*restclient.Request)
}

func (f *Framework) NewMeteringSVCPOSTRequest(endpoint string, body interface{}) *restclient.Request {
	return f.KubeClient.CoreV1().RESTClient().
		Post().
		Namespace(f.Namespace).
		Resource("services").
		SubResource("proxy").
		Name(net.JoinSchemeNamePort(f.protocol, meteringServiceName, meteringServicePortName)).
		Suf***REMOVED***x(endpoint).
		Body(body)
}
