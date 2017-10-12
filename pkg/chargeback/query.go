package chargeback

import (
	"encoding/json"
	"fmt"

	log "github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/cache"

	cbTypes "github.com/coreos-inc/kube-chargeback/pkg/apis/chargeback/v1alpha1"
	cb "github.com/coreos-inc/kube-chargeback/pkg/chargeback/v1"
)

func (c *Chargeback) runReportWorker() {
	for c.processReport() {

	}
}

func (c *Chargeback) processReport() bool {
	key, quit := c.informers.reportQueue.Get()
	if quit {
		return false
	}
	defer c.informers.reportQueue.Done(key)

	err := c.syncReport(key.(string))
	c.handleErr(err, "report", key, c.informers.reportQueue)
	return true
}

func (c *Chargeback) syncReport(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		c.logger.WithError(err).Errorf("invalid resource key :%s", key)
		return nil
	}

	report, err := c.informers.reportLister.Reports(namespace).Get(name)
	if err != nil {
		if apierrors.IsNotFound(err) {
			c.logger.Infof("Report %s does not exist anymore", key)
			return nil
		}
		return err
	}

	c.logger.Infof("syncing report %s", report.GetName())
	err = c.handleReport(report)
	if err != nil {
		c.logger.WithError(err).Errorf("error syncing report %s", report.GetName())
		return err
	}
	c.logger.Infof("successfully synced report %s", report.GetName())
	return nil
}

func (c *Chargeback) handleReport(report *cbTypes.Report) error {
	report = report.DeepCopy()

	logger := c.logger.WithFields(log.Fields{
		"name": report.Name,
	})
	switch report.Status.Phase {
	case cbTypes.ReportPhaseStarted:
		err := fmt.Errorf("unable to determine if report generation succeeded")
		c.setReportError(logger, report, err, "found already started report, report generation likely failed while processing")
		return nil
	case cbTypes.ReportPhaseFinished, cbTypes.ReportPhaseError:
		logger.Infof("ignoring report %s, status: %s", report.Name, report.Status.Phase)
		return nil
	default:
		logger.Infof("new report discovered")
	}

	// update status
	report.Status.Phase = cbTypes.ReportPhaseStarted
	newReport, err := c.chargebackClient.ChargebackV1alpha1().Reports(c.namespace).Update(report)
	if err != nil {
		logger.WithError(err).Errorf("failed to update report status to started for %q", report.Name)
		return err
	}
	report = newReport

	logger = logger.WithField("generationQuery", report.Spec.GenerationQueryName)
	genQuery, err := c.informers.reportGenerationQueryLister.ReportGenerationQueries(c.namespace).Get(report.Spec.GenerationQueryName)
	if err != nil {
		logger.WithError(err).Errorf("failed to get report generation query")
		return err
	}

	dataStore, err := c.informers.reportDataStoreLister.ReportDataStores(c.namespace).Get(genQuery.Spec.DataStoreName)
	if err != nil {
		logger.WithError(err).Errorf("failed to get report data store")
		return err
	}

	// get hive and presto connections
	if dataStore.TableName == "" {
		return fmt.Errorf("datastore table not created yet")
	}

	logger = c.logger.WithFields(log.Fields{
		"reportStart": report.Spec.ReportingStart,
		"reportEnd":   report.Spec.ReportingEnd,
	})

	rng := cb.Range{report.Spec.ReportingStart.Time, report.Spec.ReportingEnd.Time}
	results, err := generateReport(logger, report, genQuery, rng, dataStore.TableName, c.hiveConn, c.prestoConn)
	if err != nil {
		// TODO(chance): return the error and handle retrying
		c.setReportError(logger, report, err, "report execution failed")
		return nil
	}
	if c.logReport {
		resultsJSON, err := json.MarshalIndent(results, "", " ")
		if err != nil {
			logger.WithError(err).Errorf("unable to marshal report into JSON")
			return nil
		}
		logger.Debugf("results: %s", string(resultsJSON))
	}

	// update status
	report.Status.Phase = cbTypes.ReportPhaseFinished
	_, err = c.chargebackClient.ChargebackV1alpha1().Reports(c.namespace).Update(report)
	if err != nil {
		logger.WithError(err).Warnf("failed to update report status to finished for %q", report.Name)
	} else {
		logger.Infof("finished report %q", report.Name)
	}
	return nil
}

func (c *Chargeback) setReportError(logger *log.Entry, q *cbTypes.Report, err error, errMsg string) {
	logger.WithError(err).Errorf(errMsg)
	q.Status.Phase = cbTypes.ReportPhaseError
	q.Status.Output = err.Error()
	_, err = c.chargebackClient.ChargebackV1alpha1().Reports(c.namespace).Update(q)
	if err != nil {
		logger.WithError(err).Errorf("unable to update report status to error")
	}
}
