package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/gocarina/gocsv"
)

var (
	kubeconfig string
	target     string
	verbose    bool
	idOnly     bool
)

const (
	// TectonicNamespace is the namespace where all Tectonic related resources live.
	TectonicNamespace = "tectonic-system"

	// TectonicConfigName is the name of the ConfigMap holding the Tectonic cluster configuration.
	TectonicConfigName = "tectonic-config"

	// ClusterIDKey is the key in the cluster ConfigMap's data holding the cluster's ID.
	ClusterIDKey = "clusterID"
)

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", os.Getenv("KUBECONFIG"), "absolute path to the kubeconfig file")
	flag.BoolVar(&verbose, "verbose", false, "Provides detailed log information")
	flag.BoolVar(&idOnly, "id", false, "Does not run collector. Just returns cluster ID")
	flag.Parse()
	target = flag.Arg(0)
}

func main() {
	// setup Kubernetes client
	client, err := newClientSet()
	if err != nil {
		log.Fatalf("Failed to configure client: %v", err)
	}

	if idOnly {
		cfg, err := getClusterConfig(client)
		if err != nil {
			log.Fatalf("Failed to get configuration for cluster: %v", err)
		}

		fmt.Println(cfg.Data[ClusterIDKey])
		os.Exit(0)
	}

	if len(target) == 0 {
		log.Fatal("A path to write collected data to must be given. Supported schemes: hdfs://, file://")
	}

	// write to either local filesystem or HDFS
	out, err := setupOutput(target)
	if err != nil {
		log.Fatalf("Failed to setup output target: %v", err)
	}
	defer out.Close()

	// setup CSV writer
	w := csv.NewWriter(out)

	// usage data sent on channel to be written
	usageCh := make(chan interface{})
	go collectUsage(client, usageCh)

	if err = gocsv.MarshalChan(usageCh, w); err != nil {
		log.Fatalf("Failed to write usage data to file: %v", err)
	}
}

func collectUsage(client *kubernetes.Clientset, usageCh chan interface{}) {
	// get all Pods in the cluster
	pods, err := client.Pods("").List(meta.ListOptions{})
	if err != nil {
		log.Fatalf("Failed to get Pods from API server: %v", err)
	}

	for _, pod := range pods.Items {
		if pod.Status.Phase != v1.PodRunning {
			continue
		}
		log.Printf("Processing %s", pod.GetSelfLink())

		if verbose {
			log.Printf("Getting log data for node '%s'", pod.Spec.NodeName)
		}
		node, err := client.Nodes().Get(pod.Spec.NodeName, meta.GetOptions{})
		if err != nil {
			log.Fatalf("Failed to get node '%s': %v", pod.Spec.NodeName, err)
		}

		if verbose {
			log.Print("Creating pod usage...")
		}
		usage, err := NewPodUsage(&pod, node)
		if err != nil {
			log.Fatalf("Failed to create PodUsage: %v", err)
		}
		usageCh <- usage
	}
	log.Print("Done usage collection!")
	close(usageCh)
}

// getClusterConfig returns the Tectonic configuration as reported by the cluster.
func getClusterConfig(client *kubernetes.Clientset) (*v1.ConfigMap, error) {
	cfg, err := client.ConfigMaps(TectonicNamespace).Get(TectonicConfigName, meta.GetOptions{})
	if err != nil {
		err = fmt.Errorf("Failed to cluster configuration from API server: %v", err)
	}
	return cfg, err
}

// newClientSet returns a clientset configured using in-cluster discovery.
func newClientSet() (client *kubernetes.Clientset, err error) {
	var config *rest.Config
	if len(kubeconfig) != 0 {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	} else {
		config, err = rest.InClusterConfig()
	}

	if err == nil {
		client, err = kubernetes.NewForConfig(config)
	}

	if err != nil {
		err = fmt.Errorf("failed to create clientset from cluster config: %v", err)
	}
	return client, err
}
