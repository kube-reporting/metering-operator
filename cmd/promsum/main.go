package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/coreos-inc/kube-chargeback/pkg/promsum"

	"github.com/prometheus/client_golang/api"
)

var (
	before        time.Duration
	timePrecision time.Duration
	subject       string
	storageDir    string
	aggregate     bool
)

func init() {
	flag.DurationVar(&before, "before", 1*time.Hour, "duration before present to start collect billing data")
	flag.DurationVar(&timePrecision, "precision", time.Second, "unit of time used for stored amounts")
	flag.StringVar(&subject, "subject", fmt.Sprintf("%x", time.Now().Nanosecond()), "name used to group billing data")
	flag.StringVar(&storageDir, "path", "./data", "system path to read/write billing data")
	flag.BoolVar(&aggregate, "aggregate", false, "summarizes the ranged in as few billing records as possible")

	flag.Parse()
}

func main() {
	if flag.NArg() == 0 {
		log.Fatal("a pql query must be speci***REMOVED***ed")
	}

	// pql query to be executed
	query := flag.Arg(0)

	now := time.Now().UTC()
	// create range starting the given duration before now to now
	billingRng := promsum.Range{
		Start: now.Add(-before),
		End:   now,
	}

	cfg := api.Con***REMOVED***g{
		Address: "http://localhost:9090",
	}
	prom, err := NewPrometheus(cfg)
	if err != nil {
		log.Fatal("could not setup remote: ", err)
	}

	log.Println("Testing metering...")
	records, err := promsum.Meter(prom, query, billingRng, timePrecision)
	if err != nil {
		log.Fatalf("Failed to meter for %v for query '%s': %v", billingRng, query, err)
	}

	var total float64
	fmt.Println("Produced records:")
	for _, r := range records {
		log.Println("- ", r)
		total += r.Amount
	}
	fmt.Printf("Total usage over %v: %f\n", billingRng, total)

	log.Println("Testing storage...")
	store, err := promsum.NewFileStore(storageDir)
	if err != nil {
		log.Fatal("Could not setup ***REMOVED***le storage: ", err)
	}

	records, err = bill(prom, store, query, subject, billingRng, timePrecision)
	if err != nil {
		log.Fatalf("Failed to bill for period %v to %v for query '%s': %v",
			billingRng.Start, billingRng.End, query, err)
	}

	if aggregate {
		totals, err := promsum.Aggregate(records, billingRng, []string{"pod", "namespace", "container"})
		if err != nil {
			log.Fatalf("Failed to aggregate for %s over %v: %v", query, billingRng, err)
		}

		log.Println("Found aggregates:")
		for _, r := range totals {
			log.Println(r.String())
		}
	}
}
