package chargeback

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/pprof"
	"net/url"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	log "github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"

	api "github.com/operator-framework/operator-metering/pkg/apis/chargeback/v1alpha1"
	cbutil "github.com/operator-framework/operator-metering/pkg/apis/chargeback/v1alpha1/util"
	"github.com/operator-framework/operator-metering/pkg/chargeback/prestostore"
	"github.com/operator-framework/operator-metering/pkg/presto"
	"github.com/operator-framework/operator-metering/pkg/util/orderedmap"
)

var ErrReportIsRunning = errors.New("the report is still running")

type server struct {
	chargeback *Chargeback
	logger     log.FieldLogger

	// wg is used to wait on httpServer and pprofServer stopping
	wg          sync.WaitGroup
	httpServer  *http.Server
	pprofServer *http.Server

	healthCheckSingleFlight singleflight.Group
}

type requestLogger struct {
	log.FieldLogger
}

func (l *requestLogger) Print(v ...interface{}) {
	l.FieldLogger.Info(v...)
}

func newServer(c *Chargeback, logger log.FieldLogger) *server {
	router := chi.NewRouter()
	pprofMux := http.NewServeMux()

	logger = logger.WithField("component", "api")
	requestLogger := middleware.RequestLogger(&middleware.DefaultLogFormatter{Logger: &requestLogger{logger}})
	router.Use(requestLogger)

	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}
	pprofServer := &http.Server{
		Addr:    "127.0.0.1:6060",
		Handler: pprofMux,
	}

	srv := &server{
		chargeback:  c,
		logger:      logger,
		httpServer:  httpServer,
		pprofServer: pprofServer,
	}

	router.HandleFunc("/api/v1/reports/get", srv.getReportHandler)
	router.HandleFunc("/api/v2/reports/{name}/full", srv.getReportV2FullHandler)
	router.HandleFunc("/api/v2/reports/{name}/table", srv.getReportV2TableHandler)
	router.HandleFunc("/api/v1/scheduledreports/get", srv.getScheduledReportHandler)
	router.HandleFunc("/api/v1/reports/run", srv.runReportHandler)
	router.HandleFunc("/api/v1/datasources/prometheus/collect", srv.collectPromsumDataHandler)
	router.HandleFunc("/api/v1/datasources/prometheus/store/{datasourceName}", srv.storePromsumDataHandler)
	router.HandleFunc("/api/v1/datasources/prometheus/fetch/{datasourceName}", srv.fetchPromsumDataHandler)
	router.HandleFunc("/ready", srv.readinessHandler)
	router.HandleFunc("/healthy", srv.healthinessHandler)

	pprofMux.HandleFunc("/debug/pprof/", pprof.Index)
	pprofMux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	pprofMux.HandleFunc("/debug/pprof/pro***REMOVED***le", pprof.Pro***REMOVED***le)
	pprofMux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	pprofMux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	return srv
}

func (srv *server) start() {
	srv.wg.Add(2)
	go func() {
		srv.logger.Infof("HTTP API server started")
		srv.logger.WithError(srv.httpServer.ListenAndServe()).Info("HTTP API server exited")
		srv.wg.Done()
	}()
	go func() {
		srv.logger.Infof("pprof server started")
		srv.logger.WithError(srv.pprofServer.ListenAndServe()).Info("pprof server exited")
		srv.wg.Done()
	}()
}

func (srv *server) stop() {
	srv.wg.Add(2)
	go func() {
		srv.logger.Infof("stopping HTTP API server")
		err := srv.httpServer.Shutdown(context.TODO())
		if err != nil {
			srv.logger.WithError(err).Warnf("got an error shutting down HTTP API server")
		}
		srv.wg.Done()
	}()
	go func() {
		srv.logger.Infof("stopping pprof server")
		err := srv.pprofServer.Shutdown(context.TODO())
		if err != nil {
			srv.logger.WithError(err).Warnf("got an error shutting down pprof server")
		}
		srv.wg.Done()
	}()
	srv.wg.Wait()
}

