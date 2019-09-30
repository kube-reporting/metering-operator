package deploy

import (
	"fmt"
	"path/***REMOVED***lepath"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/api/core/v1"
)

func (deploy *Deployer) installNamespace() error {
	namespace, err := deploy.client.CoreV1().Namespaces().Get(deploy.con***REMOVED***g.Namespace, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		namespaceObjectMeta := metav1.ObjectMeta{
			Name: deploy.con***REMOVED***g.Namespace,
		}

		if deploy.con***REMOVED***g.Platform == "openshift" {
			namespaceObjectMeta.Labels = map[string]string{
				"openshift.io/cluster-monitoring": "true",
			}
			deploy.logger.Infof("Labeling the %s namespace with 'openshift.io/cluster-monitoring=true'", deploy.con***REMOVED***g.Namespace)
		}

		namespaceObj := &v1.Namespace{
			ObjectMeta: namespaceObjectMeta,
		}

		_, err := deploy.client.CoreV1().Namespaces().Create(namespaceObj)
		if err != nil {
			return fmt.Errorf("Failed to create %s namespace: %v", deploy.con***REMOVED***g.Namespace, err)
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
				return fmt.Errorf("Failed to add the 'openshift.io/cluster-monitoring' label to the %s namespace: %v", deploy.con***REMOVED***g.Namespace, err)
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
		_, err = deploy.meteringClient.MeteringCon***REMOVED***gs(deploy.con***REMOVED***g.Namespace).Create(&deploy.con***REMOVED***g.MeteringCon***REMOVED***g)
		if err != nil {
			return fmt.Errorf("Failed to create the MeteringCon***REMOVED***g resource: %v", err)
		}
		deploy.logger.Infof("Created the MeteringCon***REMOVED***g resource")
	} ***REMOVED*** if err == nil {
		mc.Spec = deploy.con***REMOVED***g.MeteringCon***REMOVED***g.Spec

		_, err = deploy.meteringClient.MeteringCon***REMOVED***gs(deploy.con***REMOVED***g.Namespace).Update(mc)
		if err != nil {
			return fmt.Errorf("Failed to update the MeteringCon***REMOVED***g: %v", err)
		}
		deploy.logger.Infof("The MeteringCon***REMOVED***g resource has been updated")
	} ***REMOVED*** {
		return err
	}

	return nil
}

func (deploy *Deployer) installMeteringResources() error {
	err := deploy.installMeteringDeployment(***REMOVED***lepath.Join(deploy.ansibleOperatorManifestsLocation, meteringDeploymentFile))
	if err != nil {
		return fmt.Errorf("Failed to create the metering deployment: %v", err)
	}

	err = deploy.installMeteringServiceAccount(***REMOVED***lepath.Join(deploy.ansibleOperatorManifestsLocation, meteringServiceAccountFile))
	if err != nil {
		return fmt.Errorf("Failed to create the metering service account: %v", err)
	}

	err = deploy.installMeteringRole(***REMOVED***lepath.Join(deploy.ansibleOperatorManifestsLocation, meteringRoleFile))
	if err != nil {
		return fmt.Errorf("Failed to create the metering role: %v", err)
	}

	err = deploy.installMeteringRoleBinding(***REMOVED***lepath.Join(deploy.ansibleOperatorManifestsLocation, meteringRoleBindingFile))
	if err != nil {
		return fmt.Errorf("Failed to create the metering role binding: %v", err)
	}

	err = deploy.installMeteringClusterRole(***REMOVED***lepath.Join(deploy.ansibleOperatorManifestsLocation, meteringClusterRoleFile))
	if err != nil {
		return fmt.Errorf("Failed to create the metering cluster role: %v", err)
	}

	err = deploy.installMeteringClusterRoleBinding(***REMOVED***lepath.Join(deploy.ansibleOperatorManifestsLocation, meteringClusterRoleBindingFile))
	if err != nil {
		return fmt.Errorf("Failed to create the metering cluster role binding: %v", err)
	}

	return nil
}

func (deploy *Deployer) installMeteringDeployment(deploymentName string) error {
	var res appsv1.Deployment

	err := DecodeYAMLManifestToObject(deploymentName, &res)
	if err != nil {
		return fmt.Errorf("Failed to decode the metering YAML manifest: %v", err)
	}

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
		_, err := deploy.client.AppsV1().Deployments(deploy.con***REMOVED***g.Namespace).Create(&res)
		if err != nil {
			return fmt.Errorf("Failed to create the metering deployment: %v", err)
		}
		deploy.logger.Infof("Created the metering deployment")
	} ***REMOVED*** if err == nil {
		deployment.Spec = res.Spec

		_, err = deploy.client.AppsV1().Deployments(deploy.con***REMOVED***g.Namespace).Update(deployment)
		if err != nil {
			return fmt.Errorf("Failed to update the metering deployment: %v", err)
		}
		deploy.logger.Infof("The metering deployment resource has been updated")
	} ***REMOVED*** {
		return err
	}

	return nil
}

func (deploy *Deployer) installMeteringServiceAccount(serviceAccountPath string) error {
	var res corev1.ServiceAccount

	err := DecodeYAMLManifestToObject(serviceAccountPath, &res)
	if err != nil {
		return fmt.Errorf("Failed to decode the YAML manifest: %v", err)
	}

	_, err = deploy.client.CoreV1().ServiceAccounts(deploy.con***REMOVED***g.Namespace).Get(res.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err := deploy.client.CoreV1().ServiceAccounts(deploy.con***REMOVED***g.Namespace).Create(&res)
		if err != nil {
			return fmt.Errorf("Failed to create the metering serviceaccount: %v", err)
		}
		deploy.logger.Infof("Created the metering serviceaccount")
	} ***REMOVED*** if err == nil {
		deploy.logger.Infof("The metering service account already exists")
	} ***REMOVED*** {
		return err
	}

	return nil
}

