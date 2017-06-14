package promsum

import (
	promV1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

// Meter creates a billing record for a given range and Prometheus query. It does this by summing usage
// between each Prometheus instant vector by multiplying rate against against the length of the interval.
func Meter(prom promV1.API, pqlQuery string, rng Range) (BillingRecord, error) {
	return BillingRecord{}, nil
}
