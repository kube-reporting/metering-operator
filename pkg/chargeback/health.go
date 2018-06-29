package chargeback

import (
	"net/http"

	"github.com/operator-framework/operator-metering/pkg/hive"
	"github.com/operator-framework/operator-metering/pkg/presto"
	"github.com/sirupsen/logrus"
)

type statusResponse struct {
	Status  string      `json:"status"`
	Details interface{} `json:"details"`
}

// healthinessHandler is the readiness check for the metering operator. If this
// no requests will be sent to this pod, and rolling updates will not proceed
// until the checks succeed.
func (srv *server) readinessHandler(w http.ResponseWriter, r *http.Request) {
	logger := srv.newLogger(r)
	if !srv.chargeback.isInitialized() {
		logger.Debugf("not ready: operator is not yet initialized")
		srv.writeResponseWithBody(logger, w, http.StatusInternalServerError,
			statusResponse{
				Status:  "not ready",
				Details: "not initialized",
			})
		return
	}
	if !srv.testReadFromPrestoSingleFlight(logger) {
		srv.writeResponseWithBody(logger, w, http.StatusInternalServerError,
			statusResponse{
				Status:  "not ready",
				Details: "cannot read from PrestoDB",
			})
		return
	}

	srv.writeResponseWithBody(logger, w, http.StatusOK, statusResponse{Status: "ok"})
}

// healthinessHandler is the health check for the metering operator. If this
// fails, the process will be restarted.
func (srv *server) healthinessHandler(w http.ResponseWriter, r *http.Request) {
	logger := srv.newLogger(r)
	if !srv.testWriteToPrestoSingleFlight(logger) {
		srv.writeResponseWithBody(logger, w, http.StatusInternalServerError,
			statusResponse{
				Status:  "not healthy",
				Details: "cannot write to PrestoDB",
			})
		return
	}
	srv.writeResponseWithBody(logger, w, http.StatusOK, statusResponse{Status: "ok"})
}

func (srv *server) testWriteToPrestoSingleFlight(logger logrus.FieldLogger) bool {
	const key = "presto-write"
	v, _, _ := srv.healthCheckSingleFlight.Do(key, func() (interface{}, error) {
		defer srv.healthCheckSingleFlight.Forget(key)
		healthy := srv.chargeback.testWriteToPresto(logger)
		return healthy, nil
	})
	healthy := v.(bool)
	return healthy
}

func (srv *server) testReadFromPrestoSingleFlight(logger logrus.FieldLogger) bool {
	const key = "presto-read"
	v, _, _ := srv.healthCheckSingleFlight.Do(key, func() (interface{}, error) {
		defer srv.healthCheckSingleFlight.Forget(key)
		healthy := srv.chargeback.testReadFromPresto(logger)
		return healthy, nil
	})
	healthy := v.(bool)
	return healthy
}

func (c *Chargeback) testReadFromPresto(logger logrus.FieldLogger) bool {
	_, err := presto.ExecuteSelect(c.prestoConn, "SELECT * FROM system.runtime.nodes")
	if err != nil {
		logger.WithError(err).Debugf("cannot query Presto system.runtime.nodes table")
		return false
	}
	return true
}

func (c *Chargeback) testWriteToPresto(logger logrus.FieldLogger) bool {
	logger = logger.WithField("component", "testWriteToPresto")
	const tableName = "chargeback_health_check"
	err := c.createTableForStorageNoCR(logger, nil, tableName, []hive.Column{{Name: "check_time", Type: "TIMESTAMP"}}, false)
	if err != nil {
		logger.WithError(err).Debugf("cannot create Presto table %s", tableName)
		return false
	}
	// Hive does not support timezones, and now() returns a
	// TIMESTAMP WITH TIMEZONE so we cast the return of now() to a TIMESTAMP.
	err = presto.ExecuteInsertQuery(c.prestoConn, tableName, "VALUES (cast(now() AS TIMESTAMP))")
	if err != nil {
		logger.WithError(err).Debugf("cannot insert into Presto table %s", tableName)
		return false
	}
	return true
}
