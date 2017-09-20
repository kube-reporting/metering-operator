package main

import (
	"flag"
	"log"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/coreos-inc/kube-chargeback/pkg/chargeback"
)

var (
	Kubeconfig string
	HiveHost   string
	PrestoHost string
)

func init() {
	flag.StringVar(&Kubeconfig, "kubeconfig", "", "kubeconfig path for Kubernetetes API server (debug only)")
	flag.StringVar(&HiveHost, "hive", "hive:10000", "host used to connect to Hive")
	flag.StringVar(&PrestoHost, "presto", "presto:8080", "host used to connect to Presto")
	flag.Parse()
}

func main() {
	cfg := chargeback.Config{
		HiveHost:   HiveHost,
		PrestoHost: PrestoHost,
	}

	var err error
	if len(Kubeconfig) != 0 {
		cfg.ClientCfg, err = clientcmd.BuildConfigFromFlags("", Kubeconfig)
	} else {
		cfg.ClientCfg, err = rest.InClusterConfig()
	}

	if err != nil {
		log.Fatalf("could not configure Kubernetes client: %v", err)
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
