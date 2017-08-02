package main

import (
	"github.com/coreos-inc/kube-chargeback/pkg/chargeback"
)

var (
	HiveHost   = "hive:10000"
	PrestoHost = "presto:8080"
)

func main() {
	cfg := chargeback.Con***REMOVED***g{
		HiveHost:   HiveHost,
		PrestoHost: PrestoHost,
	}

	op, err := chargeback.New(cfg)
	if err != nil {
		panic(err)
	}

	if err = op.Run(); err != nil {
		panic(err)
	}
}
