package reportingframework

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

func (rf *ReportingFramework) doRequest(endpoint, method string, body []byte, query map[string]string) (respBody []byte, code int, err error) {
	var u *url.URL

	if rf.UseRouteForReportingAPI {
		const routeName = "metering"

		// query all routes for the metering route
		meteringRoute, err := rf.RouteClient.Routes(rf.Namespace).Get(meteringRouteName, metav1.GetOptions{})
		if err != nil {
			return nil, 0, fmt.Errorf("query for metering route failed, err: %v", err)
		}

		u = &url.URL{
			Scheme: "https",
			Host:   meteringRoute.Spec.Host,
			Path:   endpoint,
		}
	} else if rf.UseKubeProxyForReportingAPI {
		u = rf.KubeAPIURL
		proto := "http"
		if rf.HTTPSAPI {
			proto = "https"
		}
		apiProxyPath := fmt.Sprintf("/api/v1/namespaces/%s/services/%s/proxy/", rf.Namespace, net.JoinSchemeNamePort(proto, reportingOperatorServiceName, reportingOperatorServicePortName))
		u.Path = path.Join(apiProxyPath, endpoint)
	} else {
		u = rf.ReportingAPIURL
		if rf.HTTPSAPI {
			u.Scheme = "https"
		}
		u.Path = endpoint
	}

	req := &http.Request{
		URL:    u,
		Method: method,
		Header: make(http.Header),
	}

	if rf.UseRouteForReportingAPI {
		// check if the bearer token to the reporting-operator serviceaccount is uninitialized
		if rf.RouteBearerToken == "" {
			return nil, 0, fmt.Errorf("use-route-for-reporting-api is set to true, but route-bearer-token is uninitialized: %v", err)
		}

		accessToken := "Bearer " + rf.RouteBearerToken
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

	resp, err := rf.HTTPClient.Do(req)
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

func (rf *ReportingFramework) ReportingOperatorRequest(endpoint string, query map[string]string) (respBody []byte, code int, err error) {
	return rf.doRequest(endpoint, "GET", nil, query)
}

func (rf *ReportingFramework) ReportingOperatorPOSTRequest(endpoint string, body []byte) (respBody []byte, code int, err error) {
	return rf.doRequest(endpoint, "POST", body, nil)
}
