package deployframework

import (
	"bufio"
	"fmt"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/clientcmd"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	metering "github.com/operator-framework/operator-metering/pkg/generated/clientset/versioned/typed/metering/v1"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	apiextclientv1beta1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/operator-framework/operator-metering/pkg/operator/deploy"
	"k8s.io/client-go/kubernetes"
)

const (
	testDirname         = "tests"
	reportResultsDir    = "report_results"
	logDir              = "logs"
	meteringconfigDir   = "meteringconfigs"
	reportsDir          = "reports"
	datasourcesDir      = "reportdatasources"
	reportqueriesDir    = "reportqueries"
	hivetablesDir       = "hivetables"
	prestotablesDir     = "prestotables"
	storagelocationsDir = "storagelocations"
)

// DeployFramework contains all the information necessary to deploy
// different metering instances and run tests against them
type DeployFramework struct {
	NamespacePrefix   string
	ManifestsDir      string
	ReportingAPIURL   string
	LoggingPath       string
	CleanupScriptPath string
	Logger            logrus.FieldLogger
	Client            kubernetes.Interface
	Config            ReportingFrameworkConfig
	APIExtClient      apiextclientv1beta1.CustomResourceDefinitionsGetter
	MeteringClient    metering.MeteringV1Interface
	Deployer          *deploy.Deployer
}

// ReportingFrameworkConfig is a structure containing information
// needed to customize a ReportingFramework object
type ReportingFrameworkConfig struct {
	Namespace                   string
	KubeConfigPath              string
	UseKubeProxyForReportingAPI bool
	UseRouteForReportingAPI     bool
	HTTPSAPI                    bool
	ReportingAPIURL             string
	RouteBearerToken            string
	ReportResultsOutputPath     string
}

// New is the constructor function that creates and returns a new DeployFramework object
func New(cfg ReportingFrameworkConfig, logger logrus.FieldLogger, nsPrefix, manifestDir, cleanupScriptPath, loggingPath string) (*DeployFramework, error) {
	config, err := clientcmd.BuildConfigFromFlags("", cfg.KubeConfigPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to build a kube config from %s: %v", cfg.KubeConfigPath, err)
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize the k8s clientset: %v", err)
	}

	apiextClient, err := apiextclientv1beta1.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize the apiextensions clientset: %v", err)
	}

	meteringClient, err := metering.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize the metering clientset: %v", err)
	}

	deployFramework := &DeployFramework{
		Logger:            logger,
		Client:            client,
		APIExtClient:      apiextClient,
		MeteringClient:    meteringClient,
		NamespacePrefix:   nsPrefix,
		ManifestsDir:      manifestDir,
		CleanupScriptPath: cleanupScriptPath,
		LoggingPath:       loggingPath,
		Config:            cfg,
	}

	return deployFramework, nil
}

// Setup handles the process of deploying metering, and waiting for all necessary resources
// to become ready in order to proceed with running the reporting tests.
func (df *DeployFramework) Setup(cfg deploy.Config, testOutputPath string, targetPods int) (*ReportingFrameworkConfig, error) {
	var err error

	cfg.OperatorResources, err = deploy.ReadMeteringAnsibleOperatorManifests(df.ManifestsDir, cfg.Platform)
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize objects from manifests: %v", err)
	}

	// randomize namespace and update namespace fields
	rand.Seed(time.Now().UnixNano())
	namespace := df.NamespacePrefix + "-" + strconv.Itoa(rand.Intn(50))

	// update the ommitted namespace fields in the test table index
	df.Config.Namespace = namespace
	cfg.Namespace = namespace
	cfg.MeteringConfig.ObjectMeta = meta.ObjectMeta{
		Name:      "operator-metering",
		Namespace: namespace,
	}

	df.Logger.Debugf("Deployer config: %+v", cfg)

	df.Deployer, err = deploy.NewDeployer(cfg, df.Logger, df.Client, df.APIExtClient, df.MeteringClient)
	if err != nil {
		return nil, fmt.Errorf("Failed to construct a new deployer object: %v", err)
	}

	err = df.Deployer.Install()
	if err != nil {
		return nil, fmt.Errorf("Failed to install metering: %v", err)
	}

	err = df.addE2ENamespaceLabel(namespace)
	if err != nil {
		return nil, fmt.Errorf("Failed to add the testing label to the %s namespace", namespace)
	}

	_, err = df.WaitForMeteringPods(targetPods, cfg.Namespace)
	if err != nil {
		df.Teardown(testOutputPath)
		return nil, fmt.Errorf("Error waiting for metering pods to become ready: %v", err)
	}

	routeBearerToken, err := df.GetRouteBearerToken(cfg.Namespace)
	if err != nil {
		return nil, fmt.Errorf("Failed to get the route bearer token: %v", err)
	}

	reportResultsPath := filepath.Join(testOutputPath, reportResultsDir)
	err = os.Mkdir(reportResultsPath, 0777)
	if err != nil {
		df.Logger.Fatalf("Failed to make the directory %s: %v", reportResultsPath, err)
	}

	df.Logger.Infof("Report results directory: %s", reportResultsPath)

	reportingFrameworkConfig := &ReportingFrameworkConfig{
		Namespace:                   cfg.Namespace,
		KubeConfigPath:              df.Config.KubeConfigPath,
		UseKubeProxyForReportingAPI: df.Config.UseKubeProxyForReportingAPI,
		UseRouteForReportingAPI:     df.Config.UseRouteForReportingAPI,
		RouteBearerToken:            routeBearerToken,
		ReportResultsOutputPath:     reportResultsPath,
	}

	return reportingFrameworkConfig, nil
}