func (srv *server) newLogger(r *http.Request) log.FieldLogger {
	return srv.logger.WithFields(log.Fields{
		"method": r.Method,
		"url":    r.URL.String(),
	}).WithFields(srv.chargeback.newLogIdenti***REMOVED***er())
}

type errorResponse struct {
	Error string `json:"error"`
}

func (srv *server) writeErrorResponse(logger log.FieldLogger, w http.ResponseWriter, r *http.Request, status int, message string, args ...interface{}) {
	msg := fmt.Sprintf(message, args...)
	srv.writeResponseWithBody(logger, w, status, errorResponse{Error: msg})
}

// writeResponseWithBody attempts to marshal an arbitrary thing to JSON then write
// it to the http.ResponseWriter
func (srv *server) writeResponseWithBody(logger log.FieldLogger, w http.ResponseWriter, code int, resp interface{}) {
	enc, err := json.Marshal(resp)
	if err != nil {
		logger.WithError(err).Error("failed JSON-encoding HTTP response")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if _, err = w.Write(enc); err != nil {
		logger.WithError(err).Error("failed writing HTTP response")
	}
}

func (srv *server) validateGetReportReq(logger log.FieldLogger, requiredQueryParams []string, w http.ResponseWriter, r *http.Request) bool {
	if r.Method != "GET" {
		srv.writeErrorResponse(logger, w, r, http.StatusNotFound, "Not found")
		return false
	}
	err := r.ParseForm()
	if err != nil {
		srv.writeErrorResponse(logger, w, r, http.StatusBadRequest, "couldn't parse URL query params: %v", err)
		return false
	}
	err = checkForFields(requiredQueryParams, r.Form)
	if err != nil {
		srv.writeErrorResponse(logger, w, r, http.StatusBadRequest, "%v", err)
		return false
	}
	format := r.Form["format"][0]
	if format == "json" || format == "csv" {
		return true
	}
	srv.writeErrorResponse(logger, w, r, http.StatusBadRequest, "format must be one of: csv, json")
	return false
}

func (srv *server) getReportHandler(w http.ResponseWriter, r *http.Request) {
	logger := srv.newLogger(r)
	if !srv.validateGetReportReq(logger, []string{"name", "format"}, w, r) {
		return
	}
	srv.getReport(logger, r.Form["name"][0], r.Form["format"][0], false, true, w, r)
}

func (srv *server) getReportV2FullHandler(w http.ResponseWriter, r *http.Request) {
	logger := srv.newLogger(r)
	name := chi.URLParam(r, "name")
	if !srv.validateGetReportReq(logger, []string{"format"}, w, r) {
		return
	}
	srv.getReport(logger, name, r.Form["format"][0], true, true, w, r)
}

func (srv *server) getReportV2TableHandler(w http.ResponseWriter, r *http.Request) {
	logger := srv.newLogger(r)
	name := chi.URLParam(r, "name")
	if !srv.validateGetReportReq(logger, []string{"format"}, w, r) {
		return
	}
	srv.getReport(logger, name, r.Form["format"][0], true, false, w, r)
}

func (srv *server) getScheduledReportHandler(w http.ResponseWriter, r *http.Request) {
	logger := srv.newLogger(r)
	if !srv.validateGetReportReq(logger, []string{"name", "format"}, w, r) {
		return
	}
	srv.getScheduledReport(logger, r.Form["name"][0], r.Form["format"][0], w, r)
}

func (srv *server) runReportHandler(w http.ResponseWriter, r *http.Request) {
	logger := srv.newLogger(r)
	if r.Method != "GET" {
		srv.writeErrorResponse(logger, w, r, http.StatusNotFound, "Not found")
		return
	}
	err := r.ParseForm()
	if err != nil {
		srv.writeErrorResponse(logger, w, r, http.StatusBadRequest, "couldn't parse URL query params: %v", err)
		return
	}
	vals := r.Form
	err = checkForFields([]string{"query", "start", "end"}, vals)
	if err != nil {
		srv.writeErrorResponse(logger, w, r, http.StatusBadRequest, "%v", err)
		return
	}
	srv.runReport(logger, vals["query"][0], vals["start"][0], vals["end"][0], w)
}

func checkForFields(***REMOVED***elds []string, vals url.Values) error {
	var missingFields []string
	for _, f := range ***REMOVED***elds {
		if len(vals[f]) == 0 || vals[f][0] == "" {
			missingFields = append(missingFields, f)
		}
	}
	if len(missingFields) != 0 {
		return fmt.Errorf("the following ***REMOVED***elds are missing or empty: %s", strings.Join(missingFields, ","))
	}
	return nil
}

func (srv *server) getScheduledReport(logger log.FieldLogger, name, format string, w http.ResponseWriter, r *http.Request) {
	// Get the scheduledReport to make sure it's isn't failed
	report, err := srv.chargeback.informers.Chargeback().V1alpha1().ScheduledReports().Lister().ScheduledReports(srv.chargeback.cfg.Namespace).Get(name)
	if err != nil {
		logger.WithError(err).Errorf("error getting scheduledReport: %v", err)
		srv.writeErrorResponse(logger, w, r, http.StatusInternalServerError, "error getting scheduledReport: %v", err)
		return
	}

	if r.FormValue("ignore_failed") != "true" {
		if cond := cbutil.GetScheduledReportCondition(report.Status, api.ScheduledReportFailure); cond != nil && cond.Status == v1.ConditionTrue {
			logger.Errorf("scheduledReport is is failed state, reason: %s, message: %s", cond.Reason, cond.Message)
			srv.writeErrorResponse(logger, w, r, http.StatusInternalServerError, "scheduledReport is is failed state, reason: %s, message: %s", cond.Reason, cond.Message)
			return
		}
	}

	reportQuery, err := srv.chargeback.informers.Chargeback().V1alpha1().ReportGenerationQueries().Lister().ReportGenerationQueries(srv.chargeback.cfg.Namespace).Get(report.Spec.GenerationQueryName)
	if err != nil {
		logger.WithError(err).Errorf("error getting report: %v", err)
		srv.writeErrorResponse(logger, w, r, http.StatusInternalServerError, "error getting report: %v", err)
		return
	}

	prestoColumns := generatePrestoColumns(reportQuery)
	tableName := scheduledReportTableName(name)
	results, err := presto.GetRows(srv.chargeback.prestoConn, tableName, prestoColumns)
	if err != nil {
		logger.WithError(err).Errorf("failed to perform presto query")
		srv.writeErrorResponse(logger, w, r, http.StatusInternalServerError, "failed to perform presto query (see chargeback logs for more details): %v", err)
		return
	}

	// Get the presto table to get actual columns in table
	prestoTable, err := srv.chargeback.informers.Chargeback().V1alpha1().PrestoTables().Lister().PrestoTables(report.Namespace).Get(prestoTableResourceNameFromKind("scheduledreport", report.Name))
	if err != nil {
		logger.WithError(err).Errorf("error getting presto table: %v", err)
		srv.writeErrorResponse(logger, w, r, http.StatusInternalServerError, "error getting presto table: %v", err)
		return
	}

	if len(results) > 0 && len(prestoTable.State.CreationParameters.Columns) != len(results[0]) {
		logger.Errorf("report results schema doesn't match expected schema, got %d columns, expected %d", len(results[0]), len(prestoTable.State.CreationParameters.Columns))
		srv.writeErrorResponse(logger, w, r, http.StatusInternalServerError, "report results schema doesn't match expected schema")
		return
	}

	srv.writeResults(logger, format, reportQuery.Spec.Columns, results, w, r)
}
func (srv *server) getReport(logger log.FieldLogger, name, format string, useNewFormat bool, full bool, w http.ResponseWriter, r *http.Request) {
	// Get the current report to make sure it's in a ***REMOVED***nished state
	report, err := srv.chargeback.informers.Chargeback().V1alpha1().Reports().Lister().Reports(srv.chargeback.cfg.Namespace).Get(name)
	if err != nil {
		logger.WithError(err).Errorf("error getting report: %v", err)
		srv.writeErrorResponse(logger, w, r, http.StatusInternalServerError, "error getting report: %v", err)
		return
	}
	switch report.Status.Phase {
	case api.ReportPhaseError:
		err := fmt.Errorf(report.Status.Output)
		logger.WithError(err).Errorf("the report encountered an error")
		srv.writeErrorResponse(logger, w, r, http.StatusInternalServerError, "the report encountered an error: %v", err)
		return
	case api.ReportPhaseFinished:
		// continue with returning the report if the report is ***REMOVED***nished
	case api.ReportPhaseWaiting, api.ReportPhaseStarted:
		fallthrough
	default:
		logger.Errorf(ErrReportIsRunning.Error())
		srv.writeErrorResponse(logger, w, r, http.StatusAccepted, ErrReportIsRunning.Error())
		return
	}

	reportQuery, err := srv.chargeback.informers.Chargeback().V1alpha1().ReportGenerationQueries().Lister().ReportGenerationQueries(srv.chargeback.cfg.Namespace).Get(report.Spec.GenerationQueryName)
	if err != nil {
		logger.WithError(err).Errorf("error getting report: %v", err)
		srv.writeErrorResponse(logger, w, r, http.StatusInternalServerError, "error getting report: %v", err)
		return
	}

	prestoColumns := generatePrestoColumns(reportQuery)
	tableName := reportTableName(name)
	results, err := presto.GetRows(srv.chargeback.prestoConn, tableName, prestoColumns)
	if err != nil {
		logger.WithError(err).Errorf("failed to perform presto query")
		srv.writeErrorResponse(logger, w, r, http.StatusInternalServerError, "failed to perform presto query (see chargeback logs for more details): %v", err)
		return
	}

	// Get the presto table to get actual columns in table
	prestoTable, err := srv.chargeback.informers.Chargeback().V1alpha1().PrestoTables().Lister().PrestoTables(report.Namespace).Get(prestoTableResourceNameFromKind("report", report.Name))
	if err != nil {
		logger.WithError(err).Errorf("error getting presto table: %v", err)
		srv.writeErrorResponse(logger, w, r, http.StatusInternalServerError, "error getting presto table: %v", err)
		return
	}

	columns := prestoTable.State.CreationParameters.Columns
	if len(results) > 0 && len(columns) != len(results[0]) {
		logger.Errorf("report results schema doesn't match expected schema, got %d columns, expected %d", len(results[0]), len(prestoTable.State.CreationParameters.Columns))
		srv.writeErrorResponse(logger, w, r, http.StatusInternalServerError, "report results schema doesn't match expected schema")
		return
	}

	if useNewFormat {
		srv.writeResultsV2(logger, full, format, reportQuery.Spec.Columns, results, w, r)
	} ***REMOVED*** {
		srv.writeResults(logger, format, reportQuery.Spec.Columns, results, w, r)
	}
}
func (srv *server) writeResultsAsCSV(logger log.FieldLogger, columns []api.ReportGenerationQueryColumn, results []presto.Row, w http.ResponseWriter, r *http.Request) {
	buf := &bytes.Buffer{}
	csvWriter := csv.NewWriter(buf)

	// Write headers
	var keys []string
	if len(results) >= 1 {
		for _, column := range columns {
			keys = append(keys, column.Name)
		}
		err := csvWriter.Write(keys)
		if err != nil {
			logger.WithError(err).Errorf("failed to write headers")
			return
		}
	}

	// Write the rest
	for _, row := range results {
		vals := make([]string, len(keys))
		for i, key := range keys {
			val, ok := row[key]
			if !ok {
				logger.Errorf("report results schema doesn't match expected schema, unexpected key: %q", key)
				srv.writeErrorResponse(logger, w, r, http.StatusInternalServerError, "report results schema doesn't match expected schema, unexpected key: %q", key)
				return
			}
			switch v := val.(type) {
			case string:
				vals[i] = v
			case []byte:
				vals[i] = string(v)
			case uint, uint8, uint16, uint32, uint64,
				int, int8, int16, int32, int64,
				float32, float64,
				complex64, complex128,
				bool:
				vals[i] = fmt.Sprintf("%v", v)
			case time.Time:
				vals[i] = v.String()
			case nil:
				vals[i] = ""
			default:
				logger.Errorf("error marshalling csv: unknown type %t for value %v", val, val)
				srv.writeErrorResponse(logger, w, r, http.StatusInternalServerError, "error marshalling csv (see chargeback logs for more details)")
				return
			}
		}
		err := csvWriter.Write(vals)
		if err != nil {
			logger.Errorf("failed to write csv row: %v", err)
			return
		}
	}

	csvWriter.Flush()
	w.Header().Set("Content-Type", "text/csv")
	w.Write(buf.Bytes())
}

func (srv *server) writeResults(logger log.FieldLogger, format string, columns []api.ReportGenerationQueryColumn, results []presto.Row, w http.ResponseWriter, r *http.Request) {
	switch format {
	case "json":
		newResults := make([]*orderedmap.OrderedMap, len(results))
		for i, item := range results {
			var err error
			newResults[i], err = orderedmap.NewFromMap(item)
			if err != nil {
				srv.writeErrorResponse(logger, w, r, http.StatusInternalServerError, "error converting results: %v", err)
				return
			}
		}
		srv.writeResponseWithBody(logger, w, http.StatusOK, newResults)
		return
	case "csv":
		srv.writeResultsAsCSV(logger, columns, results, w, r)
	}
}

type GetReportResults struct {
	Results []ReportResultEntry `json:"results"`
}

type ReportResultEntry struct {
	Values []ReportResultValues `json:"values"`
}

type ReportResultValues struct {
	Name        string      `json:"name"`
	Value       interface{} `json:"value"`
	TableHidden bool        `json:"tableHidden"`
	Unit        string      `json:"unit,omitempty"`
}

// convertsToGetReportResults converts Rows returned from `presto.ExecuteSelect` into a GetReportResults
func convertsToGetReportResults(input []presto.Row, columns []api.ReportGenerationQueryColumn) GetReportResults {
	results := GetReportResults{}
	columnsMap := make(map[string]api.ReportGenerationQueryColumn)
	for _, column := range columns {
		columnsMap[column.Name] = column
	}
	for _, row := range input {
		var valSlice ReportResultEntry
		for columnName, columnValue := range row {
			resultsValue := ReportResultValues{
				Name:        columnName,
				Value:       columnValue,
				TableHidden: columnsMap[columnName].TableHidden,
				Unit:        columnsMap[columnName].Unit,
			}
			valSlice.Values = append(valSlice.Values, resultsValue)
		}
		results.Results = append(results.Results, valSlice)
	}
	return results
}

func (srv *server) writeResultsV2(logger log.FieldLogger, full bool, format string, columns []api.ReportGenerationQueryColumn, results []presto.Row, w http.ResponseWriter, r *http.Request) {
	columnsMap := make(map[string]api.ReportGenerationQueryColumn)
	var ***REMOVED***lteredColumns []api.ReportGenerationQueryColumn
	for _, column := range columns {
		columnsMap[column.Name] = column
		showColumn := !columnsMap[column.Name].TableHidden
		// Build a new list of columns if full is false, containing only columns with TableHidden set to false
		if showColumn || full {
			***REMOVED***lteredColumns = append(***REMOVED***lteredColumns, column)
		}
	}
	// Remove columns and their values from `input` if full is false and the column's TableHidden is true
	for _, row := range results {
		for _, column := range columnsMap {
			if columnsMap[column.Name].TableHidden && !full {
				delete(row, columnsMap[column.Name].Name)
			}
		}
	}

	switch format {
	case "json":
		srv.writeResponseWithBody(logger, w, http.StatusOK, convertsToGetReportResults(results, ***REMOVED***lteredColumns))
		return

	case "csv":
		srv.writeResultsAsCSV(logger, ***REMOVED***lteredColumns, results, w, r)
	}
}

func (srv *server) runReport(logger log.FieldLogger, query, start, end string, w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("method not yet implemented"))
}

