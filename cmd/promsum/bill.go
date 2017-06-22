package main

import (
	"fmt"
	"log"
	"time"

	"github.com/coreos-inc/kube-chargeback/pkg/promsum"

	promV1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

// bill generates billing records and persists them.
func bill(prom promV1.API, store promsum.Store, query, subject string, rng promsum.Range, timePrecision time.Duration) (records []promsum.BillingRecord, err error) {
	records, err = promsum.Meter(prom, query, rng, timePrecision)
	if err != nil {
		return nil, fmt.Errorf("Failed to generate billing report for query '%s' in the range %v to %v: %v",
			query, rng.Start, rng.End, err)
	}

	for _, r := range records {
		r.Subject = subject
		err = store.Write(r)
		if err != nil {
			log.Print("Failed to record: ", err)
		}
	}
	return
}
