package deployframework

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	ansibleRunnerPath                   = "/tmp/ansible-operator/runner"
	meteringOperatorName                = "metering-ansible-operator"
	meteringOperatorLogName             = "metering-operator.log"
	reportingOperatorLogName            = "reporting-operator.log"
	destKubeConfigPath                  = "/kubeconfig"
	runReportingOperatorLocalScriptName = "run-reporting-operator-local.sh"
	cleanupScriptName                   = "run-test-cleanup.sh"

	enableDebug        = true
	disableOCPFeatures = false
)

// LocalCtx holds all the necessary information to run e2e tests locally
type LocalCtx struct {
	Namespace                      string
	BasePath                       string
	KubeConfigPath                 string
	HackScriptPath                 string
	MeteringOperatorImage          string
	ReportingAPIURL                string
	RunReportingOperatorScriptPath string
	ExtraReportingOperatorEnvVars  []string
	CmdArr                         []exec.Cmd
	Logger                         logrus.FieldLogger
}

// RunMeteringOperatorLocal is a method that runs the metering-operator locally
func (lc *LocalCtx) RunMeteringOperatorLocal() error {
	cmd := exec.Command(
		"docker", "run",
		"--rm",
		"-u", "0:0",
		"--name", meteringOperatorContainerName,
		"-v", ansibleRunnerPath,
		"-v", fmt.Sprintf("%s:%s", lc.KubeConfigPath, destKubeConfigPath),
		"-e", "KUBECONFIG="+destKubeConfigPath,
		"-e", "OPERATOR_NAME="+meteringOperatorName,
		"-e", "POD_NAME="+meteringOperatorName,
		"-e", "WATCH_NAMESPACE="+lc.Namespace,
		"-e", "ENABLE_DEBUG="+strconv.FormatBool(enableDebug),
		"-e", "DISABLE_OCP_FEATURES="+strconv.FormatBool(disableOCPFeatures),
		lc.MeteringOperatorImage,
	)

	lc.Logger.Debugf("The metering-operator container was run with the following args: %v", cmd.Args)

	logFile, err := os.Create(filepath.Join(lc.BasePath, meteringOperatorLogName))
	if err != nil {
		return fmt.Errorf("failed to create the metering-operator container output log file: %v", err)
	}
	defer logFile.Close()

	lc.Logger.Infof("Storing the metering-operator container logs to the '%s' path", logFile)

	cmd.Stdout = logFile
	cmd.Stderr = logFile

	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to run the metering-operator container: %v", err)
	}

	lc.CmdArr = append(lc.CmdArr, *cmd)

	return nil
}

// RunReportingOperatorLocal is a method that runs the reporting-operator locally
func (lc *LocalCtx) RunReportingOperatorLocal(apiListenPort, metricsListenPort, pprofListenPort int, token string) error {
	var err error

	envVarArr := []string{
		"METERING_NAMESPACE=" + lc.Namespace,
		"METERING_USE_SERVICE_ACCOUNT_AS_PROM_TOKEN=false",
		"REPORTING_OPERATOR_PROMETHEUS_BEARER_TOKEN=" + token,
		"REPORTING_OPERATOR_API_LISTEN=" + fmt.Sprintf("%s:%d", localAddr, apiListenPort),
		"REPORTING_OPERATOR_METRICS_LISTEN=" + fmt.Sprintf("%s:%d", localAddr, metricsListenPort),
		"REPORTING_OPERATOR_PPROF_LISTEN=" + fmt.Sprintf("%s:%d", localAddr, pprofListenPort),
	}

	envVarArr = append(envVarArr, lc.ExtraReportingOperatorEnvVars...)

	relPath := filepath.Join(lc.HackScriptPath, runReportingOperatorLocalScriptName)
	targetScriptDir, err := filepath.Abs(relPath)
	if err != nil {
		return fmt.Errorf("failed to get the absolute path for the '%s' path: %v", relPath, err)
	}

	_, err = os.Stat(targetScriptDir)
	if err != nil {
		return fmt.Errorf("failed to stat the '%s' path: %v", targetScriptDir, err)
	}

	cmd := exec.Command(targetScriptDir)
	cmd.Env = append(os.Environ(), envVarArr...)

	lc.Logger.Debugf("The reporting-operator-local script was run with the following args: %v", cmd.Args)

	logFile, err := os.Create(filepath.Join(lc.BasePath, reportingOperatorLogName))
	if err != nil {
		return fmt.Errorf("failed to create the local reporting-operator log file: %v", err)
	}
	defer logFile.Close()

	lc.Logger.Infof("Storing the hack/run-reporting-operator-local.sh logs to the '%s' path", logFile.Name())

	cmd.Stdout = logFile
	cmd.Stderr = logFile

	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to run the reporting-operator locally: %v", err)
	}

	lc.CmdArr = append(lc.CmdArr, *cmd)

	return nil
}

// CleanupLocal removes all resources that were created while running e2e locally
func (lc *LocalCtx) CleanupLocal() error {
	var errArr []string

	if len(lc.CmdArr) != 0 {
		err := cleanupLocalCmds(lc.Logger, lc.CmdArr...)
		if err != nil {
			errArr = append(errArr, fmt.Sprintf("failed to stop the local commands: %v", err))
		}
	}

	if len(errArr) != 0 {
		return fmt.Errorf(strings.Join(errArr, "\n"))
	}

	return nil
}
