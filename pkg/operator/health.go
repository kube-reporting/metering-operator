package operator

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
func (op *Reporting) readinessHandler(w http.ResponseWriter, r *http.Request) {
	logger := newRequestLogger(op.logger, r, op.rand)
	if !op.isInitialized() {
		logger.Debugf("not ready: operator is not yet initialized")
		writeResponseAsJSON(logger, w, http.StatusInternalServerError,
			statusResponse{
				Status:  "not ready",
				Details: "not initialized",
			})
		return
	}
	if !op.testReadFromPrestoSingleFlight(logger) {
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
func (op *Reporting) healthinessHandler(w http.ResponseWriter, r *http.Request) {
	logger := newRequestLogger(op.logger, r, op.rand)
	if !op.testWriteToPrestoSingleFlight(logger) {
		writeResponseAsJSON(logger, w, http.StatusInternalServerError,
			statusResponse{
				Status:  "not healthy",
				Details: "cannot write to PrestoDB",
			})
		return
	}
	writeResponseAsJSON(logger, w, http.StatusOK, statusResponse{Status: "ok"})
}

func (op *Reporting) testWriteToPrestoSingleFlight(logger logrus.FieldLogger) bool {
	const key = "presto-write"
	v, _, _ := op.healthCheckSingleFlight.Do(key, func() (interface{}, error) {
		defer op.healthCheckSingleFlight.Forget(key)
		healthy := op.testWriteToPresto(logger)
		return healthy, nil
	})
	healthy := v.(bool)
	return healthy
}

func (op *Reporting) testReadFromPrestoSingleFlight(logger logrus.FieldLogger) bool {
	const key = "presto-read"
	v, _, _ := op.healthCheckSingleFlight.Do(key, func() (interface{}, error) {
		defer op.healthCheckSingleFlight.Forget(key)
		healthy := op.testReadFromPresto(logger)
		return healthy, nil
	})
	healthy := v.(bool)
	return healthy
}

func (op *Reporting) testReadFromPresto(logger logrus.FieldLogger) bool {
	_, err := presto.ExecuteSelect(op.prestoConn, "SELECT * FROM system.runtime.nodes")
	if err != nil {
		logger.WithError(err).Debugf("cannot query Presto system.runtime.nodes table")
		return false
	}
	return true
}

func (op *Reporting) testWriteToPresto(logger logrus.FieldLogger) bool {
	logger = logger.WithField("component", "testWriteToPresto")
	const tableName = "operator_health_check"
	err := op.createTableForStorageNoCR(logger, nil, tableName, []hive.Column{{Name: "check_time", Type: "TIMESTAMP"}})
	if err != nil {
		logger.WithError(err).Errorf("cannot create Presto table %s", tableName)
		return false
	}
	// Hive does not support timezones, and now() returns a
	// TIMESTAMP WITH TIMEZONE so we cast the return of now() to a TIMESTAMP.
	err = presto.InsertInto(op.prestoQueryer, tableName, "VALUES (cast(now() AS TIMESTAMP))")
	if err != nil {
		logger.WithError(err).Errorf("cannot insert into Presto table %s", tableName)
		return false
	}
	return true
}
