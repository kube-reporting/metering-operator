package operator

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"

	metering "github.com/kubernetes-reporting/metering-operator/pkg/apis/metering/v1"
	meteringUtil "github.com/kubernetes-reporting/metering-operator/pkg/apis/metering/v1/util"
	listers "github.com/kubernetes-reporting/metering-operator/pkg/generated/listers/metering/v1"
	"github.com/kubernetes-reporting/metering-operator/pkg/operator/prestostore"
	"github.com/kubernetes-reporting/metering-operator/pkg/operator/reporting"
	"github.com/kubernetes-reporting/metering-operator/pkg/operator/reportingutil"
	"github.com/kubernetes-reporting/metering-operator/pkg/presto"
	"github.com/kubernetes-reporting/metering-operator/pkg/util/chiprometheus"
	"github.com/kubernetes-reporting/metering-operator/pkg/util/orderedmap"
)

var ErrReportIsRunning = errors.New("the report is still running")
var prometheusMiddleware = chiprometheus.NewMiddleware("reporting-operator")

const (
	APIV1ReportGetEndpoint         = "/api/v1/reports/get"
	APIV2ReportEndpointPrefix      = "/api/v2/reports"
	APIV2ReportQueryEndpointPrefix = "/api/v2/reportqueries"
)

type server struct {
	logger log.FieldLogger

	rand          *rand.Rand
	collectorFunc prometheusImporterFunc

	prometheusMetricsRepo prestostore.PrometheusMetricsRepo
	reportResultsGetter   prestostore.ReportResultsGetter
	dependencyResolver    DependencyResolver

	reportLister           listers.ReportLister
	reportDataSourceLister listers.ReportDataSourceLister
	reportQueryLister      listers.ReportQueryLister
	prestoTableLister      listers.PrestoTableLister
}

type requestLogger struct {
	log.FieldLogger
}

func (l *requestLogger) Print(v ...interface{}) {
	l.FieldLogger.Info(v...)
}

func newRouter(
	logger log.FieldLogger,
	rand *rand.Rand,
	prometheusMetricsRepo prestostore.PrometheusMetricsRepo,
	reportResultsGetter prestostore.ReportResultsGetter,
	depResolver DependencyResolver,
	collectorFunc prometheusImporterFunc,
	reportLister listers.ReportLister,
	reportDataSourceLister listers.ReportDataSourceLister,
	reportQueryLister listers.ReportQueryLister,
	prestoTableLister listers.PrestoTableLister,
) chi.Router {
	router := chi.NewRouter()
	logger = logger.WithField("component", "api")
	requestLogger := middleware.RequestLogger(&middleware.DefaultLogFormatter{Logger: &requestLogger{logger}})
	router.Use(requestLogger)
	router.Use(prometheusMiddleware)

	srv := &server{
		logger:                 logger,
		rand:                   rand,
		collectorFunc:          collectorFunc,
		prometheusMetricsRepo:  prometheusMetricsRepo,
		reportResultsGetter:    reportResultsGetter,
		dependencyResolver:     depResolver,
		reportLister:           reportLister,
		reportDataSourceLister: reportDataSourceLister,
		reportQueryLister:      reportQueryLister,
		prestoTableLister:      prestoTableLister,
	}

	router.HandleFunc(APIV2ReportEndpointPrefix+"/{namespace}/{name}/full", srv.getReportV2FullHandler)
	router.HandleFunc(APIV2ReportEndpointPrefix+"/{namespace}/{name}/table", srv.getReportV2TableHandler)
	router.HandleFunc(APIV2ReportQueryEndpointPrefix+"/{namespace}/{name}/render", srv.renderReportQueryV2Handler)
	router.HandleFunc(APIV1ReportGetEndpoint, srv.getReportV1Handler)
	router.HandleFunc("/api/v1/datasources/prometheus/collect/{namespace}", srv.collectPrometheusMetricsDataHandler)
	router.HandleFunc("/api/v1/datasources/prometheus/collect/{namespace}/{datasourceName}", srv.collectPrometheusMetricsDataHandler)
	router.HandleFunc("/api/v1/datasources/prometheus/store/{namespace}/{datasourceName}", srv.storePrometheusMetricsDataHandler)
	router.HandleFunc("/api/v1/datasources/prometheus/fetch/{namespace}/{datasourceName}", srv.fetchPrometheusMetricsDataHandler)

	return router
}

