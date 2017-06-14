package main

import (
	"fmt"

	"github.com/prometheus/client_golang/api"
	"github.com/prometheus/client_golang/api/prometheus/v1"
)

// NewPrometheus connects using the given con***REMOVED***guration.
func NewPrometheus(promCon***REMOVED***g api.Con***REMOVED***g) (v1.API, error) {
	client, err := api.NewClient(promCon***REMOVED***g)
	if err != nil {
		return nil, fmt.Errorf("can't connect to prometheus: %v", err)
	}
	return v1.NewAPI(client), nil
}
