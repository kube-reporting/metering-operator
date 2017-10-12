package main

import (
	"flag"
	"io/ioutil"
	"net/url"
	"path"
	"time"

	cb "github.com/coreos-inc/kube-chargeback/pkg/chargeback/v1"
	cbClientSet "github.com/coreos-inc/kube-chargeback/pkg/generated/clientset/versioned"
	"github.com/coreos-inc/kube-chargeback/pkg/promsum"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"

	"github.com/prometheus/client_golang/api"
	log "github.com/sirupsen/logrus"
)

var (
	before        time.Duration
	timePrecision time.Duration
	aggregate     bool
	promURL       string
)

func init() {
	flag.DurationVar(&before, "before", 1*time.Hour, "duration before present to start collect billing data")
	flag.DurationVar(&timePrecision, "precision", time.Second, "unit of time used for stored amounts")
	flag.StringVar(&promURL, "prom", "http://localhost:9090", "URL of the Prometheus to be queried")

	flag.Parse()

	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{ForceColors: true})
}

func main() {
	if flag.NArg() != 0 {
		log.Fatal("no arguments are expected")
	}

	log.SetLevel(log.DebugLevel)

	// Get our namespace, make a new rest client, and list all the data stores
	namespaceBytes, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		log.Fatal("could not determine namespace: ", err)
	}
	namespace := string(namespaceBytes)

	cfg, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal(err)
	}

	clientSet, err := cbClientSet.NewForConfig(cfg)
	if err != nil {
		log.Fatal(err)
	}

	dataStores, err := clientSet.ChargebackV1alpha1().ReportDataStores(namespace).List(metav1.ListOptions{})
	if err != nil {
		log.Fatal("could not list data stores: ", err)
	}

	// TODO: we should track what times we query, store the last query point in
	// a CRD, and query from there instead of a blind value like how this
	// currently works.

	// Calculate the range of time to query, and create a new prometheus client
	now := time.Now().UTC()
	billingRng := cb.Range{
		Start: now.Add(-before),
		End:   now,
	}

	promCfg := api.Config{
		Address: promURL,
	}
	prom, err := NewPrometheus(promCfg)
	if err != nil {
		log.Fatal("could not setup remote: ", err)
	}

	// For every data store, look up and perform each query, storing in the
	// designated location.
	for _, ds := range dataStores.Items {
		if ds.Spec.Promsum == nil {
			log.Debugf("datastore %q: skipping, not promsum datastore", ds.Name)
			continue
		}
		storage := ds.Spec.Promsum.Storage
		if storage == nil {
			log.Errorf("datastore %q: improperly configured datastore, storage is empty", ds.Name)
			continue
		}
		if storage.S3 == nil {
			log.Errorf("datastore %q: unsupported storage type (must be s3)", ds.Name)
			continue
		}
		storeURL := url.URL{
			Scheme: "s3",
			Path:   path.Join(storage.S3.Bucket, storage.S3.Prefix),
		}

		log.Infof("storeURL: %s", storeURL.String())
		store, err := setupStore(storeURL)
		if err != nil {
			log.Fatal("Could not setup storage: ", err)
		}

		for _, queryName := range ds.Spec.Promsum.Queries {
			query, err := clientSet.ChargebackV1alpha1().ReportPrometheusQueries(namespace).Get(queryName, metav1.GetOptions{})
			if err != nil {
				log.Fatal("Could not get prometheus query: ", err)
			}

			records, err := promsum.Meter(prom, query.Spec.Query, queryName, billingRng, timePrecision)
			if err != nil {
				log.Fatalf("Failed to generate billing report for query '%s' in the range %v to %v: %v",
					query.Name, billingRng.Start, billingRng.End, err)
			}

			err = store.Write(records)
			if err != nil {
				log.Print("Failed to record: ", err)
			}
		}
	}
}