func (srv *server) validateGetReportReq(logger log.FieldLogger, requiredQueryParams []string, w http.ResponseWriter, r *http.Request) bool {
	if r.Method != "GET" {
		writeErrorResponse(logger, w, r, http.StatusNotFound, "Not found")
		return false
	}
	err := r.ParseForm()
	if err != nil {
		writeErrorResponse(logger, w, r, http.StatusBadRequest, "couldn't parse URL query params: %v", err)
		return false
	}
	err = checkForFields(requiredQueryParams, r.Form)
	if err != nil {
		writeErrorResponse(logger, w, r, http.StatusBadRequest, "%v", err)
		return false
	}
	format := r.Form["format"][0]
	switch format {
	case "json", "csv", "tab", "tabular":
		return true
	}
	writeErrorResponse(logger, w, r, http.StatusBadRequest, "format must be one of: csv, json or tabular")
	return false
}

func (srv *server) getReportV1Handler(w http.ResponseWriter, r *http.Request) {
	logger := newRequestLogger(srv.logger, r, srv.rand)
	if !srv.validateGetReportReq(logger, []string{"name", "namespace", "format"}, w, r) {
		return
	}
	srv.getReport(logger, r.Form["name"][0], r.Form["namespace"][0], r.Form["format"][0], false, true, w, r)
}

func (srv *server) getReportV2FullHandler(w http.ResponseWriter, r *http.Request) {
	logger := newRequestLogger(srv.logger, r, srv.rand)
	name := chi.URLParam(r, "name")
	namespace := chi.URLParam(r, "namespace")
	if name == "" {
		writeErrorResponse(logger, w, r, http.StatusBadRequest, "the following fields are missing or empty: name")
		return
	}
	if !srv.validateGetReportReq(logger, []string{"format"}, w, r) {
		return
	}
	srv.getReport(logger, name, namespace, r.Form["format"][0], true, true, w, r)
}

