package chargeback

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	log "github.com/sirupsen/logrus"

	cb "github.com/coreos-inc/kube-chargeback/pkg/chargeback/v1"
	cbTypes "github.com/coreos-inc/kube-chargeback/pkg/chargeback/v1/types"
	"github.com/coreos-inc/kube-chargeback/pkg/hive"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func (c *Chargeback) handleAddReport(obj interface{}) {
	if obj == nil {
		log.Debugf("received nil object!")
		return
	}

	report := obj.(*cbTypes.Report)

	logger := log.WithFields(log.Fields{
		"name":            report.Name,
		"generationQuery": report.Spec.GenerationQueryName,
		"start":           report.Spec.ReportingStart,
		"end":             report.Spec.ReportingEnd,
	})
	logger.Infof("new report discovered")

	switch report.Status.Phase {
	case cbTypes.ReportPhaseFinished:
		fallthrough
	case cbTypes.ReportPhaseError:
		logger.Warnf("ignoring %s, status: %s", report.GetSelfLink(), report.Status.Phase)
		return
	}

	// update status
	report.Status.Phase = cbTypes.ReportPhaseStarted
	report, err := c.charge.Reports(c.namespace).Update(report)
	if err != nil {
		c.setError(logger, report, fmt.Errorf("failed to update report status for %q: %v", report.Name, err))
		return
	}

	// lookup report generation query and data store
	restClient, err := cbTypes.GetRestClient()
	if err != nil {
		c.setError(logger, report, fmt.Errorf("failed to get rest client: %v", err))
		return
	}

	genQuery, err := cbTypes.GetReportGenerationQuery(restClient, c.namespace, report.Spec.GenerationQueryName)
	if err != nil {
		c.setError(logger, report, fmt.Errorf("failed to get report generation query: %v", err))
		return
	}

	dataStore, err := cbTypes.GetReportDataStore(restClient, c.namespace, genQuery.Spec.DataStoreName)
	if err != nil {
		c.setError(logger, report, fmt.Errorf("failed to get report data store: %v", err))
		return
	}

	rng := cb.Range{report.Spec.ReportingStart.Time, report.Spec.ReportingEnd.Time}

	// get give and presto connections
	hiveCon, err := c.hiveConn()
	if err != nil {
		c.setError(logger, report, fmt.Errorf("Failed to configure Hive connection: %v", err))
		return
	}
	defer hiveCon.Close()

	prestoCon, err := c.prestoConn()
	if err != nil {
		c.setError(logger, report, fmt.Errorf("Failed to configure Presto connection: %v", err))
		return
	}
	defer prestoCon.Close()

	promsumTable := fmt.Sprintf("%s_%d", "kube_usage", rand.Int31())
	bucket, prefix := dataStore.Spec.Storage.Bucket, dataStore.Spec.Storage.Prefix
	logger.Debugf("Creating table pointing to bucket/prefix %q for promsum: %q.", bucket+"/"+prefix, promsumTable)
	if err = hive.CreatePromsumTable(hiveCon, promsumTable, bucket, prefix); err != nil {
		c.setError(logger, report, fmt.Errorf("Couldn't create table for cluster usage metric data: %v", err))
		return
	}

	results, err := generateReport(logger, report, genQuery, rng, promsumTable, hiveCon, prestoCon)
	if err != nil {
		c.setError(logger, report, fmt.Errorf("Report execution failed: %v", err))
		return
	}
	if c.logReport {
		resultsJSON, err := json.MarshalIndent(results, "", " ")
		if err != nil {
			c.setError(logger, report, fmt.Errorf("Unable to marshal report into JSON: %v", err))
			return
		}
		logger.Debugf("results: %s", string(resultsJSON))
	}

	// update status
	report.Status.Phase = cbTypes.ReportPhaseFinished
	report, err = c.charge.Reports(c.namespace).Update(report)
	if err != nil {
		logger.Warnf("failed to update report status for %q: ", report.Name, err)
	} else {
		logger.Infof("finished report %q", report.Name)
	}
}

func (c *Chargeback) setError(logger *log.Entry, q *cbTypes.Report, err error) {
	logger.WithError(err).Errorf("error encountered")
	q.Status.Phase = cbTypes.ReportPhaseError
	q.Status.Output = err.Error()
	_, err = c.charge.Reports(c.namespace).Update(q)
	if err != nil {
		logger.Errorf("FAILED TO REPORT ERROR: %v", err)
	}
}
