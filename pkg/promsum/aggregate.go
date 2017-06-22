package promsum

import (
	"fmt"
)

func Aggregate(records []BillingRecord, rng Range, mergeLabels []string) ([]BillingRecord, error) {
	// create map[string][string]BillingRecord
	recordMap := map[uint64]BillingRecord{}
	// iterate through every BillingRecord (r) within range
	for _, record := range records {
		var err error
		// check that entire range is within record, if not bill for entire range
		if rng.Within(record.Start) && rng.Within(record.End) {
			// record is wholely within the range, no prorating necessary
		} else if record.Range().Within(rng.Start) && record.Range().Within(rng.End) {
			// entire record is within range, prorate range
			record, err = record.Prorate(rng)
		} else if !rng.Within(record.Start) && rng.Within(record.End) {
			prorateRng := Range{rng.Start, record.End}
			record, err = record.Prorate(prorateRng)
		} else if !rng.Within(record.End) && rng.Within(record.Start) {
			prorateRng := Range{record.Start, rng.End}
			record, err = record.Prorate(prorateRng)
		} else {
			err = fmt.Errorf("record range (%v) is not within given range (%v)", record.Range(), rng)
		}

		if err != nil {
			fmt.Errorf("error prorating record: %v", err)
		}

		// calculate hash using subject, query and keys/values of the labels whose keys are in mergeLabels
		hashText := record.Subject + record.Query
		for _, labelKey := range mergeLabels {
			val, _ := record.Labels[labelKey]
			hashText += labelKey + ":" + val + ","
		}
		mapKey := hash(hashText)

		storedRecord, ok := recordMap[mapKey]
		if !ok {
			storedRecord = record
		} else if storedRecord.End != record.Start {
			return nil, fmt.Errorf("end of (%v) doesnt match start of (%v), first: %v, second: %v",
				storedRecord.End, record.Start, storedRecord, record)
		} else {
			storedRecord.Amount = storedRecord.Amount + record.Amount
			storedRecord.End = record.End
		}
		recordMap[mapKey] = storedRecord
	}

	var recordsOut []BillingRecord
	for _, record := range recordMap {
		recordsOut = append(recordsOut, record)
	}
	return recordsOut, nil
}
