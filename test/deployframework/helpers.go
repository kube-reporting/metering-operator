package deployframework

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	metering "github.com/operator-framework/operator-metering/pkg/apis/metering/v1"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
)

func checkPodStatus(pod v1.Pod) (bool, int) {
	if pod.Status.Phase != v1.PodRunning {
		return false, 0
	}

	var unreadyContainers int

	for _, status := range pod.Status.ContainerStatuses {
		if !status.Ready {
			unreadyContainers++
		}
	}

	return unreadyContainers == 0, len(pod.Status.ContainerStatuses) - unreadyContainers
}

func createResourceDirs(namespace, path string) ([]string, error) {
	envVarArr := []string{
		"METERING_TEST_NAMESPACE=" + namespace,
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
			return nil, fmt.Errorf("failed to create the directory %s: %v", dirPath, err)
		}

		envVarArr = append(envVarArr, env+"="+dirPath)
	}

	return envVarArr, nil
}

func logPollingSummary(logger logrus.FieldLogger, targetPods int, readyPods []string, unreadyPods []podStat) {
	logger.Infof("Poll Summary")
	logger.Infof("Current ratio of ready to target pods: %d/%d", len(readyPods), targetPods)

	for _, unreadyPod := range unreadyPods {
		if unreadyPod.Total == 0 {
			logger.Infof("Pod %s is pending", unreadyPod.PodName)
			continue
		}
		logger.Infof("Pod %s has %d/%d ready containers", unreadyPod.PodName, unreadyPod.Ready, unreadyPod.Total)
	}
}

func validateImageConfig(image metering.ImageConfig) error {
	var errArr []string

	if image.Repository == "" {
		errArr = append(errArr, "the image repository is empty")
	}
	if image.Tag == "" {
		errArr = append(errArr, "the image tag is empty")
	}

	if len(errArr) != 0 {
		return fmt.Errorf(strings.Join(errArr, "\n"))
	}

	return nil
}
