package deploy

import (
	"fmt"
	"path/***REMOVED***lepath"

	meteringv1 "github.com/operator-framework/operator-metering/pkg/apis/metering/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (deploy *Deployer) uninstallNamespace() error {
	err := deploy.client.CoreV1().Namespaces().Delete(deploy.con***REMOVED***g.Namespace, &metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		deploy.logger.Warnf("The %s namespace doesn't exist", deploy.con***REMOVED***g.Namespace)
	} ***REMOVED*** if err == nil {
		deploy.logger.Infof("Deleted the %s namespace", deploy.con***REMOVED***g.Namespace)
	} ***REMOVED*** {
		return fmt.Errorf("Failed to delete the %s namespace: %v", deploy.con***REMOVED***g.Namespace, err)
	}

	return nil
}

func (deploy *Deployer) uninstallMeteringCon***REMOVED***g() error {
	var res meteringv1.MeteringCon***REMOVED***g

	err := decodeYAMLManifestToObject(deploy.con***REMOVED***g.MeteringCR, &res)
	if err != nil {
		return fmt.Errorf("Failed while attempting to build up the MeteringCon***REMOVED***g from the YAML ***REMOVED***le, got: %v", err)
	}

	err = deploy.meteringClient.MeteringCon***REMOVED***gs(deploy.con***REMOVED***g.Namespace).Delete(res.Name, &metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		deploy.logger.Warnf("The MeteringCon***REMOVED***g resource doesn't exist")
	} ***REMOVED*** if err == nil {
		deploy.logger.Infof("Deleted the MeteringCon***REMOVED***g resource")
	} ***REMOVED*** {
		return fmt.Errorf("Failed to delete the MeteringCon***REMOVED***g resource: %v", err)
	}

	return nil
}

func (deploy *Deployer) uninstallMeteringResources() error {
	err := deploy.uninstallMeteringDeployment(***REMOVED***lepath.Join(deploy.con***REMOVED***g.ManifestLocation, meteringDeploymentFile))
	if err != nil {
		return fmt.Errorf("Failed to delete the metering service account: %v", err)
	}

	err = deploy.uninstallMeteringServiceAccount(***REMOVED***lepath.Join(deploy.con***REMOVED***g.ManifestLocation, meteringServiceAccountFile))
	if err != nil {
		return fmt.Errorf("Failed to delete the metering service account: %v", err)
	}

	err = deploy.uninstallMeteringRole(***REMOVED***lepath.Join(deploy.con***REMOVED***g.ManifestLocation, meteringRoleFile))
	if err != nil {
		return fmt.Errorf("Failed to delete the metering role: %v", err)
	}

	err = deploy.uninstallMeteringRoleBinding(***REMOVED***lepath.Join(deploy.con***REMOVED***g.ManifestLocation, meteringRoleBindingFile))
	if err != nil {
		return fmt.Errorf("Failed to delete the metering role binding: %v", err)
	}

	if deploy.con***REMOVED***g.DeleteCRB {
		err = deploy.uninstallMeteringClusterRole(***REMOVED***lepath.Join(deploy.con***REMOVED***g.ManifestLocation, meteringClusterRoleFile))
		if err != nil {
			return fmt.Errorf("Failed to delete the metering cluster role: %v", err)
		}

		err = deploy.uninstallMeteringClusterRoleBinding(***REMOVED***lepath.Join(deploy.con***REMOVED***g.ManifestLocation, meteringClusterRoleBindingFile))
		if err != nil {
			return fmt.Errorf("Failed to delete the metering cluster role binding: %v", err)
		}
	} ***REMOVED*** {
		deploy.logger.Infof("Skipped deleting the metering cluster role resources")
	}

	if deploy.con***REMOVED***g.DeletePVCs {
		err = deploy.uninstallMeteringPVCs()
		if err != nil {
			return fmt.Errorf("Failed to delete the metering PVCs: %v", err)
		}
	} ***REMOVED*** {
		deploy.logger.Infof("Skipped deleting the metering PVCs")
	}

	return nil
}

