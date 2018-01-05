package framework

import (
	"k8s.io/apimachinery/pkg/util/net"
	restclient "k8s.io/client-go/rest"
)

func (f *Framework) NewChargebackSVCRequest(ns, svcName, endpoint string, query map[string]string) *restclient.Request {
	wrapper := f.KubeClient.CoreV1().Services(ns).ProxyGet("", svcName, "8080", endpoint, query)
	return wrapper.(*restclient.Request)
}

func (f *Framework) NewChargebackSVCPOSTRequest(ns, svcName, endpoint string, body interface{}) *restclient.Request {
	return f.KubeClient.CoreV1().RESTClient().
		Post().
		Namespace(ns).
		Resource("services").
		SubResource("proxy").
		Name(net.JoinSchemeNamePort("", svcName, "8080")).
		Suf***REMOVED***x(endpoint).
		Body(body)
}
