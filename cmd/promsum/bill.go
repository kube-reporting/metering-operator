package main

import (
	"fmt"
	"log"
	"time"

	promV1 "github.com/prometheus/client_golang/api/prometheus/v1"

	cb "github.com/coreos-inc/kube-chargeback/pkg/chargeback"
	p "github.com/coreos-inc/kube-chargeback/pkg/promsum"
)

var (
	// ChargebackPodLabels are the keys of labels used to group records during aggregation.
	ChargebackPodLabels = []string{"pod", "namespace", "container"}
)

// bill generates billing records and persists them.
func bill(prom promV1.API, store p.Store, query, subject string, rng cb.Range, unit, window, rollup time.Duration) (records []p.BillingRecord, err error) {
	records, err = p.Meter(prom, query, subject, rng, unit, window)
	if err != nil {
		return nil, fmt.Errorf("Failed to generate billing report for query '%s' in the range %v to %v: %v",
			query, rng.Start, rng.End, err)
	}

	records, err = rollupRecords(records, rng, rollup)
	if err != nil {
		log.Fatal("Couldn't rollup records: ", err)
	}

	err = store.Write(records)
	if err != nil {
		log.Print("Failed to record: ", err)
	}
	return
}

func rollupRecords(records []p.BillingRecord, rng cb.Range, rollup time.Duration) (out []p.BillingRecord, err error) {
	if len(records) > 0 {
		return nil, nil
	}

	var rolledRecords []p.BillingRecord
	curRoll := cb.Range{Start: records[0].Start}
	for curRoll.Start.Before(rng.End) {
		curRoll.End = curRoll.Start.Add(rollup)

		rolledRecords, err = p.Aggregate(records, curRoll, ChargebackPodLabels)
		if err != nil {
			err = fmt.Errorf("failed to aggregate for %v in %v: %v", curRoll, rng, err)
			return
		}

		out = append(out, rolledRecords...)

		// set for next interval
		curRoll.Start = curRoll.End
	}
	return
}
