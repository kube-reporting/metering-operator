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

func (c *Chargeback) startHTTPServer() {
	http.HandleFunc("/api/v1/reports/get", c.getReportHandler)
	http.HandleFunc("/api/v1/reports/run", c.runReportHandler)
	c.logger.Fatal(http.ListenAndServe(":8080", nil))
}

func (c *Chargeback) logRequest(r *http.Request) {
	c.logger.WithFields(log.Fields{
		"method": r.Method,
		"url":    r.URL.String(),
	}).Info("new request")
}

func (c *Chargeback) reportError(w http.ResponseWriter, status int, message string, args ...interface{}) {
	w.WriteHeader(status)
	_, err := w.Write([]byte(fmt.Sprintf(message, args...)))
	if err != nil {
		c.logger.Warnf("error sending client error: %v", err)
	}
}

func (c *Chargeback) getReportHandler(w http.ResponseWriter, r *http.Request) {
	c.logRequest(r)
	if r.Method != "GET" {
		c.reportError(w, http.StatusNotFound, "Not found")
		return
	}
	err := r.ParseForm()
	if err != nil {
		c.reportError(w, http.StatusBadRequest, "couldn't parse URL query params: %v", err)
		return
	}
	vals := r.Form
	err = checkForFields([]string{"name", "format"}, vals)
	if err != nil {
		c.reportError(w, http.StatusBadRequest, "%v", err)
		return
	}
	switch vals["format"][0] {
	case "json", "csv":
		break
	default:
		c.reportError(w, http.StatusBadRequest, "format must be one of: csv, json")
		return
	}
	c.getReport(vals["name"][0], vals["format"][0], w)
}

func (c *Chargeback) runReportHandler(w http.ResponseWriter, r *http.Request) {
	c.logRequest(r)
	if r.Method != "GET" {
		c.reportError(w, http.StatusNotFound, "Not found")
		return
	}
	err := r.ParseForm()
	if err != nil {
		c.reportError(w, http.StatusBadRequest, "couldn't parse URL query params: %v", err)
		return
	}
	vals := r.Form
	err = checkForFields([]string{"query", "start", "end"}, vals)
	if err != nil {
		c.reportError(w, http.StatusBadRequest, "%v", err)
		return
	}
	c.runReport(vals["query"][0], vals["start"][0], vals["end"][0], w)
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

func (c *Chargeback) getReport(name, format string, w http.ResponseWriter) {
	prestoCon, err := c.prestoConn()
	if err != nil {
		log.Errorf("failed to con***REMOVED***gure presto connection: %v", err)
		reportError(w, http.StatusInternalServerError, "failed to con***REMOVED***gure presto connection (see chargeback logs for more details)", err)
		return
	}
	defer prestoCon.Close()

	reportTable := reportTableName(name)
	getReportQuery := fmt.Sprintf("SELECT * FROM %s", reportTable)
	results, err := presto.ExecuteSelect(prestoCon, getReportQuery)
	if err != nil {
		c.logger.Errorf("failed to perform presto query: %v", err)
		c.reportError(w, http.StatusInternalServerError, "failed to perform presto query (see chargeback logs for more details)", err)
		return
	}

	switch format {
	case "json":
		e := json.NewEncoder(w)
		err := e.Encode(results)
		if err != nil {
			c.logger.Errorf("error marshalling json: %v", err)
			return
		}
	case "csv":
		csvWriter := csv.NewWriter(w)
		defer csvWriter.Flush()

		// Write headers
		var keys []string
		if len(results) >= 1 {
			for key, _ := range results[0] {
				keys = append(keys, key)
			}
			err := csvWriter.Write(keys)
			if err != nil {
				c.logger.Errorf("failed to write headers: %v", err)
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
					c.logger.Errorf("error marshalling csv: unknown type %#T for value %v", val, val)
					c.reportError(w, http.StatusInternalServerError, "error marshalling csv (see chargeback logs for more details)", err)
					return
				}
			}
			err := csvWriter.Write(vals)
			if err != nil {
				c.logger.Errorf("failed to write csv row: %v", err)
				return
			}
		}
	}
}

func (c *Chargeback) runReport(query, start, end string, w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("method not yet implemented"))
}
