package promsum

import (
	"fmt"
	"time"
)

// BillingRecord is a receipt of a usage determined by a query within a speci***REMOVED***c time range.
type BillingRecord struct {
	Labels  map[string]string `json:"labels"`
	Query   string            `json:"query"`
	Subject string            `json:"subject"`
	Amount  float64           `json:"amount"`
	Start   time.Time         `json:"start"`
	End     time.Time         `json:"end"`
}

// Range returns the range of the billing record.
func (record BillingRecord) Range() Range {
	return Range{
		Start: record.Start,
		End:   record.End,
	}
}

// String returns a human readable representation of a BillingRecord.
func (record BillingRecord) String() string {
	return fmt.Sprintf("BillingRecord[Labels: %v, Query: %s, Subject: %s, Amount: %f, Start: %v, End: %v]",
		record.Labels, record.Query, record.Subject, record.Amount, record.Start, record.End)
}

// Gaps returns the ranges which don't yet have billing records.
func Gaps(records []BillingRecord, rng Range) ([]Range, error) {
	return nil, nil
}
