package promsum

import (
	"testing"

	"github.com/prometheus/client_golang/api"
)

func TestNewPrometheus(t *testing.T) {
	// check good configuration
	goodPromCfg := api.Config{
		Address: "http://localhost:9090",
	}
	prom, err := NewPrometheus(goodPromCfg)
	if err != nil {
		t.Error("unexpected error setting up Prometheus: ", err)
	} else if prom.api == nil {
		t.Error("failed to setup API interface (was nil)")
	}

	// check bad configuration
	badPromCfg := api.Config{
		Address: "&&*://-localhost:9090",
	}
	prom, err = NewPrometheus(badPromCfg)
	if err == nil {
		t.Error("error should have been returned for bad config")
	}
}
