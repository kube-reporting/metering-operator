package main

import (
	"github.com/coreos-inc/kube-chargeback/pkg/chargeback"
)

var (
	HiveHost   = "hive:10000"
	PrestoHost = "presto:8080"
)

func main() {
	cfg := chargeback.Config{
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
