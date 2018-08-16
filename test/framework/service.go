package framework

import (
	"k8s.io/apimachinery/pkg/util/net"
	restclient "k8s.io/client-go/rest"
)

const (
	reportingOperatorServiceName     = "reporting-operator"
	reportingOperatorServicePortName = "http"
)

func (f *Framework) NewReportingOperatorSVCRequest(endpoint string, query map[string]string) *restclient.Request {
	wrapper := f.KubeClient.CoreV1().Services(f.Namespace).ProxyGet(f.protocol, reportingOperatorServiceName, reportingOperatorServicePortName, endpoint, query)
	return wrapper.(*restclient.Request)
}

func (f *Framework) NewReportingOperatorSVCPOSTRequest(endpoint string, body interface{}) *restclient.Request {
	return f.KubeClient.CoreV1().RESTClient().
		Post().
		Namespace(f.Namespace).
		Resource("services").
		SubResource("proxy").
		Name(net.JoinSchemeNamePort(f.protocol, reportingOperatorServiceName, reportingOperatorServicePortName)).
		Suffix(endpoint).
		Body(body)
}
