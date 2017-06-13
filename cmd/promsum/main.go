package main

import (
	"flag"
	"log"
	"time"

	"github.com/coreos-inc/kube-chargeback/pkg/promsum"
)

var (
	before        time.Duration
	maxPeriodSize time.Duration
)

func init() {
	flag.Parse()

	flag.DurationVar(&before, "before", 1*time.Hour, "duration before present to start collect billing data")
	flag.DurationVar(&maxPeriodSize, "max-period", 20*time.Minute, "duration after a range gets broken into another range")
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

	// TODO: implement
	var p promsum.Promsum
	err := bill(p, query, billingRng, maxPeriodSize)
	if err != nil {
		log.Fatalf("Failed to bill for period %v to %v for query '%s': %v",
			billingRng.Start, billingRng.End, query, err)
	}
}
