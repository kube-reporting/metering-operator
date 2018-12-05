package framework

import (
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

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

	protocol                   string
	collectOnce                sync.Once
	reportStart                time.Time
	reportEnd                  time.Time
	collectPromsumDataResponse operator.CollectPromsumDataResponse
}

// New initializes a test framework and returns it.
func New(namespace, kubecon***REMOVED***g string, httpsAPI bool) (*Framework, error) {
	con***REMOVED***g, err := clientcmd.BuildCon***REMOVED***gFromFlags("", kubecon***REMOVED***g)
	if err != nil {
		return nil, fmt.Errorf("build con***REMOVED***g from flags failed: err %v", err)
	}

	kubeClient, err := kubernetes.NewForCon***REMOVED***g(con***REMOVED***g)
	if err != nil {
		return nil, fmt.Errorf("creating new kube-client failed: err %v", err)
	}

	httpc := kubeClient.CoreV1().RESTClient().(*rest.RESTClient).Client
	if err != nil {
		return nil, fmt.Errorf("creating http-client failed: err %v", err)
	}

	meteringClient, err := meteringv1alpha1.NewForCon***REMOVED***g(con***REMOVED***g)
	if err != nil {
		return nil, fmt.Errorf("creating monitoring client failed: err %v", err)
	}
	protocol := "http"
	if httpsAPI {
		protocol = "https"
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
		KubeClient:            kubeClient,
		MeteringClient:        meteringClient,
		HTTPClient:            httpc,
		Namespace:             namespace,
		ReportOutputDirectory: reportOutputDir,
		DefaultTimeout:        time.Minute,
		protocol:              protocol,
	}

	return f, nil
}
