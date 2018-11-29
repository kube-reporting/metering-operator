package presto

import (
	"context"
	"database/sql"
	"fmt"
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
	err := wait.ExponentialBackoff(backoff, cond)
	if err != nil {
		if err == wait.ErrWaitTimeout {
			return nil, fmt.Errorf("timed out while waiting to connect to presto")
		}
		return nil, err
	}

	return db, nil
}
