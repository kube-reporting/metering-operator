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
	maxPeriodSize time.Duration
	subject       string
	storageDir    string
)

func init() {
	flag.Parse()

	flag.DurationVar(&before, "before", 1*time.Hour, "duration before present to start collect billing data")
	flag.DurationVar(&maxPeriodSize, "max-period", 20*time.Minute, "duration after a range gets broken into another range")
	flag.StringVar(&subject, "subject", fmt.Sprintf("%x", time.Now().Second()), "name used to group billing data")
	flag.StringVar(&storageDir, "path", "./data", "system path to read/write billing data")
}

func main() {
	if flag.NArg() == 0 {
		log.Fatal("a pql query must be specified")
	}

	// pql query to be executed
	query := flag.Arg(0)

	now := time.Now().UTC()
	// create range starting the given duration before now to now
	billingRng := promsum.Range{
		Start: now.Add(-before),
		End:   now,
	}

	cfg := api.Config{
		Address: "http://localhost:9090",
	}
	prom, err := promsum.NewRemote(cfg)
	if err != nil {
		log.Fatal("could not setup remote: ", err)
	}

	result, err := prom.Query(query, billingRng)
	if err != nil {
		log.Fatal("could not query prom: ", err)
	}
	fmt.Print("Type of result for query is:", result)

	store, err := promsum.NewFileStore(storageDir)
	if err != nil {
		log.Fatal("Could not setup file storage: ", err)
	}

	// TODO: implement
	var p promsum.Promsum
	err = bill(p, store, query, subject, billingRng, maxPeriodSize)
	if err != nil {
		log.Fatalf("Failed to bill for period %v to %v for query '%s': %v",
			billingRng.Start, billingRng.End, query, err)
	}
}
