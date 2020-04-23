package reportingframework

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	metering "github.com/kube-reporting/metering-operator/pkg/generated/clientset/versioned/typed/metering/v1"
	"github.com/kube-reporting/metering-operator/pkg/operator"
	routev1client "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
)

type ReportingFramework struct {
	MeteringClient        metering.MeteringV1Interface
	KubeClient            kubernetes.Interface
	HTTPClient            *http.Client
	RouteClient           routev1client.RouteV1Client
	Namespace             string
	DefaultTimeout        time.Duration
	ReportOutputDirectory string

	KubeAPIURL  *url.URL
	KubeAPIPath string

	UseKubeProxyForReportingAPI bool
	UseRouteForReportingAPI     bool
	RouteBearerToken            string
	ReportingAPIURL             *url.URL
	HTTPSAPI                    bool

	collectOnce                          sync.Once
	reportStart                          time.Time
	reportEnd                            time.Time
	collectPrometheusMetricsDataResponse operator.CollectPrometheusMetricsDataResponse
}

// New initializes a test reporting framework and returns it.
func New(
	namespace,
	kubeconfig string,
	httpsAPI,
	useKubeProxyForReportingAPI,
	useRouteForReportingAPI bool,
	routeBearerToken,
	reportingAPIURL,
	reportOutputDir string,
) (*ReportingFramework, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("build config from flags failed: err %v", err)
	}

	kubeAPIURL, kubeAPIPath, err := rest.DefaultServerURL(config.Host, config.APIPath, schema.GroupVersion{}, true)
	if err != nil {
		return nil, fmt.Errorf("getting kubeAPI url failed: err %v", err)
	}

	var reportAPI *url.URL
	if reportingAPIURL != "" {
		reportAPI, err = url.Parse(reportingAPIURL)
		if err != nil {
			return nil, err
		}
	}

	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("creating new kube-client failed: err %v", err)
	}

	configCopy := *config
	transport, err := rest.TransportFor(&configCopy)
	if err != nil {
		return nil, fmt.Errorf("creating transport for HTTP client failed: err %v", err)
	}

	httpc := &http.Client{Transport: transport}
	if configCopy.Timeout > 0 {
		httpc.Timeout = configCopy.Timeout
	}

	routeClient, err := routev1client.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("creating openshift route client failed, err: %v", err)
	}

	meteringClient, err := metering.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("creating monitoring client failed: err %v", err)
	}

	rf := &ReportingFramework{
		KubeClient:                  kubeClient,
		MeteringClient:              meteringClient,
		HTTPClient:                  httpc,
		RouteClient:                 *routeClient,
		Namespace:                   namespace,
		ReportOutputDirectory:       reportOutputDir,
		DefaultTimeout:              time.Minute,
		KubeAPIURL:                  kubeAPIURL,
		KubeAPIPath:                 kubeAPIPath,
		HTTPSAPI:                    httpsAPI,
		ReportingAPIURL:             reportAPI,
		UseKubeProxyForReportingAPI: useKubeProxyForReportingAPI,
		UseRouteForReportingAPI:     useRouteForReportingAPI,
		RouteBearerToken:            routeBearerToken,
	}

	return rf, nil
}
