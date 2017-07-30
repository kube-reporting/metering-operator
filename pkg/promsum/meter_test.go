package promsum

import (
	"fmt"
	"testing"
	"time"

	cb "github.com/coreos-inc/kube-chargeback/pkg/chargeback"
)

func TestMeterQueryError(t *testing.T) {
	prom := NewMockPromAPI(t)
	subject := "unitTestingQuery"

	rng := cb.Range{Start: time.Unix(0, 0), End: time.Unix(100, 0)}
	_, err := Meter(prom, "bad query", subject, rng, PromTimePrecision)
	if err == nil {
		t.Error("metering should have failed due to error")
	}

	// check handling when prom is nil
	_, err = Meter(nil, "cluster_namespace_controller_pod_container:memory_usage:bytes", subject, rng, PromTimePrecision)
	if err == nil {
		t.Error("error should be returned if prometheus API is nil")
	}
}

func TestMeterScalarQuery(t *testing.T) {
	prom := NewMockPromAPI(t)
	subject := "unitTestingScalar"

	// track the interval 20 minutes into the past
	end := time.Now().UTC()
	start := end.Add(-20 * time.Minute)

	rng := cb.Range{
		Start: start.Round(PromTimePrecision),
		End:   end.Round(PromTimePrecision),
	}
	query := int64(2)
	timePrecision := time.Second

	// scalar queries will always return the same value
	queryStr := fmt.Sprintf("%d", query)
	records, err := Meter(prom, queryStr, subject, rng, timePrecision)
	if err != nil {
		t.Error("unexpected error: ", err)
		return
	}

	duration := rng.End.Sub(rng.Start).Nanoseconds() / int64(PromTimePrecision)
	expectedTotal := float64(duration * query)

	// adjust for desired precision
	expectedTotal = expectedTotal / float64(timePrecision/PromTimePrecision)

	var actualTotal float64
	for i := 0; i < len(records); i++ {
		record := records[i]
		if i != 0 && !records[i-1].End.Equal(record.Start) {
			t.Errorf("next segment should start when the last ends: want %v, got %v",
				records[i-1].End.Format(time.RFC3339), record.Start.Format(time.RFC3339))
		}

		if record.Query != queryStr {
			t.Errorf("returned query does not match request: want %s, got %s", queryStr, record.Query)
		}
		actualTotal += record.Amount
	}

	if actualTotal != expectedTotal {
		t.Errorf("amount billed does not match expected: want %f, got %f", expectedTotal, actualTotal)
	}
}
