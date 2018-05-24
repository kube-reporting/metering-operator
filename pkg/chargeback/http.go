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
	"github.com/operator-framework/operator-metering/pkg/db"
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

	writeToPrestoGroup singleflight.Group
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
	router.HandleFunc("/api/v2/reports/{name}/full", srv.getReportV2Handler)
	router.HandleFunc("/api/v1/scheduledreports/get", srv.getScheduledReportHandler)
	router.HandleFunc("/api/v1/reports/run", srv.runReportHandler)
	router.HandleFunc("/api/v1/datasources/prometheus/collect", srv.collectPromsumDataHandler)
	router.HandleFunc("/api/v1/datasources/prometheus/store/{datasourceName}", srv.storePromsumDataHandler)
	router.HandleFunc("/api/v1/datasources/prometheus/fetch/{datasourceName}", srv.fetchPromsumDataHandler)
	router.HandleFunc("/ready", srv.readinessHandler)
	router.HandleFunc("/healthy", srv.healthinessHandler)

	pprofMux.HandleFunc("/debug/pprof/", pprof.Index)
	pprofMux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	pprofMux.HandleFunc("/debug/pprof/profile", pprof.Profile)
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
	}).WithFields(srv.chargeback.newLogIdentifier())
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
	srv.getReport(logger, r.Form["name"][0], r.Form["format"][0], false, w, r)
}

func (srv *server) getReportV2Handler(w http.ResponseWriter, r *http.Request) {
	logger := srv.newLogger(r)
	name := chi.URLParam(r, "name")
	if !srv.validateGetReportReq(logger, []string{"format"}, w, r) {
		return
	}
	srv.getReport(logger, name, r.Form["format"][0], true, w, r)
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

func checkForFields(fields []string, vals url.Values) error {
	var missingFields []string
	for _, f := range fields {
		if len(vals[f]) == 0 || vals[f][0] == "" {
			missingFields = append(missingFields, f)
		}
	}
	if len(missingFields) != 0 {
		return fmt.Errorf("the following fields are missing or empty: %s", strings.Join(missingFields, ","))
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

	orderByColsStr := getReportOrderByColumnsString(reportQuery)
	reportTable := scheduledReportTableName(name)
	getReportQuery := fmt.Sprintf("SELECT * FROM %s ORDER BY %s", reportTable, orderByColsStr)
	results, err := presto.ExecuteSelect(srv.chargeback.prestoConn, getReportQuery)
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

	columns := prestoTable.State.CreationParameters.Columns
	if len(results) > 0 && len(columns) != len(results[0]) {
		logger.Errorf("report results schema doesn't match expected schema, got %d columns, expected %d", len(results[0]), len(prestoTable.State.CreationParameters.Columns))
		srv.writeErrorResponse(logger, w, r, http.StatusInternalServerError, "report results schema doesn't match expected schema")
		return
	}

	srv.writeResults(logger, format, columns, results, w, r)
}
func (srv *server) getReport(logger log.FieldLogger, name, format string, useNewFormat bool, w http.ResponseWriter, r *http.Request) {
	// Get the current report to make sure it's in a finished state
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
		// continue with returning the report if the report is finished
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

	orderByColsStr := getReportOrderByColumnsString(reportQuery)
	reportTable := reportTableName(name)
	getReportQuery := fmt.Sprintf("SELECT * FROM %s ORDER BY %s", reportTable, orderByColsStr)
	results, err := presto.ExecuteSelect(srv.chargeback.prestoConn, getReportQuery)
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
		srv.writeResultsV2(logger, format, columns, results, w, r)
	} else {
		srv.writeResults(logger, format, columns, results, w, r)
	}
}

func (srv *server) writeResults(logger log.FieldLogger, format string, columns []api.PrestoTableColumn, results []map[string]interface{}, w http.ResponseWriter, r *http.Request) {
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
}

type GetReportResults struct {
	Results []ReportResultEntry `json:"results"`
}

type ReportResultEntry struct {
	Values []ReportResultValues `json:"values"`
}

type ReportResultValues struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}

func convertsToGetReportResults(input []map[string]interface{}) GetReportResults {
	results := GetReportResults{}

	for _, row := range input {
		var valSlice ReportResultEntry
		for columnName, columnValue := range row {
			resultsValue := ReportResultValues{
				Name:  columnName,
				Value: columnValue,
			}
			valSlice.Values = append(valSlice.Values, resultsValue)
		}
		results.Results = append(results.Results, valSlice)
	}
	return results
}

