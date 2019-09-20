package deploy

import (
	"fmt"
	"os"
	"path/***REMOVED***lepath"
	"strings"

	metering "github.com/operator-framework/operator-metering/pkg/generated/clientset/versioned/typed/metering/v1"
	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextclientv1beta1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"

	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
)

const (
	defaultMeteringCon***REMOVED***g     = "meteringcon***REMOVED***g.yaml"
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
	meteringcon***REMOVED***gFile   = "meteringcon***REMOVED***g.crd.yaml"
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
	CRD  *apiextv1beta1.CustomResourceDe***REMOVED***nition
}

// Con***REMOVED***g contains all the information needed to handle different
// metering deployment con***REMOVED***gurations and internal states, e.g. what
// platform to deploy on, whether or not to delete the metering CRDs,
// or namespace during an install, the location to the manifests dir, etc.
type Con***REMOVED***g struct {
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
// deployment con***REMOVED***guration.
type Deployer struct {
	con***REMOVED***g                           Con***REMOVED***g
	crds                             []CRD
	ansibleOperatorManifestsLocation string
	logger                           log.FieldLogger
	client                           *kubernetes.Clientset
	apiExtClient                     apiextclientv1beta1.CustomResourceDe***REMOVED***nitionsGetter
	meteringClient                   *metering.MeteringV1Client
}

// NewDeployer creates a new reference to a deploy structure, and then calls helper
// functions that initialize the structure ***REMOVED***elds based on the value of
// environment variables and function parameters, returning a reference to this initialized
// deploy structure
func NewDeployer(
	cfg Con***REMOVED***g,
	client *kubernetes.Clientset,
	apiextClient apiextclientv1beta1.CustomResourceDe***REMOVED***nitionsGetter,
	meteringClient *metering.MeteringV1Client,
	logger log.FieldLogger,
) (*Deployer, error) {
	var err error

	deploy := &Deployer{
		client:         client,
		apiExtClient:   apiextClient,
		meteringClient: meteringClient,
		logger:         logger,
		con***REMOVED***g:         cfg,
	}

	if deploy.con***REMOVED***g.Namespace == "" {
		return deploy, fmt.Errorf("Failed to set $METERING_NAMESPACE or --namespace flag")
	}
	deploy.logger.Infof("Metering Deploy Namespace: %s", deploy.con***REMOVED***g.Namespace)

	if deploy.con***REMOVED***g.DeployManifestsDirectory == "" {
		return nil, fmt.Errorf("Failed to set the $DEPLOY_MANIFESTS_DIR or --deploy-manifests-dir flag to a non-empty value")
	}

	deployDir, err := ***REMOVED***lepath.Abs(deploy.con***REMOVED***g.DeployManifestsDirectory)
	if err != nil {
		return nil, fmt.Errorf("Failed to get the absolute path of the manifest/deploy directory %s: %v", deploy.con***REMOVED***g.DeployManifestsDirectory, err)
	}

	dirStat, err := os.Stat(deploy.con***REMOVED***g.DeployManifestsDirectory)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("Failed to get the stat the manifest/deploy directory %s: %v", deploy.con***REMOVED***g.DeployManifestsDirectory, err)
	}
	if !dirStat.IsDir() {
		return nil, fmt.Errorf("Speci***REMOVED***ed deploy directory '%s' is not a directory", deploy.con***REMOVED***g.DeployManifestsDirectory)
	}

	var ansibleOperatorManifestDir string

	switch strings.ToLower(deploy.con***REMOVED***g.Platform) {
	case "upstream":
		ansibleOperatorManifestDir = ***REMOVED***lepath.Join(deployDir, upstreamManifestDirname, manifestAnsibleOperator)
	case "openshift":
		ansibleOperatorManifestDir = ***REMOVED***lepath.Join(deployDir, openshiftManifestDirname, manifestAnsibleOperator)
	case "ocp-testing":
		ansibleOperatorManifestDir = ***REMOVED***lepath.Join(deployDir, ocpTestingManifestDirname, manifestAnsibleOperator)
	default:
		return deploy, fmt.Errorf("Failed to set $DEPLOY_PLATFORM or --platform flag to a valid value. Supported platforms: [upstream, openshift, ocp-testing]")
	}

	dirStat, err = os.Stat(ansibleOperatorManifestDir)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("Failed to stat the %s deploy platform directory '%s': %v", deploy.con***REMOVED***g.Platform, ansibleOperatorManifestDir, err)
	}
	if !dirStat.IsDir() {
		return nil, fmt.Errorf("Speci***REMOVED***ed %s deploy platform directory '%s' is not a directory", deploy.con***REMOVED***g.Platform, ansibleOperatorManifestDir)
	}

	if deploy.con***REMOVED***g.DeleteAll {
		deploy.con***REMOVED***g.DeletePVCs = true
		deploy.con***REMOVED***g.DeleteNamespace = true
		deploy.con***REMOVED***g.DeleteCRB = true
		deploy.con***REMOVED***g.DeleteCRDs = true
	}

	deploy.logger.Infof("Metering Deploy Platform: %s", deploy.con***REMOVED***g.Platform)
	deploy.ansibleOperatorManifestsLocation = ansibleOperatorManifestDir

	// initialize a slice of CRD structures and assign to the deploy.crds ***REMOVED***eld
	// this is used by the install/uninstall drivers to manage the metering CRDs
	deploy.initMeteringCRDSlice()

	return deploy, nil
}