func (srv *server) getReportV2TableHandler(w http.ResponseWriter, r *http.Request) {
	logger := newRequestLogger(srv.logger, r, srv.rand)
	name := chi.URLParam(r, "name")
	namespace := chi.URLParam(r, "namespace")
	if name == "" {
		writeErrorResponse(logger, w, r, http.StatusBadRequest, "the following fields are missing or empty: name")
		return
	}
	if !srv.validateGetReportReq(logger, []string{"format"}, w, r) {
		return
	}
	srv.getReport(logger, name, namespace, r.Form["format"][0], true, false, w, r)
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

func (srv *server) getReport(logger log.FieldLogger, name, namespace, format string, useNewFormat bool, full bool, w http.ResponseWriter, r *http.Request) {
	// Get the report to make sure it hasn't failed
	report, err := srv.reportLister.Reports(namespace).Get(name)
	if err != nil {
		code := http.StatusInternalServerError
		if k8serrors.IsNotFound(err) {
			code = http.StatusNotFound
		}
		logger.WithError(err).Errorf("error getting report: %v", err)
		writeErrorResponse(logger, w, r, code, "error getting report: %v", err)
		return
	}

	if r.FormValue("ignore_failed") != "true" {
		if cond := meteringUtil.GetReportCondition(report.Status, metering.ReportRunning); cond != nil && cond.Status == v1.ConditionFalse && cond.Reason == meteringUtil.GenerateReportFailedReason {
			logger.Errorf("report is is failed state, reason: %s, message: %s", cond.Reason, cond.Message)
			writeErrorResponse(logger, w, r, http.StatusInternalServerError, "report is is failed state, reason: %s, message: %s", cond.Reason, cond.Message)
			return
		}
	}

	reportQuery, err := srv.reportQueryLister.ReportQueries(report.Namespace).Get(report.Spec.QueryName)
	if err != nil {
		logger.WithError(err).Errorf("error getting reportQuery: %v", err)
		writeErrorResponse(logger, w, r, http.StatusInternalServerError, "error getting reportQuery: %v", err)
		return
	}

	// Get the presto table to get actual columns in table
	prestoTable, err := srv.prestoTableLister.PrestoTables(report.Namespace).Get(reportingutil.TableResourceNameFromKind("report", report.Namespace, report.Name))
	if err != nil {
		if k8serrors.IsNotFound(err) {
			writeErrorResponse(logger, w, r, http.StatusAccepted, "Report is not processed yet")
			return
		}
		logger.WithError(err).Errorf("error getting presto table: %v", err)
		writeErrorResponse(logger, w, r, http.StatusInternalServerError, "error getting presto table: %v", err)
		return
	}

	queryPrestoColumns := reportingutil.GeneratePrestoColumns(reportQuery)
	prestoColumns := prestoTable.Status.Columns

	if len(prestoColumns) == 0 {
		logger.WithError(err).Errorf("PrestoTable %s has 0 columns", prestoTable.Name)
		writeErrorResponse(logger, w, r, http.StatusInternalServerError, "PrestoTable %s has 0 columns", prestoTable.Name)
		return
	}

	if !reflect.DeepEqual(queryPrestoColumns, prestoColumns) {
		logger.Warnf("report columns and table columns don't match, ReportQuery was likely updated after the report ran")
		logger.Debugf("mismatched columns, PrestoTable columns: %v, ReportQuery columns: %v", prestoColumns, queryPrestoColumns)
	}

	tableName, err := reportingutil.FullyQualifiedTableName(prestoTable)
	if err != nil {
		logger.WithError(err).Errorf("prestoTable contains invalid Status fields")
		writeErrorResponse(logger, w, r, http.StatusInternalServerError, "invalid prestoTable.Status fields: %v", err)
		return
	}
	results, err := srv.reportResultsGetter.GetReportResults(tableName, prestoColumns)
	if err != nil {
		logger.WithError(err).Errorf("failed to perform presto query")
		writeErrorResponse(logger, w, r, http.StatusInternalServerError, "failed to perform presto query (see operator logs for more details): %v", err)
		return
	}

	if len(results) > 0 && len(prestoTable.Status.Columns) != len(results[0]) {
		logger.Errorf("report results schema doesn't match expected schema, got %d columns, expected %d", len(results[0]), len(prestoTable.Status.Columns))
		writeErrorResponse(logger, w, r, http.StatusInternalServerError, "report results schema doesn't match expected schema")
		return
	}

	if useNewFormat {
		writeResultsResponseV2(logger, full, format, reportQuery.Name, reportQuery.Spec.Columns, results, w, r)
	} else {
		writeResultsResponseV1(logger, format, reportQuery.Name, reportQuery.Spec.Columns, results, w, r)
	}
}

func writeResultsResponseAsCSV(logger log.FieldLogger, name string, columns []metering.ReportQueryColumn, results []presto.Row, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=%s.csv", name))
	err := writeResultsAsCSV(columns, results, w, ',')
	if err != nil {
		writeErrorResponse(logger, w, r, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
}

func writeResultsAsCSV(columns []metering.ReportQueryColumn, results []presto.Row, w io.Writer, delimiter rune) error {
	csvWriter := csv.NewWriter(w)
	csvWriter.Comma = delimiter

	// Write headers
	var keys []string
	if len(results) >= 1 {
		for _, column := range columns {
			keys = append(keys, column.Name)
		}
		err := csvWriter.Write(keys)
		if err != nil {
			return err
		}
	}

	// Write the rest
	for _, row := range results {
		vals := make([]string, len(keys))
		for i, key := range keys {
			val, ok := row[key]
			if !ok {
				return fmt.Errorf("report results schema doesn't match expected schema, unexpected key: %q", key)
			}
			switch v := val.(type) {
			case string:
				vals[i] = v
			case []byte:
				vals[i] = string(v)
			case uint, uint8, uint16, uint32, uint64, int, int8, int16, int32, int64:
				vals[i] = fmt.Sprintf("%d", v)
			case float32, float64, complex64, complex128:
				vals[i] = fmt.Sprintf("%f", v)
			case bool:
				vals[i] = fmt.Sprintf("%t", v)
			case time.Time:
				vals[i] = v.String()
			case nil:
				vals[i] = ""
			default:
				return fmt.Errorf("error marshalling csv: unknown type %t for value %v", val, val)
			}
		}
		err := csvWriter.Write(vals)
		if err != nil {
			return err
		}
	}

	csvWriter.Flush()
	return csvWriter.Error()
}

func writeResultsResponseAsTabular(logger log.FieldLogger, name string, columns []metering.ReportQueryColumn, results []presto.Row, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/tab-separated-values")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=%s.tsv", name))
	padding := 2
	paddingStr := r.FormValue("padding")
	if paddingStr != "" {
		var err error
		padding, err = strconv.Atoi(paddingStr)
		if err != nil {
			writeErrorResponse(logger, w, r, http.StatusInternalServerError, "invalid padding value %s, err: %s", paddingStr, err)
			return
		}
	}
	tabWriter := tabwriter.NewWriter(w, 0, 8, padding, '\t', 0)
	err := writeResultsAsCSV(columns, results, tabWriter, '\t')
	if err != nil {
		writeErrorResponse(logger, w, r, http.StatusInternalServerError, err.Error())
		return
	}
	err = tabWriter.Flush()
	if err != nil {
		writeErrorResponse(logger, w, r, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
}

func writeResultsResponseAsJSON(logger log.FieldLogger, name string, columns []metering.ReportQueryColumn, results []presto.Row, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=%s.json", name))
	newResults := make([]*orderedmap.OrderedMap, len(results))
	for i, item := range results {
		var err error
		newResults[i], err = orderedmap.NewFromMap(item)
		if err != nil {
			writeErrorResponse(logger, w, r, http.StatusInternalServerError, "error converting results: %v", err)
			return
		}
	}
	writeResponseAsJSON(logger, w, http.StatusOK, newResults)
}

func writeResultsResponse(logger log.FieldLogger, format, name string, columns []metering.ReportQueryColumn, results []presto.Row, w http.ResponseWriter, r *http.Request) {
	switch format {
	case "json":
		writeResultsResponseAsJSON(logger, name, columns, results, w, r)
	case "csv":
		writeResultsResponseAsCSV(logger, name, columns, results, w, r)
	case "tab", "tabular":
		writeResultsResponseAsTabular(logger, name, columns, results, w, r)
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
func convertsToGetReportResults(input []presto.Row, columns []metering.ReportQueryColumn) GetReportResults {
	results := GetReportResults{}
	columnsMap := make(map[string]metering.ReportQueryColumn)
	for _, column := range columns {
		columnsMap[column.Name] = column
	}
	for _, row := range input {
		var valSlice ReportResultEntry
		// iterate by columns to ensure consistent ordering of values
		for _, column := range columns {
			columnName := column.Name
			columnValue := row[columnName]
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

func writeResultsResponseV1(logger log.FieldLogger, format string, name string, columns []metering.ReportQueryColumn, results []presto.Row, w http.ResponseWriter, r *http.Request) {
	columnsMap := make(map[string]metering.ReportQueryColumn)
	var filteredColumns []metering.ReportQueryColumn

	// remove tableHidden columns and their values if the format is tabular or CSV

	// filter columns
	for _, column := range columns {
		columnsMap[column.Name] = column
		showColumn := !columnsMap[column.Name].TableHidden
		if showColumn {
			filteredColumns = append(filteredColumns, column)
		}
	}

	// filter rows
	for _, row := range results {
		for _, column := range columnsMap {
			if columnsMap[column.Name].TableHidden {
				delete(row, columnsMap[column.Name].Name)
			}
		}
	}

	writeResultsResponse(logger, format, name, filteredColumns, results, w, r)
}

func writeResultsResponseV2(logger log.FieldLogger, full bool, format string, name string, columns []metering.ReportQueryColumn, results []presto.Row, w http.ResponseWriter, r *http.Request) {
	format = strings.ToLower(format)
	isTableFormat := format == "csv" || format == "tab" || format == "tabular"
	columnsMap := make(map[string]metering.ReportQueryColumn)
	var filteredColumns []metering.ReportQueryColumn

	// Remove columns and their values from `results` if full is false and the
	// column's TableHidden is true or if TableHidden is true and we're
	// outputting tabular or CSV

	// filter the columns
	for _, column := range columns {
		columnsMap[column.Name] = column
		tableHidden := columnsMap[column.Name].TableHidden
		// skip using columns if tableHidden is true and we're outputing to
		// csv/tabular
		if tableHidden && (isTableFormat || !full) {
			continue
		}
		filteredColumns = append(filteredColumns, column)
	}

	// filter the rows
	for _, row := range results {
		for _, column := range columnsMap {
			tableHidden := columnsMap[column.Name].TableHidden
			if tableHidden && (isTableFormat || !full) {
				delete(row, columnsMap[column.Name].Name)
			}
		}
	}

	if format == "json" {
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=%s.json", name))
		writeResponseAsJSON(logger, w, http.StatusOK, convertsToGetReportResults(results, filteredColumns))
		return
	}

	writeResultsResponse(logger, format, name, filteredColumns, results, w, r)
}

func (srv *server) runReport(logger log.FieldLogger, query, start, end string, w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("method not yet implemented"))
}

type CollectPrometheusMetricsDataRequest struct {
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
}

type CollectPrometheusMetricsDataResponse struct {
	Results []*prometheusImportResults `json:"results"`
}

func (srv *server) collectPrometheusMetricsDataHandler(w http.ResponseWriter, r *http.Request) {
	logger := newRequestLogger(srv.logger, r, srv.rand)

	namespace := chi.URLParam(r, "namespace")
	dsName := chi.URLParam(r, "datasource")

	decoder := json.NewDecoder(r.Body)
	var req CollectPrometheusMetricsDataRequest
	err := decoder.Decode(&req)
	if err != nil {
		writeErrorResponse(logger, w, r, http.StatusInternalServerError, "unable to decode request as JSON: %v", err)
		return
	}

	start := req.StartTime.UTC()
	end := req.EndTime.UTC()

	logger.Debugf("collecting prometheus data for ReportDataSources in namespace %s between %s and %s", namespace, start.Format(time.RFC3339), end.Format(time.RFC3339))

	results, err := srv.collectorFunc(context.Background(), namespace, dsName, start, end)
	if err != nil {
		writeErrorResponse(logger, w, r, http.StatusInternalServerError, "unable to collect prometheus data: %v", err)
		return
	}

	writeResponseAsJSON(logger, w, http.StatusOK, CollectPrometheusMetricsDataResponse{
		Results: results,
	})
}

type StorePrometheusMetricsDataRequest []*prestostore.PrometheusMetric

func (srv *server) storePrometheusMetricsDataHandler(w http.ResponseWriter, r *http.Request) {
	logger := newRequestLogger(srv.logger, r, srv.rand)

	name := chi.URLParam(r, "datasourceName")
	namespace := chi.URLParam(r, "namespace")

	decoder := json.NewDecoder(r.Body)

	// read opening bracket
	_, err := decoder.Token()
	if err != nil {
		writeErrorResponse(logger, w, r, http.StatusInternalServerError, "unable to decode request as JSON: %v", err)
		return
	}

	var metrics []*prestostore.PrometheusMetric
	// while the array contains values
	for decoder.More() {
		var m prestostore.PrometheusMetric
		err = decoder.Decode(&m)
		if err != nil {
			writeErrorResponse(logger, w, r, http.StatusInternalServerError, "unable to decode request as JSON: %v", err)
			return
		}
		metrics = append(metrics, &m)
	}

	// read closing bracket
	_, err = decoder.Token()
	if err != nil {
		writeErrorResponse(logger, w, r, http.StatusInternalServerError, "unable to decode request as JSON: %v", err)
		return
	}

	dataSource, err := srv.reportDataSourceLister.ReportDataSources(namespace).Get(name)
	if err != nil {
		logger.WithError(err).Errorf("unable to get ReportDataSource %s: %v", name, err)
		writeErrorResponse(logger, w, r, http.StatusInternalServerError, "unable to get ReportDataSource %s: %v", name, err)
		return
	}
	if dataSource.Status.TableRef.Name == "" {
		logger.WithError(err).Errorf("ReportDataSource %s table not created yet", name)
		writeErrorResponse(logger, w, r, http.StatusInternalServerError, "ReportDataSource %s table not created yet", name)
		return
	}

	prestoTable, err := srv.prestoTableLister.PrestoTables(dataSource.Namespace).Get(dataSource.Status.TableRef.Name)
	if err != nil {
		logger.WithError(err).Errorf("unable to get PrestoTable %s: %v", dataSource.Status.TableRef.Name, err)
		writeErrorResponse(logger, w, r, http.StatusInternalServerError, "unable to get PrestoTable %s: %v", dataSource.Status.TableRef.Name, err)
		return
	}
	if prestoTable.Status.TableName == "" {
		logger.WithError(err).Errorf("PrestoTable %s table %s not created yet", prestoTable.Name, prestoTable.Spec.TableName)
		writeErrorResponse(logger, w, r, http.StatusInternalServerError, "PrestoTable %s table %s not created yet", prestoTable.Name, prestoTable.Spec.TableName)
		return
	}

	tableName, err := reportingutil.FullyQualifiedTableName(prestoTable)
	if err != nil {
		logger.WithError(err).Errorf("invalid PrestoTable %s: %v", prestoTable.Name, err)
		writeErrorResponse(logger, w, r, http.StatusInternalServerError, "invalid PrestoTable %s: %v", prestoTable.Name, err)
		return
	}

	err = srv.prometheusMetricsRepo.StorePrometheusMetrics(context.Background(), tableName, metrics)
	if err != nil {
		logger.WithError(err).Errorf("unable to store prometheus metrics: %v", err)
		writeErrorResponse(logger, w, r, http.StatusInternalServerError, "unable to store prometheus metrics: %v", err)
		return
	}

	writeResponseAsJSON(logger, w, http.StatusOK, struct{}{})
}

func (srv *server) fetchPrometheusMetricsDataHandler(w http.ResponseWriter, r *http.Request) {
	logger := newRequestLogger(srv.logger, r, srv.rand)

	name := chi.URLParam(r, "datasourceName")
	namespace := chi.URLParam(r, "namespace")
	err := r.ParseForm()
	if err != nil {
		writeErrorResponse(logger, w, r, http.StatusInternalServerError, "unable to decode body: %v", err)
		return
	}

	start := r.Form.Get("start")
	end := r.Form.Get("end")
	var startTime, endTime time.Time
	if start != "" {
		startTime, err = time.Parse(time.RFC3339, start)
		if err != nil {
			writeErrorResponse(logger, w, r, http.StatusInternalServerError, "invalid start time parameter: %v", err)
			return
		}
	}
	if end != "" {
		endTime, err = time.Parse(time.RFC3339, end)
		if err != nil {
			writeErrorResponse(logger, w, r, http.StatusInternalServerError, "invalid end time parameter: %v", err)
			return
		}
	}

	dataSource, err := srv.reportDataSourceLister.ReportDataSources(namespace).Get(name)
	if err != nil {
		logger.WithError(err).Errorf("unable to get ReportDataSource %s: %v", name, err)
		writeErrorResponse(logger, w, r, http.StatusInternalServerError, "unable to get ReportDataSource %s: %v", name, err)
		return
	}
	if dataSource.Status.TableRef.Name == "" {
		logger.WithError(err).Errorf("ReportDataSource %s table not created yet", name)
		writeErrorResponse(logger, w, r, http.StatusInternalServerError, "ReportDataSource %s table not created yet", name)
		return
	}

	prestoTable, err := srv.prestoTableLister.PrestoTables(dataSource.Namespace).Get(dataSource.Status.TableRef.Name)
	if err != nil {
		logger.WithError(err).Errorf("unable to get PrestoTable %s: %v", dataSource.Status.TableRef.Name, err)
		writeErrorResponse(logger, w, r, http.StatusInternalServerError, "unable to get PrestoTable %s: %v", dataSource.Status.TableRef.Name, err)
		return
	}
	if prestoTable.Status.TableName == "" {
		logger.WithError(err).Errorf("PrestoTable %s table %s not created yet", prestoTable.Name, prestoTable.Spec.TableName)
		writeErrorResponse(logger, w, r, http.StatusInternalServerError, "PrestoTable %s table %s not created yet", prestoTable.Name, prestoTable.Spec.TableName)
		return
	}
	tableName, err := reportingutil.FullyQualifiedTableName(prestoTable)
	if err != nil {
		logger.WithError(err).Errorf("invalid PrestoTable %s: %v", prestoTable.Name, err)
		writeErrorResponse(logger, w, r, http.StatusInternalServerError, "invalid PrestoTable %s: %v", prestoTable.Name, err)
		return
	}

	results, err := srv.prometheusMetricsRepo.GetPrometheusMetrics(tableName, startTime, endTime)
	if err != nil {
		logger.WithError(err).Errorf("error querying for datasource: %v", err)
		writeErrorResponse(logger, w, r, http.StatusInternalServerError, "error querying for datasource: %v", err)
		return
	}

	writeResponseAsJSON(logger, w, http.StatusOK, results)
}

type RenderReportQueryRequest struct {
	Inputs metering.ReportQueryInputValues `json:"inputs,omitempty"`
	Start  time.Time                       `json:"start,omitempty"`
	End    time.Time                       `json:"end,omitempty"`
}

func (srv *server) renderReportQueryV2Handler(w http.ResponseWriter, r *http.Request) {
	logger := newRequestLogger(srv.logger, r, srv.rand)

	name := chi.URLParam(r, "name")
	namespace := chi.URLParam(r, "namespace")
	err := r.ParseForm()
	if err != nil {
		writeErrorResponse(logger, w, r, http.StatusInternalServerError, "unable to decode body: %v", err)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var req RenderReportQueryRequest
	err = decoder.Decode(&req)
	if err != nil {
		writeErrorResponse(logger, w, r, http.StatusInternalServerError, "unable to decode request as JSON: %v", err)
		return
	}

	reportQuery, err := srv.reportQueryLister.ReportQueries(namespace).Get(name)
	if err != nil {
		logger.WithError(err).Errorf("error getting reportQuery: %v", err)
		writeErrorResponse(logger, w, r, http.StatusInternalServerError, "error getting reportQuery: %v", err)
		return
	}

	deps, err := srv.dependencyResolver.ResolveDependencies(namespace, reportQuery.Spec.Inputs, req.Inputs)
	if err != nil {
		logger.WithError(err).Errorf("error resolving reportQuery dependencies: %v", err)
		writeErrorResponse(logger, w, r, http.StatusInternalServerError, "error resolving reportQuery dependencies: %v", err)
		return

	}

	prestoTables, err := srv.prestoTableLister.PrestoTables(namespace).List(labels.Everything())
	if err != nil {
		logger.WithError(err).Errorf("error getting resources to render reportQuery: %v", err)
		writeErrorResponse(logger, w, r, http.StatusInternalServerError, "error getting resources to render reportQuery: %v", err)
		return

	}

	reports, err := srv.reportLister.Reports(namespace).List(labels.Everything())
	if err != nil {
		logger.WithError(err).Errorf("error getting resources to render reportQuery: %v", err)
		writeErrorResponse(logger, w, r, http.StatusInternalServerError, "error getting resources to render reportQuery: %v", err)
		return

	}

	datasources, err := srv.reportDataSourceLister.ReportDataSources(namespace).List(labels.Everything())
	if err != nil {
		logger.WithError(err).Errorf("error getting resources to render reportQuery: %v", err)
		writeErrorResponse(logger, w, r, http.StatusInternalServerError, "error getting resources to render reportQuery: %v", err)
		return

	}

	queries, err := srv.reportQueryLister.ReportQueries(namespace).List(labels.Everything())
	if err != nil {
		logger.WithError(err).Errorf("error getting resources to render reportQuery: %v", err)
		writeErrorResponse(logger, w, r, http.StatusInternalServerError, "error getting resources to render reportQuery: %v", err)
		return

	}

	requiredInputs := reportingutil.ConvertInputDefinitionsIntoInputList(reportQuery.Spec.Inputs)
	queryCtx := &reporting.ReportQueryTemplateContext{
		Namespace:         namespace,
		Query:             reportQuery.Spec.Query,
		RequiredInputs:    requiredInputs,
		Reports:           reports,
		ReportQueries:     queries,
		ReportDataSources: datasources,
		PrestoTables:      prestoTables,
	}
	tmplCtx := reporting.TemplateContext{
		Report: reporting.ReportTemplateInfo{
			ReportingStart: &req.Start,
			ReportingEnd:   &req.End,
			Inputs:         deps.InputValues,
		},
	}

	// Render the query template
	query, err := reporting.RenderQuery(queryCtx, tmplCtx)
	if err != nil {
		logger.WithError(err).Errorf("error rendering ReportQuery: %v", err)
		writeErrorResponse(logger, w, r, http.StatusInternalServerError, "error rendering ReportQuery: %v", err)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	if _, err = fmt.Fprint(w, query); err != nil {
		logger.WithError(err).Error("failed writing HTTP response")
	}
}
