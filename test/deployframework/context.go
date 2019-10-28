package deployframework

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	apiextclientv1beta1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	metering "github.com/operator-framework/operator-metering/pkg/apis/metering/v1"
	meteringclient "github.com/operator-framework/operator-metering/pkg/generated/clientset/versioned/typed/metering/v1"
	"github.com/operator-framework/operator-metering/pkg/operator/deploy"
	"github.com/operator-framework/operator-metering/test/reportingframework"
)

// DeployerCtx contains all the information needed to manage the
// full lifecycle of a single metering deployment
type DeployerCtx struct {
	TargetPodsCount    int
	Namespace          string
	KubeConfigPath     string
	TestCaseOutputPath string
	Deployer           *deploy.Deployer
	Logger             logrus.FieldLogger
	Config             *rest.Config
	Client             kubernetes.Interface
	APIExtClient       apiextclientv1beta1.CustomResourceDefinitionsGetter
	MeteringClient     meteringclient.MeteringV1Interface
}

// NewDeployerCtx constructs and returns a new DeployerCtx object
func (df *DeployFramework) NewDeployerCtx(
	namespace,
	meteringOperatorImageRepo,
	meteringOperatorImageTag,
	reportingOperatorImageRepo,
	reportingOperatorImageTag string,
	spec metering.MeteringConfigSpec,
	outputPath string,
	targetPodsCount int,
) (*DeployerCtx, error) {
	cfg, err := df.NewDeployerConfig(namespace, meteringOperatorImageRepo, meteringOperatorImageTag, reportingOperatorImageRepo, reportingOperatorImageTag, spec)
	if err != nil {
		return nil, err
	}

	df.Logger.Debugf("Deployer config: %+v", cfg)

	deployer, err := deploy.NewDeployer(*cfg, df.Logger, df.Client, df.APIExtClient, df.MeteringClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create a new deployer instance: %v", err)
	}

	return &DeployerCtx{
		Namespace:          namespace,
		TargetPodsCount:    targetPodsCount,
		TestCaseOutputPath: outputPath,
		Deployer:           deployer,
		KubeConfigPath:     df.KubeConfigPath,
		Logger:             df.Logger,
		Config:             df.Config,
		Client:             df.Client,
		APIExtClient:       df.APIExtClient,
		MeteringClient:     df.MeteringClient,
	}, nil
}

// Setup handles the process of deploying metering, and waiting for all the necessary
// resources to become ready in order to proceeed with running the reporting tests.
// This returns an initialized reportingframework object, or an error if there is any.
func (ctx *DeployerCtx) Setup() (*reportingframework.ReportingFramework, error) {
	err := ctx.Deployer.Install()
	if err != nil {
		ctx.Logger.Infof("Failed to install metering: %v", err)
		return nil, fmt.Errorf("failed to install metering: %v", err)
	}

	ctx.Logger.Infof("Waiting for the metering pods to be ready")
	start := time.Now()

	pw := &PodWaiter{
		Client: ctx.Client,
		Logger: ctx.Logger.WithField("component", "podWaiter"),
	}
	err = pw.WaitForPods(ctx.Namespace, ctx.TargetPodsCount)
	if err != nil {
		return nil, fmt.Errorf("error waiting for metering pods to become ready: %v", err)
	}

	ctx.Logger.Infof("Installing metering took %v", time.Since(start))

	ctx.Logger.Infof("Getting the service account %s", reportingOperatorServiceAccountName)

	routeBearerToken, err := GetServiceAccountToken(ctx.Client, ctx.Namespace, reportingOperatorServiceAccountName)
	if err != nil {
		return nil, fmt.Errorf("failed to get the route bearer token: %v", err)
	}

	reportResultsPath := filepath.Join(ctx.TestCaseOutputPath, reportResultsDir)
	err = os.Mkdir(reportResultsPath, 0777)
	if err != nil {
		return nil, fmt.Errorf("failed to create the report results directory %s: %v", reportResultsPath, err)
	}

	ctx.Logger.Infof("Report results directory: %s", reportResultsPath)

	// TODO create functions that determine the reportingframework fields
	// we can hardcode these values for now as we only have one
	// meteringconfig configuration that gets installed and we dont
	// support passing a METERING_CR_FILE yaml file for local testing
	useHTTPSAPI := true
	useRouteForReportingAPI := true
	useKubeProxyForReportingAPI := false
	reportingAPIURL := ""

	rf, err := reportingframework.New(
		useHTTPSAPI,
		useKubeProxyForReportingAPI,
		useRouteForReportingAPI,
		ctx.Namespace,
		routeBearerToken,
		reportingAPIURL,
		reportResultsPath,
		ctx.Config,
		ctx.Client,
		ctx.MeteringClient,
	)
	if err != nil {
		ctx.Logger.Infof("Failed to construct a reportingframework: %v", err)
		return nil, fmt.Errorf("failed to construct a reportingframework: %v", err)
	}

	return rf, nil
}

// Teardown is a method that creates the resource and container logging
// directories, then populates those directories by executing the
// @cleanupScript bash script, while streaming the script output
// to stdout. Once the cleanup script has finished execution, we can
// uninstall the metering stack and return an error if there is any.
func (ctx *DeployerCtx) Teardown(cleanupScript string) error {
	logger := ctx.Logger.WithFields(logrus.Fields{"component": "cleanup"})

	var errArr []string
	envVarArr, err := createResourceDirs(ctx.Namespace, ctx.TestCaseOutputPath)
	if err != nil {
		errArr = append(errArr, fmt.Sprintf("failed to create the resource output directories: %v", err))
	}

	cleanupCmd := exec.Command(cleanupScript)
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
			errArr = append(errArr, fmt.Sprintf("error reading output from command: %v", err))
		}
		return
	}()

	cleanupCmd.Env = append(os.Environ(), envVarArr...)
	err = cleanupCmd.Run()
	if err != nil {
		errArr = append(errArr, fmt.Sprintf("error running the cleanup script: %v", err))
	}

	err = ctx.Deployer.Uninstall()
	if err != nil {
		errArr = append(errArr, fmt.Sprintf("failed to uninstall metering: %v", err))
	}

	if len(errArr) != 0 {
		return fmt.Errorf(strings.Join(errArr, "\n"))
	}

	return nil
}
