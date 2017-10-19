package main

import (
	"io/ioutil"
	"os"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/coreos-inc/kube-chargeback/pkg/chargeback"
)

var (
	HiveHost   = "hive:10000"
	PrestoHost = "presto:8080"

	defaultPromHost = "http://prometheus.tectonic-system.svc.cluster.local:9090"

	logReport  bool
	logQueries bool

	promsumInterval  = time.Minute * 5
	promsumPrecision = time.Minute
	disablePromsum   = false
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
	if logQueriesStr := os.Getenv("LOG_QUERIES"); logQueriesStr != "" {
		var err error
		logQueries, err = strconv.ParseBool(logQueriesStr)
		if err != nil {
			logger.WithError(err).Fatalf("LOG_REPORT environment variable was not a bool, got %v", logQueriesStr)
		}
	}
	if promsumIntervalStr := os.Getenv("PROMSUM_INTERVAL"); promsumIntervalStr != "" {
		var err error
		promsumInterval, err = time.ParseDuration(promsumIntervalStr)
		if err != nil {
			logger.WithError(err).Fatalf("PROMSUM_INTERVAL environment variable was not a duration, got %q", promsumIntervalStr)
		}
	}
	if promsumPrecisionStr := os.Getenv("PROMSUM_PRECISION"); promsumPrecisionStr != "" {
		var err error
		promsumPrecision, err = time.ParseDuration(promsumPrecisionStr)
		if err != nil {
			logger.WithError(err).Fatalf("PROMSUM_PRECISION environment variable was not a duration, got %q", promsumPrecisionStr)
		}
	}
	promHost := os.Getenv("PROMETHEUS_HOST")
	if promHost == "" {
		promHost = defaultPromHost
	}
	if disablePromsumStr := os.Getenv("DISABLE_PROMSUM"); disablePromsumStr != "" {
		var err error
		disablePromsum, err = strconv.ParseBool(disablePromsumStr)
		if err != nil {
			logger.WithError(err).Fatalf("DISABLE_PROMSUM environment variable was not a bool, got %q", disablePromsumStr)
		}
	}
	namespace, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		logger.WithError(err).Fatal("could not determine namespace")
	}
	cfg := chargeback.Con***REMOVED***g{
		Namespace:        string(namespace),
		HiveHost:         HiveHost,
		PrestoHost:       PrestoHost,
		PromHost:         promHost,
		DisablePromsum:   disablePromsum,
		LogReport:        logReport,
		LogQueries:       logQueries,
		PromsumInterval:  promsumInterval,
		PromsumPrecision: promsumPrecision,
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
