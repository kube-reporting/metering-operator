package deployframework

import (
	"fmt"

	"github.com/sirupsen/logrus"
	apiextclientv1beta1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	meteringv1 "github.com/operator-framework/operator-metering/pkg/generated/clientset/versioned/typed/metering/v1"
	"github.com/operator-framework/operator-metering/pkg/operator/deploy"
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
	KubeConfigPath             string
	ReportingOperatorImageRepo string
	ReportingOperatorImageTag  string
	OperatorResources          *deploy.OperatorResources
	Logger                     logrus.FieldLogger
	Client                     kubernetes.Interface
	APIExtClient               apiextclientv1beta1.CustomResourceDefinitionsGetter
	MeteringClient             meteringv1.MeteringV1Interface
}

// New is the constructor function that creates and returns a new DeployFramework object
func New(
	logger logrus.FieldLogger,
	nsPrefix,
	manifestsDir,
	kubeconfig,
	reportingOperatorImageRepo,
	reportingOperatorImageTag string,
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

	operatorResources, err := deploy.ReadMeteringAnsibleOperatorManifests(manifestsDir, defaultPlatform)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize objects from manifests: %v", err)
	}

	deployFramework := &DeployFramework{
		OperatorResources:          operatorResources,
		KubeConfigPath:             kubeconfig,
		ReportingOperatorImageRepo: reportingOperatorImageRepo,
		ReportingOperatorImageTag:  reportingOperatorImageTag,
		Logger:         logger,
		Client:         client,
		APIExtClient:   apiextClient,
		MeteringClient: meteringClient,
	}

	return deployFramework, nil
}
