package deployframework

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"

	olmv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	olmclientv1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/clientset/versioned/typed/operators/v1"
	olmclientv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/clientset/versioned/typed/operators/v1alpha1"
	apiextclientv1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	metering "github.com/kube-reporting/metering-operator/pkg/apis/metering/v1"
	"github.com/kube-reporting/metering-operator/pkg/deploy"
	meteringclient "github.com/kube-reporting/metering-operator/pkg/generated/clientset/versioned/typed/metering/v1"
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

	defaultTargetPods        = 7
	defaultPlatform          = "openshift"
	defaultDeleteNamespace   = true
	defaultSubscriptionName  = "metering-ocp"
	defaultCatalogSourceName = "metering-dev-catalogsource"
	defaultPackageName       = "metering-ocp"

	manifestsDeployDir = "manifests/deploy"
	olmManifestsDir    = "olm_deploy/manifests/"
	hackScriptDirName  = "hack"

	registryServicePort            = 50051
	registryDeploymentManifestName = "deployment.yaml"
	registryServiceManifestName    = "service.yaml"
	registryDeployNamespace        = "openshift-marketplace"
	registryLabelSelector          = "registry.operator.metering-ansible=true"
)

// DeployFramework contains all the information necessary to deploy
// different metering instances and run tests against them
type DeployFramework struct {
	RunLocal          bool
	RunDevSetup       bool
	KubeConfigPath    string
	NamespacePrefix   string
	RepoDir           string
	RepoVersion       string
	OperatorResources *deploy.OperatorResources
	Logger            logrus.FieldLogger
	Config            *rest.Config
	Client            kubernetes.Interface
	APIExtClient      apiextclientv1.CustomResourceDefinitionsGetter
	MeteringClient    meteringclient.MeteringV1Interface
	OLMV1Client       olmclientv1.OperatorsV1Interface
	OLMV1Alpha1Client olmclientv1alpha1.OperatorsV1alpha1Interface
}

