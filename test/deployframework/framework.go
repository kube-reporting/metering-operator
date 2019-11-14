package deployframework

import (
	"fmt"
	"os"
	"path/***REMOVED***lepath"

	"github.com/sirupsen/logrus"
	apiextclientv1beta1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	metering "github.com/operator-framework/operator-metering/pkg/apis/metering/v1"
	meteringclient "github.com/operator-framework/operator-metering/pkg/generated/clientset/versioned/typed/metering/v1"
	"github.com/operator-framework/operator-metering/pkg/operator/deploy"
)

const (
	reportResultsDir                    = "report_results"
	logDir                              = "logs"
	meteringcon***REMOVED***gDir                   = "meteringcon***REMOVED***gs"
	reportsDir                          = "reports"
	datasourcesDir                      = "reportdatasources"
	reportqueriesDir                    = "reportqueries"
	hivetablesDir                       = "hivetables"
	prestotablesDir                     = "prestotables"
	storagelocationsDir                 = "storagelocations"
	testNamespaceLabel                  = "metering-testing-ns"
	meteringcon***REMOVED***gMetadataName          = "operator-metering"
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
	KubeCon***REMOVED***gPath    string
	RepoDir           string
	OperatorResources *deploy.OperatorResources
	Logger            logrus.FieldLogger
	Con***REMOVED***g            *rest.Con***REMOVED***g
	Client            kubernetes.Interface
	APIExtClient      apiextclientv1beta1.CustomResourceDe***REMOVED***nitionsGetter
	MeteringClient    meteringclient.MeteringV1Interface
}

// New is the constructor function that creates and returns a new DeployFramework object
func New(logger logrus.FieldLogger, runLocal bool, nsPre***REMOVED***x, repoDir, kubecon***REMOVED***g string) (*DeployFramework, error) {
	con***REMOVED***g, err := clientcmd.BuildCon***REMOVED***gFromFlags("", kubecon***REMOVED***g)
	if err != nil {
		return nil, fmt.Errorf("failed to build a kube con***REMOVED***g from %s: %v", kubecon***REMOVED***g, err)
	}

	client, err := kubernetes.NewForCon***REMOVED***g(con***REMOVED***g)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize the k8s clientset: %v", err)
	}

	apiextClient, err := apiextclientv1beta1.NewForCon***REMOVED***g(con***REMOVED***g)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize the apiextensions clientset: %v", err)
	}

	meteringClient, err := meteringclient.NewForCon***REMOVED***g(con***REMOVED***g)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize the metering clientset: %v", err)
	}

	manifestsDir, err := ***REMOVED***lepath.Abs(***REMOVED***lepath.Join(repoDir, manifestsDeployDir))
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
		KubeCon***REMOVED***gPath:    kubecon***REMOVED***g,
		RepoDir:           repoDir,
		RunLocal:          runLocal,
		Logger:            logger,
		Con***REMOVED***g:            con***REMOVED***g,
		Client:            client,
		APIExtClient:      apiextClient,
		MeteringClient:    meteringClient,
	}

	return deployFramework, nil
}

// NewDeployerCon***REMOVED***g handles the process of validating inputs before returning
// an initialized Deploy.Con***REMOVED***g object, or an error if there is any.
func (df *DeployFramework) NewDeployerCon***REMOVED***g(
	namespace,
	meteringOperatorImageRepo,
	meteringOperatorImageTag,
	reportingOperatorImageRepo,
	reportingOperatorImageTag string,
	spec metering.MeteringCon***REMOVED***gSpec,
) (*deploy.Con***REMOVED***g, error) {
	meteringCon***REMOVED***g := &metering.MeteringCon***REMOVED***g{
		ObjectMeta: meta.ObjectMeta{
			Name:      meteringcon***REMOVED***gMetadataName,
			Namespace: namespace,
		},
		Spec: spec,
	}

	// validate the reporting-operator image is non-empty when overrided
	if reportingOperatorImageRepo != "" || reportingOperatorImageTag != "" {
		reportingOperatorImageCon***REMOVED***g := &metering.ImageCon***REMOVED***g{
			Repository: reportingOperatorImageRepo,
			Tag:        reportingOperatorImageTag,
		}
		err := validateImageCon***REMOVED***g(*reportingOperatorImageCon***REMOVED***g)
		if err != nil {
			return nil, fmt.Errorf("invalid reporting-operator image con***REMOVED***g: %v", err)
		}
		// Ensure the repo/tag values are set on the MeteringCon***REMOVED***g
		if meteringCon***REMOVED***g.Spec.ReportingOperator == nil {
			meteringCon***REMOVED***g.Spec.ReportingOperator = &metering.ReportingOperator{}
		}
		if meteringCon***REMOVED***g.Spec.ReportingOperator.Spec == nil {
			meteringCon***REMOVED***g.Spec.ReportingOperator.Spec = &metering.ReportingOperatorSpec{}
		}
		meteringCon***REMOVED***g.Spec.ReportingOperator.Spec.Image = reportingOperatorImageCon***REMOVED***g

	}
	if meteringOperatorImageRepo != "" || meteringOperatorImageTag != "" {
		// validate both the metering operator image ***REMOVED***elds are non-empty
		meteringOperatorImageCon***REMOVED***g := &metering.ImageCon***REMOVED***g{
			Repository: meteringOperatorImageRepo,
			Tag:        meteringOperatorImageTag,
		}
		err := validateImageCon***REMOVED***g(*meteringOperatorImageCon***REMOVED***g)
		if err != nil {
			return nil, fmt.Errorf("invalid metering-operator image con***REMOVED***g: %v", err)
		}
	}

	return &deploy.Con***REMOVED***g{
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
		MeteringCon***REMOVED***g:    meteringCon***REMOVED***g,
	}, nil
}
