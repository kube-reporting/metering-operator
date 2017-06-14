package promsum

import (
	"fmt"

	"golang.org/x/net/context"

	"github.com/prometheus/client_golang/api"
	"github.com/prometheus/client_golang/api/prometheus/v1"
)

// NewPrometheus connects using the given con***REMOVED***guration.
func NewPrometheus(promCon***REMOVED***g api.Con***REMOVED***g) (prom Prometheus, err error) {
	if prom.client, err = api.NewClient(promCon***REMOVED***g); err != nil {
		err = fmt.Errorf("can't connect to prometheus: %v", err)
		return
	}
	prom.api = v1.NewAPI(prom.client)
	return
}

// Prometheus allows billing based on a remote Prometheus.
type Prometheus struct {
	client api.Client
	api    v1.API
}

// Meter creates a billing record for a given range and Prometheus query. It does this by summing usage
// between each Prometheus instant vector by multiplying rate against against the length of the interval.
func (prom Prometheus) Meter(pqlQuery string, rng Range) (BillingRecord, error) {
	return BillingRecord{}, nil
}

// Query performs the given PQL query and returns records within the given range.
func (prom Prometheus) Query(pqlQuery string, rng Range) (string, error) {
	pRng := v1.Range{
		Start: rng.Start,
		End:   rng.End,
	}
	val, err := prom.api.QueryRange(context.Background(), pqlQuery, pRng)
	if err != nil {
		return "", fmt.Errorf("could not perform PQL query '%s' on range %v to %v: %v",
			pqlQuery, rng.Start, rng.End, err)

	}
	return val.Type().String(), nil
}
