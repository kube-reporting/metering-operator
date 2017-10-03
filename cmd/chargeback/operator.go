package main

import (
	"github.com/coreos-inc/kube-chargeback/pkg/chargeback"
	"io/ioutil"
)

var (
	HiveHost   = "hive:10000"
	PrestoHost = "presto:8080"
)

func main() {
	namespace, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		panic(err)
	}
	cfg := chargeback.Con***REMOVED***g{
		Namespace:  string(namespace),
		HiveHost:   HiveHost,
		PrestoHost: PrestoHost,
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
