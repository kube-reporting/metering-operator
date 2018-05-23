package chargeback

import (
	"net/http"

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
		srv.writeResponseWithBody(logger, w, http.StatusInternalServerError,
			statusResponse{
				Status:  "not ready",
				Details: "not initialized",
			})
		return
	}
	if !srv.chargeback.testWriteToPresto(logger) {
		srv.writeResponseWithBody(logger, w, http.StatusInternalServerError,
			statusResponse{
				Status:  "not ready",
				Details: "cannot write to PrestoDB",
			})
		return
	}

	srv.writeResponseWithBody(logger, w, http.StatusOK, statusResponse{Status: "ok"})
}

// healthinessHandler is the health check for the metering operator. If this
// fails, the process will be restarted.
func (srv *server) healthinessHandler(w http.ResponseWriter, r *http.Request) {
	logger := srv.newLogger(r)
	if !srv.chargeback.testWriteToPresto(logger) {
		srv.writeResponseWithBody(logger, w, http.StatusInternalServerError,
			statusResponse{
				Status:  "not healthy",
				Details: "cannot write to PrestoDB",
			})
		return
	}
	srv.writeResponseWithBody(logger, w, http.StatusOK, statusResponse{Status: "ok"})
}

func (c *Chargeback) testWriteToPresto(logger logrus.FieldLogger) bool {
	err := c.hiveQueryer.Query("CREATE TABLE IF NOT EXISTS chargeback_health_check (check_time TIMESTAMP)")
	if err != nil {
		logger.WithError(err).Debugf("cannot create Presto table chargeback_health_check")
		return false
	}
	// Hive does not support timezones, and now() returns a
	// TIMESTAMP WITH TIMEZONE so we cast the return of now() to a TIMESTAMP.
	_, err = c.prestoConn.Query("INSERT INTO chargeback_health_check VALUES (cast(now() AS TIMESTAMP))")
	if err != nil {
		logger.WithError(err).Debugf("cannot insert into Presto table chargeback_health_check")
		return false
	}
	return true
}
