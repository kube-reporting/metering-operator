package promsum

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/api/prometheus/v1"
)

// Meter creates a billing record for a given range and Prometheus query. It does this by summing usage
// between each Prometheus instant vector by multiplying rate against against the length of the interval.
func Meter(prom v1.API, pqlQuery string, rng Range) (BillingRecord, error) {
	if prom == nil {
		return BillingRecord{}, errors.New("prometheus API was nil")
	}

	pRng := v1.Range{
		Start: rng.Start,
		End:   rng.End,
		Step:  1 * time.Minute,
	}

	_, err := prom.QueryRange(context.Background(), pqlQuery, pRng)
	if err != nil {
		return BillingRecord{}, fmt.Errorf("failed to perform billing query: %v", err)
	}
	return BillingRecord{}, nil
}
