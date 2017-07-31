package main

import (
	"fmt"
	"log"
	"time"

	promV1 "github.com/prometheus/client_golang/api/prometheus/v1"

	cb "github.com/coreos-inc/kube-chargeback/pkg/chargeback"
	"github.com/coreos-inc/kube-chargeback/pkg/promsum"
)

// bill generates billing records and persists them.
func bill(prom promV1.API, store promsum.Store, query, subject string, rng cb.Range, unit, window time.Duration) (records []promsum.BillingRecord, err error) {
	records, err = promsum.Meter(prom, query, subject, rng, unit, window)
	if err != nil {
		return nil, fmt.Errorf("Failed to generate billing report for query '%s' in the range %v to %v: %v",
			query, rng.Start, rng.End, err)
	}

	err = store.Write(records)
	if err != nil {
		log.Print("Failed to record: ", err)
	}
	return
}
