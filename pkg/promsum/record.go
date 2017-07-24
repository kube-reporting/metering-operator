package promsum

import (
	"fmt"
	"time"

	cb "github.com/coreos-inc/kube-chargeback/pkg/chargeback"
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
func (record BillingRecord) Range() cb.Range {
	return cb.Range{
		Start: record.Start,
		End:   record.End,
	}
}

// String returns a human readable representation of a BillingRecord.
func (record BillingRecord) String() string {
	return fmt.Sprintf("BillingRecord[Labels: %v, Query: %s, Subject: %s, Amount: %f, Start: %v, End: %v]",
		record.Labels, record.Query, record.Subject, record.Amount, record.Start, record.End)
}

// Prorate returns a new BillingRecord for a portion of this period. The amount is determined proportionally.
func (record BillingRecord) Prorate(rng cb.Range) (BillingRecord, error) {
	if rng.Start.Before(record.Start) || rng.End.After(record.End) {
		return BillingRecord{}, fmt.Errorf("prorate (%v) must be in range of the BillingRecord (%v)", rng, record.Range())
	}

	prorateDur := rng.End.Sub(rng.Start)
	recordDur := record.End.Sub(record.Start)
	portion := float64(prorateDur) / float64(recordDur)

	// update ***REMOVED***elds for new prorated range
	record.Start = rng.Start
	record.End = rng.End
	record.Amount = record.Amount * portion
	return record, nil
}
