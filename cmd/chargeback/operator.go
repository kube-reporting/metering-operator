package main

import (
	"github.com/coreos-inc/kube-chargeback/pkg/operator"
)

var (
	HiveHost   = "hive:10000"
	PrestoHost = "presto:8080"
)

func main() {
	cfg := operator.Config{
		HiveHost:   HiveHost,
		PrestoHost: PrestoHost,
	}

	op, err := operator.New(cfg)
	if err != nil {
		panic(err)
	}

	if err = op.Run(); err != nil {
		panic(err)
	}
}
