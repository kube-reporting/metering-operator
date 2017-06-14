package promsum

import (
	"testing"

	"github.com/prometheus/client_golang/api"
)

func TestNewPrometheus(t *testing.T) {
	// check good con***REMOVED***guration
	goodPromCfg := api.Con***REMOVED***g{
		Address: "http://localhost:9090",
	}
	prom, err := NewPrometheus(goodPromCfg)
	if err != nil {
		t.Error("unexpected error setting up Prometheus: ", err)
	} ***REMOVED*** if prom.api == nil {
		t.Error("failed to setup API interface (was nil)")
	}

	// check bad con***REMOVED***guration
	badPromCfg := api.Con***REMOVED***g{
		Address: "&&*://-localhost:9090",
	}
	prom, err = NewPrometheus(badPromCfg)
	if err == nil {
		t.Error("error should have been returned for bad con***REMOVED***g")
	}
}
