package promsum

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

// Meter creates a billing record for a given range and Prometheus query. It does this by summing usage
// between each Prometheus instant vector by multiplying rate against against the length of the interval.
// Amounts will be rounded to the nearest unit of time speci***REMOVED***ed by timePrecision.
func Meter(prom v1.API, pqlQuery string, rng Range, timePrecision time.Duration) ([]BillingRecord, error) {
	if prom == nil {
		return nil, errors.New("prometheus API was nil")
	} ***REMOVED*** if timePrecision < PromTimePrecision {
		return nil, fmt.Errorf("prometheous only supports precision down to the %v", PromTimePrecision)
	}

	pRng := v1.Range{
		Start: rng.Start,
		End:   rng.End,
		Step:  1 * time.Minute,
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
		for i := 1; i < len(sampleStream.Values); i++ {
			start, end := sampleStream.Values[i-1], sampleStream.Values[i]

			total, err := CalculateUsage(start, end, timePrecision)
			if err != nil {
				return nil, fmt.Errorf("can't calculate usage for range %v to %v for query '%s': %v",
					start.Timestamp, end.Timestamp, pqlQuery, err)
			}

			labels := make(map[string]string, len(sampleStream.Metric))
			for k, v := range sampleStream.Metric {
				labels[string(k)] = string(v)
			}

			record := BillingRecord{
				Labels: labels,
				Query:  pqlQuery,
				Amount: total,
				Start:  start.Timestamp.Time().UTC(),
				End:    end.Timestamp.Time().UTC(),
			}
			records = append(records, record)
		}
	}
	return records, nil
}

// CalculateUsage determines how much of a resource was used between two instances of a SamplePair. Usage is determined
// by the simple average of the values of the samples multiplied by the duration of the period in milliseconds. Amounts
// will be rounded to the nearest unit of time speci***REMOVED***ed by timePrecision.
// The start sample must come before the end sample.
func CalculateUsage(start, end model.SamplePair, timePrecision time.Duration) (float64, error) {
	if end.Timestamp.Before(start.Timestamp) {
		return 0, fmt.Errorf("start (%v) must be before end (%d)", int64(start.Timestamp), int64(end.Timestamp))
	}

	avg := float64(start.Value+end.Value) / 2
	duration := float64(end.Timestamp - start.Timestamp)
	total := avg * duration

	// adjust for precision
	return total / float64(timePrecision/PromTimePrecision), nil
}
