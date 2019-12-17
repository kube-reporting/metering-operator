package deployframework

import (
	"bufio"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"

	metering "github.com/operator-framework/operator-metering/pkg/apis/metering/v1"
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

type PodWaiter struct {
	InitialDelay  time.Duration
	TimeoutPeriod time.Duration
	Logger        logrus.FieldLogger
	Client        kubernetes.Interface
}

type podStat struct {
	PodName string
	Ready   int
	Total   int
}

// WaitForPods periodically polls the list of pods in the namespace
// and ensures the metering pods created are considered ready. In order to exit
// the polling loop, the number of pods listed must match the expected number
// of targetPodsCount, and all pod containers listed must report a ready status.
func (pw *PodWaiter) WaitForPods(namespace string, targetPodsCount int) error {
	err := wait.Poll(pw.InitialDelay, pw.TimeoutPeriod, func() (done bool, err error) {
		var readyPods []string
		var unreadyPods []podStat

		pods, err := pw.Client.CoreV1().Pods(namespace).List(meta.ListOptions{})
		if err != nil {
			return false, err
		}

		for _, pod := range pods.Items {
			podIsReady, readyContainers := checkPodStatus(pod)
			if podIsReady {
				readyPods = append(readyPods, pod.Name)
				continue
			}

			unreadyPods = append(unreadyPods, podStat{
				PodName: pod.Name,
				Ready:   readyContainers,
				Total:   len(pod.Status.ContainerStatuses),
			})
		}

		if pw.Logger != nil {
			logPollingSummary(pw.Logger, targetPodsCount, readyPods, unreadyPods)
		}

		return len(pods.Items) == targetPodsCount && len(unreadyPods) == 0, nil
	})
	if err != nil {
		return fmt.Errorf("the pods failed to report a ready status before the timeout period occurred: %v", err)
	}

	return nil
}

// GetServiceAccountToken queries the namespace for the service account and attempts
// to find the secret that contains the serviceAccount token and return it.
func GetServiceAccountToken(client kubernetes.Interface, initialDelay, timeoutPeriod time.Duration, namespace, serviceAccountName string) (string, error) {
	var (
		sa  *v1.ServiceAccount
		err error
	)

	err = wait.Poll(initialDelay, timeoutPeriod, func() (done bool, err error) {
		sa, err = client.CoreV1().ServiceAccounts(namespace).Get(serviceAccountName, meta.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				return false, nil
			}
			return false, err
		}

		return true, nil
	})
	if err != nil {
		return "", fmt.Errorf("error getting service account %s: %v", reportingOperatorServiceAccountName, err)
	}

	if len(sa.Secrets) == 0 {
		return "", fmt.Errorf("service account %s has no secrets", serviceAccountName)
	}

	var secretName string

	for _, secret := range sa.Secrets {
		if strings.Contains(secret.Name, "token") {
			secretName = secret.Name
			break
		}
	}

	if secretName == "" {
		return "", fmt.Errorf("%s service account has no token", serviceAccountName)
	}

	secret, err := client.CoreV1().Secrets(namespace).Get(secretName, meta.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed getting %s service account token secret: %v", serviceAccountName, err)
	}

	return string(secret.Data["token"]), nil
}

func waitForURLToReportStatusOK(logger logrus.FieldLogger, targetURL string, timeout time.Duration) error {
	u, err := url.Parse(targetURL)
	if err != nil {
		return fmt.Errorf("failed to parse the %s URL: %v", targetURL, err)
	}

	logger.Debugf("Waiting for the %s url to report a 200 status", u)

	err = wait.Poll(10*time.Second, timeout, func() (done bool, err error) {
		resp, err := http.Get(u.String())
		if err != nil {
			return false, nil
		}
		defer resp.Body.Close()

		return resp.StatusCode == http.StatusOK, nil
	})
	if err != nil {
		return fmt.Errorf("timed-out while waiting for the %s url to report a 200 status code: %v", u, err)
	}

	logger.Infof("The %s url reported a 200 status code", u)

	return nil
}

func runCleanupScript(logger logrus.FieldLogger, namespace, outputPath, scriptPath string) error {
	var errArr []string
	envVarArr, err := createResourceDirs(namespace, outputPath)
	if err != nil {
		errArr = append(errArr, fmt.Sprintf("failed to create the resource output directories: %v", err))
	}

	cleanupCmd := exec.Command(scriptPath)
	cleanupStdout, err := cleanupCmd.StdoutPipe()
	if err != nil {
		errArr = append(errArr, fmt.Sprintf("failed to create a pipe from command output to stdout: %v", err))
	}

	go func() {
		scanner := bufio.NewScanner(cleanupStdout)
		for scanner.Scan() {
			line := scanner.Text()
			logger.Infof(line)
		}
		if err := scanner.Err(); err != nil {
			errArr = append(errArr, fmt.Sprintf("failed to read the command output: %v", err))
		}
	}()

	cleanupCmd.Env = append(os.Environ(), envVarArr...)
	err = cleanupCmd.Run()
	if err != nil {
		errArr = append(errArr, fmt.Sprintf("failed to successfully run the cleanup script: %v", err))
	}

	if len(errArr) != 0 {
		return fmt.Errorf(strings.Join(errArr, "\n"))
	}

	return nil
}

func cleanupLocalCmds(logger logrus.FieldLogger, commands ...exec.Cmd) error {
	var errArr []string

	for _, cmd := range commands {
		logger.Infof("Sending an interrupt to the %s command (pid %d)", cmd.Path, cmd.Process.Pid)

		err := cmd.Process.Signal(os.Interrupt)
		if err != nil {
			errArr = append(errArr, fmt.Sprintf("failed to interrupt pid %d: %v", cmd.Process.Pid, err))
		}

		err = cmd.Wait()
		if err != nil {
			_, ok := err.(*exec.ExitError)
			if !ok {
				logger.Infof("There was an error while waiting for the %s command to finish running: %v", cmd.Path, err)
				errArr = append(errArr, fmt.Sprintf("failed to wait for the %s command to finish running: %v", cmd.Path, err))
			}
		}
	}

	if len(errArr) != 0 {
		return fmt.Errorf(strings.Join(errArr, "\n"))
	}

	return nil
}
