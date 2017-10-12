package main

import (
	"io/ioutil"
	"os"
	"strconv"

	log "github.com/sirupsen/logrus"

	"github.com/coreos-inc/kube-chargeback/pkg/chargeback"
)

var (
	HiveHost   = "hive:10000"
	PrestoHost = "presto:8080"

	logReport bool
)

func init() {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{ForceColors: true})
}

func main() {
	logger := log.WithFields(log.Fields{
		"app": "chargeback-operator",
	})
	if logReportEnv := os.Getenv("LOG_REPORT"); logReportEnv != "" {
		var err error
		logReport, err = strconv.ParseBool(logReportEnv)
		if err != nil {
			logger.WithError(err).Fatalf("LOG_REPORT environment variable was not a bool, got %v", logReportEnv)
		}
	}
	namespace, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		logger.WithError(err).Fatal("could not determine namespace")
	}
	cfg := chargeback.Config{
		Namespace:  string(namespace),
		HiveHost:   HiveHost,
		PrestoHost: PrestoHost,
		LogReport:  logReport,
	}

	op, err := chargeback.New(logger, cfg)
	if err != nil {
		logger.WithError(err).Fatal("unable to setup Chargeback operator")
	}

	stopCh := make(<-chan struct{})
	if err = op.Run(stopCh); err != nil {
		logger.WithError(err).Fatalf("error occurred while the Chargeback operator was running")
	}
}
