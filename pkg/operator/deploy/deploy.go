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
	"k8s.io/client-go/tools/clientcmd"
)

const (
	defaultMeteringConfig     = "meteringconfig.yaml"
	manifestDeployDirname     = "manifests/deploy"
	manifestAnsibleOperator   = "metering-ansible-operator"
	upstreamManifestDirname   = "upstream"
	openshiftManifestDirname  = "openshift"
	ocpTestingManifestDirname = "ocp-testing"

	hivetableFile        = "hive.crd.yaml"
	prestotableFile      = "prestotable.crd.yaml"
	meteringconfigFile   = "meteringconfig.crd.yaml"
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
	CRD  *apiextv1beta1.CustomResourceDefinition
}

// Config contains all the information needed to handle different
// metering deployment configurations and internal states, e.g. what
// platform to deploy on, whether or not to delete the metering CRDs,
// or namespace during an install, the location to the manifests dir, etc.
type Config struct {
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
	MeteringConfig         *meteringv1.MeteringConfig
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
// deployment configuration.
type Deployer struct {
	config         Config
	logger         log.FieldLogger
	client         kubernetes.Interface
	apiExtClient   apiextclientv1beta1.CustomResourceDefinitionsGetter
	meteringClient metering.MeteringV1Interface
}

// NewDeployer creates a new reference to a deploy structure, and then calls helper
// functions that initialize the structure fields based on the value of
// environment variables and function parameters, returning a reference to this initialized
// deploy structure
func NewDeployer(
	cfg Config,
	logger log.FieldLogger,
) (*Deployer, error) {
	kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)

	restconfig, err := kubeconfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize the kubernetes client config: %v", err)
	}

	client, err := kubernetes.NewForConfig(restconfig)
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize the kubernetes clientset: %v", err)
	}

	apiextClient, err := apiextclientv1beta1.NewForConfig(restconfig)
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize the apiextensions clientset: %v", err)
	}

	meteringClient, err := metering.NewForConfig(restconfig)
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize the metering clientset: %v", err)
	}

	deploy := &Deployer{
		client:         client,
		apiExtClient:   apiextClient,
		meteringClient: meteringClient,
		logger:         logger,
		config:         cfg,
	}

	deploy.logger.Infof("Metering Deploy Namespace: %s", deploy.config.Namespace)
	deploy.logger.Infof("Metering Deploy Platform: %s", deploy.config.Platform)

	if deploy.config.DeleteAll {
		deploy.config.DeletePVCs = true
		deploy.config.DeleteNamespace = true
		deploy.config.DeleteCRB = true
		deploy.config.DeleteCRDs = true
	}

	if deploy.config.Namespace == "" {
		return deploy, fmt.Errorf("Failed to set $METERING_NAMESPACE or --namespace flag")
	}

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
