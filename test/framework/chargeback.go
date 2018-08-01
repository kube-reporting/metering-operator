package framework

import (
	"k8s.io/apimachinery/pkg/util/net"
	restclient "k8s.io/client-go/rest"
)

const (
	chargebackServiceName     = "metering"
	chargebackServicePortName = "http"
)

func (f *Framework) NewChargebackSVCRequest(endpoint string, query map[string]string) *restclient.Request {
	wrapper := f.KubeClient.CoreV1().Services(f.Namespace).ProxyGet("", chargebackServiceName, chargebackServicePortName, endpoint, query)
	return wrapper.(*restclient.Request)
}

func (f *Framework) NewChargebackSVCPOSTRequest(endpoint string, body interface{}) *restclient.Request {
	return f.KubeClient.CoreV1().RESTClient().
		Post().
		Namespace(f.Namespace).
		Resource("services").
		SubResource("proxy").
		Name(net.JoinSchemeNamePort("", chargebackServiceName, chargebackServicePortName)).
		Suffix(endpoint).
		Body(body)
}
