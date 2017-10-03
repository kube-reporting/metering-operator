package promsum

import (
	"fmt"
	"time"
)

// BillingRecord is a receipt of a usage determined by a query within a speci***REMOVED***c time range.
type BillingRecord struct {
	Labels        map[string]string `json:"labels"`
	QueryName     string            `json:"query"`
	Amount        float64           `json:"amount"`
	TimePrecision time.Duration     `json:"timePrecision"`
	Timestamp     time.Time         `json:"timestamp"`
}

// String returns a human readable representation of a BillingRecord.
func (record BillingRecord) String() string {
	return fmt.Sprintf("BillingRecord[Labels: %v, QueryName: %s, Amount: %f, TimePrecision: %v, Timestamp: %v]",
		record.Labels, record.QueryName, record.Amount, record.TimePrecision, record.Timestamp)
}
