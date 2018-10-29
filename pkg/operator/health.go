package operator

import (
	"net/http"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/singleflight"

	"github.com/operator-framework/operator-metering/pkg/db"
	"github.com/operator-framework/operator-metering/pkg/hive"
	"github.com/operator-framework/operator-metering/pkg/presto"
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

type prestoHealthChecker struct {
	logger       logrus.FieldLogger
	queryer      db.Queryer
	tableManager TableManager

	tableProperties hive.TableProperties
	// ensures only at most a single testRead query is running against Presto
	// at one time
	healthCheckSingleFlight singleflight.Group
}

func NewPrestoHealthChecker(logger logrus.FieldLogger, queryer db.Queryer, tableManager TableManager, tableProperties hive.TableProperties) *prestoHealthChecker {
	return &prestoHealthChecker{
		logger:          logger,
		queryer:         queryer,
		tableManager:    tableManager,
		tableProperties: tableProperties,
	}
}

func (checker *prestoHealthChecker) TestWriteToPrestoSingleFlight() bool {
	const key = "presto-write"
	v, _, _ := checker.healthCheckSingleFlight.Do(key, func() (interface{}, error) {
		defer checker.healthCheckSingleFlight.Forget(key)
		healthy := checker.TestWriteToPresto()
		return healthy, nil
	})
	healthy := v.(bool)
	return healthy
}

func (checker *prestoHealthChecker) TestReadFromPrestoSingleFlight() bool {
	const key = "presto-read"
	v, _, _ := checker.healthCheckSingleFlight.Do(key, func() (interface{}, error) {
		defer checker.healthCheckSingleFlight.Forget(key)
		healthy := checker.TestReadFromPresto()
		return healthy, nil
	})
	healthy := v.(bool)
	return healthy
}

func (checker *prestoHealthChecker) TestReadFromPresto() bool {
	_, err := presto.ExecuteSelect(checker.queryer, "SELECT * FROM system.runtime.nodes")
	if err != nil {
		checker.logger.WithError(err).Debugf("cannot query Presto system.runtime.nodes table")
		return false
	}
	return true
}

func (checker *prestoHealthChecker) TestWriteToPresto() bool {
	logger := checker.logger.WithField("component", "testWriteToPresto")
	const tableName = "operator_health_check"
	columns := []hive.Column{{Name: "check_time", Type: "TIMESTAMP"}}

	params := hive.TableParameters{
		Name:         tableName,
		Columns:      columns,
		IgnoreExists: true,
	}
	err := checker.tableManager.CreateTable(params, checker.tableProperties)
	if err != nil {
		logger.WithError(err).Errorf("cannot create Presto table %s", tableName)
		return false
	}

	// Hive does not support timezones, and now() returns a
	// TIMESTAMP WITH TIMEZONE so we cast the return of now() to a TIMESTAMP.
	err = presto.InsertInto(checker.queryer, tableName, "VALUES (cast(now() AS TIMESTAMP))")
	if err != nil {
		logger.WithError(err).Errorf("cannot insert into Presto table %s", tableName)
		return false
	}
	return true
}
