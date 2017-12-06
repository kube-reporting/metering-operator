package chargeback

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	api "github.com/coreos-inc/kube-chargeback/pkg/apis/chargeback/v1alpha1"
	"github.com/coreos-inc/kube-chargeback/pkg/presto"
)

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
	}).WithFields(newLogIdentifier())
}

func (srv *server) logRequest(r *http.Request) {
	srv.newLogger(r).Infof("%s %s", r.Method, r.URL.String())
}

func (srv *server) reportError(w http.ResponseWriter, r *http.Request, status int, message string, args ...interface{}) {
	logger := srv.newLogger(r)
	w.WriteHeader(status)
	_, err := w.Write([]byte(fmt.Sprintf(message, args...)))
	if err != nil {
		logger.WithError(err).Warnf("error sending client error")
	}
}

func (srv *server) getReportHandler(w http.ResponseWriter, r *http.Request) {
	srv.logRequest(r)
	if r.Method != "GET" {
		srv.reportError(w, r, http.StatusNotFound, "Not found")
		return
	}
	err := r.ParseForm()
	if err != nil {
		srv.reportError(w, r, http.StatusBadRequest, "couldn't parse URL query params: %v", err)
		return
	}
	vals := r.Form
	err = checkForFields([]string{"name", "format"}, vals)
	if err != nil {
		srv.reportError(w, r, http.StatusBadRequest, "%v", err)
		return
	}
	switch vals["format"][0] {
	case "json", "csv":
		break
	default:
		srv.reportError(w, r, http.StatusBadRequest, "format must be one of: csv, json")
		return
	}
	srv.getReport(vals["name"][0], vals["format"][0], w, r)
}

func (srv *server) runReportHandler(w http.ResponseWriter, r *http.Request) {
	srv.logRequest(r)
	if r.Method != "GET" {
		srv.reportError(w, r, http.StatusNotFound, "Not found")
		return
	}
	err := r.ParseForm()
	if err != nil {
		srv.reportError(w, r, http.StatusBadRequest, "couldn't parse URL query params: %v", err)
		return
	}
	vals := r.Form
	err = checkForFields([]string{"query", "start", "end"}, vals)
	if err != nil {
		srv.reportError(w, r, http.StatusBadRequest, "%v", err)
		return
	}
	srv.runReport(vals["query"][0], vals["start"][0], vals["end"][0], w)
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

func (srv *server) getReport(name, format string, w http.ResponseWriter, r *http.Request) {
	// Get the current report to make sure it's in a finished state
	report, err := srv.chargeback.informers.reportLister.Reports(srv.chargeback.namespace).Get(name)
	if err != nil {
		srv.logger.WithError(err).Errorf("error getting report: %v", err)
		srv.reportError(w, r, http.StatusInternalServerError, "error getting report: %v", err)
		return
	}
	switch report.Status.Phase {
	case api.ReportPhaseError:
		err := fmt.Errorf(report.Status.Output)
		srv.logger.WithError(err).Errorf("the report encountered an error")
		srv.reportError(w, r, http.StatusInternalServerError, "the report encountered an error: %v", err)
		return
	case api.ReportPhaseFinished:
		// continue with returning the report if the report is finished
	case api.ReportPhaseWaiting, api.ReportPhaseStarted:
		fallthrough
	default:
		srv.logger.Errorf("the report is still running")
		srv.reportError(w, r, http.StatusBadRequest, "the report is still running")
		return
	}

	reportTable := reportTableName(name)
	getReportQuery := fmt.Sprintf("SELECT * FROM %s", reportTable)
	results, err := presto.ExecuteSelect(srv.chargeback.prestoConn, getReportQuery)
	if err != nil {
		srv.logger.WithError(err).Errorf("failed to perform presto query")
		srv.reportError(w, r, http.StatusInternalServerError, "failed to perform presto query (see chargeback logs for more details): %v", err)
		return
	}

	switch format {
	case "json":
		e := json.NewEncoder(w)
		err := e.Encode(results)
		if err != nil {
			srv.logger.Errorf("error marshalling json: %v", err)
			return
		}
	case "csv":
		// Get generation query to get the list of columns
		genQuery, err := srv.chargeback.informers.reportGenerationQueryLister.ReportGenerationQueries(report.Namespace).Get(report.Spec.GenerationQueryName)
		if err != nil {
			srv.logger.WithError(err).Errorf("error getting report generation query: %v", err)
			srv.reportError(w, r, http.StatusInternalServerError, "error getting report generation query: %v", err)
			return
		}

		if len(results) > 0 && len(genQuery.Spec.Columns) != len(results[0]) {
			srv.logger.WithError(err).Errorf("report results schema doesn't match expected schema")
			srv.reportError(w, r, http.StatusInternalServerError, "report results schema doesn't match expected schema")
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
				srv.logger.WithError(err).Errorf("failed to write headers")
				return
			}
		}

		// Write the rest
		for _, row := range results {
			vals := make([]string, len(keys))
			for i, key := range keys {
				val, ok := row[key]
				if !ok {
					srv.logger.WithError(err).Errorf("report results schema doesn't match expected schema, unexpected key: %q", key)
					srv.reportError(w, r, http.StatusInternalServerError, "report results schema doesn't match expected schema, unexpected key: %q", key)
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
					srv.logger.Errorf("error marshalling csv: unknown type %#T for value %v", val, val)
					srv.reportError(w, r, http.StatusInternalServerError, "error marshalling csv (see chargeback logs for more details)", err)
					return
				}
			}
			err := csvWriter.Write(vals)
			if err != nil {
				srv.logger.Errorf("failed to write csv row: %v", err)
				return
			}
		}

		csvWriter.Flush()
		w.Write(buf.Bytes())
	}
}

func (srv *server) runReport(query, start, end string, w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("method not yet implemented"))
}
