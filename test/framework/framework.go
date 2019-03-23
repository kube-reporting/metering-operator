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
func New(namespace, kubecon***REMOVED***g string, httpsAPI, useKubeProxyForReportingAPI bool, reportingAPIURL string) (*Framework, error) {
	con***REMOVED***g, err := clientcmd.BuildCon***REMOVED***gFromFlags("", kubecon***REMOVED***g)
	if err != nil {
		return nil, fmt.Errorf("build con***REMOVED***g from flags failed: err %v", err)
	}

	kubeAPIURL, kubeAPIPath, err := rest.DefaultServerURL(con***REMOVED***g.Host, con***REMOVED***g.APIPath, schema.GroupVersion{}, true)
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

	kubeClient, err := kubernetes.NewForCon***REMOVED***g(con***REMOVED***g)
	if err != nil {
		return nil, fmt.Errorf("creating new kube-client failed: err %v", err)
	}

	con***REMOVED***gCopy := *con***REMOVED***g
	transport, err := rest.TransportFor(&con***REMOVED***gCopy)
	if err != nil {
		return nil, fmt.Errorf("creating transport for HTTP client failed: err %v", err)
	}

	httpc := &http.Client{Transport: transport}
	if con***REMOVED***gCopy.Timeout > 0 {
		httpc.Timeout = con***REMOVED***gCopy.Timeout
	}

	meteringClient, err := meteringv1alpha1.NewForCon***REMOVED***g(con***REMOVED***g)
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
