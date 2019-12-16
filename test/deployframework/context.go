package deployframework

import (
	"fmt"
	"os"
	"path/***REMOVED***lepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	apiextclientv1beta1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	metering "github.com/operator-framework/operator-metering/pkg/apis/metering/v1"
	"github.com/operator-framework/operator-metering/pkg/deploy"
	meteringclient "github.com/operator-framework/operator-metering/pkg/generated/clientset/versioned/typed/metering/v1"
	"github.com/operator-framework/operator-metering/test/reportingframework"
)

var (
	httpScheme = "http"
	localAddr  = "127.0.0.1"

	apiPort     = 8100
	metricsPort = 8101
	pprofPort   = 8102

	healthCheckEndpoint      = "healthy"
	healthCheckTimeoutPeriod = 5 * time.Minute

	waitingForPodsTimeoutPeriod        = 20 * time.Minute
	waitingForServiceAccountTimePeriod = 10 * time.Minute

	useHTTPSAPI                 = true
	useRouteForReportingAPI     = true
	useKubeProxyForReportingAPI = false
	reportingAPIURL             string
)

// DeployerCtx contains all the information needed to manage the
// full lifecycle of a single metering deployment
type DeployerCtx struct {
	TargetPodsCount           int
	Namespace                 string
	KubeCon***REMOVED***gPath            string
	TestCaseOutputPath        string
	HackScriptPath            string
	MeteringOperatorImageRepo string
	MeteringOperatorImageTag  string
	RunTestLocal              bool
	RunDevSetup               bool
	ExtraLocalEnvVars         []string
	LocalCtx                  *LocalCtx
	Deployer                  *deploy.Deployer
	Logger                    logrus.FieldLogger
	Con***REMOVED***g                    *rest.Con***REMOVED***g
	Client                    kubernetes.Interface
	APIExtClient              apiextclientv1beta1.CustomResourceDe***REMOVED***nitionsGetter
	MeteringClient            meteringclient.MeteringV1Interface
}

// NewDeployerCtx constructs and returns a new DeployerCtx object
func (df *DeployFramework) NewDeployerCtx(
	namespace,
	meteringOperatorImageRepo,
	meteringOperatorImageTag,
	reportingOperatorImageRepo,
	reportingOperatorImageTag,
	outputPath string,
	extraLocalEnvVars []string,
	spec metering.MeteringCon***REMOVED***gSpec,
) (*DeployerCtx, error) {
	cfg, err := df.NewDeployerCon***REMOVED***g(namespace, meteringOperatorImageRepo, meteringOperatorImageTag, reportingOperatorImageRepo, reportingOperatorImageTag, spec)
	if err != nil {
		return nil, err
	}

	hackScriptDir, err := ***REMOVED***lepath.Abs(***REMOVED***lepath.Join(df.RepoDir, hackScriptDirName))
	if err != nil {
		return nil, fmt.Errorf("failed to get the absolute path to the hack script directory: %v", err)
	}
	_, err = os.Stat(hackScriptDir)
	if err != nil {
		return nil, fmt.Errorf("failed to successfully stat the %s path to the hack script directory: %v", hackScriptDir, err)
	}

	targetPodCount := defaultTargetPods

	if df.RunLocal {
		var replicas int32

		if cfg.MeteringCon***REMOVED***g.Spec.ReportingOperator != nil && cfg.MeteringCon***REMOVED***g.Spec.ReportingOperator.Spec != nil {
			cfg.MeteringCon***REMOVED***g.Spec.ReportingOperator.Spec.Replicas = &replicas
		}
		targetPodCount -= 2
	}

	df.Logger.Debugf("Deployer con***REMOVED***g: %+v", cfg)

	deployer, err := deploy.NewDeployer(*cfg, df.Logger, df.Client, df.APIExtClient, df.MeteringClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create a new deployer instance: %v", err)
	}

	return &DeployerCtx{
		Namespace:                 namespace,
		TargetPodsCount:           targetPodCount,
		TestCaseOutputPath:        outputPath,
		MeteringOperatorImageRepo: meteringOperatorImageRepo,
		MeteringOperatorImageTag:  meteringOperatorImageTag,
		Deployer:                  deployer,
		HackScriptPath:            hackScriptDir,
		ExtraLocalEnvVars:         extraLocalEnvVars,
		RunTestLocal:              df.RunLocal,
		RunDevSetup:               df.RunDevSetup,
		KubeCon***REMOVED***gPath:            df.KubeCon***REMOVED***gPath,
		Logger:                    df.Logger,
		Con***REMOVED***g:                    df.Con***REMOVED***g,
		Client:                    df.Client,
		APIExtClient:              df.APIExtClient,
		MeteringClient:            df.MeteringClient,
	}, nil
}

// NewLocalCtx returns a new LocalCtx object
func (ctx *DeployerCtx) NewLocalCtx() *LocalCtx {
	meteringOperatorImage := fmt.Sprintf("%s:%s", ctx.MeteringOperatorImageRepo, ctx.MeteringOperatorImageTag)

	return &LocalCtx{
		Namespace:                     ctx.Namespace,
		KubeCon***REMOVED***gPath:                ctx.KubeCon***REMOVED***gPath,
		BasePath:                      ctx.TestCaseOutputPath,
		HackScriptPath:                ctx.HackScriptPath,
		ExtraReportingOperatorEnvVars: ctx.ExtraLocalEnvVars,
		Logger:                        ctx.Logger,
		MeteringOperatorImage:         meteringOperatorImage,
	}
}

