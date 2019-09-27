package deployframework

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (df *DeployFramework) checkPodStatus(pod v1.Pod) (bool, int) {
	if pod.Status.Phase != v1.PodRunning {
		return false, 0
	}

	var unreadyContainers int

	for _, status := range pod.Status.ContainerStatuses {
		if !status.Ready {
			unreadyContainers++
		}
	}

	return (unreadyContainers == 0), (len(pod.Status.ContainerStatuses) - unreadyContainers)
}

func (df *DeployFramework) addE2ENamespaceLabel(namespace string) error {
	ns, err := df.Client.CoreV1().Namespaces().Get(namespace, meta.GetOptions{})
	if err != nil {
		return fmt.Errorf("Failed to get the %s namespace: %v", namespace, err)
	}

	if ns.ObjectMeta.Labels != nil {
		ns.ObjectMeta.Labels["name"] = "e2e-testing"
		df.Logger.Infof("Updated the 'name=e2e-testing' label to the %s namespace", namespace)
	} else {
		ns.ObjectMeta.Labels = map[string]string{
			"name": "e2e-testing",
		}
		df.Logger.Infof("Added the 'name=e2e-testing' label to the %s namespace", namespace)
	}

	_, err = df.Client.CoreV1().Namespaces().Update(ns)
	if err != nil {
		return fmt.Errorf("Failed to add the 'name=e2e-testing' label to the %s namespace: %v", namespace, err)
	}

	return nil
}
