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
	kubecon***REMOVED***g string
	target     string
	verbose    bool
	idOnly     bool
)

const (
	// TectonicNamespace is the namespace where all Tectonic related resources live.
	TectonicNamespace = "tectonic-system"

	// TectonicCon***REMOVED***gName is the name of the Con***REMOVED***gMap holding the Tectonic cluster con***REMOVED***guration.
	TectonicCon***REMOVED***gName = "tectonic-con***REMOVED***g"

	// ClusterIDKey is the key in the cluster Con***REMOVED***gMap's data holding the cluster's ID.
	ClusterIDKey = "clusterID"
)

func init() {
	flag.StringVar(&kubecon***REMOVED***g, "kubecon***REMOVED***g", os.Getenv("KUBECONFIG"), "absolute path to the kubecon***REMOVED***g ***REMOVED***le")
	flag.BoolVar(&verbose, "verbose", false, "Provides detailed log information")
	flag.BoolVar(&idOnly, "id", false, "Does not run collector. Just returns cluster ID")
	flag.Parse()
	target = flag.Arg(0)
}

func main() {
	// setup Kubernetes client
	client, err := newClientSet()
	if err != nil {
		log.Fatalf("Failed to con***REMOVED***gure client: %v", err)
	}

	if idOnly {
		cfg, err := getClusterCon***REMOVED***g(client)
		if err != nil {
			log.Fatalf("Failed to get con***REMOVED***guration for cluster: %v", err)
		}

		fmt.Println(cfg.Data[ClusterIDKey])
		os.Exit(0)
	}

	if len(target) == 0 {
		log.Fatal("A path to write collected data to must be given. Supported schemes: hdfs://, ***REMOVED***le://")
	}

	// write to either local ***REMOVED***lesystem or HDFS
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
		log.Fatalf("Failed to write usage data to ***REMOVED***le: %v", err)
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

// getClusterCon***REMOVED***g returns the Tectonic con***REMOVED***guration as reported by the cluster.
func getClusterCon***REMOVED***g(client *kubernetes.Clientset) (*v1.Con***REMOVED***gMap, error) {
	cfg, err := client.Con***REMOVED***gMaps(TectonicNamespace).Get(TectonicCon***REMOVED***gName, meta.GetOptions{})
	if err != nil {
		err = fmt.Errorf("Failed to cluster con***REMOVED***guration from API server: %v", err)
	}
	return cfg, err
}

// newClientSet returns a clientset con***REMOVED***gured using in-cluster discovery.
func newClientSet() (client *kubernetes.Clientset, err error) {
	var con***REMOVED***g *rest.Con***REMOVED***g
	if len(kubecon***REMOVED***g) != 0 {
		con***REMOVED***g, err = clientcmd.BuildCon***REMOVED***gFromFlags("", kubecon***REMOVED***g)
	} ***REMOVED*** {
		con***REMOVED***g, err = rest.InClusterCon***REMOVED***g()
	}

	if err == nil {
		client, err = kubernetes.NewForCon***REMOVED***g(con***REMOVED***g)
	}

	if err != nil {
		err = fmt.Errorf("failed to create clientset from cluster con***REMOVED***g: %v", err)
	}
	return client, err
}
