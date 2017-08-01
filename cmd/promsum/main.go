package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	cb "github.com/coreos-inc/kube-chargeback/pkg/chargeback"

	"github.com/prometheus/client_golang/api"
)

var (
	before        time.Duration
	timePrecision time.Duration
	window        time.Duration
	rollup        time.Duration
	subject       string
	storeURL      string
	aggregate     bool
	promURL       string
)

func init() {
	flag.DurationVar(&before, "before", 1*time.Hour, "duration before present to start collect billing data")
	flag.DurationVar(&timePrecision, "precision", time.Second, "unit of time used for stored amounts")
	flag.DurationVar(&window, "window", 20*time.Second, "interval Prometheus should return query results")
	flag.DurationVar(&rollup, "rollup", 10*time.Minute, "interval at which billing records are bundled into a single record")
	flag.StringVar(&subject, "subject", fmt.Sprintf("%x", time.Now().Nanosecond()), "name used to group billing data")
	flag.StringVar(&storeURL, "path", "file://data", "URL to the location that records should be stored")
	flag.StringVar(&promURL, "prom", "http://localhost:9090", "URL of the Prometheus to be queried")
	flag.BoolVar(&aggregate, "aggregate", false, "summarizes the ranged in as few billing records as possible")

	flag.Parse()
}

func main() {
	if flag.NArg() == 0 {
		log.Fatal("a pql query must be specified")
	}

	// pql query to be executed
	query := flag.Arg(0)

	now := time.Now().UTC()
	// create range starting the given duration before now to now
	billingRng := cb.Range{
		Start: now.Add(-before),
		End:   now,
	}

	cfg := api.Config{
		Address: promURL,
	}
	prom, err := NewPrometheus(cfg)
	if err != nil {
		log.Fatal("could not setup remote: ", err)
	}

	log.Println("Testing storage...")
	store, err := setupStore(storeURL)
	if err != nil {
		log.Fatal("Could not setup storage: ", err)
	}

	_, err = bill(prom, store, query, subject, billingRng, timePrecision, window, rollup)
	if err != nil {
		log.Fatalf("Failed to bill for period %v to %v for query '%s': %v",
			billingRng.Start, billingRng.End, query, err)
	}
}
