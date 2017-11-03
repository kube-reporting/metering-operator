package chargeback

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/coreos-inc/kube-chargeback/pkg/presto"
)

type server struct {
	chargeback *Chargeback
	logger     log.FieldLogger
}

func newServer(c *Chargeback, logger log.FieldLogger) *server {
	logger = logger.WithField("component", "api")
	return &server{
		chargeback: c,
		logger:     logger,
	}
}

func (srv *server) start() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/reports/get", srv.getReportHandler)
	mux.HandleFunc("/api/v1/reports/run", srv.runReportHandler)
	srv.logger.Fatal(http.ListenAndServe(":8080", mux))
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
		csvWriter := csv.NewWriter(w)
		defer csvWriter.Flush()

		// Write headers
		var keys []string
		if len(results) >= 1 {
			for key := range results[0] {
				keys = append(keys, key)
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
				val := row[key]
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
	}
}

func (srv *server) runReport(query, start, end string, w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("method not yet implemented"))
}
