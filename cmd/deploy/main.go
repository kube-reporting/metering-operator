package main

import (
	log "github.com/sirupsen/logrus"
	"os"

	meteringclientv1 "github.com/operator-framework/operator-metering/pkg/generated/clientset/versioned/typed/metering/v1"
	"github.com/operator-framework/operator-metering/pkg/operator/deploy"
	apiextclientv1beta1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"

	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	logger := setupLogger("info")
	con***REMOVED***g := deploy.Con***REMOVED***g{}

	kubecon***REMOVED***g := clientcmd.NewNonInteractiveDeferredLoadingClientCon***REMOVED***g(
		clientcmd.NewDefaultClientCon***REMOVED***gLoadingRules(),
		&clientcmd.Con***REMOVED***gOverrides{},
	)

	restcon***REMOVED***g, err := kubecon***REMOVED***g.ClientCon***REMOVED***g()
	if err != nil {
		logger.Fatalf("Failed to initialize the kubernetes client con***REMOVED***g: %v", err)
	}

	client, err := kubernetes.NewForCon***REMOVED***g(restcon***REMOVED***g)
	if err != nil {
		logger.Fatalf("Failed to initialize the kubernetes clientset: %v", err)
	}

	apiextClient, err := apiextclientv1beta1.NewForCon***REMOVED***g(restcon***REMOVED***g)
	if err != nil {
		logger.Fatalf("Failed to initialize the apiextensions clientset: %v", err)
	}

	meteringClient, err := meteringclientv1.NewForCon***REMOVED***g(restcon***REMOVED***g)
	if err != nil {
		logger.Fatalf("Failed to initialize the metering clientset: %v", err)
	}

	deployObj, err := deploy.NewDeployer(con***REMOVED***g, client, apiextClient, meteringClient, logger)
	if err != nil {
		logger.Fatalf("Failed to deploy metering: %v", err)
	}

	deployType := os.Getenv("DEPLOY_TYPE")
	if deployType == "" {
		logger.Fatalf("error: you need to set the $DEPLOY_TYPE env var")
	}

	if deployType == "install" {
		err = deployObj.Install()
		if err != nil {
			logger.Fatalf("Failed to install metering resources: %v", err)
		}
		logger.Infof("Finished installing metering")
	} ***REMOVED*** if deployType == "uninstall" {
		err = deployObj.Uninstall()
		if err != nil {
			logger.Fatalf("Failed to uninstall metering resources: %v", err)
		}
		logger.Infof("Finished uninstalling metering")
	}

	logger.Infof("Finished deploying metering")
}

func setupLogger(logLevelStr string) log.FieldLogger {
	var err error

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "01-02-2006 15:04:05",
	})

	logger := log.WithFields(log.Fields{
		"app": "deploy",
	})
	logLevel, err := log.ParseLevel(logLevelStr)
	if err != nil {
		logger.WithError(err).Fatalf("invalid log level: %s", logLevel)
	}
	logger.Infof("Setting the log level to %s", logLevel.String())
	logger.Logger.Level = logLevel

	return logger
}
