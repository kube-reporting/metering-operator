package framework

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
	"k8s.io/client-go/tools/clientcmd"

	meteringv1alpha1 "github.com/operator-framework/operator-metering/pkg/generated/clientset/versioned/typed/metering/v1alpha1"
	"github.com/operator-framework/operator-metering/pkg/operator"
)

type Framework struct {
	MeteringClient        meteringv1alpha1.MeteringV1alpha1Interface
	KubeClient            kubernetes.Interface
	HTTPClient            *http.Client
	Namespace             string
	DefaultTimeout        time.Duration
	ReportOutputDirectory string

	KubeAPIURL  *url.URL
	KubeAPIPath string

	UseKubeProxyForReportingAPI bool
	ReportingAPIURL             *url.URL
	HTTPSAPI                    bool

	collectOnce                sync.Once
	reportStart                time.Time
	reportEnd                  time.Time
	collectPromsumDataResponse operator.CollectPromsumDataResponse
}

// New initializes a test framework and returns it.
func New(namespace, kubeconfig string, httpsAPI, useKubeProxyForReportingAPI bool, reportingAPIURL string) (*Framework, error) {
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

	meteringClient, err := meteringv1alpha1.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("creating monitoring client failed: err %v", err)
	}

	reportOutputDir := os.Getenv("TEST_RESULT_REPORT_OUTPUT_DIRECTORY")
	if reportOutputDir == "" {
		return nil, fmt.Errorf("$TEST_RESULT_REPORT_OUTPUT_DIRECTORY must be set")
	}

	err = os.MkdirAll(reportOutputDir, 0777)
	if err != nil {
		return nil, fmt.Errorf("error making directory %s, err: %s", reportOutputDir, err)
	}

	f := &Framework{
		KubeClient:                  kubeClient,
		MeteringClient:              meteringClient,
		HTTPClient:                  httpc,
		Namespace:                   namespace,
		ReportOutputDirectory:       reportOutputDir,
		DefaultTimeout:              time.Minute,
		KubeAPIURL:                  kubeAPIURL,
		KubeAPIPath:                 kubeAPIPath,
		HTTPSAPI:                    httpsAPI,
		ReportingAPIURL:             reportAPI,
		UseKubeProxyForReportingAPI: useKubeProxyForReportingAPI,
	}

	return f, nil
}
