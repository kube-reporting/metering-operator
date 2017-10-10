package chargeback

import (
	"encoding/json"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"

	cbTypes "github.com/coreos-inc/kube-chargeback/pkg/apis/chargeback/v1alpha1"
	cb "github.com/coreos-inc/kube-chargeback/pkg/chargeback/v1"
	"github.com/coreos-inc/kube-chargeback/pkg/hive"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Chargeback) runReportWorker() {
	for c.processReport() {

	}
}

func (c *Chargeback) processReport() bool {
	key, quit := c.reportQueue.Get()
	if quit {
		return false
	}
	defer c.reportQueue.Done(key)

	err := c.syncReport(key.(string))
	c.handleErr(err, key)
	return true
}

func (c *Chargeback) syncReport(key string) error {
	indexer := c.reportInformer.GetIndexer()
	obj, exists, err := indexer.GetByKey(key)
	if err != nil {
		log.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		log.Infof("Report %s does not exist anymore", key)
	} else {
		report := obj.(*cbTypes.Report)
		log.Infof("Sync/Add/Update for Report %s", report.GetName())
		c.handleReport(report)
	}
	return nil
}

// handleErr checks if an error happened and makes sure we will retry later.
func (c *Chargeback) handleErr(err error, key interface{}) {
	if err == nil {
		c.reportQueue.Forget(key)
		return
	}

	// This controller retries 5 times if something goes wrong. After that, it stops trying.
	if c.reportQueue.NumRequeues(key) < 5 {
		log.WithError(err).Error("Error syncing report %v", key)

		c.reportQueue.AddRateLimited(key)
		return
	}

	c.reportQueue.Forget(key)
	log.WithError(err).Infof("Dropping report %q out of the queue", key)
}

func (c *Chargeback) handleReport(report *cbTypes.Report) {
	report = report.DeepCopy()

	logger := log.WithFields(log.Fields{
		"name":            report.Name,
		"generationQuery": report.Spec.GenerationQueryName,
		"start":           report.Spec.ReportingStart,
		"end":             report.Spec.ReportingEnd,
	})

	switch report.Status.Phase {
	case cbTypes.ReportPhaseStarted, cbTypes.ReportPhaseFinished, cbTypes.ReportPhaseError:
		logger.Warnf("ignoring report %s, status: %s", report.Name, report.Status.Phase)
		return
	default:
		logger.Infof("new report discovered")
	}

	// update status
	report.Status.Phase = cbTypes.ReportPhaseStarted
	report, err := c.chargebackClient.ChargebackV1alpha1().Reports(c.namespace).Update(report)
	if err != nil {
		c.setError(logger, report, fmt.Errorf("failed to update report status for %q: %v", report.Name, err))
		return
	}

	genQuery, err := c.chargebackClient.ChargebackV1alpha1().ReportGenerationQueries(c.namespace).Get(report.Spec.GenerationQueryName, metav1.GetOptions{})
	if err != nil {
		c.setError(logger, report, fmt.Errorf("failed to get report generation query: %v", err))
		return
	}

	dataStore, err := c.chargebackClient.ChargebackV1alpha1().ReportDataStores(c.namespace).Get(genQuery.Spec.DataStoreName, metav1.GetOptions{})
	if err != nil {
		c.setError(logger, report, fmt.Errorf("failed to get report data store: %v", err))
		return
	}

	rng := cb.Range{report.Spec.ReportingStart.Time, report.Spec.ReportingEnd.Time}

	// get hive and presto connections
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

	replacer := strings.NewReplacer("-", "_", ".", "_")
	datastoreTable := fmt.Sprintf("datastore_%s", replacer.Replace(dataStore.Name))
	bucket, prefix := dataStore.Spec.Storage.Bucket, dataStore.Spec.Storage.Prefix
	logger.Debugf("Creating table %s pointing to s3 bucket %s at prefix %s", datastoreTable, bucket, prefix)
	if err = hive.CreatePromsumTable(hiveCon, datastoreTable, bucket, prefix); err != nil {
		c.setError(logger, report, fmt.Errorf("Couldn't create table for cluster usage metric data: %v", err))
		return
	}

	results, err := generateReport(logger, report, genQuery, rng, datastoreTable, hiveCon, prestoCon)
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
	report, err = c.chargebackClient.ChargebackV1alpha1().Reports(c.namespace).Update(report)
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
	_, err = c.chargebackClient.ChargebackV1alpha1().Reports(c.namespace).Update(q)
	if err != nil {
		logger.Errorf("FAILED TO REPORT ERROR: %v", err)
	}
}
