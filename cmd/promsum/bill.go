package main

import (
	"fmt"
	"log"
	"time"

	"github.com/coreos-inc/kube-chargeback/pkg/promsum"
)

// bill generates billing records and persists them.
func bill(p promsum.Promsum, store promsum.Store, query, subject string, rng promsum.Range, maxPeriod time.Duration) (err error) {
	var billingRecords []promsum.BillingRecord
	for {
		billingRecords, err = store.Read(rng, query, subject)
		if err != nil {
			log.Printf("Couldn't retrieve billing records, will save entire period (from %v to %v)",
				rng.Start, rng.End)
		}

		// Determine what gaps exist in data that need to be filled
		gaps, err := promsum.Gaps(billingRecords, rng)
		if err != nil {
			return fmt.Errorf("Failed to determine which records are remaining to be generated: %v", err)
		}

		// exit if there nothing left to do
		if len(gaps) == 0 {
			return nil
		}

		// attempt to create billing record for every period
		for _, rng := range gaps {
			record, err := p.Meter(query, rng)
			if err != nil {
				log.Printf("Failed to generate billing report for query '%s' in the range %v to %v: %v",
					query, rng.Start, rng.End, err)
				continue
			}

			err = store.Write(record)
			if err != nil {
				log.Print("Failed to record: ", err)
			}
		}
	}
}
