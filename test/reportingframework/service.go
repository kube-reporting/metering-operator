package reportingframework

import (
	"bytes"
	"context"
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

func (rf *ReportingFramework) determineMeteringAPIUrl(endpoint string) (*url.URL, error) {
	u := rf.ReportingAPIURL

	if rf.UseRouteForReportingAPI {
		meteringRoute, err := rf.RouteClient.Routes(rf.Namespace).Get(context.Background(), meteringRouteName, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}

		return &url.URL{
			Scheme: "https",
			Host:   meteringRoute.Spec.Host,
			Path:   endpoint,
		}, nil
	}
	if rf.UseKubeProxyForReportingAPI {
		u = rf.KubeAPIURL
		scheme := "http"

		if rf.HTTPSAPI {
			scheme = "https"
		}
		apiProxyPath := fmt.Sprintf("/api/v1/namespaces/%s/services/%s/proxy/", rf.Namespace, net.JoinSchemeNamePort(scheme, reportingOperatorServiceName, reportingOperatorServicePortName))
		u.Path = path.Join(apiProxyPath, endpoint)
		return u, nil
	}
	if rf.HTTPSAPI {
		u.Scheme = "https"
	}
	u.Path = endpoint

	return u, nil
}

func (rf *ReportingFramework) doRequest(endpoint, method string, body []byte, query map[string]string) (respBody []byte, code int, err error) {
	u, err := rf.determineMeteringAPIUrl(endpoint)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to successfully return an initialized url.URL object: %v", err)
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

// ReportingOperatorGetRequest is a reportingframework method that performs
// a single GET request to the metering API.
func (rf *ReportingFramework) ReportingOperatorGetRequest(endpoint string, query map[string]string) (respBody []byte, code int, err error) {
	return rf.doRequest(endpoint, "GET", nil, query)
}

// ReportingOperatorPOSTRequest is a reportingframework method that performs
// a single POST request to the metering API.
func (rf *ReportingFramework) ReportingOperatorPOSTRequest(endpoint string, body []byte) (respBody []byte, code int, err error) {
	return rf.doRequest(endpoint, "POST", body, nil)
}
