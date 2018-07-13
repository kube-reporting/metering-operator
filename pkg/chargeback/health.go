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
func (c *Chargeback) readinessHandler(w http.ResponseWriter, r *http.Request) {
	logger := newRequestLogger(c.logger, r, c.rand)
	if !c.isInitialized() {
		logger.Debugf("not ready: operator is not yet initialized")
		writeResponseAsJSON(logger, w, http.StatusInternalServerError,
			statusResponse{
				Status:  "not ready",
				Details: "not initialized",
			})
		return
	}
	if !c.testReadFromPrestoSingleFlight(logger) {
		writeResponseAsJSON(logger, w, http.StatusInternalServerError,
			statusResponse{
				Status:  "not ready",
				Details: "cannot read from PrestoDB",
			})
		return
	}

	writeResponseAsJSON(logger, w, http.StatusOK, statusResponse{Status: "ok"})
}

// healthinessHandler is the health check for the metering operator. If this
// fails, the process will be restarted.
func (c *Chargeback) healthinessHandler(w http.ResponseWriter, r *http.Request) {
	logger := newRequestLogger(c.logger, r, c.rand)
	if !c.testWriteToPrestoSingleFlight(logger) {
		writeResponseAsJSON(logger, w, http.StatusInternalServerError,
			statusResponse{
				Status:  "not healthy",
				Details: "cannot write to PrestoDB",
			})
		return
	}
	writeResponseAsJSON(logger, w, http.StatusOK, statusResponse{Status: "ok"})
}

func (c *Chargeback) testWriteToPrestoSingleFlight(logger logrus.FieldLogger) bool {
	const key = "presto-write"
	v, _, _ := c.healthCheckSingleFlight.Do(key, func() (interface{}, error) {
		defer c.healthCheckSingleFlight.Forget(key)
		healthy := c.testWriteToPresto(logger)
		return healthy, nil
	})
	healthy := v.(bool)
	return healthy
}

func (c *Chargeback) testReadFromPrestoSingleFlight(logger logrus.FieldLogger) bool {
	const key = "presto-read"
	v, _, _ := c.healthCheckSingleFlight.Do(key, func() (interface{}, error) {
		defer c.healthCheckSingleFlight.Forget(key)
		healthy := c.testReadFromPresto(logger)
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
	err := c.createTableForStorageNoCR(logger, nil, tableName, []hive.Column{{Name: "check_time", Type: "TIMESTAMP"}})
	if err != nil {
		logger.WithError(err).Errorf("cannot create Presto table %s", tableName)
		return false
	}
	// Hive does not support timezones, and now() returns a
	// TIMESTAMP WITH TIMEZONE so we cast the return of now() to a TIMESTAMP.
	err = presto.InsertInto(c.prestoQueryer, tableName, "VALUES (cast(now() AS TIMESTAMP))")
	if err != nil {
		logger.WithError(err).Errorf("cannot insert into Presto table %s", tableName)
		return false
	}
	return true
}
