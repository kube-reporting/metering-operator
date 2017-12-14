package framework

import restclient "k8s.io/client-go/rest"

func (f *Framework) NewChargebackSVCRequest(ns, svcName, endpoint string, query map[string]string) *restclient.Request {
	wrapper := f.KubeClient.CoreV1().Services(ns).ProxyGet("", svcName, "8080", endpoint, query)
	return wrapper.(*restclient.Request)
}
