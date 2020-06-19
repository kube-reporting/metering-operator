package deploy

import (
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (deploy *Deployer) uninstallNamespace() error {
	err := deploy.client.CoreV1().Namespaces().Delete(context.TODO(), deploy.config.Namespace, metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		deploy.logger.Warnf("The %s namespace doesn't exist", deploy.config.Namespace)
	} else if err == nil {
		deploy.logger.Infof("Deleted the %s namespace", deploy.config.Namespace)
	} else {
		return err
	}

	return nil
}

func (deploy *Deployer) uninstallMeteringConfig() error {
	err := deploy.meteringClient.MeteringConfigs(deploy.config.Namespace).Delete(context.TODO(), deploy.config.MeteringConfig.Name, metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		deploy.logger.Warnf("The MeteringConfig resource doesn't exist")
	} else if err == nil {
		deploy.logger.Infof("Deleted the MeteringConfig resource")
	} else {
		return err
	}

	return nil
}

func (deploy *Deployer) uninstallMeteringOperatorGroup() error {
	opgrp, err := deploy.olmV1Client.OperatorGroups(deploy.config.Namespace).Get(context.TODO(), deploy.config.Namespace, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		deploy.logger.Warnf("The metering OperatorGroup resource does not exist")
	} else if err == nil {
		err := deploy.olmV1Client.OperatorGroups(deploy.config.Namespace).Delete(context.TODO(), opgrp.Name, metav1.DeleteOptions{})
		if err != nil {
			return fmt.Errorf("failed to delete the metering OperatorGroup resource: %v", err)
		}
		deploy.logger.Infof("Deleted the metering OperatorGroup resource")
	} else {
		return err
	}

	return nil
}

func (deploy *Deployer) uninstallMeteringSubscription() error {
	sub, err := deploy.olmV1Alpha1Client.Subscriptions(deploy.config.Namespace).Get(context.TODO(), "metering-ocp", metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		deploy.logger.Warnf("The metering Subscription does not exist")
	} else if err == nil {
		err := deploy.olmV1Alpha1Client.Subscriptions(deploy.config.Namespace).Delete(context.TODO(), sub.Name, metav1.DeleteOptions{})
		if err != nil {
			return fmt.Errorf("failed to delete the metering Subscription: %v", err)
		}
		deploy.logger.Infof("Deleted the metering Subscription resource")
	} else {
		return err
	}

	return nil
}

func (deploy *Deployer) uninstallMeteringCSV() error {
	// attempt to query for the metering subscription as we don't have a way of knowing
	// what the CSV's name is beforehand without exposing more configurable flags.
	// in the case where the subscription resource does not already exist, exit early
	// and hope that the user is re-running the olm-uninstall command.
	sub, err := deploy.olmV1Alpha1Client.Subscriptions(deploy.config.Namespace).Get(context.TODO(), deploy.config.SubscriptionName, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to get the metering subscription: %v", err)
	}

	if sub.Status.CurrentCSV == "" {
		return fmt.Errorf("failed to get the currentCSV stored in the %s metering Subscription", deploy.config.SubscriptionName)
	}
	csvName := sub.Status.CurrentCSV
	deploy.logger.Infof("Found existing metering subscription, attempting to delete the %s CSV", csvName)

	csv, err := deploy.olmV1Alpha1Client.ClusterServiceVersions(deploy.config.Namespace).Get(context.TODO(), csvName, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		deploy.logger.Warnf("The metering CSV does not exist")
	} else if err == nil {
		err := deploy.olmV1Alpha1Client.ClusterServiceVersions(deploy.config.Namespace).Delete(context.TODO(), csv.Name, metav1.DeleteOptions{})
		if err != nil {
			return fmt.Errorf("failed to delete the metering Subscription: %v", err)
		}
		deploy.logger.Infof("Deleted the metering Subscription resource")
	} else {
		return err
	}

	return nil
}

func (deploy *Deployer) uninstallMeteringResources() error {
	err := deploy.uninstallMeteringDeployment()
	if err != nil {
		return fmt.Errorf("failed to delete the metering service account: %v", err)
	}

	err = deploy.uninstallMeteringServiceAccount()
	if err != nil {
		return fmt.Errorf("failed to delete the metering service account: %v", err)
	}

	err = deploy.uninstallMeteringRole()
	if err != nil {
		return fmt.Errorf("failed to delete the metering role: %v", err)
	}

	err = deploy.uninstallMeteringRoleBinding()
	if err != nil {
		return fmt.Errorf("failed to delete the metering role binding: %v", err)
	}

	if deploy.config.DeleteCRB {
		err = deploy.uninstallMeteringClusterRole()
		if err != nil {
			return fmt.Errorf("failed to delete the metering cluster role: %v", err)
		}

		err = deploy.uninstallMeteringClusterRoleBinding()
		if err != nil {
			return fmt.Errorf("failed to delete the metering cluster role binding: %v", err)
		}
	} else {
		deploy.logger.Infof("Skipped deleting the metering cluster role resources")
	}

	if deploy.config.DeletePVCs {
		err = deploy.uninstallMeteringPVCs()
		if err != nil {
			return fmt.Errorf("failed to delete the metering PVCs: %v", err)
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
	pvcs, err := deploy.client.CoreV1().PersistentVolumeClaims(deploy.config.Namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: "app in (hdfs,hive-metastore)",
	})
	if err != nil {
		return fmt.Errorf("failed to list all the metering PVCs in the %s namespace: %v", deploy.config.Namespace, err)
	}

	if len(pvcs.Items) == 0 {
		deploy.logger.Warnf("The Hive/HDFS PVCs don't exist")
		return nil
	}

	for _, pvc := range pvcs.Items {
		err = deploy.client.CoreV1().PersistentVolumeClaims(deploy.config.Namespace).Delete(context.TODO(), pvc.Name, metav1.DeleteOptions{})
		if err != nil {
			return fmt.Errorf("failed to delete the %s PVC: %v", pvc.Name, err)
		}
	}

	deploy.logger.Infof("Deleted the PVCs managed by metering")

	return nil
}

func (deploy *Deployer) uninstallMeteringDeployment() error {
	err := deploy.client.AppsV1().Deployments(deploy.config.Namespace).Delete(context.TODO(), deploy.config.OperatorResources.Deployment.Name, metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		deploy.logger.Warnf("The metering deployment doesn't exist")
	} else if err == nil {
		deploy.logger.Infof("Deleted the metering deployment")
	} else {
		return err
	}

	return nil
}

func (deploy *Deployer) uninstallMeteringServiceAccount() error {
	err := deploy.client.CoreV1().ServiceAccounts(deploy.config.Namespace).Delete(context.TODO(), deploy.config.OperatorResources.ServiceAccount.Name, metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		deploy.logger.Warnf("The metering service account doesn't exist")
	} else if err == nil {
		deploy.logger.Infof("Deleted the metering serviceaccount")
	} else {
		return err
	}

	return nil
}

func (deploy *Deployer) uninstallMeteringRoleBinding() error {
	res := deploy.config.OperatorResources.RoleBinding

	res.Name = deploy.config.Namespace + "-" + res.Name
	res.RoleRef.Name = res.Name
	res.Namespace = deploy.config.Namespace

	for index := range res.Subjects {
		res.Subjects[index].Namespace = deploy.config.Namespace
	}

	err := deploy.client.RbacV1().RoleBindings(deploy.config.Namespace).Delete(context.TODO(), res.Name, metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		deploy.logger.Warnf("The metering role binding doesn't exist")
	} else if err == nil {
		deploy.logger.Infof("Deleted the metering role binding")
	} else {
		return err
	}

	return nil
}

func (deploy *Deployer) uninstallMeteringRole() error {
	res := deploy.config.OperatorResources.Role

	res.Name = deploy.config.Namespace + "-" + res.Name
	res.Namespace = deploy.config.Namespace

	err := deploy.client.RbacV1().Roles(deploy.config.Namespace).Delete(context.TODO(), res.Name, metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		deploy.logger.Warnf("The metering role doesn't exist")
	} else if err == nil {
		deploy.logger.Infof("Deleted the metering role")
	} else {
		return err
	}

	return nil
}

func (deploy *Deployer) uninstallMeteringClusterRole() error {
	res := deploy.config.OperatorResources.ClusterRole

	res.Name = deploy.config.Namespace + "-" + res.Name

	err := deploy.client.RbacV1().ClusterRoles().Delete(context.TODO(), res.Name, metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		deploy.logger.Warnf("The metering cluster role doesn't exist")
	} else if err == nil {
		deploy.logger.Infof("Deleted the metering cluster role")
	} else {
		return err
	}

	// attempt to delete any of the clusterroles the reporting-operator creates
	err = deploy.client.RbacV1().ClusterRoles().DeleteCollection(context.TODO(), metav1.DeleteOptions{}, metav1.ListOptions{
		LabelSelector: "app=reporting-operator",
	})
	if err != nil {
		return fmt.Errorf("failed to list all the reporting-operator clusterroles in the %s namespace: %v", deploy.config.Namespace, err)
	}
	deploy.logger.Infof("Deleted the 'app=reporting-operator' cluster roles")

	return nil
}

func (deploy *Deployer) uninstallMeteringClusterRoleBinding() error {
	res := deploy.config.OperatorResources.ClusterRoleBinding

	res.Name = deploy.config.Namespace + "-" + res.Name
	res.RoleRef.Name = res.Name

	for index := range res.Subjects {
		res.Subjects[index].Namespace = deploy.config.Namespace
	}

	err := deploy.client.RbacV1().ClusterRoleBindings().Delete(context.TODO(), res.Name, metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		deploy.logger.Warnf("The metering cluster role binding doesn't exist")
	} else if err == nil {
		deploy.logger.Infof("Deleted the metering cluster role binding")
	} else {
		return err
	}

	// attempt to delete any of the clusterrolebindings the reporting-operator creates
	err = deploy.client.RbacV1().ClusterRoleBindings().DeleteCollection(context.TODO(), metav1.DeleteOptions{}, metav1.ListOptions{
		LabelSelector: "app=reporting-operator",
	})
	if err != nil {
		return fmt.Errorf("failed to list all the reporting-operator clusterrolebindings in the %s namespace: %v", deploy.config.Namespace, err)
	}
	deploy.logger.Infof("Deleted the 'app=reporting-operator' cluster role bindings")

	return nil
}

func (deploy *Deployer) uninstallMeteringCRDs() error {
	for _, crd := range deploy.config.OperatorResources.CRDs {
		err := deploy.uninstallMeteringCRD(crd)
		if err != nil {
			return fmt.Errorf("failed to delete a CRD while looping: %v", err)
		}
	}

	return nil
}

func (deploy *Deployer) uninstallMeteringCRD(resource CRD) error {
	err := deploy.apiExtClient.CustomResourceDefinitions().Delete(context.TODO(), resource.Name, metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		deploy.logger.Warnf("The %s CRD doesn't exist", resource.Name)
	} else if err == nil {
		deploy.logger.Infof("Deleted the %s CRD", resource.Name)
	} else {
		return fmt.Errorf("failed to remove the %s CRD: %v", resource.Name, err)
	}

	return nil
}
