package framework

import (
	"k8s.io/apimachinery/pkg/util/net"
	restclient "k8s.io/client-go/rest"
)

func (f *Framework) NewChargebackSVCRequest(svcName, endpoint string, query map[string]string) *restclient.Request {
	wrapper := f.KubeClient.CoreV1().Services(f.Namespace).ProxyGet("", svcName, "8080", endpoint, query)
	return wrapper.(*restclient.Request)
}

func (f *Framework) NewChargebackSVCPOSTRequest(svcName, endpoint string, body interface{}) *restclient.Request {
	return f.KubeClient.CoreV1().RESTClient().
		Post().
		Namespace(f.Namespace).
		Resource("services").
		SubResource("proxy").
		Name(net.JoinSchemeNamePort("", svcName, "8080")).
		Suf***REMOVED***x(endpoint).
		Body(body)
}
