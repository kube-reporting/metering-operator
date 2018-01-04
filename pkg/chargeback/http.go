package chargeback

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	api "github.com/coreos-inc/kube-chargeback/pkg/apis/chargeback/v1alpha1"
	"github.com/coreos-inc/kube-chargeback/pkg/presto"
)

var ErrReportIsRunning = errors.New("the report is still running")

type server struct {
	chargeback *Chargeback
	logger     log.FieldLogger
	httpServer *http.Server
}

func newServer(c *Chargeback, logger log.FieldLogger) *server {
	logger = logger.WithField("component", "api")
	mux := http.NewServeMux()
	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	srv := &server{
		chargeback: c,
		logger:     logger,
		httpServer: httpServer,
	}
	mux.HandleFunc("/api/v1/reports/get", srv.getReportHandler)
	mux.HandleFunc("/api/v1/reports/run", srv.runReportHandler)
	mux.HandleFunc("/ready", srv.readinessHandler)
	return srv
}

func (srv *server) start() {
	srv.logger.Infof("HTTP server started")
	srv.logger.WithError(srv.httpServer.ListenAndServe()).Info("HTTP server exited")
}

func (srv *server) stop() error {
	return srv.httpServer.Shutdown(context.TODO())
}

func (srv *server) newLogger(r *http.Request) log.FieldLogger {
	return srv.logger.WithFields(log.Fields{
		"method": r.Method,
		"url":    r.URL.String(),
	}).WithFields(newLogIdenti***REMOVED***er())
}

func (srv *server) logRequest(logger log.FieldLogger, r *http.Request) {
	logger.Infof("%s %s", r.Method, r.URL.String())
}

type reportErrorResponse struct {
	Error string `json:"error"`
}

func (srv *server) reportError(logger log.FieldLogger, w http.ResponseWriter, r *http.Request, status int, message string, args ...interface{}) {
	msg := fmt.Sprintf(message, args...)
	srv.writeResponseWithBody(logger, w, status, reportErrorResponse{Error: msg})
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

func (srv *server) getReportHandler(w http.ResponseWriter, r *http.Request) {
	logger := srv.newLogger(r)
	srv.logRequest(logger, r)
	if r.Method != "GET" {
		srv.reportError(logger, w, r, http.StatusNotFound, "Not found")
		return
	}
	err := r.ParseForm()
	if err != nil {
		srv.reportError(logger, w, r, http.StatusBadRequest, "couldn't parse URL query params: %v", err)
		return
	}
	vals := r.Form
	err = checkForFields([]string{"name", "format"}, vals)
	if err != nil {
		srv.reportError(logger, w, r, http.StatusBadRequest, "%v", err)
		return
	}
	switch vals["format"][0] {
	case "json", "csv":
		break
	default:
		srv.reportError(logger, w, r, http.StatusBadRequest, "format must be one of: csv, json")
		return
	}
	srv.getReport(logger, vals["name"][0], vals["format"][0], w, r)
}

func (srv *server) runReportHandler(w http.ResponseWriter, r *http.Request) {
	logger := srv.newLogger(r)
	srv.logRequest(logger, r)
	if r.Method != "GET" {
		srv.reportError(logger, w, r, http.StatusNotFound, "Not found")
		return
	}
	err := r.ParseForm()
	if err != nil {
		srv.reportError(logger, w, r, http.StatusBadRequest, "couldn't parse URL query params: %v", err)
		return
	}
	vals := r.Form
	err = checkForFields([]string{"query", "start", "end"}, vals)
	if err != nil {
		srv.reportError(logger, w, r, http.StatusBadRequest, "%v", err)
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

func (srv *server) getReport(logger log.FieldLogger, name, format string, w http.ResponseWriter, r *http.Request) {
	// Get the current report to make sure it's in a ***REMOVED***nished state
	report, err := srv.chargeback.informers.reportLister.Reports(srv.chargeback.namespace).Get(name)
	if err != nil {
		logger.WithError(err).Errorf("error getting report: %v", err)
		srv.reportError(logger, w, r, http.StatusInternalServerError, "error getting report: %v", err)
		return
	}
	switch report.Status.Phase {
	case api.ReportPhaseError:
		err := fmt.Errorf(report.Status.Output)
		logger.WithError(err).Errorf("the report encountered an error")
		srv.reportError(logger, w, r, http.StatusInternalServerError, "the report encountered an error: %v", err)
		return
	case api.ReportPhaseFinished:
		// continue with returning the report if the report is ***REMOVED***nished
	case api.ReportPhaseWaiting, api.ReportPhaseStarted:
		fallthrough
	default:
		logger.Errorf(ErrReportIsRunning.Error())
		srv.reportError(logger, w, r, http.StatusAccepted, ErrReportIsRunning.Error())
		return
	}

	reportTable := reportTableName(name)
	getReportQuery := fmt.Sprintf("SELECT * FROM %s", reportTable)
	results, err := presto.ExecuteSelect(srv.chargeback.prestoConn, getReportQuery)
	if err != nil {
		logger.WithError(err).Errorf("failed to perform presto query")
		srv.reportError(logger, w, r, http.StatusInternalServerError, "failed to perform presto query (see chargeback logs for more details): %v", err)
		return
	}

	switch format {
	case "json":
		srv.writeResponseWithBody(logger, w, http.StatusOK, results)
		return
	case "csv":
		// Get generation query to get the list of columns
		genQuery, err := srv.chargeback.informers.reportGenerationQueryLister.ReportGenerationQueries(report.Namespace).Get(report.Spec.GenerationQueryName)
		if err != nil {
			logger.WithError(err).Errorf("error getting report generation query: %v", err)
			srv.reportError(logger, w, r, http.StatusInternalServerError, "error getting report generation query: %v", err)
			return
		}

		if len(results) > 0 && len(genQuery.Spec.Columns) != len(results[0]) {
			logger.WithError(err).Errorf("report results schema doesn't match expected schema")
			srv.reportError(logger, w, r, http.StatusInternalServerError, "report results schema doesn't match expected schema")
			return
		}

		buf := &bytes.Buffer{}
		csvWriter := csv.NewWriter(buf)

		// Write headers
		var keys []string
		if len(results) >= 1 {
			for _, column := range genQuery.Spec.Columns {
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
					logger.WithError(err).Errorf("report results schema doesn't match expected schema, unexpected key: %q", key)
					srv.reportError(logger, w, r, http.StatusInternalServerError, "report results schema doesn't match expected schema, unexpected key: %q", key)
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
					logger.Errorf("error marshalling csv: unknown type %#T for value %v", val, val)
					srv.reportError(logger, w, r, http.StatusInternalServerError, "error marshalling csv (see chargeback logs for more details)", err)
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

func (srv *server) runReport(logger log.FieldLogger, query, start, end string, w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("method not yet implemented"))
}
