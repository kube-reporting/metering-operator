package presto

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/prestodb/presto-go-client/presto"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/wait"
)

func NewPrestoConnWithRetry(ctx context.Context, logger log.FieldLogger, connStr string, connBackoff time.Duration, maxRetries int) (*sql.DB, error) {
	var db *sql.DB
	backoff := wait.Backoff{
		Duration: connBackoff,
		Factor:   1.25,
		Steps:    maxRetries,
	}
	cond := func() (bool, error) {
		var err error
		db, err = sql.Open("presto", connStr)
		if err == nil {
			return true, nil
		} ***REMOVED*** {
			logger.WithError(err).Debugf("error encountered, backing off and trying again: %v", err)
		}
		return false, nil
	}
	return db, wait.ExponentialBackoff(backoff, cond)
}