// Teardown is a method that dumps the container and resource logs before uninstalling
// the metering resource provisioned by the df.Deployer instance
func (df *DeployFramework) Teardown(path string) error {
	cmdEnvVarArr := []string{
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
			return fmt.Errorf("Failed to create the directory %s: %v", dirPath, err)
		}

		cmdEnvVarArr = append(cmdEnvVarArr, env+"="+dirPath)
	}

	logger := df.Logger.WithFields(logrus.Fields{"component": "cleanup"})

	cleanupCmd := exec.Command(df.CleanupScriptPath)
	stdout, err := cleanupCmd.StdoutPipe()
	if err != nil {
		logger.Warnf("Failed to create pipe from command output to Stdout: %v", err)
	}

	scanner := bufio.NewScanner(stdout)
	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			logger.Infof(line)
		}
		if err := scanner.Err(); err != nil {
			logger.Errorf("error reading output from command: %v", err)
		}
		return
	}()

	cleanupCmd.Env = append(os.Environ(), cmdEnvVarArr...)
	err = cleanupCmd.Run()
	if err != nil {
		return fmt.Errorf("Failed to run the cleanup script: %v", err)
	}

	return df.Deployer.Uninstall()
}

type podStat struct {
	PodName string
	Ready   int
	Total   int
}

// WaitForMeteringPods periodically polls the list of pods in the @namespace
// and ensures the metering pods created are considered ready. In order to exit
// the polling loop, the number of pods listed must match the expected number
// of @targetPods, and all pod containers listed must report a ready status.
func (df *DeployFramework) WaitForMeteringPods(targetPods int, namespace string) (bool, error) {
	var readyPods []string
	var unreadyPods []podStat

	start := time.Now()

	df.Logger.Infof("Waiting for all metering pods to be ready")
	err := wait.Poll(10*time.Second, 20*time.Minute, func() (done bool, err error) {
		unreadyPods = nil
		readyPods = nil

		pods, err := df.Client.CoreV1().Pods(namespace).List(meta.ListOptions{})
		if err != nil {
			return false, err
		}

		if len(pods.Items) == 0 {
			return false, fmt.Errorf("The number of pods in the %s namespace should exceed zero", namespace)
		}

		for _, pod := range pods.Items {
			podIsReady, readyContainers := df.checkPodStatus(pod)
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

		df.logPollingSummary(targetPods, readyPods, unreadyPods)

		return ((len(pods.Items) == targetPods) && len(unreadyPods) == 0), nil
	})
	if err != nil {
		return false, fmt.Errorf("The metering pods failed to report a ready status before the timeout period occurred: %v", err)
	}

	df.Logger.Infof("Installing metering took %v", time.Since(start))

	return true, nil
}

// GetRouteBearerToken queries the @namespace for the reporting-operator serviceaccount and attempts
// to find the secret that contains the reporting-operator token. If that secret exists, return the
// string representation of the token (key) byte slice (value), or return an error.
func (df *DeployFramework) GetRouteBearerToken(namespace string) (string, error) {
	var secretName string
	var err error
	var sa *v1.ServiceAccount

	reportingOperatorName := "reporting-operator"

	df.Logger.Infof("Waiting for the reporting-operator service account to be created")
	err = wait.Poll(5*time.Second, 5*time.Minute, func() (done bool, err error) {
		sa, err = df.Client.CoreV1().ServiceAccounts(namespace).Get(reportingOperatorName, meta.GetOptions{})
		if err != nil {
			return false, nil
		}

		df.Logger.Infof("The reporting-operator service account has been created")
		return true, nil
	})
	if err != nil {
		return "", fmt.Errorf("Failed to get the reporting-operator service account before timeout has occurred: %v", err)
	}

	if len(sa.Secrets) == 0 {
		return "", fmt.Errorf("Failed to return a list of secrets in the reporting-operator service account")
	}

	for _, secret := range sa.Secrets {
		if strings.Contains(secret.Name, "token") {
			secretName = secret.Name
		}
	}

	if secretName == "" {
		return "", fmt.Errorf("Failed to get the secret token for the reporting-operator serviceaccount")
	}

	secret, err := df.Client.CoreV1().Secrets(namespace).Get(secretName, meta.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("Failed to get the secret containing the reporting-operator service account token: %v", err)
	}

	return string(secret.Data["token"]), nil
}
