package main

import (
	"io/ioutil"
	"os"
	"strconv"

	"github.com/coreos-inc/kube-chargeback/pkg/chargeback"
)

var (
	HiveHost   = "hive:10000"
	PrestoHost = "presto:8080"

	logReport bool
)

func main() {
	if logReportEnv := os.Getenv("LOG_REPORT"); logReportEnv != "" {
		var err error
		logReport, err = strconv.ParseBool(logReportEnv)
		if err != nil {
			panic(err)
		}
	}
	namespace, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		panic(err)
	}
	cfg := chargeback.Config{
		Namespace:  string(namespace),
		HiveHost:   HiveHost,
		PrestoHost: PrestoHost,
		LogReport:  logReport,
	}

	op, err := chargeback.New(cfg)
	if err != nil {
		panic(err)
	}

	stopCh := make(<-chan struct{})
	if err = op.Run(stopCh); err != nil {
		panic(err)
	}
}
