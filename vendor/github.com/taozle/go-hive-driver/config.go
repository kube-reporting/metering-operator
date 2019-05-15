package hive

import (
	"net/url"
	"strconv"
)

type config struct {
	batchSize int64
	auth      string
}

func parseConfigFromQuery(s url.Values) *config {
	batchSize := int64(1000)
	if v, err := strconv.ParseInt(s.Get("batch"), 10, 64); err == nil {
		batchSize = v
	}

	auth := s.Get("auth")

	return &config{
		batchSize: batchSize,
		auth:      auth,
	}
}