type CollectPromsumDataRequest struct {
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
}

func (srv *server) collectPromsumDataHandler(w http.ResponseWriter, r *http.Request) {
	logger := srv.newLogger(r)

	decoder := json.NewDecoder(r.Body)
	var req CollectPromsumDataRequest
	err := decoder.Decode(&req)
	if err != nil {
		srv.writeErrorResponse(logger, w, r, http.StatusInternalServerError, "unable to decode response as JSON: %v", err)
		return
	}

	start := req.StartTime.UTC()
	end := req.EndTime.UTC()

	logger.Debugf("collecting promsum data between %s and %s", start.Format(time.RFC3339), end.Format(time.RFC3339))

	err = srv.chargeback.triggerPrometheusImporterForTimeRange(context.Background(), start, end)
	if err != nil {
		srv.writeErrorResponse(logger, w, r, http.StatusInternalServerError, "unable to collect prometheus data: %v", err)
		return
	}

	srv.writeResponseWithBody(logger, w, http.StatusOK, struct{}{})
}

type StorePromsumDataRequest []*prestostore.PrometheusMetric

func (srv *server) storePromsumDataHandler(w http.ResponseWriter, r *http.Request) {
	logger := srv.newLogger(r)

	name := chi.URLParam(r, "datasourceName")

	decoder := json.NewDecoder(r.Body)
	var req StorePromsumDataRequest
	err := decoder.Decode(&req)
	if err != nil {
		srv.writeErrorResponse(logger, w, r, http.StatusInternalServerError, "unable to decode response as JSON: %v", err)
		return
	}

	err = prestostore.StorePrometheusMetrics(context.Background(), srv.chargeback.prestoConn, dataSourceTableName(name), []*prestostore.PrometheusMetric(req))
	if err != nil {
		srv.writeErrorResponse(logger, w, r, http.StatusInternalServerError, "unable to store promsum metrics: %v", err)
		return
	}

	srv.writeResponseWithBody(logger, w, http.StatusOK, struct{}{})
}

