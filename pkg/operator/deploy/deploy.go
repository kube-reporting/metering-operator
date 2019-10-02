package deploy

import (
	"fmt"

	meteringv1 "github.com/operator-framework/operator-metering/pkg/apis/metering/v1"
	metering "github.com/operator-framework/operator-metering/pkg/generated/clientset/versioned/typed/metering/v1"
	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextclientv1beta1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

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

	hivetableFile        = "hive.crd.yaml"
	prestotableFile      = "prestotable.crd.yaml"
	meteringcon***REMOVED***gFile   = "meteringcon***REMOVED***g.crd.yaml"
	reportFile           = "report.crd.yaml"
	reportdatasourceFile = "reportdatasource.crd.yaml"
	reportqueryFile      = "reportquery.crd.yaml"
	storagelocationFile  = "storagelocation.crd.yaml"

	meteringDeploymentFile         = "metering-operator-deployment.yaml"
	meteringServiceAccountFile     = "metering-operator-service-account.yaml"
	meteringRoleBindingFile        = "metering-operator-rolebinding.yaml"
	meteringRoleFile               = "metering-operator-role.yaml"
	meteringClusterRoleBindingFile = "metering-operator-clusterrolebinding.yaml"
	meteringClusterRoleFile        = "metering-operator-clusterrole.yaml"
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
	SkipMeteringDeployment bool
	DeleteCRDs             bool
	DeleteCRB              bool
	DeleteNamespace        bool
	DeletePVCs             bool
	DeleteAll              bool
	Namespace              string
	Platform               string
	Repo                   string
	Tag                    string
	OperatorResources      *OperatorResources
	MeteringCon***REMOVED***g         *meteringv1.MeteringCon***REMOVED***g
}

// OperatorResources contains all the objects that make up the
// Metering Ansible Operator
type OperatorResources struct {
	CRDs               []CRD
	Deployment         *appsv1.Deployment
	ServiceAccount     *corev1.ServiceAccount
	RoleBinding        *rbacv1.RoleBinding
	Role               *rbacv1.Role
	ClusterRoleBinding *rbacv1.ClusterRoleBinding
	ClusterRole        *rbacv1.ClusterRole
}

// Deployer holds all the information needed to handle the deployment
// process of the metering stack. This includes the clientsets needed
// to provision and remove all the metering resources, and a customized
// deployment con***REMOVED***guration.
type Deployer struct {
	con***REMOVED***g         Con***REMOVED***g
	logger         log.FieldLogger
	client         kubernetes.Interface
	apiExtClient   apiextclientv1beta1.CustomResourceDe***REMOVED***nitionsGetter
	meteringClient metering.MeteringV1Interface
}

// NewDeployer creates a new reference to a deploy structure, and then calls helper
// functions that initialize the structure ***REMOVED***elds based on the value of
// environment variables and function parameters, returning a reference to this initialized
// deploy structure
func NewDeployer(
	cfg Con***REMOVED***g,
	logger log.FieldLogger,
	client kubernetes.Interface,
	apiextClient apiextclientv1beta1.CustomResourceDe***REMOVED***nitionsGetter,
	meteringClient metering.MeteringV1Interface,
) (*Deployer, error) {
	deploy := &Deployer{
		client:         client,
		apiExtClient:   apiextClient,
		meteringClient: meteringClient,
		logger:         logger,
		con***REMOVED***g:         cfg,
	}

	deploy.logger.Infof("Metering Deploy Namespace: %s", deploy.con***REMOVED***g.Namespace)
	deploy.logger.Infof("Metering Deploy Platform: %s", deploy.con***REMOVED***g.Platform)

	if deploy.con***REMOVED***g.DeleteAll {
		deploy.con***REMOVED***g.DeletePVCs = true
		deploy.con***REMOVED***g.DeleteNamespace = true
		deploy.con***REMOVED***g.DeleteCRB = true
		deploy.con***REMOVED***g.DeleteCRDs = true
	}

	if deploy.con***REMOVED***g.Namespace == "" {
		return deploy, fmt.Errorf("Failed to set $METERING_NAMESPACE or --namespace flag")
	}

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
