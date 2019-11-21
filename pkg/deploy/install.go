package deploy

import (
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "k8s.io/api/core/v1"
)

func (deploy *Deployer) installNamespace() error {
	namespace, err := deploy.client.CoreV1().Namespaces().Get(deploy.con***REMOVED***g.Namespace, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		namespaceObjectMeta := metav1.ObjectMeta{
			Name: deploy.con***REMOVED***g.Namespace,
		}

		labels := make(map[string]string)

		for key, val := range deploy.con***REMOVED***g.ExtraNamespaceLabels {
			labels[key] = val
			deploy.logger.Infof("Labeling the %s namespace with '%s=%s'", deploy.con***REMOVED***g.Namespace, key, val)
		}

		if deploy.con***REMOVED***g.Platform == "openshift" {
			labels["openshift.io/cluster-monitoring"] = "true"
			deploy.logger.Infof("Labeling the %s namespace with 'openshift.io/cluster-monitoring=true'", deploy.con***REMOVED***g.Namespace)
		}

		namespaceObjectMeta.Labels = labels
		namespaceObj := &v1.Namespace{
			ObjectMeta: namespaceObjectMeta,
		}

		_, err := deploy.client.CoreV1().Namespaces().Create(namespaceObj)
		if err != nil {
			return err
		}
		deploy.logger.Infof("Created the %s namespace", deploy.con***REMOVED***g.Namespace)
	} ***REMOVED*** if err == nil {
		// check if we need to add/update the cluster-monitoring label for Openshift installs.
		if deploy.con***REMOVED***g.Platform == "openshift" {
			if namespace.ObjectMeta.Labels != nil {
				namespace.ObjectMeta.Labels["openshift.io/cluster-monitoring"] = "true"
				deploy.logger.Infof("Updated the 'openshift.io/cluster-monitoring' label to the %s namespace", deploy.con***REMOVED***g.Namespace)
			} ***REMOVED*** {
				namespace.ObjectMeta.Labels = map[string]string{
					"openshift.io/cluster-monitoring": "true",
				}
				deploy.logger.Infof("Added the 'openshift.io/cluster-monitoring' label to the %s namespace", deploy.con***REMOVED***g.Namespace)
			}

			_, err := deploy.client.CoreV1().Namespaces().Update(namespace)
			if err != nil {
				return fmt.Errorf("failed to add the 'openshift.io/cluster-monitoring' label to the %s namespace: %v", deploy.con***REMOVED***g.Namespace, err)
			}
		} ***REMOVED*** {
			deploy.logger.Infof("The %s namespace already exists", deploy.con***REMOVED***g.Namespace)
		}
	} ***REMOVED*** {
		return err
	}

	return nil
}

func (deploy *Deployer) installMeteringCon***REMOVED***g() error {
	mc, err := deploy.meteringClient.MeteringCon***REMOVED***gs(deploy.con***REMOVED***g.Namespace).Get(deploy.con***REMOVED***g.MeteringCon***REMOVED***g.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = deploy.meteringClient.MeteringCon***REMOVED***gs(deploy.con***REMOVED***g.Namespace).Create(deploy.con***REMOVED***g.MeteringCon***REMOVED***g)
		if err != nil {
			return err
		}
		deploy.logger.Infof("Created the MeteringCon***REMOVED***g resource")
	} ***REMOVED*** if err == nil {
		mc.Spec = deploy.con***REMOVED***g.MeteringCon***REMOVED***g.Spec

		_, err = deploy.meteringClient.MeteringCon***REMOVED***gs(deploy.con***REMOVED***g.Namespace).Update(mc)
		if err != nil {
			return fmt.Errorf("failed to update the MeteringCon***REMOVED***g: %v", err)
		}
		deploy.logger.Infof("The MeteringCon***REMOVED***g resource has been updated")
	} ***REMOVED*** {
		return err
	}

	return nil
}

