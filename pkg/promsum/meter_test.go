package promsum

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/prometheus/common/model"
)

func TestMeterQueryError(t *testing.T) {
	prom := NewMockPromAPI(t)
	prom.responseCh <- mockPromResponse{
		error: errors.New("any error"),
	}

	rng := Range{Start: time.Unix(0, 0), End: time.Unix(100, 0)}
	_, err := Meter(prom, "bad query", rng)
	if err == nil {
		t.Error("metering should have failed due to error")
	}

	// check handling when prom is nil
	_, err = Meter(nil, "cluster_namespace_controller_pod_container:memory_usage:bytes", rng)
	if err == nil {
		t.Error("error should be returned if prometheus API is nil")
	}
}

func TestMeterScalarQuery(t *testing.T) {
	prom := NewMockPromAPI(t)

	end := time.Now().UTC()
	rng := Range{
		Start: end.Add(-20 * time.Minute),
		End:   end,
	}
	query := 2

	// scalar queries will always return the same value
	prom.responseCh <- mockPromResponse{
		Value: model.Matrix{
			{
				Values: []model.SamplePair{
					{
						Timestamp: model.Time(rng.Start.Unix()),
						Value:     model.SampleValue(query),
					},
					{
						Timestamp: model.Time(rng.End.Unix()),
						Value:     model.SampleValue(query),
					},
				},
			},
		},
	}
	queryStr := fmt.Sprintf("%d", query)
	record, err := Meter(prom, queryStr, rng)
	if err != nil {
		t.Error("unexpected error: ", err)
		return
	}

	if !record.Start.Equal(rng.Start) {
		t.Errorf("unexpected start time: want %v, got %v", rng.Start, record.Start)
	}

	if !record.End.Equal(rng.End) {
		t.Errorf("unexpected end time: want %v, got %v", rng.End, record.End)
	}

	if record.Query != queryStr {
		t.Errorf("returned query does not match request: want %s, got %s", queryStr, record.Query)
	}

	expectedTotal := rng.End.Sub(rng.Start).Seconds() * float64(query)
	if record.Amount != expectedTotal {
		t.Errorf("amount billed does not match expected: want %f, got %f", expectedTotal, record.Amount)
	}
}
