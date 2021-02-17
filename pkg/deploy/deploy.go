package deploy

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	metering "github.com/kube-reporting/metering-operator/pkg/apis/metering/v1"
	meteringclient "github.com/kube-reporting/metering-operator/pkg/generated/clientset/versioned/typed/metering/v1"
	olmclientv1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/clientset/versioned/typed/operators/v1"
	olmclientv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/clientset/versioned/typed/operators/v1alpha1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextclientv1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	manifestAnsibleOperator   = "metering-ansible-operator"
	upstreamManifestDirname   = "upstream"
	openshiftManifestDirname  = "openshift"
	ocpTestingManifestDirname = "ocp-testing"

	packageName            = "metering-ocp"
	catalogSourceName      = "redhat-operators"
	catalogSourceNamespace = "openshift-marketplace"

	crdPollTimeout = 5 * time.Minute
	crdInitialPoll = 1 * time.Second

	hivetableFile         = "hive.crd.yaml"
	prestotableFile       = "prestotable.crd.yaml"
	meteringconfigFile    = "meteringconfig.crd.yaml"
	reportFile            = "report.crd.yaml"
	reportdatasourceFile  = "reportdatasource.crd.yaml"
	reportqueryFile       = "reportquery.crd.yaml"
	storagelocationFile   = "storagelocation.crd.yaml"
	meteringconfigCRDName = "meteringconfigs.metering.openshift.io"

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
	CRD  *apiextv1.CustomResourceDefinition
}

// Config contains all the information needed to handle different
// metering deployment configurations and internal states, e.g. what
// platform to deploy on, whether or not to delete the metering CRDs,
// or namespace during an install, the location to the manifests dir, etc.
type Config struct {
	SkipMeteringDeployment   bool
	RunMeteringOperatorLocal bool
	DeleteCRDs               bool
	DeleteCRBs               bool
	DeleteNamespace          bool
	DeletePVCs               bool
	DeleteAll                bool
	Namespace                string
	Platform                 string
	Repo                     string
	Tag                      string
	Channel                  string
	PackageName              string
	CatalogSourceName        string
	CatalogSourceNamespace   string
	SubscriptionName         string
	ExtraNamespaceLabels     map[string]string
	OperatorResources        *OperatorResources
	MeteringConfig           *metering.MeteringConfig
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
	Config            Config
	Logger            logrus.FieldLogger
	Client            kubernetes.Interface
	APIExtClient      apiextclientv1.CustomResourceDefinitionsGetter
	MeteringClient    meteringclient.MeteringV1Interface
	OLMV1Client       olmclientv1.OperatorsV1Interface
	OLMV1Alpha1Client olmclientv1alpha1.OperatorsV1alpha1Interface
}

// NewDeployer creates a new reference to a deploy structure, and then calls helper
// functions that initialize the structure fields based on the value of
// environment variables and function parameters, returning a reference to this initialized
// deploy structure
func NewDeployer(
	cfg Config,
	logger logrus.FieldLogger,
	client kubernetes.Interface,
	apiextClient apiextclientv1.CustomResourceDefinitionsGetter,
	meteringClient meteringclient.MeteringV1Interface,
	olmV1Client olmclientv1.OperatorsV1Interface,
	olmV1Alpha1Client olmclientv1alpha1.OperatorsV1alpha1Interface,
) (*Deployer, error) {
	deploy := &Deployer{
		Client:            client,
		APIExtClient:      apiextClient,
		MeteringClient:    meteringClient,
		OLMV1Client:       olmV1Client,
		OLMV1Alpha1Client: olmV1Alpha1Client,
		Logger:            logger,
		Config:            cfg,
	}

	deploy.Logger.Infof("Metering Deploy Namespace: %s", deploy.Config.Namespace)
	deploy.Logger.Infof("Metering Deploy Platform: %s", deploy.Config.Platform)

	if deploy.Config.DeleteAll {
		deploy.Config.DeletePVCs = true
		deploy.Config.DeleteNamespace = true
		deploy.Config.DeleteCRBs = true
		deploy.Config.DeleteCRDs = true
	}
	if deploy.Config.Namespace == "" {
		return deploy, fmt.Errorf("failed to set $METERING_NAMESPACE or --namespace flag")
	}

	return deploy, nil
}

// InstallOLM is the driver function that manages the process of creating
// all of the OLM-related resources that are needed to deploy a Metering
// installation.
func (deploy *Deployer) InstallOLM() error {
	err := deploy.installNamespace()
	if err != nil {
		return fmt.Errorf("failed to create the %s namespace: %v", deploy.Config.Namespace, err)
	}

	err = deploy.installMeteringOperatorGroup()
	if err != nil {
		return fmt.Errorf("failed to create the metering OperatorGroup: %v", err)
	}

	err = deploy.installMeteringSubscription()
	if err != nil {
		return fmt.Errorf("failed to create the metering Subscription: %v", err)
	}

	err = deploy.installMeteringConfig()
	if err != nil {
		return fmt.Errorf("failed to create the MeteringConfig resource: %v", err)
	}

	return nil
}

