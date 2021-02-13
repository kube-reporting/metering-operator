package deploy

import (
	"context"
	"fmt"

	olmv1 "github.com/operator-framework/api/pkg/operators/v1"
	olmv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"

	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

func (deploy *Deployer) installNamespace() error {
	// TODO: we can further cleanup this method by separating the
	// creation of the namespace object and the label/annotation
	// handling by always treating the latter as an update:
	// https://github.com/kube-reporting/metering-operator/pull/1270#discussion_r444436226
	namespace, err := deploy.Client.CoreV1().Namespaces().Get(context.TODO(), deploy.Config.Namespace, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		namespaceObjectMeta := metav1.ObjectMeta{
			Name: deploy.Config.Namespace,
		}

		labels := make(map[string]string)

		for key, val := range deploy.Config.ExtraNamespaceLabels {
			labels[key] = val
			deploy.Logger.Infof("Labeling the %s namespace with '%s=%s'", deploy.Config.Namespace, key, val)
		}

		/*
			In the case where the platform is set to Openshift (the default value),
			we need to make a few modifications to the namespace metadata.
			The 'openshift.io/cluster-monitoring' labels tells the cluster-monitoring
			operator to scrape Prometheus metrics for the installed Metering namespace.
			The 'openshift.io/node-selector' annotation is a way to control where Pods
			get scheduled in a specific namespace. If this annotation is set to an empty
			label, that means that Pods for this namespace can be scheduled on any nodes.
			In the case where a cluster administrator has configured a value for the
			defaultNodeSelector field in the cluster's Scheduler object, we need to set
			this namespace annotation in order to avoid a collision with what the user
			has supplied in their MeteringConfig custom resource. This implies that whenever
			a cluster has been configured to schedule Pods using a default node selector,
			those changes must also be propogated to the MeteringConfig custom resource, else
			the Pods in Metering namespace will be scheduled on any available node.
		*/
		if deploy.Config.Platform == "openshift" {
			labels["openshift.io/cluster-monitoring"] = "true"
			deploy.Logger.Infof("Labeling the %s namespace with 'openshift.io/cluster-monitoring=true'", deploy.Config.Namespace)
			namespaceObjectMeta.Annotations = map[string]string{
				"openshift.io/node-selector": "",
			}
			deploy.Logger.Infof("Annotating the %s namespace with 'openshift.io/node-selector=''", deploy.Config.Namespace)
		}

		namespaceObjectMeta.Labels = labels
		namespaceObj := &v1.Namespace{
			ObjectMeta: namespaceObjectMeta,
		}

		_, err := deploy.Client.CoreV1().Namespaces().Create(context.TODO(), namespaceObj, metav1.CreateOptions{})
		if err != nil {
			return err
		}
		deploy.Logger.Infof("Created the %s namespace", deploy.Config.Namespace)
		return nil
	}
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}

	// check if we need to add/update the cluster-monitoring label for Openshift installs.
	if deploy.Config.Platform == "openshift" {
		if namespace.ObjectMeta.Labels != nil {
			namespace.ObjectMeta.Labels["openshift.io/cluster-monitoring"] = "true"
			deploy.Logger.Infof("Updated the 'openshift.io/cluster-monitoring' label to the %s namespace", deploy.Config.Namespace)
		} else {
			namespace.ObjectMeta.Labels = map[string]string{
				"openshift.io/cluster-monitoring": "true",
			}
			deploy.Logger.Infof("Added the 'openshift.io/cluster-monitoring' label to the %s namespace", deploy.Config.Namespace)
		}
		if namespace.ObjectMeta.Annotations != nil {
			namespace.ObjectMeta.Annotations["openshift.io/node-selector"] = ""
			deploy.Logger.Infof("Updated the 'openshift.io/node-selector' annotation to the %s namespace", deploy.Config.Namespace)
		} else {
			namespace.ObjectMeta.Annotations = map[string]string{
				"openshift.io/node-selector": "",
			}
			deploy.Logger.Infof("Added the empty 'openshift.io/node-selector' annotation to the %s namespace", deploy.Config.Namespace)
		}

		_, err := deploy.Client.CoreV1().Namespaces().Update(context.TODO(), namespace, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("failed to add the 'openshift.io/cluster-monitoring' label to the %s namespace: %v", deploy.Config.Namespace, err)
		}
		return nil
	}

	// TODO: handle updating the namespace for non-openshift installations
	deploy.Logger.Infof("The %s namespace already exists", deploy.Config.Namespace)

	return nil
}

