package promsum

import (
	"time"
)

// BillingRecord is a receipt of a usage determined by a query within a speci***REMOVED***c time range.
type BillingRecord struct {
	Query  string
	Amount float64
	Unit   string
	Start  time.Time
	End    time.Time
}

// Gaps returns the ranges which don't yet have billing records.
func Gaps(records []BillingRecord, rng Range) ([]Range, error) {
	return nil, nil
}
