package promsum

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/promql"
)

// mockPromAPI implements the Prometheus API interface.
var _ v1.API = mockPromAPI{}

// NewMockPromAPI initializes a mock of the Prometheus API.
func NewMockPromAPI(t *testing.T, input ...string) mockPromAPI {
	test, err := promql.NewTest(t, strings.Join(input, "\n"))
	if err != nil {
		t.Fatal("Could not setup Prometheus mock: ", err)
	}
	return mockPromAPI{
		promTest: test,
		T:        t,
	}
}

// mockPromAPI allows Prometheus API responses to be injected for testing purposes.
type mockPromAPI struct {
	promTest *promql.Test
	// for test logging
	*testing.T
	// counts API calls for use in hashing
	counter uint32
}

// QueryRange returns the next mockPromResponse as the response to any query.
func (a mockPromAPI) QueryRange(ctx context.Context, query string, r v1.Range) (model.Value, error) {
	id := a.id(query, r.Start, r.End)
	a.Logf("%s: Mock Prometheus queried with '%s' for %v to %v.", id, query, r.Start.Unix(), r.End.Unix())

	// perform query
	start, end := model.TimeFromUnixNano(r.Start.UnixNano()), model.TimeFromUnixNano(r.End.UnixNano())
	pQuery, err := a.promTest.QueryEngine().NewRangeQuery(query, start, end, r.Step)
	if err != nil {
		return nil, err
	}

	response := pQuery.Exec(ctx)
	a.Logf("%s: Mock Prometheus responded with value='%v', error='%v'", id, response.Value, response.Err)
	return response.Value, response.Err
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
