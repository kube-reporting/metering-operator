package framework

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/util/net"
)

const (
	reportingOperatorServiceName     = "reporting-operator"
	reportingOperatorServicePortName = "http"
	meteringRouteName                = "metering"
)

func (f *Framework) doRequest(endpoint, method string, body []byte, query map[string]string) (respBody []byte, code int, err error) {
	var u *url.URL

	if f.UseRouteForReportingAPI {
		const routeName = "metering"

		// query all routes for the metering route
		meteringRoute, err := f.RouteClient.Routes(f.Namespace).Get(meteringRouteName, metav1.GetOptions{})
		if err != nil {
			return nil, 0, fmt.Errorf("query for metering route failed, err: %v", err)
		}

		u = &url.URL{
			Scheme: "https",
			Host:   meteringRoute.Spec.Host,
			Path:   endpoint,
		}
	} ***REMOVED*** if f.UseKubeProxyForReportingAPI {
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
		Header: make(http.Header),
	}

	if f.UseRouteForReportingAPI {
		// check if the bearer token to the reporting-operator serviceaccount is uninitialized
		if f.RouteBearerToken == "" {
			return nil, 0, fmt.Errorf("use-route-for-reporting-api is set to true, but route-bearer-token is uninitialized: %v", err)
		}

		accessToken := "Bearer " + f.RouteBearerToken
		req.Header.Set("Authorization", accessToken)
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
