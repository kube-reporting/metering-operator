package promsum

import (
	"testing"

	"github.com/prometheus/client_golang/api"
)

func TestNewRemote(t *testing.T) {
	// check good con***REMOVED***guration
	goodPromCfg := api.Con***REMOVED***g{
		Address: "http://localhost:9090",
	}
	remote, err := NewRemote(goodPromCfg)
	if err != nil {
		t.Error("unexpected error setting up Prometheus: ", err)
	} ***REMOVED*** if remote.api == nil {
		t.Error("failed to setup API interface (was nil)")
	}

	// check bad con***REMOVED***guration
	badPromCfg := api.Con***REMOVED***g{
		Address: "&&*://-localhost:9090",
	}
	remote, err = NewRemote(badPromCfg)
	if err == nil {
		t.Error("error should have been returned for bad con***REMOVED***g")
	}
}
