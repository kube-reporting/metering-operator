package promsum

import (
	"context"
	"errors"
	"time"

	"github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

// mockPromAPI implements the Prometheus API interface.
var _ v1.API = mockPromAPI{}

// NewMockPromAPI initializes a mock of the Prometheus API.
func NewMockPromAPI() mockPromAPI {
	return mockPromAPI{
		responseCh: make(chan mockPromResponse, 10),
	}
}

// mockPromAPI allows Prometheus API responses to be injected for testing purposes.
type mockPromAPI struct {
	// responseCh holds values that are returned when API queries are made.
	responseCh chan mockPromResponse
}

// mockPromResponse contains both a Value and Error to be returned in mocking.
type mockPromResponse struct {
	model.Value
	error
}

// QueryRange returns the next mockPromResponse as the response to any query.
func (a mockPromAPI) QueryRange(ctx context.Context, query string, r v1.Range) (model.Value, error) {
	response := <-a.responseCh
	return response.Value, response.error
}

// satisfy interface
func (mockPromAPI) Query(ctx context.Context, query string, ts time.Time) (model.Value, error) {
	return nil, errors.New("not implemented")
}
func (mockPromAPI) LabelValues(ctx context.Context, label string) (model.LabelValues, error) {
	return nil, errors.New("not implemented")
}