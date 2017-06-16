package promsum

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

// mockPromAPI implements the Prometheus API interface.
var _ v1.API = mockPromAPI{}

// NewMockPromAPI initializes a mock of the Prometheus API.
func NewMockPromAPI(t *testing.T) mockPromAPI {
	return mockPromAPI{
		responseCh: make(chan mockPromResponse, 10),
		T:          t,
	}
}

// mockPromAPI allows Prometheus API responses to be injected for testing purposes.
type mockPromAPI struct {
	// responseCh holds values that are returned when API queries are made.
	responseCh chan mockPromResponse
	// for test logging
	*testing.T
	// counts API calls for use in hashing
	counter uint32
}

// mockPromResponse contains both a Value and Error to be returned in mocking.
type mockPromResponse struct {
	model.Value
	error
}

// QueryRange returns the next mockPromResponse as the response to any query.
func (a mockPromAPI) QueryRange(ctx context.Context, query string, r v1.Range) (model.Value, error) {
	id := a.id(query, r.Start, r.End)
	a.Logf("%s: Mock Prometheus queried with '%s' for %v to %v.", id, query, r.Start.Unix(), r.End.Unix())
	response := <-a.responseCh
	a.Logf("%s: Mock Prometheus responded with value='%v', error='%v'", id, response.Value, response.error)
	return response.Value, response.error
}

// id returns a unique identifier for the query based on query text and optionally a series of times.
func (a mockPromAPI) id(query string, times ...time.Time) string {
	count := atomic.AddUint32(&a.counter, 1)
	num := int64(0)
	for _, t := range times {
		num += t.Unix()
	}
	h := hash(fmt.Sprint(num, query))
	return fmt.Sprintf("%x(%d)", h, count)
}

// satisfy interface
func (mockPromAPI) Query(ctx context.Context, query string, ts time.Time) (model.Value, error) {
	return nil, errors.New("not implemented")
}
func (mockPromAPI) LabelValues(ctx context.Context, label string) (model.LabelValues, error) {
	return nil, errors.New("not implemented")
}
