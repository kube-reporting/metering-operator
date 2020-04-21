package deployframework

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	apiextclientv1beta1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	metering "github.com/kubernetes-reporting/metering-operator/pkg/apis/metering/v1"
	"github.com/kubernetes-reporting/metering-operator/pkg/deploy"
	meteringclient "github.com/kubernetes-reporting/metering-operator/pkg/generated/clientset/versioned/typed/metering/v1"
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

	manifestsDeployDir = "manifests/deploy"
	hackScriptDirName  = "hack"

	defaultTargetPods             = 7
	meteringOperatorContainerName = "metering-operator-e2e"
)

// DeployFramework contains all the information necessary to deploy
// different metering instances and run tests against them
type DeployFramework struct {
	RunLocal          bool
	RunDevSetup       bool
	KubeConfigPath    string
	RepoDir           string
	OperatorResources *deploy.OperatorResources
	Logger            logrus.FieldLogger
	Config            *rest.Config
	Client            kubernetes.Interface
	APIExtClient      apiextclientv1beta1.CustomResourceDefinitionsGetter
	MeteringClient    meteringclient.MeteringV1Interface
}

// New is the constructor function that creates and returns a new DeployFramework object
func New(logger logrus.FieldLogger, runLocal, runDevSetup bool, nsPrefix, repoDir, kubeconfig string) (*DeployFramework, error) {
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

	meteringClient, err := meteringclient.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize the metering clientset: %v", err)
	}

	manifestsDir, err := filepath.Abs(filepath.Join(repoDir, manifestsDeployDir))
	if err != nil {
		return nil, fmt.Errorf("failed to get the absolute path to the manifest/deploy directory: %v", err)
	}
	_, err = os.Stat(manifestsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to stat the %s path to the manifest/deploy directory: %v", manifestsDir, err)
	}

	operatorResources, err := deploy.ReadMeteringAnsibleOperatorManifests(manifestsDir, defaultPlatform)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize objects from manifests: %v", err)
	}

	deployFramework := &DeployFramework{
		OperatorResources: operatorResources,
		KubeConfigPath:    kubeconfig,
		RepoDir:           repoDir,
		RunLocal:          runLocal,
		RunDevSetup:       runDevSetup,
		Logger:            logger,
		Config:            config,
		Client:            client,
		APIExtClient:      apiextClient,
		MeteringClient:    meteringClient,
	}

	return deployFramework, nil
}

// NewDeployerConfig handles the process of validating inputs before returning
// an initialized Deploy.Config object, or an error if there is any.
func (df *DeployFramework) NewDeployerConfig(
	namespace,
	meteringOperatorImageRepo,
	meteringOperatorImageTag,
	reportingOperatorImageRepo,
	reportingOperatorImageTag string,
	spec metering.MeteringConfigSpec,
) (*deploy.Config, error) {
	meteringConfig := &metering.MeteringConfig{
		ObjectMeta: meta.ObjectMeta{
			Name:      meteringconfigMetadataName,
			Namespace: namespace,
		},
		Spec: spec,
	}

	// validate the reporting-operator image is non-empty when overrided
	if reportingOperatorImageRepo != "" || reportingOperatorImageTag != "" {
		reportingOperatorImageConfig := &metering.ImageConfig{
			Repository: reportingOperatorImageRepo,
			Tag:        reportingOperatorImageTag,
		}
		err := validateImageConfig(*reportingOperatorImageConfig)
		if err != nil {
			return nil, fmt.Errorf("invalid reporting-operator image config: %v", err)
		}
		// Ensure the repo/tag values are set on the MeteringConfig
		if meteringConfig.Spec.ReportingOperator == nil {
			meteringConfig.Spec.ReportingOperator = &metering.ReportingOperator{}
		}
		if meteringConfig.Spec.ReportingOperator.Spec == nil {
			meteringConfig.Spec.ReportingOperator.Spec = &metering.ReportingOperatorSpec{}
		}
		meteringConfig.Spec.ReportingOperator.Spec.Image = reportingOperatorImageConfig

	}
	if meteringOperatorImageRepo != "" || meteringOperatorImageTag != "" {
		// validate both the metering operator image fields are non-empty
		meteringOperatorImageConfig := &metering.ImageConfig{
			Repository: meteringOperatorImageRepo,
			Tag:        meteringOperatorImageTag,
		}
		err := validateImageConfig(*meteringOperatorImageConfig)
		if err != nil {
			return nil, fmt.Errorf("invalid metering-operator image config: %v", err)
		}
	}

	return &deploy.Config{
		Namespace:                namespace,
		Repo:                     meteringOperatorImageRepo,
		RunMeteringOperatorLocal: df.RunLocal,
		Tag:                      meteringOperatorImageTag,
		Platform:                 defaultPlatform,
		DeleteNamespace:          defaultDeleteNamespace,
		ExtraNamespaceLabels: map[string]string{
			"name": testNamespaceLabel,
		},
		OperatorResources: df.OperatorResources,
		MeteringConfig:    meteringConfig,
	}, nil
}