func (deploy *Deployer) installMeteringResources() error {
	if !deploy.con***REMOVED***g.RunMeteringOperatorLocal {
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
	res := deploy.con***REMOVED***g.OperatorResources.Deployment

	// check if the metering operator image needs to be updated
	// TODO: implement support for METERING_OPERATOR_ALL_NAMESPACES and METERING_OPERATOR_TARGET_NAMESPACES
	if deploy.con***REMOVED***g.Repo != "" && deploy.con***REMOVED***g.Tag != "" {
		newImage := deploy.con***REMOVED***g.Repo + ":" + deploy.con***REMOVED***g.Tag

		for index := range res.Spec.Template.Spec.Containers {
			res.Spec.Template.Spec.Containers[index].Image = newImage
		}

		deploy.logger.Infof("Overriding the default image with %s", newImage)
	}

	deployment, err := deploy.client.AppsV1().Deployments(deploy.con***REMOVED***g.Namespace).Get(res.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err := deploy.client.AppsV1().Deployments(deploy.con***REMOVED***g.Namespace).Create(res)
		if err != nil {
			return err
		}
		deploy.logger.Infof("Created the metering deployment")
	} ***REMOVED*** if err == nil {
		deployment.Spec = res.Spec

		_, err = deploy.client.AppsV1().Deployments(deploy.con***REMOVED***g.Namespace).Update(deployment)
		if err != nil {
			return fmt.Errorf("failed to update the metering deployment: %v", err)
		}
		deploy.logger.Infof("The metering deployment resource has been updated")
	} ***REMOVED*** {
		return err
	}

	return nil
}

func (deploy *Deployer) installMeteringServiceAccount() error {
	_, err := deploy.client.CoreV1().ServiceAccounts(deploy.con***REMOVED***g.Namespace).Get(deploy.con***REMOVED***g.OperatorResources.ServiceAccount.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err := deploy.client.CoreV1().ServiceAccounts(deploy.con***REMOVED***g.Namespace).Create(deploy.con***REMOVED***g.OperatorResources.ServiceAccount)
		if err != nil {
			return err
		}
		deploy.logger.Infof("Created the metering serviceaccount")
	} ***REMOVED*** if err == nil {
		deploy.logger.Infof("The metering service account already exists")
	} ***REMOVED*** {
		return err
	}

	return nil
}

func (deploy *Deployer) installMeteringRoleBinding() error {
	res := deploy.con***REMOVED***g.OperatorResources.RoleBinding

	// TODO: implement support for METERING_OPERATOR_TARGET_NAMESPACES
	res.Name = deploy.con***REMOVED***g.Namespace + "-" + res.Name
	res.RoleRef.Name = res.Name
	res.Namespace = deploy.con***REMOVED***g.Namespace

	for index := range res.Subjects {
		res.Subjects[index].Namespace = deploy.con***REMOVED***g.Namespace
	}

	_, err := deploy.client.RbacV1().RoleBindings(deploy.con***REMOVED***g.Namespace).Get(res.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err := deploy.client.RbacV1().RoleBindings(deploy.con***REMOVED***g.Namespace).Create(res)
		if err != nil {
			return err
		}
		deploy.logger.Infof("Created the metering role binding")
	} ***REMOVED*** if err == nil {
		deploy.logger.Infof("The metering role binding already exists")
	} ***REMOVED*** {
		return err
	}

	return nil
}

func (deploy *Deployer) installMeteringRole() error {
	res := deploy.con***REMOVED***g.OperatorResources.Role

	res.Name = deploy.con***REMOVED***g.Namespace + "-" + res.Name
	res.Namespace = deploy.con***REMOVED***g.Namespace

	_, err := deploy.client.RbacV1().Roles(deploy.con***REMOVED***g.Namespace).Get(res.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err := deploy.client.RbacV1().Roles(deploy.con***REMOVED***g.Namespace).Create(res)
		if err != nil {
			return err
		}
		deploy.logger.Infof("Created the metering role")
	} ***REMOVED*** if err == nil {
		deploy.logger.Infof("The metering role already exists")
	} ***REMOVED*** {
		return err
	}

	return nil
}

func (deploy *Deployer) installMeteringClusterRoleBinding() error {
	res := deploy.con***REMOVED***g.OperatorResources.ClusterRoleBinding

	res.Name = deploy.con***REMOVED***g.Namespace + "-" + res.Name
	res.RoleRef.Name = res.Name

	for index := range res.Subjects {
		res.Subjects[index].Namespace = deploy.con***REMOVED***g.Namespace
	}

	_, err := deploy.client.RbacV1().ClusterRoleBindings().Get(res.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err := deploy.client.RbacV1().ClusterRoleBindings().Create(res)
		if err != nil {
			return err
		}
		deploy.logger.Infof("Created the metering cluster role binding")
	} ***REMOVED*** if err == nil {
		deploy.logger.Infof("The metering cluster role binding already exists")
	} ***REMOVED*** {
		return err
	}

	return nil
}

func (deploy *Deployer) installMeteringClusterRole() error {
	res := deploy.con***REMOVED***g.OperatorResources.ClusterRole

	res.Name = deploy.con***REMOVED***g.Namespace + "-" + res.Name

	_, err := deploy.client.RbacV1().ClusterRoles().Get(res.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err := deploy.client.RbacV1().ClusterRoles().Create(res)
		if err != nil {
			return err
		}
		deploy.logger.Infof("Created the metering cluster role")
	} ***REMOVED*** if err == nil {
		deploy.logger.Infof("The metering cluster role already exists")
	} ***REMOVED*** {
		return err
	}

	return nil
}

func (deploy *Deployer) installMeteringCRDs() error {
	for _, crd := range deploy.con***REMOVED***g.OperatorResources.CRDs {
		err := deploy.installMeteringCRD(crd)
		if err != nil {
			return fmt.Errorf("failed to create a CRD while looping: %v", err)
		}
	}

	return nil
}

func (deploy *Deployer) installMeteringCRD(resource CRD) error {
	crd, err := deploy.apiExtClient.CustomResourceDe***REMOVED***nitions().Get(resource.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err := deploy.apiExtClient.CustomResourceDe***REMOVED***nitions().Create(resource.CRD)
		if err != nil {
			return err
		}
		deploy.logger.Infof("Created the %s CRD", resource.Name)
	} ***REMOVED*** if err == nil {
		crd.Spec = resource.CRD.Spec

		_, err := deploy.apiExtClient.CustomResourceDe***REMOVED***nitions().Update(crd)
		if err != nil {
			return fmt.Errorf("failed to update the %s CRD: %v", resource.CRD.Name, err)
		}
		deploy.logger.Infof("Updated the %s CRD", resource.CRD.Name)
	} ***REMOVED*** {
		return err
	}

	return nil
}
