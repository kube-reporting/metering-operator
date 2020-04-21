package reportingframework

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/client-go/rest"

	metering "github.com/kubernetes-reporting/metering-operator/pkg/generated/clientset/versioned/typed/metering/v1"
	"github.com/kubernetes-reporting/metering-operator/pkg/operator"
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
	httpsAPI,
	useKubeProxyForReportingAPI,
	useRouteForReportingAPI bool,
	namespace,
	routeBearerToken,
	reportingAPIURL,
	reportOutputDir string,
	kubeconfig *rest.Config,
	kubeClient kubernetes.Interface,
	meteringClient metering.MeteringV1Interface,
) (*ReportingFramework, error) {
	kubeAPIURL, kubeAPIPath, err := rest.DefaultServerURL(kubeconfig.Host, kubeconfig.APIPath, schema.GroupVersion{}, true)
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

	configCopy := *kubeconfig
	transport, err := rest.TransportFor(&configCopy)
	if err != nil {
		return nil, fmt.Errorf("creating transport for HTTP client failed: err %v", err)
	}

	httpc := &http.Client{Transport: transport}
	if configCopy.Timeout > 0 {
		httpc.Timeout = configCopy.Timeout
	}

	routeClient, err := routev1client.NewForConfig(kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("creating openshift route client failed, err: %v", err)
	}

	stat, err := os.Stat(reportOutputDir)
	if err != nil {
		return nil, fmt.Errorf("failed to stat the path to the report results directory %s: %v", reportOutputDir, err)
	}
	if !stat.IsDir() {
		return nil, fmt.Errorf("the %s path to the report results directory is not a directory", reportOutputDir)
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
