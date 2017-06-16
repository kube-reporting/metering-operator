package main

import (
	"fmt"

	"github.com/prometheus/client_golang/api"
	"github.com/prometheus/client_golang/api/prometheus/v1"
)

// NewPrometheus connects using the given configuration.
func NewPrometheus(promConfig api.Config) (v1.API, error) {
	client, err := api.NewClient(promConfig)
	if err != nil {
		return nil, fmt.Errorf("can't connect to prometheus: %v", err)
	}
	return v1.NewAPI(client), nil
}
