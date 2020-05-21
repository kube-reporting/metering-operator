package deploy

import (
	"fmt"
	"time"

	olmv1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1"
	olmv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"

	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

func (deploy *Deployer) installNamespace() error {
	namespace, err := deploy.client.CoreV1().Namespaces().Get(deploy.config.Namespace, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		namespaceObjectMeta := metav1.ObjectMeta{
			Name: deploy.config.Namespace,
		}

		labels := make(map[string]string)

		for key, val := range deploy.config.ExtraNamespaceLabels {
			labels[key] = val
			deploy.logger.Infof("Labeling the %s namespace with '%s=%s'", deploy.config.Namespace, key, val)
		}

		if deploy.config.Platform == "openshift" {
			labels["openshift.io/cluster-monitoring"] = "true"
			deploy.logger.Infof("Labeling the %s namespace with 'openshift.io/cluster-monitoring=true'", deploy.config.Namespace)
		}

		namespaceObjectMeta.Labels = labels
		namespaceObj := &v1.Namespace{
			ObjectMeta: namespaceObjectMeta,
		}

		_, err := deploy.client.CoreV1().Namespaces().Create(namespaceObj)
		if err != nil {
			return err
		}
		deploy.logger.Infof("Created the %s namespace", deploy.config.Namespace)
	} else if err == nil {
		// check if we need to add/update the cluster-monitoring label for Openshift installs.
		if deploy.config.Platform == "openshift" {
			if namespace.ObjectMeta.Labels != nil {
				namespace.ObjectMeta.Labels["openshift.io/cluster-monitoring"] = "true"
				deploy.logger.Infof("Updated the 'openshift.io/cluster-monitoring' label to the %s namespace", deploy.config.Namespace)
			} else {
				namespace.ObjectMeta.Labels = map[string]string{
					"openshift.io/cluster-monitoring": "true",
				}
				deploy.logger.Infof("Added the 'openshift.io/cluster-monitoring' label to the %s namespace", deploy.config.Namespace)
			}

			_, err := deploy.client.CoreV1().Namespaces().Update(namespace)
			if err != nil {
				return fmt.Errorf("failed to add the 'openshift.io/cluster-monitoring' label to the %s namespace: %v", deploy.config.Namespace, err)
			}
		} else {
			deploy.logger.Infof("The %s namespace already exists", deploy.config.Namespace)
		}
	} else {
		return err
	}

	return nil
}

func (deploy *Deployer) installMeteringConfig() error {
	if deploy.config.MeteringConfig == nil {
		return fmt.Errorf("invalid deploy configuration: MeteringConfig object is nil")
	}
	if deploy.config.MeteringConfig.Name == "" {
		return fmt.Errorf("invalid deploy configuration: metadata.Name is unset")
	}

	// ensure the MeteringConfig CRD has already been created to avoid
	// any errors while instantiating a MeteringConfig custom resource
	err := wait.Poll(crdInitialPoll, crdPollTimeout, func() (done bool, err error) {
		deploy.logger.Infof("Waiting for the MeteringConfig CRD to be created")

		_, err = deploy.apiExtClient.CustomResourceDefinitions().Get(meteringconfigCRDName, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			// re-poll as the MeteringConfig CRD does not already exist
			return false, nil
		}
		if err != nil {
			return false, err
		}

		return true, nil
	})
	if err != nil {
		return fmt.Errorf("failed to wait for the MeteringConfig CRD to be created: %v", err)
	}
	deploy.logger.Infof("The MeteringConfig CRD exists")

	mc, err := deploy.meteringClient.MeteringConfigs(deploy.config.Namespace).Get(deploy.config.MeteringConfig.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		// This was needed to fix https://bugzilla.redhat.com/show_bug.cgi?id=1838652
		// In short, despite checking if the MeteringConfig CRD exists before creating the CR,
		// we were still getting the occasionally apiserver error. If we instead attempt
		// to create this CR over the course of the next minute, we reduce the number
		// of apiserver (and overall e2e suite) flakes dramatically.
		err := wait.PollImmediate(time.Second, time.Minute, func() (done bool, err error) {
			_, err = deploy.meteringClient.MeteringConfigs(deploy.config.Namespace).Create(deploy.config.MeteringConfig)
			deploy.logger.Infof("Waiting the MeteringConfig CR to be created")
			if err != nil {
				return false, nil
			}
			return true, nil
		})
		if err != nil {
			return err
		}

		deploy.logger.Infof("Created the MeteringConfig resource")
	} else if err == nil {
		mc.Spec = deploy.config.MeteringConfig.Spec

		_, err = deploy.meteringClient.MeteringConfigs(deploy.config.Namespace).Update(mc)
		if err != nil {
			return fmt.Errorf("failed to update the MeteringConfig: %v", err)
		}
		deploy.logger.Infof("The MeteringConfig resource has been updated")
	} else {
		return err
	}

	return nil
}

func (deploy *Deployer) installMeteringOperatorGroup() error {
	opgrp, err := deploy.olmV1Client.OperatorGroups(deploy.config.Namespace).Get(deploy.config.Namespace, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		opgrp := &olmv1.OperatorGroup{
			ObjectMeta: metav1.ObjectMeta{
				Name:      deploy.config.Namespace,
				Namespace: deploy.config.Namespace,
			},
			Spec: olmv1.OperatorGroupSpec{
				TargetNamespaces: []string{
					deploy.config.Namespace,
				},
			},
		}

		_, err = deploy.olmV1Client.OperatorGroups(deploy.config.Namespace).Create(opgrp)
		if err != nil {
			return err
		}
		deploy.logger.Infof("Created the %s metering OperatorGroup", opgrp.Name)
	} else if err == nil {
		deploy.logger.Infof("The %s metering OperatorGroup resource already exists", opgrp.Name)
	} else {
		return err
	}

	return nil
}

func (deploy *Deployer) installMeteringSubscription() error {
	_, err := deploy.olmV1Alpha1Client.Subscriptions(deploy.config.Namespace).Get(deploy.config.SubscriptionName, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		sub := &olmv1alpha1.Subscription{
			ObjectMeta: metav1.ObjectMeta{
				Name:      deploy.config.SubscriptionName,
				Namespace: deploy.config.Namespace,
			},
			Spec: &olmv1alpha1.SubscriptionSpec{
				CatalogSource:          catalogSourceName,
				CatalogSourceNamespace: catalogSourceNamespace,
				Package:                packageName,
				Channel:                deploy.config.Channel,
				InstallPlanApproval:    olmv1alpha1.ApprovalAutomatic,
			},
		}

		_, err := deploy.olmV1Alpha1Client.Subscriptions(deploy.config.Namespace).Create(sub)
		if err != nil {
			return fmt.Errorf("failed to create the %s Subscription: %v", deploy.config.SubscriptionName, err)
		}
		deploy.logger.Infof("Created the metering Subscription")
	} else if err == nil {
		deploy.logger.Infof("The metering Subscription already exists")
	} else {
		return err
	}

	return nil
}

func (deploy *Deployer) installMeteringResources() error {
	if !deploy.config.RunMeteringOperatorLocal {
		err := deploy.installMeteringDeployment()
		if err != nil {
			return fmt.Errorf("failed to create the metering deployment: %v", err)
		}
	}

	err := deploy.installMeteringServiceAccount()
	if err != nil {
		return fmt.Errorf("failed to create the metering service account: %v", err)
	}

	err = deploy.installMeteringRole()
	if err != nil {
		return fmt.Errorf("failed to create the metering role: %v", err)
	}

	err = deploy.installMeteringRoleBinding()
	if err != nil {
		return fmt.Errorf("failed to create the metering role binding: %v", err)
	}

	err = deploy.installMeteringClusterRole()
	if err != nil {
		return fmt.Errorf("failed to create the metering cluster role: %v", err)
	}

	err = deploy.installMeteringClusterRoleBinding()
	if err != nil {
		return fmt.Errorf("failed to create the metering cluster role binding: %v", err)
	}

	return nil
}

func (deploy *Deployer) installMeteringDeployment() error {
	res := deploy.config.OperatorResources.Deployment

	// check if the metering operator image needs to be updated
	// TODO: implement support for METERING_OPERATOR_ALL_NAMESPACES and METERING_OPERATOR_TARGET_NAMESPACES
	if deploy.config.Repo != "" && deploy.config.Tag != "" {
		newImage := deploy.config.Repo + ":" + deploy.config.Tag

		for index := range res.Spec.Template.Spec.Containers {
			res.Spec.Template.Spec.Containers[index].Image = newImage
		}

		deploy.logger.Infof("Overriding the default image with %s", newImage)
	}

	deployment, err := deploy.client.AppsV1().Deployments(deploy.config.Namespace).Get(res.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err := deploy.client.AppsV1().Deployments(deploy.config.Namespace).Create(res)
		if err != nil {
			return err
		}
		deploy.logger.Infof("Created the metering deployment")
	} else if err == nil {
		deployment.Spec = res.Spec

		_, err = deploy.client.AppsV1().Deployments(deploy.config.Namespace).Update(deployment)
		if err != nil {
			return fmt.Errorf("failed to update the metering deployment: %v", err)
		}
		deploy.logger.Infof("The metering deployment resource has been updated")
	} else {
		return err
	}

	return nil
}

func (deploy *Deployer) installMeteringServiceAccount() error {
	_, err := deploy.client.CoreV1().ServiceAccounts(deploy.config.Namespace).Get(deploy.config.OperatorResources.ServiceAccount.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err := deploy.client.CoreV1().ServiceAccounts(deploy.config.Namespace).Create(deploy.config.OperatorResources.ServiceAccount)
		if err != nil {
			return err
		}
		deploy.logger.Infof("Created the metering serviceaccount")
	} else if err == nil {
		deploy.logger.Infof("The metering service account already exists")
	} else {
		return err
	}

	return nil
}

func (deploy *Deployer) installMeteringRoleBinding() error {
	res := deploy.config.OperatorResources.RoleBinding

	// TODO: implement support for METERING_OPERATOR_TARGET_NAMESPACES
	res.Name = deploy.config.Namespace + "-" + res.Name
	res.RoleRef.Name = res.Name
	res.Namespace = deploy.config.Namespace

	for index := range res.Subjects {
		res.Subjects[index].Namespace = deploy.config.Namespace
	}

	_, err := deploy.client.RbacV1().RoleBindings(deploy.config.Namespace).Get(res.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err := deploy.client.RbacV1().RoleBindings(deploy.config.Namespace).Create(res)
		if err != nil {
			return err
		}
		deploy.logger.Infof("Created the metering role binding")
	} else if err == nil {
		deploy.logger.Infof("The metering role binding already exists")
	} else {
		return err
	}

	return nil
}

func (deploy *Deployer) installMeteringRole() error {
	res := deploy.config.OperatorResources.Role

	res.Name = deploy.config.Namespace + "-" + res.Name
	res.Namespace = deploy.config.Namespace

	_, err := deploy.client.RbacV1().Roles(deploy.config.Namespace).Get(res.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err := deploy.client.RbacV1().Roles(deploy.config.Namespace).Create(res)
		if err != nil {
			return err
		}
		deploy.logger.Infof("Created the metering role")
	} else if err == nil {
		deploy.logger.Infof("The metering role already exists")
	} else {
		return err
	}

	return nil
}

func (deploy *Deployer) installMeteringClusterRoleBinding() error {
	res := deploy.config.OperatorResources.ClusterRoleBinding

	res.Name = deploy.config.Namespace + "-" + res.Name
	res.RoleRef.Name = res.Name

	for index := range res.Subjects {
		res.Subjects[index].Namespace = deploy.config.Namespace
	}

	_, err := deploy.client.RbacV1().ClusterRoleBindings().Get(res.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err := deploy.client.RbacV1().ClusterRoleBindings().Create(res)
		if err != nil {
			return err
		}
		deploy.logger.Infof("Created the metering cluster role binding")
	} else if err == nil {
		_, err = deploy.client.RbacV1().ClusterRoleBindings().Update(res)
		if err != nil {
			return fmt.Errorf("failed to update the metering clusterrolebinding: %v", err)
		}
		deploy.logger.Infof("The metering cluster role binding has been updated")
	} else {
		return err
	}

	return nil
}

func (deploy *Deployer) installMeteringClusterRole() error {
	res := deploy.config.OperatorResources.ClusterRole

	res.Name = deploy.config.Namespace + "-" + res.Name

	clusterRole, err := deploy.client.RbacV1().ClusterRoles().Get(res.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err := deploy.client.RbacV1().ClusterRoles().Create(res)
		if err != nil {
			return err
		}
		deploy.logger.Infof("Created the metering cluster role")
	} else if err == nil {
		clusterRole.Rules = res.Rules

		_, err = deploy.client.RbacV1().ClusterRoles().Update(clusterRole)
		if err != nil {
			return fmt.Errorf("failed to update the metering clusterrole: %v", err)
		}
		deploy.logger.Infof("The metering clusterrole has been updated")
	} else {
		return err
	}

	return nil
}

func (deploy *Deployer) installMeteringCRDs() error {
	for _, crd := range deploy.config.OperatorResources.CRDs {
		err := deploy.installMeteringCRD(crd)
		if err != nil {
			return fmt.Errorf("failed to create a CRD while looping: %v", err)
		}
	}

	return nil
}

func (deploy *Deployer) installMeteringCRD(resource CRD) error {
	crd, err := deploy.apiExtClient.CustomResourceDefinitions().Get(resource.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err := deploy.apiExtClient.CustomResourceDefinitions().Create(resource.CRD)
		if err != nil {
			return err
		}
		deploy.logger.Infof("Created the %s CRD", resource.Name)
	} else if err == nil {
		crd.Spec = resource.CRD.Spec

		_, err := deploy.apiExtClient.CustomResourceDefinitions().Update(crd)
		if err != nil {
			return fmt.Errorf("failed to update the %s CRD: %v", resource.CRD.Name, err)
		}
		deploy.logger.Infof("Updated the %s CRD", resource.CRD.Name)
	} else {
		return err
	}

	return nil
}