func (deploy *Deployer) installMeteringConfig() error {
	if deploy.Config.MeteringConfig == nil {
		return fmt.Errorf("invalid deploy configuration: MeteringConfig object is nil")
	}
	if deploy.Config.MeteringConfig.Name == "" {
		return fmt.Errorf("invalid deploy configuration: metadata.Name is unset")
	}

	// ensure the MeteringConfig CRD has already been created to avoid
	// any errors while instantiating a MeteringConfig custom resource
	err := wait.Poll(crdInitialPoll, crdPollTimeout, func() (done bool, err error) {
		mc, err := deploy.APIExtClient.CustomResourceDefinitions().Get(context.TODO(), meteringconfigCRDName, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			deploy.Logger.Infof("Waiting for the MeteringConfig CRD to be created")
			return false, nil
		}
		if err != nil {
			return false, err
		}
		// in order to handle the following error, ensure the Status field has a populated entry for the "plural" CRD name:
		// the server could not find the requested resource (post meteringconfigs.metering.openshift.io)
		if mc.Status.AcceptedNames.Plural != "meteringconfigs" {
			deploy.Logger.Infof("Waiting for the MeteringConfig CRD to be ready")
			return false, nil
		}

		return true, nil
	})
	if err != nil {
		return fmt.Errorf("failed to wait for the MeteringConfig CRD to be created: %v", err)
	}
	deploy.Logger.Infof("The MeteringConfig CRD exists")

	mc, err := deploy.MeteringClient.MeteringConfigs(deploy.Config.Namespace).Get(context.TODO(), deploy.Config.MeteringConfig.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = deploy.MeteringClient.MeteringConfigs(deploy.Config.Namespace).Create(context.TODO(), deploy.Config.MeteringConfig, metav1.CreateOptions{})
		if err != nil {
			return err
		}
		deploy.Logger.Infof("Created the MeteringConfig resource")
		return nil
	}
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}

	mc.Spec = deploy.Config.MeteringConfig.Spec

	_, err = deploy.MeteringClient.MeteringConfigs(deploy.Config.Namespace).Update(context.TODO(), mc, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update the MeteringConfig: %v", err)
	}
	deploy.Logger.Infof("The MeteringConfig resource has been updated")

	return nil
}

func (deploy *Deployer) installMeteringOperatorGroup() error {
	deployNamespace := deploy.Config.Namespace

	opgrp, err := deploy.OLMV1Client.OperatorGroups(deployNamespace).Get(context.TODO(), deployNamespace, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		opgrp := &olmv1.OperatorGroup{
			ObjectMeta: metav1.ObjectMeta{
				Name:      deployNamespace,
				Namespace: deployNamespace,
			},
			Spec: olmv1.OperatorGroupSpec{
				TargetNamespaces: []string{
					deployNamespace,
				},
			},
		}

		_, err = deploy.OLMV1Client.OperatorGroups(deployNamespace).Create(context.TODO(), opgrp, metav1.CreateOptions{})
		if err != nil {
			return err
		}
		deploy.Logger.Infof("Created the %s metering OperatorGroup in the %s namespace", opgrp.Name, deployNamespace)
		return nil
	}
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}
	deploy.Logger.Infof("The %s metering OperatorGroup resource already exists", opgrp.Name)

	return nil
}