func (srv *server) fetchPromsumDataHandler(w http.ResponseWriter, r *http.Request) {
	logger := srv.newLogger(r)

	name := chi.URLParam(r, "datasourceName")
	err := r.ParseForm()
	if err != nil {
		srv.writeErrorResponse(logger, w, r, http.StatusInternalServerError, "unable to decode body: %v", err)
		return
	}

	datasourceTable := dataSourceTableName(name)
	start := r.Form.Get("start")
	end := r.Form.Get("end")
	var startTime, endTime time.Time
	if start != "" {
		startTime, err = time.Parse(time.RFC3339, start)
		if err != nil {
			srv.writeErrorResponse(logger, w, r, http.StatusInternalServerError, "invalid start time parameter: %v", err)
			return
		}
	}
	if end != "" {
		endTime, err = time.Parse(time.RFC3339, end)
		if err != nil {
			srv.writeErrorResponse(logger, w, r, http.StatusInternalServerError, "invalid end time parameter: %v", err)
			return
		}
	}
	results, err := prestostore.GetPrometheusMetrics(srv.chargeback.prestoConn, datasourceTable, startTime, endTime)
	if err != nil {
		srv.writeErrorResponse(logger, w, r, http.StatusInternalServerError, "error querying for datasource: %v", err)
		return
	}

	srv.writeResponseWithBody(logger, w, http.StatusOK, results)
}
