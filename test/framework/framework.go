package framework

import (
	"fmt"
	"net/http"
	"time"

	"k8s.io/client-go/kubernetes"
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
}

// New initializes a test framework and returns it.
func New(namespace, kubeconfig string) (*Framework, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("build config from flags failed: err %v", err)
	}

	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("creating new kube-client failed: err %v", err)
	}

	httpc := kubeClient.CoreV1().RESTClient().(*rest.RESTClient).Client
	if err != nil {
		return nil, fmt.Errorf("creating http-client failed: err %v", err)
	}

	chargebackClient, err := chargebackv1alpha1.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("creating monitoring client failed: err %v", err)
	}

	f := &Framework{
		KubeClient:       kubeClient,
		ChargebackClient: chargebackClient,
		HTTPClient:       httpc,
		Namespace:        namespace,
		DefaultTimeout:   time.Minute,
	}

	return f, nil
}
