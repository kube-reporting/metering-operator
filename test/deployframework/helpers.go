package deployframework

import (
	"fmt"
	"os"
	"path/filepath"

	v1 "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (df *DeployFramework) createResourceDirs(path string) ([]string, error) {
	envVarArr := []string{
		"METERING_TEST_NAMESPACE=" + df.Config.Namespace,
		"TEST_OUTPUT_DIR=" + path,
	}

	testDirsMap := map[string]string{
		logDir:              "LOG_DIR",
		reportsDir:          "REPORTS_DIR",
		meteringconfigDir:   "METERINGCONFIGS_DIR",
		datasourcesDir:      "DATASOURCES_DIR",
		reportqueriesDir:    "REPORTQUERIES_DIR",
		hivetablesDir:       "HIVETABLES_DIR",
		prestotablesDir:     "PRESTOTABLES_DIR",
		storagelocationsDir: "STORAGELOCATIONS_DIR",
	}

	for dirname, env := range testDirsMap {
		dirPath := filepath.Join(path, dirname)

		err := os.MkdirAll(dirPath, 0777)
		if err != nil {
			return nil, fmt.Errorf("Failed to create the directory %s: %v", dirPath, err)
		}

		envVarArr = append(envVarArr, env+"="+dirPath)
	}

	return envVarArr, nil
}

func (df *DeployFramework) logPollingSummary(targetPods int, readyPods []string, unreadyPods []podStat) {
	df.Logger.Infof("Poll Summary")
	df.Logger.Infof("Current ratio of ready to target pods: %d/%d", len(readyPods), targetPods)

	for _, unreadyPod := range unreadyPods {
		if unreadyPod.Total == 0 {
			df.Logger.Infof("Pod %s is pending", unreadyPod.PodName)
			continue
		}
		df.Logger.Infof("Pod %s has %d/%d ready containers", unreadyPod.PodName, unreadyPod.Ready, unreadyPod.Total)
	}
}

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
