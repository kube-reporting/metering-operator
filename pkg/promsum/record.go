package promsum

import (
	"time"
)

// BillingRecord is a receipt of a usage determined by a query within a specific time range.
type BillingRecord struct {
	Query   string    `json:"query"`
	Subject string    `json:"subject"`
	Amount  float64   `json:"amount"`
	Start   time.Time `json:"start"`
	End     time.Time `json:"end"`
}

// Range returns the range of the billing record.
func (record BillingRecord) Range() Range {
	return Range{
		Start: record.Start,
		End:   record.End,
	}
}

// Gaps returns the ranges which don't yet have billing records.
func Gaps(records []BillingRecord, rng Range) ([]Range, error) {
	return nil, nil
}