// Setup handles the process of deploying metering, and waiting for all the necessary
// resources to become ready in order to proceeed with running the reporting tests.
// This returns an initialized reportingframework object, or an error if there is any.
func (ctx *DeployerCtx) Setup(expectInstallErr bool) (*reportingframework.ReportingFramework, error) {
	var (
		installErrMsg    string
		routeBearerToken string
		installErr       bool
	)

	// If we expect an install error, and there was an install error, then delay returning
	// that error message until after the reportingframework has been constructed.
	err := ctx.Deployer.Install()
	if err != nil {
		installErr = true
		installErrMsg = fmt.Sprintf("failed to install metering: %v", err)

		if !expectInstallErr {
			return nil, fmt.Errorf(installErrMsg)
		}
	}

	if !installErr {
		if ctx.RunTestLocal {
			ctx.LocalCtx = ctx.NewLocalCtx()

			err = ctx.LocalCtx.RunMeteringOperatorLocal()
			if err != nil {
				return nil, fmt.Errorf("failed to run the metering-operator docker container: %v", err)
			}
		}

		ctx.Logger.Infof("Waiting for the metering pods to be ready")
		start := time.Now()

		initialDelay := 10 * time.Second

		if ctx.RunDevSetup {
			initialDelay = 1 * time.Second
		}

		pw := &PodWaiter{
			InitialDelay:  initialDelay,
			TimeoutPeriod: waitingForPodsTimeoutPeriod,
			Client:        ctx.Client,
			Logger:        ctx.Logger.WithField("component", "podWaiter"),
		}
		err = pw.WaitForPods(ctx.Namespace, ctx.TargetPodsCount)
		if err != nil {
			return nil, fmt.Errorf("error waiting for metering pods to become ready: %v", err)
		}

		ctx.Logger.Infof("Installing metering took %v", time.Since(start))

		ctx.Logger.Infof("Getting the service account %s", reportingOperatorServiceAccountName)

		routeBearerToken, err = GetServiceAccountToken(
			ctx.Client,
			initialDelay,
			waitingForServiceAccountTimePeriod,
			ctx.Namespace,
			reportingOperatorServiceAccountName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to get the route bearer token: %v", err)
		}

		if ctx.RunTestLocal {
			useHTTPSAPI = false
			useRouteForReportingAPI = false
			useKubeProxyForReportingAPI = false
			reportingAPIURL = fmt.Sprintf("%s://%s:%d", httpScheme, localAddr, apiPort)

			err = ctx.LocalCtx.RunReportingOperatorLocal(apiPort, metricsPort, pprofPort, routeBearerToken)
			if err != nil {
				return nil, fmt.Errorf("failed to run the reporting-operator locally: %v", err)
			}

			reportingAPIHealthCheckURL := fmt.Sprintf("%s/%s", reportingAPIURL, healthCheckEndpoint)

			err := waitForURLToReportStatusOK(ctx.Logger, reportingAPIHealthCheckURL, healthCheckTimeoutPeriod)
			if err != nil {
				return nil, fmt.Errorf("failed to wait for the reporting-operator to become healthy: %v", err)
			}
		}
	}

	reportResultsPath := ***REMOVED***lepath.Join(ctx.TestCaseOutputPath, reportResultsDir)
	err = os.Mkdir(reportResultsPath, 0777)
	if err != nil {
		return nil, fmt.Errorf("failed to create the report results directory %s: %v", reportResultsPath, err)
	}

	ctx.Logger.Infof("Report results directory: %s", reportResultsPath)

	rf, err := reportingframework.New(
		useHTTPSAPI,
		useKubeProxyForReportingAPI,
		useRouteForReportingAPI,
		ctx.Namespace,
		routeBearerToken,
		reportingAPIURL,
		reportResultsPath,
		ctx.Con***REMOVED***g,
		ctx.Client,
		ctx.MeteringClient,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to construct a reportingframework: %v", err)
	}

	if installErrMsg != "" {
		return rf, fmt.Errorf(installErrMsg)
	}

	return rf, nil
}

// Teardown is a method that creates the resource and container logging
// directories, then populates those directories by executing the
// cleanup bash script, while streaming the script output
// to stdout. Once the cleanup script has ***REMOVED***nished execution, we can
// uninstall the metering stack and return an error if there is any.
func (ctx *DeployerCtx) Teardown() error {
	var errArr []string

	logger := ctx.Logger.WithFields(logrus.Fields{"component": "cleanup"})

	if ctx.RunTestLocal && ctx.LocalCtx != nil {
		err := ctx.LocalCtx.CleanupLocal()
		if err != nil {
			errArr = append(errArr, fmt.Sprintf("failed to successfully cleanup local resources: %v", err))
		}
	}

	// Check if the user wants to run the E2E suite using the developer setup.
	// If true, we skip the process of deleting and logging of the metering
	// resources that were provisioned during the manual install.
	if !ctx.RunDevSetup {
		relPath := ***REMOVED***lepath.Join(ctx.HackScriptPath, cleanupScriptName)
		targetScriptDir, err := ***REMOVED***lepath.Abs(relPath)
		if err != nil {
			errArr = append(errArr, fmt.Sprintf("failed to get the absolute path from '%s': %v", relPath, err))
		}

		_, err = os.Stat(targetScriptDir)
		if err != nil {
			errArr = append(errArr, fmt.Sprintf("failed to stat the '%s' path: %v", targetScriptDir, err))
		}

		err = runCleanupScript(logger, ctx.Namespace, ctx.TestCaseOutputPath, targetScriptDir)
		if err != nil {
			errArr = append(errArr, fmt.Sprintf("failed to successfully run the cleanup script: %v", err))
		}

		err = ctx.Deployer.Uninstall()
		if err != nil {
			errArr = append(errArr, fmt.Sprintf("failed to uninstall metering: %v", err))
		}
	}

	if len(errArr) != 0 {
		return fmt.Errorf(strings.Join(errArr, "\n"))
	}

	return nil
}
