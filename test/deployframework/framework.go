package deployframework

import (
	"fmt"

	"github.com/sirupsen/logrus"
	apiextclientv1beta1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	meteringv1 "github.com/operator-framework/operator-metering/pkg/generated/clientset/versioned/typed/metering/v1"
)

const (
	reportResultsDir                    = "report_results"
	logDir                              = "logs"
	meteringconfigDir                   = "meteringconfigs"
	reportsDir                          = "reports"
	datasourcesDir                      = "reportdatasources"
	reportqueriesDir                    = "reportqueries"
	hivetablesDir                       = "hivetables"
	prestotablesDir                     = "prestotables"
	storagelocationsDir                 = "storagelocations"
	testNamespaceLabel                  = "metering-testing-ns"
	meteringconfigMetadataName          = "operator-metering"
	reportingOperatorServiceAccountName = "reporting-operator"
	defaultPlatform                     = "openshift"
	defaultDeleteNamespace              = true
)

// DeployFramework contains all the information necessary to deploy
// different metering instances and run tests against them
type DeployFramework struct {
	NamespacePrefix   string
	ManifestsDir      string
	LoggingPath       string
	CleanupScriptPath string
	KubeConfigPath    string
	Logger            logrus.FieldLogger
	Client            kubernetes.Interface
	APIExtClient      apiextclientv1beta1.CustomResourceDefinitionsGetter
	MeteringClient    meteringv1.MeteringV1Interface
}

// New is the constructor function that creates and returns a new DeployFramework object
func New(
	logger logrus.FieldLogger,
	nsPrefix,
	manifestDir,
	kubeconfig,
	cleanupScriptPath,
	loggingPath string,
) (*DeployFramework, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to build a kube config from %s: %v", kubeconfig, err)
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize the k8s clientset: %v", err)
	}

	apiextClient, err := apiextclientv1beta1.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize the apiextensions clientset: %v", err)
	}

	meteringClient, err := meteringv1.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize the metering clientset: %v", err)
	}

	deployFramework := &DeployFramework{
		NamespacePrefix:   nsPrefix,
		ManifestsDir:      manifestDir,
		CleanupScriptPath: cleanupScriptPath,
		KubeConfigPath:    kubeconfig,
		LoggingPath:       loggingPath,
		Logger:            logger,
		Client:            client,
		APIExtClient:      apiextClient,
		MeteringClient:    meteringClient,
	}

	return deployFramework, nil
}
