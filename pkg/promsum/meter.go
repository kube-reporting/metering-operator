package promsum

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"

	cb "github.com/coreos-inc/kube-chargeback/pkg/chargeback/v1"
)

// Meter creates a billing record for a given range and Prometheus query. It does this by summing usage
// between each Prometheus instant vector by multiplying rate against against the length of the interval.
// Amounts will be rounded to the nearest unit of time specified by timePrecision.
func Meter(prom v1.API, pqlQuery, queryName string, rng cb.Range, timePrecision time.Duration) ([]BillingRecord, error) {
	if prom == nil {
		return nil, errors.New("prometheus API was nil")
	} else if timePrecision < PromTimePrecision {
		return nil, fmt.Errorf("prometheous only supports precision down to the %v", PromTimePrecision)
	}

	pRng := v1.Range{
		Start: rng.Start,
		End:   rng.End,
		Step:  5 * time.Minute,
	}

	pVal, err := prom.QueryRange(context.Background(), pqlQuery, pRng)
	if err != nil {
		return nil, fmt.Errorf("failed to perform billing query: %v", err)
	}

	matrix, ok := pVal.(model.Matrix)
	if !ok {
		return nil, fmt.Errorf("expected a matrix in response to query, got a %v", pVal.Type())
	}

	records := []BillingRecord{}
	// iterate over segments of contiguous billing records
	for _, sampleStream := range matrix {
		for _, value := range sampleStream.Values {
			labels := make(map[string]string, len(sampleStream.Metric))
			for k, v := range sampleStream.Metric {
				labels[string(k)] = string(v)
			}

			record := BillingRecord{
				Labels:        labels,
				QueryName:     queryName,
				Amount:        float64(value.Value),
				TimePrecision: timePrecision,
				Timestamp:     value.Timestamp.Time().UTC(),
			}
			records = append(records, record)
		}
	}
	return records, nil
}