func (deploy *Deployer) installMeteringRoleBinding(roleBindingPath string) error {
	var res rbacv1.RoleBinding

	err := DecodeYAMLManifestToObject(roleBindingPath, &res)
	if err != nil {
		return fmt.Errorf("Failed to decode the YAML manifest: %v", err)
	}

	// TODO: implement support for METERING_OPERATOR_TARGET_NAMESPACES
	res.Name = deploy.con***REMOVED***g.Namespace + "-" + res.Name
	res.RoleRef.Name = res.Name
	res.Namespace = deploy.con***REMOVED***g.Namespace

	for index := range res.Subjects {
		res.Subjects[index].Namespace = deploy.con***REMOVED***g.Namespace
	}

	_, err = deploy.client.RbacV1().RoleBindings(deploy.con***REMOVED***g.Namespace).Get(res.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err := deploy.client.RbacV1().RoleBindings(deploy.con***REMOVED***g.Namespace).Create(&res)
		if err != nil {
			return fmt.Errorf("Failed to create the metering role binding: %v", err)
		}
		deploy.logger.Infof("Created the metering role binding")
	} ***REMOVED*** if err == nil {
		deploy.logger.Infof("The metering role binding already exists")
	} ***REMOVED*** {
		return err
	}

	return nil
}

func (deploy *Deployer) installMeteringRole(rolePath string) error {
	var res rbacv1.Role

	err := DecodeYAMLManifestToObject(rolePath, &res)
	if err != nil {
		return fmt.Errorf("Failed to decode the YAML manifest: %v", err)
	}

	res.Name = deploy.con***REMOVED***g.Namespace + "-" + res.Name
	res.Namespace = deploy.con***REMOVED***g.Namespace

	_, err = deploy.client.RbacV1().Roles(deploy.con***REMOVED***g.Namespace).Get(res.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err := deploy.client.RbacV1().Roles(deploy.con***REMOVED***g.Namespace).Create(&res)
		if err != nil {
			return fmt.Errorf("Failed to create the metering role: %v", err)
		}
		deploy.logger.Infof("Created the metering role")
	} ***REMOVED*** if err == nil {
		deploy.logger.Infof("The metering role already exists")
	} ***REMOVED*** {
		return err
	}

	return nil
}

func (deploy *Deployer) installMeteringClusterRoleBinding(clusterrolebindingFile string) error {
	var res rbacv1.ClusterRoleBinding

	err := DecodeYAMLManifestToObject(clusterrolebindingFile, &res)
	if err != nil {
		return fmt.Errorf("Failed to decode the YAML manifest: %v", err)
	}

	res.Name = deploy.con***REMOVED***g.Namespace + "-" + res.Name
	res.RoleRef.Name = res.Name

	for index := range res.Subjects {
		res.Subjects[index].Namespace = deploy.con***REMOVED***g.Namespace
	}

	_, err = deploy.client.RbacV1().ClusterRoleBindings().Get(res.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err := deploy.client.RbacV1().ClusterRoleBindings().Create(&res)
		if err != nil {
			return fmt.Errorf("Failed to create the metering cluster role, got: %v", err)
		}
		deploy.logger.Infof("Created the metering cluster role binding")
	} ***REMOVED*** if err == nil {
		deploy.logger.Infof("The metering cluster role binding already exists")
	} ***REMOVED*** {
		return err
	}

	return nil
}

func (deploy *Deployer) installMeteringClusterRole(clusterrolePath string) error {
	var res rbacv1.ClusterRole

	err := DecodeYAMLManifestToObject(clusterrolePath, &res)
	if err != nil {
		return fmt.Errorf("Failed to decode the YAML manifest: %v", err)
	}

	res.Name = deploy.con***REMOVED***g.Namespace + "-" + res.Name

	_, err = deploy.client.RbacV1().ClusterRoles().Get(res.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err := deploy.client.RbacV1().ClusterRoles().Create(&res)
		if err != nil {
			return fmt.Errorf("Failed to create the metering cluster role: %v", err)
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
	for _, crd := range deploy.crds {
		err := deploy.installMeteringCRD(crd)
		if err != nil {
			return fmt.Errorf("Failed to create a CRD while looping: %v", err)
		}
	}

	return nil
}

func (deploy *Deployer) installMeteringCRD(resource CRD) error {
	err := DecodeYAMLManifestToObject(resource.Path, resource.CRD)
	if err != nil {
		return fmt.Errorf("Failed to decode the YAML manifest: %v", err)
	}

	crd, err := deploy.apiExtClient.CustomResourceDe***REMOVED***nitions().Get(resource.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err := deploy.apiExtClient.CustomResourceDe***REMOVED***nitions().Create(resource.CRD)
		if err != nil {
			return fmt.Errorf("Failed to create the %s CRD: %v", resource.CRD.Name, err)
		}
		deploy.logger.Infof("Created the %s CRD", resource.Name)
	} ***REMOVED*** if err == nil {
		crd.Spec = resource.CRD.Spec

		_, err := deploy.apiExtClient.CustomResourceDe***REMOVED***nitions().Update(crd)
		if err != nil {
			return fmt.Errorf("Failed to update the %s CRD: %v", resource.CRD.Name, err)
		}
		deploy.logger.Infof("Updated the %s CRD", resource.CRD.Name)
	} ***REMOVED*** {
		return err
	}

	return nil
}
