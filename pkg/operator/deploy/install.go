package deploy

import (
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/api/core/v1"
)

func (deploy *Deployer) installNamespace() error {
	namespace, err := deploy.client.CoreV1().Namespaces().Get(deploy.config.Namespace, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		namespaceObjectMeta := metav1.ObjectMeta{
			Name: deploy.config.Namespace,
		}

		if deploy.config.Platform == "openshift" {
			namespaceObjectMeta.Labels = map[string]string{
				"openshift.io/cluster-monitoring": "true",
			}
			deploy.logger.Infof("Labeling the %s namespace with 'openshift.io/cluster-monitoring=true'", deploy.config.Namespace)
		}

		namespaceObj := &v1.Namespace{
			ObjectMeta: namespaceObjectMeta,
		}

		_, err := deploy.client.CoreV1().Namespaces().Create(namespaceObj)
		if err != nil {
			return fmt.Errorf("Failed to create %s namespace: %v", deploy.config.Namespace, err)
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
				return fmt.Errorf("Failed to add the 'openshift.io/cluster-monitoring' label to the %s namespace: %v", deploy.config.Namespace, err)
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
	mc, err := deploy.meteringClient.MeteringConfigs(deploy.config.Namespace).Get(deploy.config.Resources.MeteringConfig.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = deploy.meteringClient.MeteringConfigs(deploy.config.Namespace).Create(deploy.config.Resources.MeteringConfig)
		if err != nil {
			return fmt.Errorf("Failed to create the MeteringConfig resource: %v", err)
		}
		deploy.logger.Infof("Created the MeteringConfig resource")
	} else if err == nil {
		mc.Spec = deploy.config.Resources.MeteringConfig.Spec

		_, err = deploy.meteringClient.MeteringConfigs(deploy.config.Namespace).Update(mc)
		if err != nil {
			return fmt.Errorf("Failed to update the MeteringConfig: %v", err)
		}
		deploy.logger.Infof("The MeteringConfig resource has been updated")
	} else {
		return err
	}

	return nil
}

func (deploy *Deployer) installMeteringResources() error {
	err := deploy.installMeteringDeployment()
	if err != nil {
		return fmt.Errorf("Failed to create the metering deployment: %v", err)
	}

	err = deploy.installMeteringServiceAccount()
	if err != nil {
		return fmt.Errorf("Failed to create the metering service account: %v", err)
	}

	err = deploy.installMeteringRole()
	if err != nil {
		return fmt.Errorf("Failed to create the metering role: %v", err)
	}

	err = deploy.installMeteringRoleBinding()
	if err != nil {
		return fmt.Errorf("Failed to create the metering role binding: %v", err)
	}

	err = deploy.installMeteringClusterRole()
	if err != nil {
		return fmt.Errorf("Failed to create the metering cluster role: %v", err)
	}

	err = deploy.installMeteringClusterRoleBinding()
	if err != nil {
		return fmt.Errorf("Failed to create the metering cluster role binding: %v", err)
	}

	return nil
}

func (deploy *Deployer) installMeteringDeployment() error {
	res := deploy.config.Resources.Deployment

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
			return fmt.Errorf("Failed to create the metering deployment: %v", err)
		}
		deploy.logger.Infof("Created the metering deployment")
	} else if err == nil {
		deployment.Spec = res.Spec

		_, err = deploy.client.AppsV1().Deployments(deploy.config.Namespace).Update(deployment)
		if err != nil {
			return fmt.Errorf("Failed to update the metering deployment: %v", err)
		}
		deploy.logger.Infof("The metering deployment resource has been updated")
	} else {
		return err
	}

	return nil
}

func (deploy *Deployer) installMeteringServiceAccount() error {
	_, err := deploy.client.CoreV1().ServiceAccounts(deploy.config.Namespace).Get(deploy.config.Resources.ServiceAccount.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err := deploy.client.CoreV1().ServiceAccounts(deploy.config.Namespace).Create(deploy.config.Resources.ServiceAccount)
		if err != nil {
			return fmt.Errorf("Failed to create the metering serviceaccount: %v", err)
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
	res := deploy.config.Resources.RoleBinding

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
			return fmt.Errorf("Failed to create the metering role binding: %v", err)
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
	res := deploy.config.Resources.Role

	res.Name = deploy.config.Namespace + "-" + res.Name
	res.Namespace = deploy.config.Namespace

	_, err := deploy.client.RbacV1().Roles(deploy.config.Namespace).Get(res.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err := deploy.client.RbacV1().Roles(deploy.config.Namespace).Create(res)
		if err != nil {
			return fmt.Errorf("Failed to create the metering role: %v", err)
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
	res := deploy.config.Resources.ClusterRoleBinding

	res.Name = deploy.config.Namespace + "-" + res.Name
	res.RoleRef.Name = res.Name

	for index := range res.Subjects {
		res.Subjects[index].Namespace = deploy.config.Namespace
	}

	_, err := deploy.client.RbacV1().ClusterRoleBindings().Get(res.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err := deploy.client.RbacV1().ClusterRoleBindings().Create(res)
		if err != nil {
			return fmt.Errorf("Failed to create the metering cluster role, got: %v", err)
		}
		deploy.logger.Infof("Created the metering cluster role binding")
	} else if err == nil {
		deploy.logger.Infof("The metering cluster role binding already exists")
	} else {
		return err
	}

	return nil
}

func (deploy *Deployer) installMeteringClusterRole() error {
	res := deploy.config.Resources.ClusterRole

	res.Name = deploy.config.Namespace + "-" + res.Name

	_, err := deploy.client.RbacV1().ClusterRoles().Get(res.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err := deploy.client.RbacV1().ClusterRoles().Create(res)
		if err != nil {
			return fmt.Errorf("Failed to create the metering cluster role: %v", err)
		}
		deploy.logger.Infof("Created the metering cluster role")
	} else if err == nil {
		deploy.logger.Infof("The metering cluster role already exists")
	} else {
		return err
	}

	return nil
}

func (deploy *Deployer) installMeteringCRDs() error {
	for _, crd := range deploy.config.Resources.CRDs {
		err := deploy.installMeteringCRD(crd)
		if err != nil {
			return fmt.Errorf("Failed to create a CRD while looping: %v", err)
		}
	}

	return nil
}

func (deploy *Deployer) installMeteringCRD(resource CRD) error {
	crd, err := deploy.apiExtClient.CustomResourceDefinitions().Get(resource.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err := deploy.apiExtClient.CustomResourceDefinitions().Create(resource.CRD)
		if err != nil {
			return fmt.Errorf("Failed to create the %s CRD: %v", resource.CRD.Name, err)
		}
		deploy.logger.Infof("Created the %s CRD", resource.Name)
	} else if err == nil {
		crd.Spec = resource.CRD.Spec

		_, err := deploy.apiExtClient.CustomResourceDefinitions().Update(crd)
		if err != nil {
			return fmt.Errorf("Failed to update the %s CRD: %v", resource.CRD.Name, err)
		}
		deploy.logger.Infof("Updated the %s CRD", resource.CRD.Name)
	} else {
		return err
	}

	return nil
}