func (srv *server) writeResultsV2(logger log.FieldLogger, format string, columns []api.PrestoTableColumn, results []map[string]interface{}, w http.ResponseWriter, r *http.Request) {
	switch format {
	case "json":
		srv.writeResponseWithBody(logger, w, http.StatusOK, convertsToGetReportResults(results))
		return
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

	logger.Debugf("collecting promsum data between %s and %s", req.StartTime.Format(time.RFC3339), req.EndTime.Format(time.RFC3339))

	timeBoundsGetter := promsumDataSourceTimeBoundsGetter(func(dataSource *api.ReportDataSource) (startTime, endTime time.Time, err error) {
		return req.StartTime.UTC(), req.EndTime.UTC(), nil
	})

	err = srv.chargeback.collectPromsumData(context.Background(), logger, timeBoundsGetter, -1, true)
	if err != nil {
		srv.writeErrorResponse(logger, w, r, http.StatusInternalServerError, "unable to collect prometheus data: %v", err)
		return
	}

	srv.writeResponseWithBody(logger, w, http.StatusOK, struct{}{})
}

type StorePromsumDataRequest []*PromsumRecord

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

	err = srv.chargeback.promsumStoreRecords(context.Background(), logger, dataSourceTableName(name), []*PromsumRecord(req))
	if err != nil {
		srv.writeErrorResponse(logger, w, r, http.StatusInternalServerError, "unable to store promsum records: %v", err)
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

	results, err := queryPromsumDatasource(logger, srv.chargeback.prestoConn, datasourceTable, startTime, endTime)
	if err != nil {
		srv.writeErrorResponse(logger, w, r, http.StatusInternalServerError, "error querying for datasource: %v", err)
		return
	}

	srv.writeResponseWithBody(logger, w, http.StatusOK, results)
}

func queryPromsumDatasource(logger log.FieldLogger, queryer db.Queryer, datasourceTable string, start, end time.Time) ([]*PromsumRecord, error) {
	whereClause := ""
	if !start.IsZero() {
		whereClause += fmt.Sprintf(`WHERE "timestamp" >= timestamp '%s' `, prestoTimestamp(start))
	}
	if !end.IsZero() {
		if !start.IsZero() {
			whereClause += " AND "
		} else {
			whereClause += " WHERE "
		}
		whereClause += fmt.Sprintf(`"timestamp" <= timestamp '%s'`, prestoTimestamp(end))
	}

	// we use map_entries for ordering on the labels because maps are
	// unorderable in Presto.
	query := fmt.Sprintf(`SELECT labels, amount, timeprecision, "timestamp" FROM %s %s ORDER BY "timestamp", map_entries(labels), amount, timeprecision ASC`, datasourceTable, whereClause)
	rows, err := queryer.Query(query)
	if err != nil {
		return nil, err
	}

	var results []*PromsumRecord
	for rows.Next() {
		var dbRecord promsumDBRecord
		if err := rows.Scan(&dbRecord.Labels, &dbRecord.Amount, &dbRecord.TimePrecision, &dbRecord.Timestamp); err != nil {
			return nil, err
		}
		labels := make(map[string]string)
		for key, value := range dbRecord.Labels {
			var ok bool
			labels[key], ok = value.(string)
			if !ok {
				logger.Errorf("invalid label %s, valueType: %T, value: %+v", key, value, value)
			}
		}
		record := PromsumRecord{
			Labels:    labels,
			Amount:    dbRecord.Amount,
			StepSize:  time.Duration(dbRecord.TimePrecision) * time.Second,
			Timestamp: dbRecord.Timestamp,
		}
		results = append(results, &record)
	}
	return results, nil
}

type promsumDBRecord struct {
	Labels        map[string]interface{}
	Amount        float64
	TimePrecision float64
	Timestamp     time.Time
}

func getReportOrderByColumnsString(query *api.ReportGenerationQuery) string {
	cols := query.Spec.Columns
	var quotedColumns []string
	for _, col := range cols {
		colName := col.Name
		// if we detect a map(...) in the column, use map_entries to do
		// ordering. we detect a map column using a best effort approach by
		// checking if the column type contains the string "map(" , and is
		// followed by a ")" after that string.
		if mapIndex := strings.Index(col.Type, "map("); mapIndex != -1 && strings.Index(col.Type, ")") > mapIndex {
			quotedColumns = append(quotedColumns, fmt.Sprintf(`map_entries("%s")`, colName))
		} else {
			quotedColumns = append(quotedColumns, fmt.Sprintf(`"%s"`, colName))
		}
	}
	return fmt.Sprintf("period_start, period_end, %s ASC", strings.Join(quotedColumns, ", "))
}