// New is the constructor function that creates and returns a new DeployFramework object
func New(logger logrus.FieldLogger, runLocal, runDevSetup bool, nsPrefix, repoDir, repoVersion, kubeconfig string) (*DeployFramework, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to build a kube config from %s: %v", kubeconfig, err)
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize the k8s clientset: %v", err)
	}

	apiextClient, err := apiextclientv1.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize the apiextensions clientset: %v", err)
	}

	meteringClient, err := meteringclient.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize the metering clientset: %v", err)
	}

	olmV1Client, err := olmclientv1.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize the v1 OLM clientset: %v", err)
	}

	olmV1Alpha1Client, err := olmclientv1alpha1.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize the v1alpha OLM clientset: %v", err)
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
		RepoVersion:       repoVersion,
		NamespacePrefix:   nsPrefix,
		RunLocal:          runLocal,
		RunDevSetup:       runDevSetup,
		Logger:            logger,
		Config:            config,
		Client:            client,
		APIExtClient:      apiextClient,
		MeteringClient:    meteringClient,
		OLMV1Client:       olmV1Client,
		OLMV1Alpha1Client: olmV1Alpha1Client,
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
	reportingOperatorImageTag,
	catalogSourceName,
	catalogSourceNamespace,
	subscriptionChannel string,
	spec metering.MeteringConfigSpec,
) (*deploy.Config, error) {
	meteringConfig := &metering.MeteringConfig{
		ObjectMeta: metav1.ObjectMeta{
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
		Namespace:              namespace,
		Repo:                   meteringOperatorImageRepo,
		Tag:                    meteringOperatorImageTag,
		Platform:               defaultPlatform,
		DeleteNamespace:        defaultDeleteNamespace,
		SubscriptionName:       defaultSubscriptionName,
		PackageName:            defaultPackageName,
		CatalogSourceName:      catalogSourceName,
		CatalogSourceNamespace: catalogSourceNamespace,
		Channel:                subscriptionChannel,
		ExtraNamespaceLabels: map[string]string{
			"name": df.NamespacePrefix + "-" + testNamespaceLabel,
		},
		OperatorResources:        df.OperatorResources,
		RunMeteringOperatorLocal: df.RunLocal,
		MeteringConfig:           meteringConfig,
	}, nil
}

// CreateRegistryResources is a deployframework method responsible
// for instantiating a new CatalogSource that can be used
// throughout individual Metering installations.
func (df *DeployFramework) CreateRegistryResources(registryImage, meteringOperatorImage, reportingOperatorImage string) (string, string, error) {
	// Create the registry Service object responsible for exposing the 50051 grpc port.
	// We're interested in the spec.ClusterIP for this object as we need that value to
	// use in the `spec.addr` field of the CatalogSource we're creating later.
	serviceManifestPath := filepath.Join(df.RepoDir, olmManifestsDir, registryServiceManifestName)
	addr, err := CreateRegistryService(df.Logger, df.Client, df.NamespacePrefix, serviceManifestPath, registryDeployNamespace)
	if err != nil {
		return "", "", fmt.Errorf("failed to create the registry service manifest in the %s namespace: %v", err, registryDeployNamespace)
	}
	if addr == "" {
		return "", "", fmt.Errorf("the registry service spec.ClusterIP returned is empty")
	}

	deploymentManifestPath := filepath.Join(df.RepoDir, olmManifestsDir, registryDeploymentManifestName)
	err = CreateRegistryDeployment(df.Logger, df.Client, df.NamespacePrefix, deploymentManifestPath, registryImage, meteringOperatorImage, reportingOperatorImage, registryDeployNamespace)
	if err != nil {
		return "", "", fmt.Errorf("failed to create the registry deployment manifest in the %s namespace: %v", err, registryDeployNamespace)
	}

	var catalogSource *olmv1alpha1.CatalogSource
	catalogSourceName := df.NamespacePrefix + "-" + defaultCatalogSourceName

	catalogSource, err = df.OLMV1Alpha1Client.CatalogSources(registryDeployNamespace).Get(context.TODO(), catalogSourceName, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return "", "", err
	}
	if apierrors.IsNotFound(err) {
		catsrc := &olmv1alpha1.CatalogSource{
			ObjectMeta: metav1.ObjectMeta{
				Name:      catalogSourceName,
				Namespace: registryDeployNamespace,
				Labels: map[string]string{
					"name": df.NamespacePrefix + "-" + testNamespaceLabel,
				},
			},
			Spec: olmv1alpha1.CatalogSourceSpec{
				SourceType:  olmv1alpha1.SourceTypeGrpc,
				Address:     fmt.Sprintf("%s:%d", addr, registryServicePort),
				Publisher:   "Red Hat",
				DisplayName: "Metering Dev",
			},
		}

		catalogSource, err = df.OLMV1Alpha1Client.CatalogSources(registryDeployNamespace).Create(context.TODO(), catsrc, metav1.CreateOptions{})
		if err != nil {
			return "", "", err
		}
		df.Logger.Infof("Created the metering CatalogSource using the %s registry image in the %s namespace", registryImage, registryDeployNamespace)
	}

	if catalogSource.ObjectMeta.Name == "" || catalogSource.ObjectMeta.Namespace == "" {
		return "", "", fmt.Errorf("failed to get a non-empty catalogsource name and namespace")
	}

	return catalogSource.ObjectMeta.Name, catalogSource.ObjectMeta.Namespace, nil
}

// DeleteRegistryResources is a deployframework method responsible
// for cleaning up any registry resources that were created during the
// execution of the testing suite. Note: we add a label to the registry
// service and deployment manifests to help distinguish between resources
// created by a particular developer, which is reflected in the label
// selector that we pass to the helper functions that do the heavy-lifting.
func (df *DeployFramework) DeleteRegistryResources(name, namespace string) error {
	var errArr []string

	// Start building up the label selectors for searching for the registry
	// resources that we created. We inject a testing label to both of those
	// resources, e.g. `name=tflannag-metering-testing-ns`.
	testingRegistryLabelSelector := fmt.Sprintf("name=%s-%s", df.NamespacePrefix, testNamespaceLabel)
	registryLabelSelector := fmt.Sprintf("%s,%s", registryLabelSelector, testingRegistryLabelSelector)

	err := DeleteRegistryDeployment(df.Logger, df.Client, namespace, registryLabelSelector)
	if err != nil {
		errArr = append(errArr, fmt.Sprintf("failed to successfully delete the registry deployments(s): %v", err))
	}

	err = DeleteRegistryService(df.Logger, df.Client, namespace, registryLabelSelector)
	if err != nil {
		errArr = append(errArr, fmt.Sprintf("failed to successfully delete the registry service(s): %v", err))
	}

	catsrc, err := df.OLMV1Alpha1Client.CatalogSources(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil || apierrors.IsNotFound(err) {
		errArr = append(errArr, fmt.Sprintf("failed to successfully get the %s CatalogSource in the %s namespace: %v", name, namespace, err))
	}

	err = df.OLMV1Alpha1Client.CatalogSources(catsrc.Namespace).Delete(context.TODO(), catsrc.Name, metav1.DeleteOptions{})
	if err != nil {
		errArr = append(errArr, fmt.Sprintf("failed to successfully delete the %s CatalogSource in the %s namespace: %v", name, namespace, err))
	}
	df.Logger.Infof("Deleted the %s CatalogSource in the %s namespace", catsrc.Name, catsrc.Namespace)

	if len(errArr) != 0 {
		return fmt.Errorf(strings.Join(errArr, "\n"))
	}

	return nil
}
