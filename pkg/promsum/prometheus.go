package promsum

import (
	"fmt"

	"golang.org/x/net/context"

	"github.com/prometheus/client_golang/api"
	"github.com/prometheus/client_golang/api/prometheus/v1"
)

// NewRemote connects to the Prometheus using the given configuration.
func NewRemote(promConfig api.Config) (remote Remote, err error) {
	if remote.client, err = api.NewClient(promConfig); err != nil {
		err = fmt.Errorf("can't connect to prometheus: %v", err)
		return
	}
	remote.api = v1.NewAPI(remote.client)
	return
}

// RemotePromsum allows billing based on a remote Prometheus.
type Remote struct {
	client api.Client
	api    v1.API
}

// Remote implements Promsum.
var _ Promsum = Remote{}

func (r Remote) Meter(pqlQuery string, rng Range) (BillingRecord, error) {
	return BillingRecord{}, nil
}

// Query performs the given PQL query and returns records within the given range.
func (r Remote) Query(pqlQuery string, rng Range) (string, error) {
	pRng := v1.Range{
		Start: rng.Start,
		End:   rng.End,
	}
	val, err := r.api.QueryRange(context.Background(), pqlQuery, pRng)
	if err != nil {
		return "", fmt.Errorf("could not perform PQL query '%s' on range %v to %v: %v",
			pqlQuery, rng.Start, rng.End, err)

	}
	return val.Type().String(), nil
}