// uninstallMeteringPVCs gets a list of all the PVCs associated with the hdfs and hive-metastore
// pods in the $METERING_NAMESPACE namespace, and attempts to delete all the PVCs that match that list criteria
func (deploy *Deployer) uninstallMeteringPVCs() error {
	// Attempt to get a list of PVCs that match the hdfs or hive labels
	pvcs, err := deploy.client.CoreV1().PersistentVolumeClaims(deploy.con***REMOVED***g.Namespace).List(metav1.ListOptions{
		LabelSelector: "app in (hdfs,hive)",
	})
	if err != nil {
		return fmt.Errorf("Failed to list all the metering PVCs in the %s namespace: %v", deploy.con***REMOVED***g.Namespace, err)
	}

	if len(pvcs.Items) == 0 {
		deploy.logger.Warnf("The Hive/HDFS PVCs don't exist")
		return nil
	}

	for _, pvc := range pvcs.Items {
		err = deploy.client.CoreV1().PersistentVolumeClaims(deploy.con***REMOVED***g.Namespace).Delete(pvc.Name, &metav1.DeleteOptions{})
		if err != nil {
			return fmt.Errorf("Failed to delete the PVC %s: %v", pvc.Name, err)
		}
	}

	deploy.logger.Infof("Deleted the PVCs managed by metering")

	return nil
}

func (deploy *Deployer) uninstallMeteringDeployment(deploymentName string) error {
	var res appsv1.Deployment

	err := decodeYAMLManifestToObject(deploymentName, &res)
	if err != nil {
		return fmt.Errorf("Failed to decode the YAML manifest: %v", err)
	}

	err = deploy.client.AppsV1().Deployments(deploy.con***REMOVED***g.Namespace).Delete(res.Name, &metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		deploy.logger.Warnf("The metering deployment doesn't exist")
	} ***REMOVED*** if err == nil {
		deploy.logger.Infof("Deleted the metering deployment")
	} ***REMOVED*** {
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

	err = deploy.client.CoreV1().ServiceAccounts(deploy.con***REMOVED***g.Namespace).Delete(res.Name, &metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		deploy.logger.Warnf("The metering service account doesn't exist")
	} ***REMOVED*** if err == nil {
		deploy.logger.Infof("Deleted the metering serviceaccount")
	} ***REMOVED*** {
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

	res.Name = deploy.con***REMOVED***g.Namespace + "-" + res.Name
	res.RoleRef.Name = res.Name
	res.Namespace = deploy.con***REMOVED***g.Namespace

	for index := range res.Subjects {
		res.Subjects[index].Namespace = deploy.con***REMOVED***g.Namespace
	}

	err = deploy.client.RbacV1().RoleBindings(deploy.con***REMOVED***g.Namespace).Delete(res.Name, &metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		deploy.logger.Warnf("The metering role binding doesn't exist")
	} ***REMOVED*** if err == nil {
		deploy.logger.Infof("Deleted the metering role binding")
	} ***REMOVED*** {
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

	res.Name = deploy.con***REMOVED***g.Namespace + "-" + res.Name
	res.Namespace = deploy.con***REMOVED***g.Namespace

	err = deploy.client.RbacV1().Roles(deploy.con***REMOVED***g.Namespace).Delete(res.Name, &metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		deploy.logger.Warnf("The metering role doesn't exist")
	} ***REMOVED*** if err == nil {
		deploy.logger.Infof("Deleted the metering role")
	} ***REMOVED*** {
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

	res.Name = deploy.con***REMOVED***g.Namespace + "-" + res.Name

	err = deploy.client.RbacV1().ClusterRoles().Delete(res.Name, &metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		deploy.logger.Warnf("The metering cluster role doesn't exist")
	} ***REMOVED*** if err == nil {
		deploy.logger.Infof("Deleted the metering cluster role")
	} ***REMOVED*** {
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

	res.Name = deploy.con***REMOVED***g.Namespace + "-" + res.Name
	res.RoleRef.Name = res.Name

	for index := range res.Subjects {
		res.Subjects[index].Namespace = deploy.con***REMOVED***g.Namespace
	}

	err = deploy.client.RbacV1().ClusterRoleBindings().Delete(res.Name, &metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		deploy.logger.Warnf("The metering cluster role binding doesn't exist")
	} ***REMOVED*** if err == nil {
		deploy.logger.Infof("Deleted the metering cluster role binding")
	} ***REMOVED*** {
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
	err := decodeYAMLManifestToObject(resource.Path, resource.CRD)
	if err != nil {
		return fmt.Errorf("Failed to decode the YAML manifest: %v", err)
	}

	err = deploy.apiExtClient.CustomResourceDe***REMOVED***nitions().Delete(resource.Name, &metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		deploy.logger.Warnf("The %s CRD doesn't exist", resource.Name)
	} ***REMOVED*** if err == nil {
		deploy.logger.Infof("Deleted the %s CRD", resource.Name)
	} ***REMOVED*** {
		return fmt.Errorf("Failed to remove the %s CRD: %v", resource.Name, err)
	}

	return nil
}
