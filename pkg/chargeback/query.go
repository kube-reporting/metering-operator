package chargeback

import (
	"encoding/json"
	"fmt"

	log "github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		// If it's started, query the API to get the most up to date resource,
		// as it's possible it's ***REMOVED***nished, but we haven't gotten it yet.
		newReport, err := c.chargebackClient.ChargebackV1alpha1().Reports(report.Namespace).Get(report.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		if report.UID != newReport.UID {
			logger.Warn("started report has different UUID in API than in cache, waiting for resync to process")
			return nil
		}

		c.informers.reportInformer.GetIndexer().Update(newReport)
		if err != nil {
			logger.WithError(err).Warnf("unable to update report cache with updated report")
			// if we cannot update it, don't re queue it
			return nil
		}

		// It's no longer started, requeue it
		if newReport.Status.Phase != cbTypes.ReportPhaseStarted {
			key, err := cache.MetaNamespaceKeyFunc(newReport)
			if err == nil {
				c.informers.reportQueue.AddRateLimited(key)
			}
			return nil
		}

		err = fmt.Errorf("unable to determine if report generation succeeded")
		c.setReportError(logger, report, err, "found already started report, report generation likely failed while processing")
		return nil
	case cbTypes.ReportPhaseFinished, cbTypes.ReportPhaseError:
		logger.Infof("ignoring report %s, status: %s", report.Name, report.Status.Phase)
		return nil
	default:
		logger.Infof("new report discovered")
	}

	logger = logger.WithField("generationQuery", report.Spec.GenerationQueryName)
	genQuery, err := c.informers.reportGenerationQueryLister.ReportGenerationQueries(report.Namespace).Get(report.Spec.GenerationQueryName)
	if err != nil {
		logger.WithError(err).Errorf("failed to get report generation query")
		return err
	}

	dataStore, err := c.informers.reportDataStoreLister.ReportDataStores(report.Namespace).Get(genQuery.Spec.DataStoreName)
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

	// update status
	report.Status.Phase = cbTypes.ReportPhaseStarted
	newReport, err := c.chargebackClient.ChargebackV1alpha1().Reports(report.Namespace).Update(report)
	if err != nil {
		logger.WithError(err).Errorf("failed to update report status to started for %q", report.Name)
		return err
	}
	report = newReport

	rng := cb.Range{report.Spec.ReportingStart.Time, report.Spec.ReportingEnd.Time}
	results, err := c.generateReport(logger, report, genQuery, rng, dataStore.TableName)
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
	_, err = c.chargebackClient.ChargebackV1alpha1().Reports(report.Namespace).Update(report)
	if err != nil {
		logger.WithError(err).Warnf("failed to update report status to ***REMOVED***nished for %q", report.Name)
	} ***REMOVED*** {
		logger.Infof("***REMOVED***nished report %q", report.Name)
	}
	return nil
}

func (c *Chargeback) setReportError(logger *log.Entry, report *cbTypes.Report, err error, errMsg string) {
	logger.WithError(err).Errorf(errMsg)
	report.Status.Phase = cbTypes.ReportPhaseError
	report.Status.Output = err.Error()
	_, err = c.chargebackClient.ChargebackV1alpha1().Reports(report.Namespace).Update(report)
	if err != nil {
		logger.WithError(err).Errorf("unable to update report status to error")
	}
}
