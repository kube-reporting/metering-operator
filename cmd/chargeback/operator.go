package main

import (
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"k8s.io/apimachinery/pkg/util/clock"

	log "github.com/sirupsen/logrus"

	"github.com/coreos-inc/kube-chargeback/pkg/chargeback"
)

var (
	HiveHost   = "hive:10000"
	PrestoHost = "presto:8080"

	defaultPromHost = "http://prometheus.tectonic-system.svc.cluster.local:9090"

	// logReport logs out report results after creating a report result table
	logReport bool
	// logDMLQueries controls if we log data manipulation queries made via
	// Presto (SELECT, INSERT, etc)
	logDMLQueries bool
	// logDDLQueries controls if we log data de***REMOVED***nition language queries made
	// via Hive (CREATE TABLE, DROP TABLE, etc).
	logDDLQueries bool

	// promsumInterval is how often we poll prometheus
	promsumInterval = time.Minute * 5
	// promsumStepSize is the query step size for Promethus query. This
	// controls resolution of results.
	promsumStepSize = time.Minute
	// promsumChunkSize controls how much the range query window size
	// by limiting the range query to a range of time no longer than this
	// duration.
	promsumChunkSize = time.Minute * 5
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
	if logDMLQueriesStr := os.Getenv("LOG_DML_QUERIES"); logDMLQueriesStr != "" {
		var err error
		logDMLQueries, err = strconv.ParseBool(logDMLQueriesStr)
		if err != nil {
			logger.WithError(err).Fatalf("LOG_DML_REPORT environment variable was not a bool, got %v", logDMLQueriesStr)
		}
	}
	if logDDLQueriesStr := os.Getenv("LOG_DDL_QUERIES"); logDDLQueriesStr != "" {
		var err error
		logDDLQueries, err = strconv.ParseBool(logDDLQueriesStr)
		if err != nil {
			logger.WithError(err).Fatalf("LOG_DDL_REPORT environment variable was not a bool, got %v", logDDLQueriesStr)
		}
	}
	if promsumIntervalStr := os.Getenv("PROMSUM_INTERVAL"); promsumIntervalStr != "" {
		var err error
		promsumInterval, err = time.ParseDuration(promsumIntervalStr)
		if err != nil {
			logger.WithError(err).Fatalf("PROMSUM_INTERVAL environment variable was not a duration, got %q", promsumIntervalStr)
		}
	}
	if promsumStepSizeStr := os.Getenv("PROMSUM_STEP_SIZE"); promsumStepSizeStr != "" {
		var err error
		promsumStepSize, err = time.ParseDuration(promsumStepSizeStr)
		if err != nil {
			logger.WithError(err).Fatalf("PROMSUM_STEP_SIZE environment variable was not a duration, got %q", promsumStepSizeStr)
		}
	}
	if promsumChunkSizeStr := os.Getenv("PROMSUM_CHUNK_SIZE"); promsumChunkSizeStr != "" {
		var err error
		promsumChunkSize, err = time.ParseDuration(promsumChunkSizeStr)
		if err != nil {
			logger.WithError(err).Fatalf("PROMSUM_CHUNK_SIZE environment variable was not a duration, got %q", promsumChunkSize)
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
		LogDMLQueries:    logDMLQueries,
		LogDDLQueries:    logDDLQueries,
		PromsumInterval:  promsumInterval,
		PromsumStepSize:  promsumStepSize,
		PromsumChunkSize: promsumChunkSize,
	}

	clock := clock.RealClock{}
	op, err := chargeback.New(logger, cfg, clock)
	if err != nil {
		logger.WithError(err).Fatal("unable to setup Chargeback operator")
	}

	sigs := make(chan os.Signal, 1)
	stopCh := make(chan struct{})
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		logger.Infof("got signal %s, performing shutdown", sig)
		close(stopCh)
	}()

	if err = op.Run(stopCh); err != nil {
		logger.WithError(err).Fatal("error occurred while the Chargeback operator was running")
	}
}
