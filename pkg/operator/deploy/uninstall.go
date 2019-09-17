package deploy

import (
	"fmt"
	"path/filepath"

	meteringv1 "github.com/operator-framework/operator-metering/pkg/apis/metering/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (deploy *Deployer) uninstallNamespace() error {
	err := deploy.Client.CoreV1().Namespaces().Delete(deploy.Namespace, &metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		return fmt.Errorf("The %s namespace doesn't exist", deploy.Namespace)
	} else if err == nil {
		deploy.Logger.Infof("Deleted the %s namespace", deploy.Namespace)
	} else {
		return fmt.Errorf("Failed to delete the %s namespace: %v", deploy.Namespace, err)
	}

	return nil
}

func (deploy *Deployer) uninstallMeteringConfig() error {
	var res meteringv1.MeteringConfig

	err := decodeYAMLManifestToObject(deploy.MeteringCR, &res)
	if err != nil {
		return fmt.Errorf("Failed while attempting to build up the MeteringConfig from the YAML file, got: %v", err)
	}

	err = deploy.MeteringClient.MeteringConfigs(deploy.Namespace).Delete(res.Name, &metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		deploy.Logger.Warnf("The MeteringConfig resource doesn't exist")
	} else if err == nil {
		deploy.Logger.Infof("Deleted the MeteringConfig resource")
	} else {
		return fmt.Errorf("Failed to delete the MeteringConfig resource: %v", err)
	}

	return nil
}

func (deploy *Deployer) uninstallMeteringResources() error {
	err := deploy.uninstallMeteringDeployment(filepath.Join(deploy.ManifestLocation, meteringDeploymentFile))
	if err != nil {
		return fmt.Errorf("Failed to delete the metering service account: %v", err)
	}

	err = deploy.uninstallMeteringServiceAccount(filepath.Join(deploy.ManifestLocation, meteringServiceAccountFile))
	if err != nil {
		return fmt.Errorf("Failed to delete the metering service account: %v", err)
	}

	err = deploy.uninstallMeteringRole(filepath.Join(deploy.ManifestLocation, meteringRoleFile))
	if err != nil {
		return fmt.Errorf("Failed to delete the metering role: %v", err)
	}

	err = deploy.uninstallMeteringRoleBinding(filepath.Join(deploy.ManifestLocation, meteringRoleBindingFile))
	if err != nil {
		return fmt.Errorf("Failed to delete the metering role binding: %v", err)
	}

	if deploy.DeleteCRB {
		err = deploy.uninstallMeteringClusterRole(filepath.Join(deploy.ManifestLocation, meteringClusterRoleFile))
		if err != nil {
			return fmt.Errorf("Failed to delete the metering cluster role: %v", err)
		}

		err = deploy.uninstallMeteringClusterRoleBinding(filepath.Join(deploy.ManifestLocation, meteringClusterRoleBindingFile))
		if err != nil {
			return fmt.Errorf("Failed to delete the metering cluster role binding: %v", err)
		}
	} else {
		deploy.Logger.Infof("Skipped deleting the metering cluster role resources")
	}

	if deploy.DeletePVCs {
		err = deploy.uninstallMeteringPVCs()
		if err != nil {
			return fmt.Errorf("Failed to delete the metering PVCs: %v", err)
		}
	} else {
		deploy.Logger.Infof("Skipped deleting the metering PVCs")
	}

	return nil
}

// uninstallMeteringPVCs gets a list of all the PVCs associated with the hdfs and hive-metastore
// pods in the $METERING_NAMESPACE namespace, and attempts to delete all the PVCs that match that list criteria
func (deploy *Deployer) uninstallMeteringPVCs() error {
	// Attempt to get a list of PVCs that match the hdfs or hive labels
	pvcs, err := deploy.Client.CoreV1().PersistentVolumeClaims(deploy.Namespace).List(metav1.ListOptions{
		LabelSelector: "app in (hdfs,hive)",
	})
	if err != nil {
		return fmt.Errorf("Failed to list all the metering PVCs in the %s namespace: %v", deploy.Namespace, err)
	}

	if len(pvcs.Items) == 0 {
		deploy.Logger.Warnf("The Hive/HDFS PVCs don't exist")
		return nil
	}

	for _, pvc := range pvcs.Items {
		err = deploy.Client.CoreV1().PersistentVolumeClaims(deploy.Namespace).Delete(pvc.Name, &metav1.DeleteOptions{})
		if err != nil {
			return fmt.Errorf("Failed to delete the PVC %s: %v", pvc.Name, err)
		}
	}

	deploy.Logger.Infof("Deleted the PVCs managed by metering")

	return nil
}

func (deploy *Deployer) uninstallMeteringDeployment(deploymentName string) error {
	var res appsv1.Deployment

	err := decodeYAMLManifestToObject(deploymentName, &res)
	if err != nil {
		return fmt.Errorf("Failed to decode the YAML manifest: %v", err)
	}

	err = deploy.Client.AppsV1().Deployments(deploy.Namespace).Delete(res.Name, &metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		deploy.Logger.Warnf("The metering deployment doesn't exist")
	} else if err == nil {
		deploy.Logger.Infof("Deleted the metering deployment")
	} else {
		return fmt.Errorf("Failed to delete the metering deployment: %v", err)
	}

	return nil
}