// Install is the driver function that manages the process of creating all
// the resources that metering needs to install: namespace, CRDs, etc.
func (deploy *Deployer) Install() error {
	err := deploy.installNamespace()
	if err != nil {
		return fmt.Errorf("Failed to create the %s namespace: %v", deploy.con***REMOVED***g.Namespace, err)
	}

	err = deploy.installMeteringCRDs()
	if err != nil {
		return fmt.Errorf("Failed to create the Metering CRDs: %v", err)
	}

	if !deploy.con***REMOVED***g.SkipMeteringDeployment {
		err = deploy.installMeteringResources()
		if err != nil {
			return fmt.Errorf("Failed to create the metering resources: %v", err)
		}
	}

	err = deploy.installMeteringCon***REMOVED***g()
	if err != nil {
		return fmt.Errorf("Failed to create the MeteringCon***REMOVED***g resource: %v", err)
	}

	return nil
}

// Uninstall is the driver function that manages deleting all the resources that
// metering had created. Depending on the con***REMOVED***guration of the deploy structure,
// resources like the metering CRDs, PVCs, or cluster role/role binding may be skipped
func (deploy *Deployer) Uninstall() error {
	err := deploy.uninstallMeteringCon***REMOVED***g()
	if err != nil {
		return fmt.Errorf("Failed to delete the MeteringCon***REMOVED***g resource: %v", err)
	}

	err = deploy.uninstallMeteringResources()
	if err != nil {
		return fmt.Errorf("Failed to delete the metering resources: %v", err)
	}

	if deploy.con***REMOVED***g.DeleteCRDs {
		err = deploy.uninstallMeteringCRDs()
		if err != nil {
			return fmt.Errorf("Failed to delete the Metering CRDs: %v", err)
		}
	} ***REMOVED*** {
		deploy.logger.Infof("Skipped deleting the metering CRDs")
	}

	if deploy.con***REMOVED***g.DeleteNamespace {
		err = deploy.uninstallNamespace()
		if err != nil {
			return fmt.Errorf("Failed to delete the %s namespace: %v", deploy.con***REMOVED***g.Namespace, err)
		}
	} ***REMOVED*** {
		deploy.logger.Infof("Skipped deleting the %s namespace", deploy.con***REMOVED***g.Namespace)
	}

	return nil
}