func (deploy *Deployer) installMeteringSubscription() error {
	deployNamespace := deploy.Config.Namespace
	subName := deploy.Config.SubscriptionName

	_, err := deploy.OLMV1Alpha1Client.Subscriptions(deployNamespace).Get(context.TODO(), subName, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		sub := &olmv1alpha1.Subscription{
			ObjectMeta: metav1.ObjectMeta{
				Name:      subName,
				Namespace: deployNamespace,
			},
			Spec: &olmv1alpha1.SubscriptionSpec{
				CatalogSource:          deploy.Config.CatalogSourceName,
				CatalogSourceNamespace: deploy.Config.CatalogSourceNamespace,
				Package:                deploy.Config.PackageName,
				Channel:                deploy.Config.Channel,
				InstallPlanApproval:    olmv1alpha1.ApprovalAutomatic,
			},
		}

		_, err := deploy.OLMV1Alpha1Client.Subscriptions(deployNamespace).Create(context.TODO(), sub, metav1.CreateOptions{})
		if err != nil {
			return err
		}
		deploy.Logger.Infof("Created the %s metering Subscription in the %s namespace", subName, deployNamespace)
		return nil
	}
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}
	deploy.Logger.Infof("The %s metering Subscription in the %s namespace already exists", subName, deployNamespace)

	return nil
}

func (deploy *Deployer) installMeteringResources() error {
	if !deploy.Config.RunMeteringOperatorLocal {
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
	res := deploy.Config.OperatorResources.Deployment

	// check if the metering operator image needs to be updated
	// TODO: implement support for METERING_OPERATOR_ALL_NAMESPACES and METERING_OPERATOR_TARGET_NAMESPACES
	if deploy.Config.Repo != "" && deploy.Config.Tag != "" {
		newImage := deploy.Config.Repo + ":" + deploy.Config.Tag

		for index := range res.Spec.Template.Spec.Containers {
			res.Spec.Template.Spec.Containers[index].Image = newImage
		}

		deploy.Logger.Infof("Overriding the default image with %s", newImage)
	}

	deployment, err := deploy.Client.AppsV1().Deployments(deploy.Config.Namespace).Get(context.TODO(), res.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err := deploy.Client.AppsV1().Deployments(deploy.Config.Namespace).Create(context.TODO(), res, metav1.CreateOptions{})
		if err != nil {
			return err
		}
		deploy.Logger.Infof("Created the metering deployment")
		return nil
	}
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}

	deployment.Spec = res.Spec

	_, err = deploy.Client.AppsV1().Deployments(deploy.Config.Namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update the metering deployment: %v", err)
	}
	deploy.Logger.Infof("The metering deployment resource has been updated")

	return nil
}

func (deploy *Deployer) installMeteringServiceAccount() error {
	_, err := deploy.Client.CoreV1().ServiceAccounts(deploy.Config.Namespace).Get(context.TODO(), deploy.Config.OperatorResources.ServiceAccount.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err := deploy.Client.CoreV1().ServiceAccounts(deploy.Config.Namespace).Create(context.TODO(), deploy.Config.OperatorResources.ServiceAccount, metav1.CreateOptions{})
		if err != nil {
			return err
		}
		deploy.Logger.Infof("Created the metering serviceaccount")
		return nil
	}
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}

	deploy.Logger.Infof("The metering service account already exists")

	return nil
}

func (deploy *Deployer) installMeteringRoleBinding() error {
	res := deploy.Config.OperatorResources.RoleBinding
	// TODO: implement support for METERING_OPERATOR_TARGET_NAMESPACES
	res.Name = deploy.Config.Namespace + "-" + res.Name
	res.RoleRef.Name = res.Name
	res.Namespace = deploy.Config.Namespace

	for index := range res.Subjects {
		res.Subjects[index].Namespace = deploy.Config.Namespace
	}

	_, err := deploy.Client.RbacV1().RoleBindings(deploy.Config.Namespace).Get(context.TODO(), res.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err := deploy.Client.RbacV1().RoleBindings(deploy.Config.Namespace).Create(context.TODO(), res, metav1.CreateOptions{})
		if err != nil {
			return err
		}
		deploy.Logger.Infof("Created the metering role binding")
		return nil
	}
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}

	deploy.Logger.Infof("The metering role binding already exists")

	return nil
}

