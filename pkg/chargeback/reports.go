package chargeback

import (
	"encoding/json"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"

	cbTypes "github.com/operator-framework/operator-metering/pkg/apis/chargeback/v1alpha1"
)

var (
	defaultGracePeriod = metav1.Duration{Duration: time.Minute * 5}
)

func (c *Chargeback) runReportWorker() {
	logger := c.logger.WithField("component", "reportWorker")
	logger.Infof("Report worker started")
	for c.processReport(logger) {

	}
}

func (c *Chargeback) processReport(logger log.FieldLogger) bool {
	if c.queues.reportQueue.ShuttingDown() {
		logger.Infof("queue is shutting down")
	}
	obj, quit := c.queues.reportQueue.Get()
	if quit {
		logger.Infof("queue is shutting down, exiting worker")
		return false
	}
	defer c.queues.reportQueue.Done(obj)

	logger = logger.WithFields(c.newLogIdentifier())
	if key, ok := c.getKeyFromQueueObj(logger, "report", obj, c.queues.reportQueue); ok {
		err := c.syncReport(logger, key)
		c.handleErr(logger, err, "report", obj, c.queues.reportQueue)
	}
	return true
}

func (c *Chargeback) syncReport(logger log.FieldLogger, key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		logger.WithError(err).Errorf("invalid resource key :%s", key)
		return nil
	}

	logger = logger.WithField("report", name)
	report, err := c.informers.Chargeback().V1alpha1().Reports().Lister().Reports(namespace).Get(name)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Infof("Report %s does not exist anymore", key)
			return nil
		}
		return err
	}

	logger.Infof("syncing report %s", report.GetName())
	err = c.handleReport(logger, report)
	if err != nil {
		logger.WithError(err).Errorf("error syncing report %s", report.GetName())
		return err
	}
	logger.Infof("successfully synced report %s", report.GetName())
	return nil
}

func (c *Chargeback) handleReport(logger log.FieldLogger, report *cbTypes.Report) error {
	report = report.DeepCopy()

	switch report.Status.Phase {
	case cbTypes.ReportPhaseStarted:
		// If it's started, query the API to get the most up to date resource,
		// as it's possible it's finished, but we haven't gotten it yet.
		newReport, err := c.chargebackClient.ChargebackV1alpha1().Reports(report.Namespace).Get(report.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		if report.UID != newReport.UID {
			logger.Warn("started report has different UUID in API than in cache, waiting for resync to process")
			return nil
		}

		err = c.informers.Chargeback().V1alpha1().Reports().Informer().GetIndexer().Update(newReport)
		if err != nil {
			logger.WithError(err).Warnf("unable to update report cache with updated report")
			// if we cannot update it, don't re queue it
			return nil
		}

		// It's no longer started, requeue it
		if newReport.Status.Phase != cbTypes.ReportPhaseStarted {
			key, err := cache.MetaNamespaceKeyFunc(newReport)
			if err == nil {
				c.queues.reportQueue.AddRateLimited(key)
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

	// set the default grace period
	if report.Spec.GracePeriod == nil {
		report.Spec.GracePeriod = &defaultGracePeriod
	}

	// If we're waiting until the end and we're not past the end time + grace
	// period, ignore this report
	if !report.Spec.RunImmediately && report.Spec.ReportingEnd.Add(report.Spec.GracePeriod.Duration).After(c.clock.Now()) {
		logger.Infof("report %s not past grace period yet, ignoring for now", report.Name)
		return nil
	}

	logger = logger.WithField("generationQuery", report.Spec.GenerationQueryName)
	genQuery, err := c.informers.Chargeback().V1alpha1().ReportGenerationQueries().Lister().ReportGenerationQueries(report.Namespace).Get(report.Spec.GenerationQueryName)
	if err != nil {
		logger.WithError(err).Errorf("failed to get report generation query")
		return err
	}

	logger = logger.WithFields(log.Fields{
		"reportStart": report.Spec.ReportingStart,
		"reportEnd":   report.Spec.ReportingEnd,
	})

	if valid, err := c.validateGenerationQuery(logger, genQuery, true); err != nil {
		c.setReportError(logger, report, err, "report is invalid")
		return nil
	} else if !valid {
		logger.Warnf("cannot start report, it has uninitialized dependencies")
		return nil
	}

	logger.Debug("updating report status to started")
	// update status
	report.Status.Phase = cbTypes.ReportPhaseStarted
	newReport, err := c.chargebackClient.ChargebackV1alpha1().Reports(report.Namespace).Update(report)
	if err != nil {
		logger.WithError(err).Errorf("failed to update report status to started for %q", report.Name)
		return err
	}

	report = newReport
	tableName := reportTableName(report.Name)

	results, err := c.generateReport(
		logger,
		report,
		"report",
		report.Name,
		tableName,
		report.Spec.ReportingStart.Time,
		report.Spec.ReportingEnd.Time,
		report.Spec.Output,
		genQuery,
		true,
	)
	if err != nil {
		c.setReportError(logger, report, err, "report execution failed")
		return err
	}
	if c.cfg.LogReport {
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
		logger.WithError(err).Warnf("failed to update report status to finished for %q", report.Name)
	} else {
		logger.Infof("finished report %q", report.Name)
	}
	return nil
}

func (c *Chargeback) setReportError(logger log.FieldLogger, report *cbTypes.Report, err error, errMsg string) {
	logger.WithField("report", report.Name).WithError(err).Errorf(errMsg)
	report.Status.Phase = cbTypes.ReportPhaseError
	report.Status.Output = err.Error()
	_, err = c.chargebackClient.ChargebackV1alpha1().Reports(report.Namespace).Update(report)
	if err != nil {
		logger.WithError(err).Errorf("unable to update report status to error")
	}
}