// UninstallOLM is the driver function that manages deleting all of the OLM-related
// metering resources that get created: the OperatorGroup, Subscription, CSV, etc.
// Deleting the ClusterServiceVersion object will delete the resources that were
// provisioned from the CSV, so deleting the MeteringConfig custom resource is still
// needed to fully cleanup any non-CRD resources, e.g. the reporting-operator pod.
func (deploy *Deployer) UninstallOLM() error {
	err := deploy.uninstallMeteringConfig()
	if err != nil {
		return fmt.Errorf("failed to delete the MeteringConfig resource: %v", err)
	}

	err = deploy.uninstallMeteringCSV()
	if err != nil {
		return fmt.Errorf("failed to delete the metering ClusterServiceVersion: %v", err)
	}

	err = deploy.uninstallMeteringSubscription()
	if err != nil {
		return fmt.Errorf("failed to uninstall the metering Subscription: %v", err)
	}

	err = deploy.uninstallMeteringOperatorGroup()
	if err != nil {
		return fmt.Errorf("failed to uninstall the metering OperatorGroup: %v", err)
	}

	if deploy.Config.DeletePVCs {
		err = deploy.uninstallMeteringPVCs()
		if err != nil {
			return fmt.Errorf("failed to uninstall the Metering PVCs: %v", err)
		}
	}
	if deploy.Config.DeleteCRDs {
		err = deploy.uninstallMeteringCRDs()
		if err != nil {
			return fmt.Errorf("failed to uninstall the metering-related CRDs: %v", err)
		}
	}
	if deploy.Config.DeleteNamespace {
		err = deploy.uninstallNamespace()
		if err != nil {
			return fmt.Errorf("failed to uninstall the %s metering namespace: %v", deploy.Config.Namespace, err)
		}
	}
	if deploy.Config.DeleteCRBs {
		err = deploy.uninstallReportingOperatorClusterRole()
		if err != nil {
			return fmt.Errorf("failed to delete the reporting-operator ClusterRole resources: %v", err)
		}
		err = deploy.uninstallReportingOperatorClusterRoleBinding()
		if err != nil {
			return fmt.Errorf("failed to delete the reporting-operator ClusterRoleBinding resources: %v", err)
		}
	}

	return nil
}

// Install is the driver function that manages the process of creating all
// the resources that metering needs to install: namespace, CRDs, etc.
func (deploy *Deployer) Install() error {
	err := deploy.installNamespace()
	if err != nil {
		return fmt.Errorf("failed to create the %s namespace: %v", deploy.Config.Namespace, err)
	}

	err = deploy.installMeteringCRDs()
	if err != nil {
		return fmt.Errorf("failed to create the Metering CRDs: %v", err)
	}

	if !deploy.Config.SkipMeteringDeployment {
		err = deploy.installMeteringResources()
		if err != nil {
			return fmt.Errorf("failed to create the metering resources: %v", err)
		}
	}

	err = deploy.installMeteringConfig()
	if err != nil {
		return fmt.Errorf("failed to create the MeteringConfig resource: %v", err)
	}

	return nil
}

// Uninstall is the driver function that manages deleting all the resources that
// metering had created. Depending on the configuration of the deploy structure,
// resources like the metering CRDs, PVCs, or cluster role/role binding may be skipped
func (deploy *Deployer) Uninstall() error {
	err := deploy.uninstallMeteringConfig()
	if err != nil {
		return fmt.Errorf("failed to delete the MeteringConfig resource: %v", err)
	}

	err = deploy.uninstallMeteringResources()
	if err != nil {
		return fmt.Errorf("failed to delete the metering resources: %v", err)
	}

	if deploy.Config.DeleteCRDs {
		err = deploy.uninstallMeteringCRDs()
		if err != nil {
			return fmt.Errorf("failed to delete the Metering CRDs: %v", err)
		}
	} else {
		deploy.Logger.Infof("Skipped deleting the metering CRDs")
	}

	if deploy.Config.DeleteNamespace {
		err = deploy.uninstallNamespace()
		if err != nil {
			return fmt.Errorf("failed to delete the %s namespace: %v", deploy.Config.Namespace, err)
		}
	} else {
		deploy.Logger.Infof("Skipped deleting the %s namespace", deploy.Config.Namespace)
	}

	return nil
}
