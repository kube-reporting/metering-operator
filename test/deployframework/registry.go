package deployframework

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	olmv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	packageManifestClientV1 "github.com/operator-framework/operator-lifecycle-manager/pkg/package-server/client/clientset/versioned/typed/operators/v1"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

func (df *DeployFramework) CreateCatalogSourceFromIndex(indexImage string) (string, string, error) {
	catalogSourceName := df.NamespacePrefix + "-" + DefaultCatalogSourceName

	catalogSource, err := df.OLMV1Alpha1Client.CatalogSources(registryDeployNamespace).Get(context.TODO(), catalogSourceName, metav1.GetOptions{})
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
				Image:       indexImage,
				Publisher:   "Red Hat",
				DisplayName: "Metering Dev",
			},
		}

		catalogSource, err = df.OLMV1Alpha1Client.CatalogSources(registryDeployNamespace).Create(context.TODO(), catsrc, metav1.CreateOptions{})
		if err != nil {
			return "", "", err
		}
		df.Logger.Infof("Created the metering CatalogSource using the %s index image in the %s namespace", indexImage, registryDeployNamespace)
	}

	if catalogSource.ObjectMeta.Name == "" || catalogSource.ObjectMeta.Namespace == "" {
		return "", "", fmt.Errorf("failed to get a non-empty catalogsource name and namespace")
	}

	return catalogSource.ObjectMeta.Name, catalogSource.ObjectMeta.Namespace, nil
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
	catalogSourceName := df.NamespacePrefix + "-" + DefaultCatalogSourceName

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
func (df *DeployFramework) DeleteRegistryResources(registryProvisioned bool, name, namespace string) error {
	var errArr []string

	if registryProvisioned {
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

// WaitForPackageManifest is a deployframework method responsible for ensuring
// the packagemanifest that gets created as a result of df.CreateRegistryResources.
// We can define "readiness" here by polling until the metering-ocp packagemanifest
// has the @subscriptionChannel present in one of the channels listed in the package.
// Note: in the case where an invalid subscription channel has been passed to the e2e
// suite, this would essentially act as a verification check as well.
func (df *DeployFramework) WaitForPackageManifest(name, namespace, subscriptionChannel string) error {
	// Start build up the packagemanifest typed clientset so we can
	// list off any packagemanifests that match our label selector
	packageClient, err := packageManifestClientV1.NewForConfig(df.Config)
	if err != nil {
		return fmt.Errorf("failed to initialize the packagemanifest clientset: %v", err)
	}

	labelSelector := fmt.Sprintf("catalog=%s,catalog-namespace=%s", name, namespace)
	err = wait.Poll(3*time.Second, 5*time.Minute, func() (done bool, err error) {
		df.Logger.Infof("Waiting for the metering-ocp packagemanifest to become ready")
		packages, err := packageClient.PackageManifests(namespace).List(context.TODO(), metav1.ListOptions{
			LabelSelector: labelSelector,
		})
		if len(packages.Items) == 0 {
			df.Logger.Debugf("No packages matched the %s label selector, re-polling...", labelSelector)
			return false, nil
		}

		var ready bool
		for _, p := range packages.Items {
			for _, channel := range p.Status.Channels {
				if channel.Name == subscriptionChannel {
					ready = true
				}
			}
		}
		if !ready {
			df.Logger.Warnf("The metering-ocp packagemanifest is present but the %s channel is not present", subscriptionChannel)
		}

		return ready, nil
	})
	if err != nil {
		return err
	}
	df.Logger.Infof("The metering-ocp packagemanifest is ready")

	return nil
}
