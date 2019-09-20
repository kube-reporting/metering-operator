package deploy

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	metering "github.com/operator-framework/operator-metering/pkg/generated/clientset/versioned/typed/metering/v1"
	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextclientv1beta1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"

	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
)

const (
	defaultMeteringConfig     = "meteringconfig.yaml"
	manifestDeployDirname     = "manifests/deploy"
	manifestAnsibleOperator   = "metering-ansible-operator"
	upstreamManifestDirname   = "upstream"
	openshiftManifestDirname  = "openshift"
	ocpTestingManifestDirname = "ocp-testing"

	meteringDeploymentFile         = "metering-operator-deployment.yaml"
	meteringServiceAccountFile     = "metering-operator-service-account.yaml"
	meteringRoleBindingFile        = "metering-operator-rolebinding.yaml"
	meteringRoleFile               = "metering-operator-role.yaml"
	meteringClusterRoleBindingFile = "metering-operator-clusterrolebinding.yaml"
	meteringClusterRoleFile        = "metering-operator-clusterrole.yaml"

	hivetableFile        = "hive.crd.yaml"
	prestotableFile      = "prestotable.crd.yaml"
	meteringconfigFile   = "meteringconfig.crd.yaml"
	reportFile           = "report.crd.yaml"
	reportdatasourceFile = "reportdatasource.crd.yaml"
	reportqueryFile      = "reportquery.crd.yaml"
	storagelocationFile  = "storagelocation.crd.yaml"
)

// CRD is a structure that holds the information needed to install
// these resources, like the name of the CRD, the path to the CRD
// in the manifests directory, and a pointer to the apiextensions CRD type.
type CRD struct {
	Name string
	Path string
	CRD  *apiextv1beta1.CustomResourceDefinition
}

// Config contains all the information needed to handle different
// metering deployment configurations and internal states, e.g. what
// platform to deploy on, whether or not to delete the metering CRDs,
// or namespace during an install, the location to the manifests dir, etc.
type Config struct {
	Namespace                string
	Platform                 string
	DeployManifestsDirectory string
	MeteringCR               string
	SkipMeteringDeployment   bool
	DeleteCRDs               bool
	DeleteCRB                bool
	DeleteNamespace          bool
	DeletePVCs               bool
	DeleteAll                bool
	Repo                     string
	Tag                      string
}

// Deployer holds all the information needed to handle the deployment
// process of the metering stack. This includes the clientsets needed
// to provision and remove all the metering resources, and a customized
// deployment configuration.
type Deployer struct {
	config                           Config
	crds                             []CRD
	ansibleOperatorManifestsLocation string
	logger                           log.FieldLogger
	client                           *kubernetes.Clientset
	apiExtClient                     apiextclientv1beta1.CustomResourceDefinitionsGetter
	meteringClient                   *metering.MeteringV1Client
}

