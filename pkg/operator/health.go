package operator

import (
	"net/http"
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
	if !op.testReadFromPrestoFunc() {
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
	if !op.testWriteToPrestoFunc() {
		writeResponseAsJSON(logger, w, http.StatusInternalServerError,
			statusResponse{
				Status:  "not healthy",
				Details: "cannot write to PrestoDB",
			})
		return
	}
	writeResponseAsJSON(logger, w, http.StatusOK, statusResponse{Status: "ok"})
}