func (deploy *Deployer) installMeteringRole() error {
	res := deploy.Config.OperatorResources.Role
	res.Name = deploy.Config.Namespace + "-" + res.Name
	res.Namespace = deploy.Config.Namespace

	_, err := deploy.Client.RbacV1().Roles(deploy.Config.Namespace).Get(context.TODO(), res.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err := deploy.Client.RbacV1().Roles(deploy.Config.Namespace).Create(context.TODO(), res, metav1.CreateOptions{})
		if err != nil {
			return err
		}
		deploy.Logger.Infof("Created the metering role")
		return nil
	}
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}

	deploy.Logger.Infof("The metering role already exists")

	return nil
}

func (deploy *Deployer) installMeteringClusterRoleBinding() error {
	res := deploy.Config.OperatorResources.ClusterRoleBinding
	res.Name = deploy.Config.Namespace + "-" + res.Name
	res.RoleRef.Name = res.Name

	for index := range res.Subjects {
		res.Subjects[index].Namespace = deploy.Config.Namespace
	}

	_, err := deploy.Client.RbacV1().ClusterRoleBindings().Get(context.TODO(), res.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err := deploy.Client.RbacV1().ClusterRoleBindings().Create(context.TODO(), res, metav1.CreateOptions{})
		if err != nil {
			return err
		}
		deploy.Logger.Infof("Created the metering cluster role binding")
		return nil
	}
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}

	_, err = deploy.Client.RbacV1().ClusterRoleBindings().Update(context.TODO(), res, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update the metering clusterrolebinding: %v", err)
	}
	deploy.Logger.Infof("The metering cluster role binding has been updated")

	return nil
}

func (deploy *Deployer) installMeteringClusterRole() error {
	res := deploy.Config.OperatorResources.ClusterRole
	res.Name = deploy.Config.Namespace + "-" + res.Name

	clusterRole, err := deploy.Client.RbacV1().ClusterRoles().Get(context.TODO(), res.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err := deploy.Client.RbacV1().ClusterRoles().Create(context.TODO(), res, metav1.CreateOptions{})
		if err != nil {
			return err
		}
		deploy.Logger.Infof("Created the metering cluster role")
		return nil
	}
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}

	clusterRole.Rules = res.Rules
	_, err = deploy.Client.RbacV1().ClusterRoles().Update(context.TODO(), clusterRole, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update the metering clusterrole: %v", err)
	}
	deploy.Logger.Infof("The metering clusterrole has been updated")

	return err
}

func (deploy *Deployer) installMeteringCRDs() error {
	for _, crd := range deploy.Config.OperatorResources.CRDs {
		err := deploy.installMeteringCRD(crd)
		if err != nil {
			return fmt.Errorf("failed to create a CRD while looping: %v", err)
		}
	}

	return nil
}

func (deploy *Deployer) installMeteringCRD(resource CRD) error {
	crd, err := deploy.APIExtClient.CustomResourceDefinitions().Get(context.TODO(), resource.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err := deploy.APIExtClient.CustomResourceDefinitions().Create(context.TODO(), resource.CRD, metav1.CreateOptions{})
		if err != nil {
			return err
		}
		deploy.Logger.Infof("Created the %s CRD", resource.Name)
		return nil
	}
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}

	crd.Spec = resource.CRD.Spec
	_, err = deploy.APIExtClient.CustomResourceDefinitions().Update(context.TODO(), crd, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update the %s CRD: %v", resource.CRD.Name, err)
	}
	deploy.Logger.Infof("Updated the %s CRD", resource.CRD.Name)

	return nil
}