// NewDeployer creates a new reference to a deploy structure, and then calls helper
// functions that initialize the structure fields based on the value of
// environment variables and function parameters, returning a reference to this initialized
// deploy structure
func NewDeployer(
	cfg Config,
	client *kubernetes.Clientset,
	apiextClient apiextclientv1beta1.CustomResourceDefinitionsGetter,
	meteringClient *metering.MeteringV1Client,
	logger log.FieldLogger,
) (*Deployer, error) {
	var err error

	deploy := &Deployer{
		client:         client,
		apiExtClient:   apiextClient,
		meteringClient: meteringClient,
		logger:         logger,
		config:         cfg,
	}

	if deploy.config.Namespace == "" {
		return deploy, fmt.Errorf("Failed to set $METERING_NAMESPACE or --namespace flag")
	}
	deploy.logger.Infof("Metering Deploy Namespace: %s", deploy.config.Namespace)

	if deploy.config.DeployManifestsDirectory == "" {
		return nil, fmt.Errorf("Failed to set the $DEPLOY_MANIFESTS_DIR or --deploy-manifests-dir flag to a non-empty value")
	}

	deployDir, err := filepath.Abs(deploy.config.DeployManifestsDirectory)
	if err != nil {
		return nil, fmt.Errorf("Failed to get the absolute path of the manifest/deploy directory %s: %v", deploy.config.DeployManifestsDirectory, err)
	}

	dirStat, err := os.Stat(deploy.config.DeployManifestsDirectory)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("Failed to get the stat the manifest/deploy directory %s: %v", deploy.config.DeployManifestsDirectory, err)
	}
	if !dirStat.IsDir() {
		return nil, fmt.Errorf("Specified deploy directory '%s' is not a directory", deploy.config.DeployManifestsDirectory)
	}

	var ansibleOperatorManifestDir string

	switch strings.ToLower(deploy.config.Platform) {
	case "upstream":
		ansibleOperatorManifestDir = filepath.Join(deployDir, upstreamManifestDirname, manifestAnsibleOperator)
	case "openshift":
		ansibleOperatorManifestDir = filepath.Join(deployDir, openshiftManifestDirname, manifestAnsibleOperator)
	case "ocp-testing":
		ansibleOperatorManifestDir = filepath.Join(deployDir, ocpTestingManifestDirname, manifestAnsibleOperator)
	default:
		return deploy, fmt.Errorf("Failed to set $DEPLOY_PLATFORM or --platform flag to a valid value. Supported platforms: [upstream, openshift, ocp-testing]")
	}

	dirStat, err = os.Stat(ansibleOperatorManifestDir)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("Failed to stat the %s deploy platform directory '%s': %v", deploy.config.Platform, ansibleOperatorManifestDir, err)
	}
	if !dirStat.IsDir() {
		return nil, fmt.Errorf("Specified %s deploy platform directory '%s' is not a directory", deploy.config.Platform, ansibleOperatorManifestDir)
	}

	if deploy.config.DeleteAll {
		deploy.config.DeletePVCs = true
		deploy.config.DeleteNamespace = true
		deploy.config.DeleteCRB = true
		deploy.config.DeleteCRDs = true
	}

	deploy.logger.Infof("Metering Deploy Platform: %s", deploy.config.Platform)
	deploy.ansibleOperatorManifestsLocation = ansibleOperatorManifestDir

	// initialize a slice of CRD structures and assign to the deploy.crds field
	// this is used by the install/uninstall drivers to manage the metering CRDs
	deploy.initMeteringCRDSlice()

	return deploy, nil
}

// Install is the driver function that manages the process of creating all
// the resources that metering needs to install: namespace, CRDs, etc.
func (deploy *Deployer) Install() error {
	err := deploy.installNamespace()
	if err != nil {
		return fmt.Errorf("Failed to create the %s namespace: %v", deploy.config.Namespace, err)
	}

	err = deploy.installMeteringCRDs()
	if err != nil {
		return fmt.Errorf("Failed to create the Metering CRDs: %v", err)
	}

	if !deploy.config.SkipMeteringDeployment {
		err = deploy.installMeteringResources()
		if err != nil {
			return fmt.Errorf("Failed to create the metering resources: %v", err)
		}
	}

	err = deploy.installMeteringConfig()
	if err != nil {
		return fmt.Errorf("Failed to create the MeteringConfig resource: %v", err)
	}

	return nil
}

// Uninstall is the driver function that manages deleting all the resources that
// metering had created. Depending on the configuration of the deploy structure,
// resources like the metering CRDs, PVCs, or cluster role/role binding may be skipped
func (deploy *Deployer) Uninstall() error {
	err := deploy.uninstallMeteringConfig()
	if err != nil {
		return fmt.Errorf("Failed to delete the MeteringConfig resource: %v", err)
	}

	err = deploy.uninstallMeteringResources()
	if err != nil {
		return fmt.Errorf("Failed to delete the metering resources: %v", err)
	}

	if deploy.config.DeleteCRDs {
		err = deploy.uninstallMeteringCRDs()
		if err != nil {
			return fmt.Errorf("Failed to delete the Metering CRDs: %v", err)
		}
	} else {
		deploy.logger.Infof("Skipped deleting the metering CRDs")
	}

	if deploy.config.DeleteNamespace {
		err = deploy.uninstallNamespace()
		if err != nil {
			return fmt.Errorf("Failed to delete the %s namespace: %v", deploy.config.Namespace, err)
		}
	} else {
		deploy.logger.Infof("Skipped deleting the %s namespace", deploy.config.Namespace)
	}

	return nil
}
