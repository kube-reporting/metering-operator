package deploy

import (
	"fmt"
	"path/filepath"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (deploy *Deployer) uninstallNamespace() error {
	err := deploy.client.CoreV1().Namespaces().Delete(deploy.config.Namespace, &metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		deploy.logger.Warnf("The %s namespace doesn't exist", deploy.config.Namespace)
	} else if err == nil {
		deploy.logger.Infof("Deleted the %s namespace", deploy.config.Namespace)
	} else {
		return fmt.Errorf("Failed to delete the %s namespace: %v", deploy.config.Namespace, err)
	}

	return nil
}

func (deploy *Deployer) uninstallMeteringConfig() error {
	err := deploy.meteringClient.MeteringConfigs(deploy.config.Namespace).Delete(deploy.config.MeteringConfig.Name, &metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		deploy.logger.Warnf("The MeteringConfig resource doesn't exist")
	} else if err == nil {
		deploy.logger.Infof("Deleted the MeteringConfig resource")
	} else {
		return fmt.Errorf("Failed to delete the MeteringConfig resource: %v", err)
	}

	return nil
}

func (deploy *Deployer) uninstallMeteringResources() error {
	err := deploy.uninstallMeteringDeployment(filepath.Join(deploy.ansibleOperatorManifestsLocation, meteringDeploymentFile))
	if err != nil {
		return fmt.Errorf("Failed to delete the metering service account: %v", err)
	}

	err = deploy.uninstallMeteringServiceAccount(filepath.Join(deploy.ansibleOperatorManifestsLocation, meteringServiceAccountFile))
	if err != nil {
		return fmt.Errorf("Failed to delete the metering service account: %v", err)
	}

	err = deploy.uninstallMeteringRole(filepath.Join(deploy.ansibleOperatorManifestsLocation, meteringRoleFile))
	if err != nil {
		return fmt.Errorf("Failed to delete the metering role: %v", err)
	}

	err = deploy.uninstallMeteringRoleBinding(filepath.Join(deploy.ansibleOperatorManifestsLocation, meteringRoleBindingFile))
	if err != nil {
		return fmt.Errorf("Failed to delete the metering role binding: %v", err)
	}

	if deploy.config.DeleteCRB {
		err = deploy.uninstallMeteringClusterRole(filepath.Join(deploy.ansibleOperatorManifestsLocation, meteringClusterRoleFile))
		if err != nil {
			return fmt.Errorf("Failed to delete the metering cluster role: %v", err)
		}

		err = deploy.uninstallMeteringClusterRoleBinding(filepath.Join(deploy.ansibleOperatorManifestsLocation, meteringClusterRoleBindingFile))
		if err != nil {
			return fmt.Errorf("Failed to delete the metering cluster role binding: %v", err)
		}
	} else {
		deploy.logger.Infof("Skipped deleting the metering cluster role resources")
	}

	if deploy.config.DeletePVCs {
		err = deploy.uninstallMeteringPVCs()
		if err != nil {
			return fmt.Errorf("Failed to delete the metering PVCs: %v", err)
		}
	} else {
		deploy.logger.Infof("Skipped deleting the metering PVCs")
	}

	return nil
}

// uninstallMeteringPVCs gets a list of all the PVCs associated with the hdfs and hive-metastore
// pods in the $METERING_NAMESPACE namespace, and attempts to delete all the PVCs that match that list criteria
func (deploy *Deployer) uninstallMeteringPVCs() error {
	// Attempt to get a list of PVCs that match the hdfs or hive labels
	pvcs, err := deploy.client.CoreV1().PersistentVolumeClaims(deploy.config.Namespace).List(metav1.ListOptions{
		LabelSelector: "app in (hdfs,hive)",
	})
	if err != nil {
		return fmt.Errorf("Failed to list all the metering PVCs in the %s namespace: %v", deploy.config.Namespace, err)
	}

	if len(pvcs.Items) == 0 {
		deploy.logger.Warnf("The Hive/HDFS PVCs don't exist")
		return nil
	}

	for _, pvc := range pvcs.Items {
		err = deploy.client.CoreV1().PersistentVolumeClaims(deploy.config.Namespace).Delete(pvc.Name, &metav1.DeleteOptions{})
		if err != nil {
			return fmt.Errorf("Failed to delete the PVC %s: %v", pvc.Name, err)
		}
	}

	deploy.logger.Infof("Deleted the PVCs managed by metering")

	return nil
}

func (deploy *Deployer) uninstallMeteringDeployment(deploymentName string) error {
	var res appsv1.Deployment

	err := DecodeYAMLManifestToObject(deploymentName, &res)
	if err != nil {
		return fmt.Errorf("Failed to decode the YAML manifest: %v", err)
	}

	err = deploy.client.AppsV1().Deployments(deploy.config.Namespace).Delete(res.Name, &metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		deploy.logger.Warnf("The metering deployment doesn't exist")
	} else if err == nil {
		deploy.logger.Infof("Deleted the metering deployment")
	} else {
		return fmt.Errorf("Failed to delete the metering deployment: %v", err)
	}

	return nil
}

func (deploy *Deployer) uninstallMeteringServiceAccount(serviceAccountPath string) error {
	var res corev1.ServiceAccount

	err := DecodeYAMLManifestToObject(serviceAccountPath, &res)
	if err != nil {
		return fmt.Errorf("Failed to decode the YAML manifest: %v", err)
	}

	err = deploy.client.CoreV1().ServiceAccounts(deploy.config.Namespace).Delete(res.Name, &metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		deploy.logger.Warnf("The metering service account doesn't exist")
	} else if err == nil {
		deploy.logger.Infof("Deleted the metering serviceaccount")
	} else {
		return fmt.Errorf("Failed to delete the metering serviceaccount: %v", err)
	}

	return nil
}

func (deploy *Deployer) uninstallMeteringRoleBinding(roleBindingPath string) error {
	var res rbacv1.RoleBinding

	err := DecodeYAMLManifestToObject(roleBindingPath, &res)
	if err != nil {
		return fmt.Errorf("Failed to decode the YAML manifest: %v", err)
	}

	res.Name = deploy.config.Namespace + "-" + res.Name
	res.RoleRef.Name = res.Name
	res.Namespace = deploy.config.Namespace

	for index := range res.Subjects {
		res.Subjects[index].Namespace = deploy.config.Namespace
	}

	err = deploy.client.RbacV1().RoleBindings(deploy.config.Namespace).Delete(res.Name, &metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		deploy.logger.Warnf("The metering role binding doesn't exist")
	} else if err == nil {
		deploy.logger.Infof("Deleted the metering role binding")
	} else {
		return fmt.Errorf("Failed to delete the metering role binding: %v", err)
	}

	return nil
}

func (deploy *Deployer) uninstallMeteringRole(rolePath string) error {
	var res rbacv1.Role

	err := DecodeYAMLManifestToObject(rolePath, &res)
	if err != nil {
		return fmt.Errorf("Failed to decode the YAML manifest: %v", err)
	}

	res.Name = deploy.config.Namespace + "-" + res.Name
	res.Namespace = deploy.config.Namespace

	err = deploy.client.RbacV1().Roles(deploy.config.Namespace).Delete(res.Name, &metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		deploy.logger.Warnf("The metering role doesn't exist")
	} else if err == nil {
		deploy.logger.Infof("Deleted the metering role")
	} else {
		return fmt.Errorf("Failed to delete the metering role: %v", err)
	}

	return nil
}

func (deploy *Deployer) uninstallMeteringClusterRole(clusterrolePath string) error {
	var res rbacv1.ClusterRole

	err := DecodeYAMLManifestToObject(clusterrolePath, &res)
	if err != nil {
		return fmt.Errorf("Failed to decode the YAML manifest: %v", err)
	}

	res.Name = deploy.config.Namespace + "-" + res.Name

	err = deploy.client.RbacV1().ClusterRoles().Delete(res.Name, &metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		deploy.logger.Warnf("The metering cluster role doesn't exist")
	} else if err == nil {
		deploy.logger.Infof("Deleted the metering cluster role")
	} else {
		return fmt.Errorf("Failed to delete the metering cluster role: %v", err)
	}

	return nil
}

func (deploy *Deployer) uninstallMeteringClusterRoleBinding(meteringClusterRoleBindingFile string) error {
	var res rbacv1.ClusterRoleBinding

	err := DecodeYAMLManifestToObject(meteringClusterRoleBindingFile, &res)
	if err != nil {
		return fmt.Errorf("Failed to decode the YAML manifest: %v", err)
	}

	res.Name = deploy.config.Namespace + "-" + res.Name
	res.RoleRef.Name = res.Name

	for index := range res.Subjects {
		res.Subjects[index].Namespace = deploy.config.Namespace
	}

	err = deploy.client.RbacV1().ClusterRoleBindings().Delete(res.Name, &metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		deploy.logger.Warnf("The metering cluster role binding doesn't exist")
	} else if err == nil {
		deploy.logger.Infof("Deleted the metering cluster role binding")
	} else {
		return fmt.Errorf("Failed to delete the metering cluster role binding: %v", err)
	}

	return nil
}

func (deploy *Deployer) uninstallMeteringCRDs() error {
	for _, crd := range deploy.crds {
		err := deploy.uninstallMeteringCRD(crd)
		if err != nil {
			return fmt.Errorf("Failed to delete a CRD while looping: %v", err)
		}
	}

	return nil
}

func (deploy *Deployer) uninstallMeteringCRD(resource CRD) error {
	err := DecodeYAMLManifestToObject(resource.Path, resource.CRD)
	if err != nil {
		return fmt.Errorf("Failed to decode the YAML manifest: %v", err)
	}

	err = deploy.apiExtClient.CustomResourceDefinitions().Delete(resource.Name, &metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		deploy.logger.Warnf("The %s CRD doesn't exist", resource.Name)
	} else if err == nil {
		deploy.logger.Infof("Deleted the %s CRD", resource.Name)
	} else {
		return fmt.Errorf("Failed to remove the %s CRD: %v", resource.Name, err)
	}

	return nil
}
