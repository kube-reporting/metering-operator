package framework

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"

	"k8s.io/apimachinery/pkg/util/net"
)

const (
	reportingOperatorServiceName     = "reporting-operator"
	reportingOperatorServicePortName = "http"
)

func (f *Framework) doRequest(endpoint, method string, body []byte, query map[string]string) (respBody []byte, code int, err error) {
	var u *url.URL
	if f.UseKubeProxyForReportingAPI {
		u = f.KubeAPIURL
		proto := "http"
		if f.HTTPSAPI {
			proto = "https"
		}
		apiProxyPath := fmt.Sprintf("/api/v1/namespaces/%s/services/%s/proxy/", f.Namespace, net.JoinSchemeNamePort(proto, reportingOperatorServiceName, reportingOperatorServicePortName))
		u.Path = path.Join(apiProxyPath, endpoint)
	} ***REMOVED*** {
		u = f.ReportingAPIURL
		if f.HTTPSAPI {
			u.Scheme = "https"
		}
		u.Path = endpoint
	}

	req := &http.Request{
		URL:    u,
		Method: method,
	}

	if body != nil {
		req.Body = ioutil.NopCloser(bytes.NewReader(body))
	}

	q := req.URL.Query()
	for key, val := range query {
		q.Set(key, val)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := f.HTTPClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	respBody, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	return respBody, resp.StatusCode, nil
}

func (f *Framework) ReportingOperatorRequest(endpoint string, query map[string]string) (respBody []byte, code int, err error) {
	return f.doRequest(endpoint, "GET", nil, query)
}

func (f *Framework) ReportingOperatorPOSTRequest(endpoint string, body []byte) (respBody []byte, code int, err error) {
	return f.doRequest(endpoint, "POST", body, nil)
}