func (deploy *Deployer) uninstallMeteringServiceAccount(serviceAccountPath string) error {
	var res corev1.ServiceAccount

	err := decodeYAMLManifestToObject(serviceAccountPath, &res)
	if err != nil {
		return fmt.Errorf("Failed to decode the YAML manifest: %v", err)
	}

	err = deploy.Client.CoreV1().ServiceAccounts(deploy.Namespace).Delete(res.Name, &metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		deploy.Logger.Warnf("The metering service account doesn't exist")
	} else if err == nil {
		deploy.Logger.Infof("Deleted the metering serviceaccount")
	} else {
		return fmt.Errorf("Failed to delete the metering serviceaccount: %v", err)
	}

	return nil
}

func (deploy *Deployer) uninstallMeteringRoleBinding(roleBindingPath string) error {
	var res rbacv1.RoleBinding

	err := decodeYAMLManifestToObject(roleBindingPath, &res)
	if err != nil {
		return fmt.Errorf("Failed to decode the YAML manifest: %v", err)
	}

	res.Name = deploy.Namespace + "-" + res.Name
	res.RoleRef.Name = res.Name
	res.Namespace = deploy.Namespace

	for index := range res.Subjects {
		res.Subjects[index].Namespace = deploy.Namespace
	}

	err = deploy.Client.RbacV1().RoleBindings(deploy.Namespace).Delete(res.Name, &metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		deploy.Logger.Warnf("The metering role binding doesn't exist")
	} else if err == nil {
		deploy.Logger.Infof("Deleted the metering role binding")
	} else {
		return fmt.Errorf("Failed to delete the metering role binding: %v", err)
	}

	return nil
}

func (deploy *Deployer) uninstallMeteringRole(rolePath string) error {
	var res rbacv1.Role

	err := decodeYAMLManifestToObject(rolePath, &res)
	if err != nil {
		return fmt.Errorf("Failed to decode the YAML manifest: %v", err)
	}

	res.Name = deploy.Namespace + "-" + res.Name
	res.Namespace = deploy.Namespace

	err = deploy.Client.RbacV1().Roles(deploy.Namespace).Delete(res.Name, &metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		deploy.Logger.Warnf("The metering role doesn't exist")
	} else if err == nil {
		deploy.Logger.Infof("Deleted the metering role")
	} else {
		return fmt.Errorf("Failed to delete the metering role: %v", err)
	}

	return nil
}

func (deploy *Deployer) uninstallMeteringClusterRole(clusterrolePath string) error {
	var res rbacv1.ClusterRole

	err := decodeYAMLManifestToObject(clusterrolePath, &res)
	if err != nil {
		return fmt.Errorf("Failed to decode the YAML manifest: %v", err)
	}

	res.Name = deploy.Namespace + "-" + res.Name

	err = deploy.Client.RbacV1().ClusterRoles().Delete(res.Name, &metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		deploy.Logger.Warnf("The metering cluster role doesn't exist")
	} else if err == nil {
		deploy.Logger.Infof("Deleted the metering cluster role")
	} else {
		return fmt.Errorf("Failed to delete the metering cluster role: %v", err)
	}

	return nil
}

func (deploy *Deployer) uninstallMeteringClusterRoleBinding(meteringClusterRoleBindingFile string) error {
	var res rbacv1.ClusterRoleBinding

	err := decodeYAMLManifestToObject(meteringClusterRoleBindingFile, &res)
	if err != nil {
		return fmt.Errorf("Failed to decode the YAML manifest: %v", err)
	}

	res.Name = deploy.Namespace + "-" + res.Name
	res.RoleRef.Name = res.Name

	for index := range res.Subjects {
		res.Subjects[index].Namespace = deploy.Namespace
	}

	err = deploy.Client.RbacV1().ClusterRoleBindings().Delete(res.Name, &metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		deploy.Logger.Warnf("The metering cluster role binding doesn't exist")
	} else if err == nil {
		deploy.Logger.Infof("Deleted the metering cluster role binding")
	} else {
		return fmt.Errorf("Failed to delete the metering cluster role binding: %v", err)
	}

	return nil
}

func (deploy *Deployer) uninstallMeteringCRDs() error {
	for _, crd := range deploy.CRDs {
		err := deploy.uninstallMeteringCRD(crd)
		if err != nil {
			return fmt.Errorf("Failed to delete a CRD while looping: %v", err)
		}
	}

	return nil
}

func (deploy *Deployer) uninstallMeteringCRD(resource CRD) error {
	err := decodeYAMLManifestToObject(resource.Path, resource.CRD)
	if err != nil {
		return fmt.Errorf("Failed to decode the YAML manifest: %v", err)
	}

	err = deploy.APIExtClient.CustomResourceDefinitions().Delete(resource.Name, &metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		deploy.Logger.Warnf("The %s CRD doesn't exist", resource.Name)
	} else if err == nil {
		deploy.Logger.Infof("Deleted the %s CRD", resource.Name)
	} else {
		return fmt.Errorf("Failed to remove the %s CRD: %v", resource.Name, err)
	}

	return nil
}
