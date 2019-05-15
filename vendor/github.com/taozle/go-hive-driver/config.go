package hive

import (
	"net/url"
	"strconv"
)

type con***REMOVED***g struct {
	batchSize int64
	auth      string
}

func parseCon***REMOVED***gFromQuery(s url.Values) *con***REMOVED***g {
	batchSize := int64(1000)
	if v, err := strconv.ParseInt(s.Get("batch"), 10, 64); err == nil {
		batchSize = v
	}

	auth := s.Get("auth")

	return &con***REMOVED***g{
		batchSize: batchSize,
		auth:      auth,
	}
}
