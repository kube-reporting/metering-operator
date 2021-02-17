package deployframework

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"

	metering "github.com/kube-reporting/metering-operator/pkg/apis/metering/v1"
	"github.com/kube-reporting/metering-operator/pkg/deploy"
	meteringclient "github.com/kube-reporting/metering-operator/pkg/generated/clientset/versioned/typed/metering/v1"
	olmv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	olmclientv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/clientset/versioned/typed/operators/v1alpha1"
)

func checkPodStatus(pod corev1.Pod) (bool, int) {
	if pod.Status.Phase != corev1.PodRunning {
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
	OLMClient     olmclientv1alpha1.OperatorsV1alpha1Interface
}

type podStat struct {
	PodName string
	Ready   int
	Total   int
}

// ErrInstallPlanFailed represents a failing InstallPlan during an
// individual metering installation instantiated by OLM.
var ErrInstallPlanFailed = errors.New("detected failing InstallPlan")

// handleInstallPlanFailures is a helper function responsible for managing
// any potential InstallPlan custom resources that are "problematic" to avoid
// unnecessary e2e testing flakes. We define "problematic" here as any InstallPlan
// custom resource that is reporting a `Failed` status, in addition to containing
// at least one conditions index message alluding to limitations outside of our
// control. This is a workaround for the https://bugzilla.redhat.com/show_bug.cgi?id=1923111
// OLM bug, where any failure in an InstallPlan is treated as a terminal error,
// regardless of the context of the error.
//
// TODO(tflannag): This function will no longer be necessary once that above BZ
// is addressed and backported to the release branches we care about.
func handleInstallPlanFailures(client olmclientv1alpha1.OperatorsV1alpha1Interface, logger logrus.FieldLogger, namespace string) error {
	ips, err := client.InstallPlans(namespace).List(context.Background(), metav1.ListOptions{
		LabelSelector: fmt.Sprintf("operators.coreos.com/metering-ocp.%s=", namespace),
	})
	if err != nil {
		return nil
	}

	// Handle the potential case where no InstallPlan has been created yet, e.g.
	// we just created the Subscription for the metering-* package.
	if len(ips.Items) == 0 {
		return nil
	}
	// Handle case where there's multiple InstallPlans in a single namespace
	// and exit early. We can make a relatively safe assumption that we'll
	// likely only encounter a single replica for failed installations.
	if len(ips.Items) != 1 {
		logger.Info("skipping any potential InstallPlan failures as there are multiple replicas in the %s namespace", namespace)
		return nil
	}

	// Continue filtering out any non-problematic, at least currently,
	// InstallPlan custom resources and check again at the next poll.
	ip := ips.Items[0]
	if ip.Status.Phase != olmv1alpha1.InstallPlanPhaseFailed {
		return nil
	}

	var olmBugDetected bool
	for _, condition := range ip.Status.Conditions {
		if strings.Contains(condition.Message, "object has been modified") {
			olmBugDetected = true
			break
		}
	}
	if !olmBugDetected {
		return nil
	}

	logger.Infof("Identified the failing %s InstallPlan due to object modification", ip.Name)
	err = client.InstallPlans(namespace).Delete(context.Background(), ip.Name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	logger.Infof("Deleted the failing %s InstallPlan due to object modification", ip.Name)

	return ErrInstallPlanFailed
}

// WaitForPods periodically polls the list of pods in the namespace
// and ensures the metering pods created are considered ready. In order to exit
// the polling loop, the number of pods listed must match the expected number
// of targetPodsCount, and all pod containers listed must report a ready status.
func (pw *PodWaiter) WaitForPods(namespace string, targetPodsCount int) error {
	// TODO: generalize this more and pass a metav1.ListOptions parameter
	err := wait.Poll(pw.InitialDelay, pw.TimeoutPeriod, func() (done bool, err error) {
		var readyPods []string
		var unreadyPods []podStat

		pods, err := pw.Client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
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

		// Check if all of the Pods are reporting a Ready status, exit early.
		// In the case where there are still unready Pods, but there's at least a single
		// Pod that's ready, we can reasonable assume that we aren't dealing with a failed
		// metering installation, at least in the context of an OLM deployment, so aavoid
		// checking if there are any problematic InstallPlan custom resources. Else, we
		// do need to list and filter any InstallPlan(s) that are reporting a failed status,
		// and that failed status is due to something out of our control, like object modification
		// being reported by the apiserver and OLM not handling that error correctly.
		ready := len(pods.Items) == targetPodsCount && len(unreadyPods) == 0
		if ready {
			return ready, nil
		}
		if len(unreadyPods) > 0 {
			return false, nil
		}
		err = handleInstallPlanFailures(pw.OLMClient, pw.Logger, namespace)
		if err != nil {
			return false, err
		}

		return ready, nil
	})
	if err != nil {
		if err == ErrInstallPlanFailed {
			pw.Logger.Infof("Encountered failing installation -- requeuing")
			return ErrInstallPlanFailed
		}
		return fmt.Errorf("the pods failed to report a ready status before the timeout period occurred: %v", err)
	}

	return nil
}

// GetServiceAccountToken queries the namespace for the service account and attempts
// to find the secret that contains the serviceAccount token and return it.
func GetServiceAccountToken(client kubernetes.Interface, initialDelay, timeoutPeriod time.Duration, namespace, serviceAccountName string) (string, error) {
	var (
		sa  *corev1.ServiceAccount
		err error
	)
	err = wait.Poll(initialDelay, timeoutPeriod, func() (done bool, err error) {
		sa, err = client.CoreV1().ServiceAccounts(namespace).Get(context.TODO(), serviceAccountName, metav1.GetOptions{})
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

	secret, err := client.CoreV1().Secrets(namespace).Get(context.TODO(), secretName, metav1.GetOptions{})
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

	scanner := bufio.NewScanner(cleanupStdout)
	go func() {
		for scanner.Scan() {
			logger.Infof(scanner.Text())
		}
	}()

	cleanupCmd.Env = append(os.Environ(), envVarArr...)
	err = cleanupCmd.Run()
	if err != nil {
		// TODO(tflannag): we need to add more flexibility to this
		// function, especially in the case where we expect that a
		// test case will fail, and it did fail, but the gather test
		// install artifacts scripts will return a non-zero exit code
		// as it cannot successfully log any resources. The workaround
		// for now is to log the error, but don't return an error.
		logger.Infof("%v", err)
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

func CreateCatalogSourceFromConfigMap(logger logrus.FieldLogger, name, namespace, configMapName string, client olmclientv1alpha1.OperatorsV1alpha1Interface) error {
	// check if the @name CatalogSource already exists and if true, exit early.
	// If no CatalogSource exists by that name, start building up that object
	// and attempt to create it through the OLM v1alpha1 client.
	_, err := client.CatalogSources(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		catsrc := &olmv1alpha1.CatalogSource{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: olmv1alpha1.CatalogSourceSpec{
				SourceType:  olmv1alpha1.SourceTypeConfigmap,
				ConfigMap:   configMapName,
				DisplayName: configMapName,
				Publisher:   "Red Hat",
			},
		}
		_, err := client.CatalogSources(namespace).Create(context.TODO(), catsrc, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create the %s CatalogSource for metering: %v", name, err)
		}
		logger.Infof("Created the %s CatalogSource", name)
	} else if err != nil {
		return err
	}

	return nil
}

// CreateCatalogSourceFromImage is responsible for attempting to create a
// CatalogSource custom resource based on an index image source type.
func CreateCatalogSourceFromImage(logger logrus.FieldLogger, client olmclientv1alpha1.OperatorsV1alpha1Interface, name, namespace, imageName string) (*olmv1alpha1.CatalogSource, error) {
	// check if the @name CatalogSource already exists and if true, exit early.
	// If no CatalogSource exists by that name, start building up that object
	// and attempt to create it through the OLM v1alpha1 client.
	catsrc, err := client.CatalogSources(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return nil, err
	}
	if apierrors.IsNotFound(err) {
		catsrc := &olmv1alpha1.CatalogSource{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
				Labels: map[string]string{
					"name": fmt.Sprintf("%s-%s", namespace, testNamespaceLabel),
				},
			},
			Spec: olmv1alpha1.CatalogSourceSpec{
				SourceType:  olmv1alpha1.SourceTypeGrpc,
				Image:       imageName,
				DisplayName: "Custom metering-ansible-operator",
				Publisher:   "Metering Dev",
			},
		}
		_, err := client.CatalogSources(namespace).Create(context.TODO(), catsrc, metav1.CreateOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create the %s CatalogSource for metering: %v", name, err)
		}
		logger.Infof("Created the %s CatalogSource", name)

		return catsrc, nil
	}

	return catsrc, err
}

// VerifyCatalogSourcePod is a deployframework helper function that checks the @namespace
// and verifies that there's a ready Pod that was created by an OLM CatalogSource resource.
func VerifyCatalogSourcePod(logger logrus.FieldLogger, client kubernetes.Interface, packageName, namespace string) error {
	// polling every three seconds, list all of the Pods in the @namespace, checking
	// if any of those Pods match the `olm.catalogSource=@packageName` label selector.
	// Continue polling until a single Pod is returned by that label selector query
	// and that Pod is reporting a Ready stauts, or stop when the timeout period is reached.
	err := wait.Poll(3*time.Second, 1*time.Minute, func() (done bool, err error) {
		pods, err := client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
			LabelSelector: fmt.Sprintf("olm.catalogSource=%s", packageName),
		})
		if err != nil {
			return false, err
		}
		if len(pods.Items) != 1 {
			return false, nil
		}

		for _, pod := range pods.Items {
			podIsReady, _ := checkPodStatus(pod)
			if !podIsReady {
				logger.Infof("Waiting for the %s Pod to become Ready", pod.Name)
				return false, nil
			}
		}

		return true, nil
	})
	if err != nil {
		return fmt.Errorf("failed to wait for the %s catalogsource Pod to become ready: %v", packageName, err)
	}

	return nil
}

// CreateUpgradeConfigMap is a helper function responsible for creating a ConfigMap
// that contains the current version of the repositories' CRDs, CSV and metering-ocp package
// which OLM can then consume through a CatalogSource. In order to create this ConfigMap,
// we execute a bash script that handles the heavy-lifting, overriding any of the environment
// variables that the script uses, to match our current deployment context.
func CreateUpgradeConfigMap(logger logrus.FieldLogger, name, namespace, scriptPath string) error {
	/*
		Check if we are running in CI by getting the value of the
		IMAGE_FORMAT environment variable that CI builds and exposes
		for our job. If this value is non-empty, then the "update
		configmap" script will override the containerImage field in the CSV.
		Else, the containerImage will use the default origin images.

		More information:
		https://github.com/openshift/ci-tools/blob/master/TEMPLATES.md#image_format
	*/
	imageOverride := os.Getenv("IMAGE_FORMAT")
	if imageOverride != "" {
		imageOverride = strings.Replace(imageOverride, "${component}", "metering-ansible-operator", 1)
	}

	envVarArr := []string{
		"IMAGE_OVERRIDE=" + imageOverride,
		"NAMESPACE=" + namespace,
		"NAME=" + name,
	}

	// build up the path to the ./hack/@scriptPath and stat that path,
	// verifying it exists before running that bash script
	relPath := filepath.Join(scriptPath, createUpgradeConfigMapScriptName)
	createConfigMapScript, err := filepath.Abs(relPath)
	if err != nil {
		return fmt.Errorf("failed to get the absolute path for the '%s' path: %v", relPath, err)
	}
	_, err = os.Stat(createConfigMapScript)
	if err != nil {
		return fmt.Errorf("failed to stat the '%s' path: %v", createConfigMapScript, err)
	}

	cmd := exec.Command(createConfigMapScript)
	cmd.Env = append(os.Environ(), envVarArr...)
	stderr, _ := cmd.StderrPipe()

	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start running the %s script", createConfigMapScript)
	}

	scanner := bufio.NewScanner(stderr)
	scanner.Split(bufio.ScanWords)
	for scanner.Scan() {
		m := scanner.Text()
		fmt.Println(m)
	}

	// TODO(tflannag): add a timeout function that kills the cmd.Process
	// https://medium.com/@vCabbage/go-timeout-commands-with-os-exec-commandcontext-ba0c861ed738
	// https://github.com/golang/go/issues/9580#issuecomment-69724465
	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("failed to wait until the %s script has finished running", createConfigMapScript)
	}

	return nil
}

// VerifyConfigMap is a helper function that polls until the @name ConfigMap
// has been created in the @namespace namespace.
func VerifyConfigMap(logger logrus.FieldLogger, client kubernetes.Interface, name, namespace string) error {
	err := wait.Poll(1*time.Second, 45*time.Second, func() (done bool, err error) {
		_, err = client.CoreV1().ConfigMaps(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		if err != nil {
			return false, err
		}

		return true, nil
	})
	if err != nil {
		return fmt.Errorf("failed to wait for the %s configmap to be created in the %s namespace: %v", name, namespace, err)
	}
	logger.Infof("The %s ConfigMap has been created in the %s namespace", name, namespace)

	return nil
}

// UpdateExistingSubscription is a helper function responsible for upgrading an existing metering-ocp Subscription
// to use the newest payload and verify that the Subscription object is reporting a successful upgrade status.
func UpdateExistingSubscription(logger logrus.FieldLogger, client olmclientv1alpha1.OperatorsV1alpha1Interface, name, namespace, catalogSourceName, catalogSourceNamespace, upgradeSubscriptionChannel string) error {
	sub, err := client.Subscriptions(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		return fmt.Errorf("the %s subscription does not exist", name)
	}
	if err != nil {
		return err
	}

	// update the Subscription to use the most recent channel listed in the package.yaml
	// and change the Subscription source type to use the contents of a CatalogSource.
	sub.Spec.CatalogSource = catalogSourceName
	sub.Spec.CatalogSourceNamespace = catalogSourceNamespace
	sub.Spec.Channel = upgradeSubscriptionChannel
	_, err = client.Subscriptions(namespace).Update(context.TODO(), sub, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	logger.Infof("Updated the %s Subscription to use the %s channel", name, upgradeSubscriptionChannel)

	// after updating the metering-ocp Subscription to use a newer channel,
	// wait until this object is reporting a successful upgrade state before
	// transferring control back to the function call site.
	err = wait.Poll(3*time.Second, 1*time.Minute, func() (done bool, err error) {
		sub, err := client.Subscriptions(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		if err != nil {
			return false, err
		}

		logger.Infof("Waiting for the %s Subscription to finish upgrading", name)
		if !strings.Contains(sub.Status.CurrentCSV, upgradeSubscriptionChannel) {
			logger.Infof("Subscription status does not report metering-operator-v%s as the currentCSV", upgradeSubscriptionChannel)
			return false, nil
		}
		if sub.Status.State != olmv1alpha1.SubscriptionStateAtLatest {
			logger.Infof("Subscription status has not reported AtLatestKnown yet")
			return false, nil
		}

		return true, nil
	})
	if err != nil {
		return fmt.Errorf("failed to wait for the %s subscription to finish updating in the %s namespace: %v", name, namespace, err)
	}
	return nil
}

// WaitForMeteringOperatorDeployment is a helper function that will poll for the @name
// deployment every ten seconds, waiting until that deployment reports a single signed
// 32-bit integer for both of the UpdatedReplicas and Replicas status fields, which will
// indicate a successful upgrade status.
func WaitForMeteringOperatorDeployment(logger logrus.FieldLogger, client kubernetes.Interface, name, namespace string) error {
	err := wait.Poll(10*time.Second, 10*time.Minute, func() (done bool, err error) {
		deployment, err := client.AppsV1().Deployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		if err != nil {
			return false, err
		}

		logger.Infof("Waiting for the %s Deployment status to report a successful upgrade.", deployment.Name)
		return deployment.Status.UpdatedReplicas == int32(1) && deployment.Status.Replicas == int32(1), nil
	})
	if err != nil {
		return fmt.Errorf("failed to wait for the %s Deployment to finish updating in the %s namespace: %v", name, namespace, err)
	}
	logger.Infof("The %s Deployment has reported a successful upgrade status", name)

	return nil
}

// WaitForReportingOperatorDeployment is a helper function that will poll for the @name
// deployment every twenty seconds, waiting until that deployment reports a successful
// upgrade status. Note: the reporting-operator deployment uses a RollingUpdate strategy
// which means we need to be careful about marking a deployment as "Ready" when there's
// two reporting-operator Pods in the @namespace. This means we should instead keep
// polling until there's a single replica.
func WaitForReportingOperatorDeployment(logger logrus.FieldLogger, client kubernetes.Interface, name, namespace string) error {
	err := wait.Poll(20*time.Second, 10*time.Minute, func() (done bool, err error) {
		deployment, err := client.AppsV1().Deployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		if err != nil {
			return false, err
		}

		logger.Infof("Waiting for the %s Deployment status to report a successful upgrade.", deployment.Name)
		return deployment.Status.UpdatedReplicas == int32(1) && deployment.Status.Replicas == int32(1) && deployment.Status.ObservedGeneration == int64(2), nil
	})
	if err != nil {
		return fmt.Errorf("failed to wait for the %s Deployment to finish updating in the %s namespace: %v", name, namespace, err)
	}
	logger.Infof("The %s Deployment has reported a successful upgrade status", name)

	return nil
}

func WaitForReportDataSources(logger logrus.FieldLogger, client meteringclient.MeteringV1Interface, namespace string) error {
	err := wait.Poll(10*time.Second, 5*time.Minute, func() (done bool, err error) {
		dataSources, err := client.ReportDataSources(namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return false, err
		}

		logger.Infof("Waiting for the ReportDataSoures to exist in the %s namespace", namespace)
		return len(dataSources.Items) != 0, nil
	})
	if err != nil {
		return fmt.Errorf("failed to wait %s namespace to existing ReportDataSources: %v", namespace, err)
	}
	logger.Infof("The %s namespace has ReportDataSources present", namespace)

	return nil
}

func DeleteAllTestReports(logger logrus.FieldLogger, client meteringclient.MeteringV1Interface, namespace string) error {
	err := client.Reports(namespace).DeleteCollection(context.TODO(), metav1.DeleteOptions{}, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete all the Reports in the %s namespace: %v", namespace, err)
	}
	logger.Infof("Deleted all of the Reports in the %s namespace", namespace)

	return nil
}

func DeleteAllReportDataSources(logger logrus.FieldLogger, client meteringclient.MeteringV1Interface, namespace string) error {
	err := client.ReportDataSources(namespace).DeleteCollection(context.TODO(), metav1.DeleteOptions{}, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete all the ReportDataSources in the %s namespace: %v", namespace, err)
	}
	logger.Infof("Deleted all of the ReportDataSources in the %s namespace", namespace)

	return nil
}

// CreateRegistryDeployment is a helper function responsible for building up
// and creating a local registry Deployment based on the @manifestPath YAML manifest
// object. In the case where the metering-operator or reporting-operator images have
// been overrided, we need to do some manipulation of the decoded deployment object
// to ensure we propagate those custom images when manipulating the CSV images.
func CreateRegistryDeployment(
	logger logrus.FieldLogger,
	client kubernetes.Interface,
	namespacePrefix,
	manifestPath,
	registryImage,
	meteringOperatorImage,
	reportingOperatorImage,
	namespace string,
) error {
	if registryImage == "" {
		return fmt.Errorf("error: the registry image parameter is empty")
	}

	var deployment *appsv1.Deployment
	err := deploy.DecodeYAMLManifestToObject(manifestPath, &deployment)
	if err != nil {
		return fmt.Errorf("failed to decode the %s YAML deployment manifest: %v", manifestPath, err)
	}

	// Add a namespace prefix to the decoded object to avoid any collisions with other e2e instances
	deployment.Name = namespacePrefix + "-" + deployment.Name
	deployment.Namespace = namespace
	if deployment.Labels != nil {
		deployment.Labels["name"] = namespacePrefix + "-" + testNamespaceLabel
		deployment.Labels["tenant"] = namespacePrefix
	}

	// Update the invalid manifest registry image with the @registryImage provided
	deployment.Spec.Template.Spec.Containers[0].Image = registryImage
	deployment.Spec.Template.Spec.InitContainers[0].Image = registryImage

	/*
		In the case where we're running the testing suite in CI,
		or if a dev has specified the METERING_OPERATOR_IMAGE_* or
		REPORTING_OPERATOR_IMAGE_* environment variables, then update
		the deployment manifest with these values.

		Note: we can somewhat safely make the assumption that the environment
		variable for the metering-operator is going to be at the first
		index in that initContainer .Env array, and the reporting-operator
		is going to be in the second index, but this is a bit gross, so we
		need to figure out a better way of manipulating this manifest without
		blindly indexing into this the environment array.
	*/
	if meteringOperatorImage != "" {
		logger.Infof("Overriding the default metering-ansible-operator image with %s", meteringOperatorImage)
		deployment.Spec.Template.Spec.InitContainers[0].Env[0].Value = meteringOperatorImage
	}
	if reportingOperatorImage != "" {
		logger.Infof("Overriding the default metering-reporting-operator image with %s", reportingOperatorImage)
		deployment.Spec.Template.Spec.InitContainers[0].Env[1].Value = reportingOperatorImage
	}
	// Populate the tenant label throughout the deployment manifest
	// that we're currently build up. By default, that label's value
	// is set to "", so we're injecting the namespacePrefix before creation.
	deployment.Spec.Selector.MatchLabels["tenant"] = namespacePrefix
	deployment.Spec.Template.ObjectMeta.Labels["tenant"] = namespacePrefix

	_, err = client.AppsV1().Deployments(namespace).Get(context.TODO(), deployment.Name, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	if apierrors.IsNotFound(err) {
		deployment, err := client.AppsV1().Deployments(namespace).Create(context.TODO(), deployment, metav1.CreateOptions{})
		if err != nil {
			return err
		}
		logger.Infof("Created the %s registry deployment in the %s namespace", deployment.Name, deployment.Namespace)
	}

	return nil
}

// CreateRegistryService is a helper function responsible for decoding the
// @manifestPath registry Service YAML manifest into a Go object, which we
// then attempt to create in the @namespace. We're namely interested in the
// spec.ClusterIP that gets populated once the service object has been created,
// which we can then reference in a CatalogSource custom resource (`spec.addr`).
func CreateRegistryService(logger logrus.FieldLogger, client kubernetes.Interface, namespacePrefix, manifestPath, namespace string) (string, error) {
	var service *corev1.Service
	err := deploy.DecodeYAMLManifestToObject(manifestPath, &service)
	if err != nil {
		return "", fmt.Errorf("failed to decode the %s YAML service manifest: %v", manifestPath, err)
	}

	// Add a namespace prefix to the decoded object to avoid any collisions with other e2e instances
	service.Name = namespacePrefix + "-" + service.Name
	service.Namespace = namespace
	if service.Labels != nil {
		service.Labels["name"] = namespacePrefix + "-" + testNamespaceLabel
	}

	svc, err := client.CoreV1().Services(namespace).Get(context.TODO(), service.Name, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return "", err
	}
	if apierrors.IsNotFound(err) {
		// Add tenancy to the label selector to ensure we're only targeting a
		// single metering installation. Note: we can make the assumption
		// that we're indexing into a populated selector map as we decoded
		// the service YAML manifest, which already contains a label selector.
		service.Spec.Selector["tenant"] = namespacePrefix
		svc, err = client.CoreV1().Services(namespace).Create(context.TODO(), service, metav1.CreateOptions{})
		if err != nil {
			return "", err
		}
		logger.Infof("Created the %s registry service in the %s namespace", service.Name, service.Namespace)
	}

	// Check if the Spec.ClusterIP has been populated yet, and if not, continue to poll
	// until that string is non-empty before returning control back to the call site
	if svc.Spec.ClusterIP == "" {
		err := wait.Poll(time.Second, 10*time.Second, func() (done bool, err error) {
			svc, err := client.CoreV1().Services(namespace).Get(context.TODO(), service.Name, metav1.GetOptions{})
			if err != nil {
				return false, err
			}

			return svc.Spec.ClusterIP != "", nil
		})
		if err != nil {
			return "", fmt.Errorf("failed to wait for the %s registry service in the %s namespace to have a populated ClusterIP field: %v", service.Name, service.Namespace, err)
		}
	}

	return svc.Spec.ClusterIP, nil
}

// DeleteRegistryService is a helper function responsible for listing
// all of the registry services in the @namespace that match the
// @labelSelector parameters, and deleting any services that get
// returned from that list
func DeleteRegistryService(logger logrus.FieldLogger, client kubernetes.Interface, namespace, labelSelector string) error {
	var errArr []string

	services, err := client.CoreV1().Services(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if len(services.Items) == 0 {
		errArr = append(errArr, fmt.Sprintf("failed to list the registry service: %v", err))
	}
	for _, service := range services.Items {
		err = client.CoreV1().Services(service.Namespace).Delete(context.TODO(), service.Name, metav1.DeleteOptions{})
		if err != nil {
			errArr = append(errArr, fmt.Sprintf("failed to successfully delete the %s registry service: %v", service.Name, err))
		}
		logger.Infof("Deleted the %s registry service in the %s namespace", service.Name, service.Namespace)
	}

	if len(errArr) != 0 {
		return fmt.Errorf(strings.Join(errArr, "\n"))
	}

	return nil
}

// DeleteRegistryDeployment is a helper function responsible for listing
// all of the registry deployments in the @namespace that match the
// @labelSelector label selector, and deleting any deployments that get
// returned from that list.
func DeleteRegistryDeployment(logger logrus.FieldLogger, client kubernetes.Interface, namespace, labelSelector string) error {
	var errArr []string

	deployments, err := client.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		errArr = append(errArr, fmt.Sprintf("failed to successfully delete the the registry deployment: %v", err))
	}
	if len(deployments.Items) == 0 {
		errArr = append(errArr, fmt.Sprintf("failed to list the registry deployment: %v", err))
	}
	for _, deployment := range deployments.Items {
		err = client.AppsV1().Deployments(namespace).Delete(context.TODO(), deployment.Name, metav1.DeleteOptions{})
		if err != nil {
			errArr = append(errArr, fmt.Sprintf("failed to successfully delete the %s registry deployment in the %s namespace: %v", deployment.Name, deployment.Namespace, err))
		}
		logger.Infof("Deleted the %s registry deployment in the %s namespace", deployment.Name, deployment.Namespace)
	}

	if len(errArr) != 0 {
		return fmt.Errorf(strings.Join(errArr, "\n"))
	}

	return nil
}
