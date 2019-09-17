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
	Namespace              string
	Platform               string
	ManifestLocation       string
	MeteringCR             string
	SkipMeteringDeployment bool
	DeleteCRDs             bool
	DeleteCRB              bool
	DeleteNamespace        bool
	DeletePVCs             bool
	DeleteAll              bool
	Repo                   string
	Tag                    string
}

// Deployer holds all the information needed to handle the deployment
// process of the metering stack. This includes the clientsets needed
// to provision and remove all the metering resources, and a customized
// deployment configuration.
type Deployer struct {
	config         Config
	crds           []CRD
	logger         log.FieldLogger
	client         *kubernetes.Clientset
	apiExtClient   apiextclientv1beta1.CustomResourceDefinitionsGetter
	meteringClient *metering.MeteringV1Client
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

	meteringNamespace := os.Getenv("METERING_NAMESPACE")
	if meteringNamespace == "" {
		return nil, fmt.Errorf("Failed to set $METERING_NAMESPACE")
	}

	deploy.config.Namespace = meteringNamespace
	deploy.logger.Infof("Metering Deploy Namespace: %s", deploy.config.Namespace)

	manifestOverrideLocation := os.Getenv("INSTALLER_MANIFESTS_DIR")
	if manifestOverrideLocation != "" {
		deploy.config.ManifestLocation, err = filepath.Abs(manifestOverrideLocation)
		if err != nil {
			return nil, fmt.Errorf("Failed to override the manifest location: %v", err)
		}

		deploy.logger.Infof("Overrided manifest location: %s", deploy.config.ManifestLocation)
	} else {
		deployPlatform := os.Getenv("DEPLOY_PLATFORM")
		if deployPlatform == "" {
			deploy.config.Platform = "openshift"
		} else {
			deploy.config.Platform = deployPlatform
		}

		defaultManifestBase, err := filepath.Abs(manifestDeployDirname)
		if err != nil {
			return nil, fmt.Errorf("Failed to get the absolute path of the manifest/deploy directory: %v", err)
		}

		switch strings.ToLower(deploy.config.Platform) {
		case "upstream":
			deploy.config.ManifestLocation = filepath.Join(defaultManifestBase, upstreamManifestDirname, manifestAnsibleOperator)
		case "openshift":
			deploy.config.ManifestLocation = filepath.Join(defaultManifestBase, openshiftManifestDirname, manifestAnsibleOperator)
		case "ocp-testing":
			deploy.config.ManifestLocation = filepath.Join(defaultManifestBase, ocpTestingManifestDirname, manifestAnsibleOperator)
		default:
			return nil, fmt.Errorf("Failed to set $DEPLOY_PLATFORM to an invalid value. Supported types: [upstream, openshift, ocp-testing]")
		}

		deploy.logger.Infof("Metering Deploy Platform: %s", deploy.config.Platform)
	}

	meteringCRFile := os.Getenv("METERING_CR_FILE")
	if meteringCRFile == "" {
		deploy.logger.Info("The $METERING_CR_FILE env var is unset, using the default MeteringConfig manifest")
		deploy.config.MeteringCR = deploy.config.ManifestLocation + defaultMeteringConfig
	} else {
		deploy.config.MeteringCR = meteringCRFile
	}

	deploy.config.DeleteAll, err = getBoolEnv("METERING_DELETE_ALL", false)
	if err != nil {
		return nil, fmt.Errorf("Failed to read $METERING_DELETE_ALL: %v", err)
	}

	if deploy.config.DeleteAll {
		deploy.config.DeleteCRDs = true
		deploy.config.DeleteCRB = true
		deploy.config.DeleteNamespace = true
		deploy.config.DeletePVCs = true
	} else {
		deploy.config.DeleteCRDs, err = getBoolEnv("METERING_DELETE_CRDS", false)
		if err != nil {
			return nil, fmt.Errorf("Failed to read $METERING_DELETE_CRDS: %v", err)
		}

		deploy.config.DeleteCRB, err = getBoolEnv("METERING_DELETE_CRB", false)
		if err != nil {
			return nil, fmt.Errorf("Failed to read $METERING_DELETE_CRB: %v", err)
		}

		deploy.config.DeleteNamespace, err = getBoolEnv("METERING_DELETE_NAMESPACE", false)
		if err != nil {
			return nil, fmt.Errorf("Failed to read $METERING_DELETE_NAMESPACE: %v", err)
		}

		deploy.config.DeletePVCs, err = getBoolEnv("METERING_DELETE_PVCS", true)
		if err != nil {
			return nil, fmt.Errorf("Failed to read $METERING_DELETE_PVCS: %v", err)
		}
	}

	deploy.config.SkipMeteringDeployment, err = getBoolEnv("SKIP_METERING_OPERATOR_DEPLOYMENT", false)
	if err != nil {
		return nil, fmt.Errorf("Failed to read $SKIP_METERING_OPERATOR_DEPLOYMENT: %v", err)
	}

	imageRepo := os.Getenv("METERING_OPERATOR_IMAGE_REPO")
	if imageRepo != "" {
		deploy.config.Repo = imageRepo
	}

	imageTag := os.Getenv("METERING_OPERATOR_IMAGE_TAG")
	if imageTag != "" {
		deploy.config.Tag = imageTag
	}

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
