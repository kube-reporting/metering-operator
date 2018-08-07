package framework

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	chargebackv1alpha1 "github.com/operator-framework/operator-metering/pkg/generated/clientset/versioned/typed/chargeback/v1alpha1"
)

type Framework struct {
	ChargebackClient chargebackv1alpha1.ChargebackV1alpha1Interface
	KubeClient       kubernetes.Interface
	HTTPClient       *http.Client
	Namespace        string
	DefaultTimeout   time.Duration

	protocol    string
	collectOnce sync.Once
	reportStart time.Time
	reportEnd   time.Time
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

	chargebackClient, err := chargebackv1alpha1.NewForCon***REMOVED***g(con***REMOVED***g)
	if err != nil {
		return nil, fmt.Errorf("creating monitoring client failed: err %v", err)
	}
	protocol := "http"
	if httpsAPI {
		protocol = "https"
	}

	f := &Framework{
		KubeClient:       kubeClient,
		ChargebackClient: chargebackClient,
		HTTPClient:       httpc,
		Namespace:        namespace,
		DefaultTimeout:   time.Minute,
		protocol:         protocol,
	}

	return f, nil
}
