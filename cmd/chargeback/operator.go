package main

import (
	"github.com/coreos-inc/kube-chargeback/pkg/operator"
)

var (
	HiveHost   = "hive:10000"
	PrestoHost = "presto:8080"
)

func main() {
	cfg := operator.Con***REMOVED***g{
		HiveHost:   HiveHost,
		PrestoHost: PrestoHost,
	}

	_, err := operator.New(cfg)
	if err != nil {
		panic(err)
	}
}
