package main

import (
	"fmt"
	"log"
	"time"

	"github.com/coreos-inc/kube-chargeback/pkg/promsum"
)

// bill generates billing records and persists them.
func bill(p promsum.Promsum, pqlQuery string, billingRng promsum.Range, maxPeriodSize time.Duration) (err error) {
	var billingRecords []promsum.BillingRecord
	for {
		billingRecords, err = p.Read(billingRng)
		if err != nil {
			log.Printf("Couldn't retrieve billing records, will save entire period (from %v to %v)",
				billingRng.Start, billingRng.End)
		}

		// Determine what gaps exist in data that need to be filled
		gaps, err := promsum.Gaps(billingRecords, billingRng)
		if err != nil {
			return fmt.Errorf("Failed to determine which records are remaining to be generated: %v", err)
		}

		// exit if there nothing left to do
		if len(gaps) == 0 {
			return nil
		}

		// divide gaps that are larger than the max period size
		gaps, err = promsum.Segment(gaps, maxPeriodSize)
		if err != nil {
			log.Fatal("Unable to determine billing periods from data gaps: ", err)
		}

		// attempt to create billing record for every period
		for _, rng := range gaps {
			record, err := p.Meter(pqlQuery, rng)
			if err != nil {
				log.Printf("Failed to generate billing report for query '%s' in the range %v to %v: %v",
					pqlQuery, rng.Start, rng.End, err)
				continue
			}

			err = p.Write(record)
			if err != nil {
				log.Print("Failed to record: ", err)
			}
		}
	}
}
