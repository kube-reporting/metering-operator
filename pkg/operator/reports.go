package operator

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"

	cbTypes "github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1"
	"github.com/operator-framework/operator-metering/pkg/hive"
	"github.com/operator-framework/operator-metering/pkg/operator/reporting"
)

var (
	defaultGracePeriod = metav1.Duration{Duration: time.Minute * 5}

	reportPrometheusMetricLabels = []string{"report", "reportgenerationquery", "table_name"}

	generateReportTotalCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: prometheusMetricNamespace,
			Name:      "generate_reports_total",
			Help:      "Duration to generate a Report.",
		},
		reportPrometheusMetricLabels,
	)

	generateReportFailedCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: prometheusMetricNamespace,
			Name:      "generate_reports_failed_total",
			Help:      "Duration to generate a Report.",
		},
		reportPrometheusMetricLabels,
	)

	generateReportDurationHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: prometheusMetricNamespace,
			Name:      "generate_report_duration_seconds",
			Help:      "Duration to generate a Report.",
			Buckets:   []float64{60.0, 300.0, 600.0},
		},
		reportPrometheusMetricLabels,
	)
)

func init() {
	prometheus.MustRegister(generateReportFailedCounter)
	prometheus.MustRegister(generateReportTotalCounter)
	prometheus.MustRegister(generateReportDurationHistogram)
}

func (op *Reporting) runReportWorker() {
	logger := op.logger.WithField("component", "reportWorker")
	logger.Infof("Report worker started")
	for op.processResource(logger, op.syncReport, "Report", op.queues.reportQueue) {
	}
}

func (op *Reporting) syncReport(logger log.FieldLogger, key string) error {
	startTime := op.clock.Now()
	defer func() {
		logger.Debugf("report sync for %q took %v", key, op.clock.Since(startTime))
	}()

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		logger.WithError(err).Errorf("invalid resource key :%s", key)
		return nil
	}

	logger = logger.WithField("report", name)
	report, err := op.informers.Metering().V1alpha1().Reports().Lister().Reports(namespace).Get(name)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Infof("Report %s does not exist anymore", key)
			return nil
		}
		return err
	}

	return op.handleReport(logger, report)
}

func (op *Reporting) handleReport(logger log.FieldLogger, report *cbTypes.Report) error {
	report = report.DeepCopy()

	tableName := reportTableName(report.Name)
	metricLabels := prometheus.Labels{
		"report":                report.Name,
		"reportgenerationquery": report.Spec.GenerationQueryName,
		"table_name":            tableName,
	}

	genReportFailedCounter := generateReportFailedCounter.With(metricLabels)
	genReportTotalCounter := generateReportTotalCounter.With(metricLabels)
	genReportDurationObserver := generateReportDurationHistogram.With(metricLabels)

	switch report.Status.Phase {
	case cbTypes.ReportPhaseStarted:
		// If it's started, query the API to get the most up to date resource,
		// as it's possible it's ***REMOVED***nished, but we haven't gotten it yet.
		newReport, err := op.meteringClient.MeteringV1alpha1().Reports(report.Namespace).Get(report.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		if report.UID != newReport.UID {
			return fmt.Errorf("started report has different UUID in API than in cache, skipping processing until next reconcile")
		}

		err = op.informers.Metering().V1alpha1().Reports().Informer().GetIndexer().Update(newReport)
		if err != nil {
			logger.WithError(err).Warnf("unable to update report cache with updated report")
			// if we cannot update it, don't re queue it
			return err
		}

		// It's no longer started, requeue it
		if newReport.Status.Phase != cbTypes.ReportPhaseStarted {
			op.enqueueReportRateLimited(newReport)
			return nil
		}

		err = fmt.Errorf("unable to determine if report generation succeeded")
		op.setReportError(logger, report, err, "found already started report, report generation likely failed while processing")
		return nil
	case cbTypes.ReportPhaseFinished, cbTypes.ReportPhaseError:
		logger.Infof("ignoring report %s, status: %s", report.Name, report.Status.Phase)
		return nil
	default:
		logger.Infof("new report discovered")
	}

	logger = logger.WithFields(log.Fields{
		"reportStart": report.Spec.ReportingStart,
		"reportEnd":   report.Spec.ReportingEnd,
	})

	now := op.clock.Now()

	var gracePeriod time.Duration
	if report.Spec.GracePeriod != nil {
		gracePeriod = report.Spec.GracePeriod.Duration
	} ***REMOVED*** {
		gracePeriod = op.getDefaultReportGracePeriod()
		logger.Debugf("Report has no gracePeriod con***REMOVED***gured, falling back to defaultGracePeriod: %s", gracePeriod)
	}

	var waitTime time.Duration
	nextRunTime := report.Spec.ReportingEnd.Add(gracePeriod)
	reportGracePeriodUnmet := nextRunTime.After(now)
	waitTime = nextRunTime.Sub(now)

	if report.Spec.RunImmediately {
		logger.Infof("report con***REMOVED***gured to run immediately with %s until periodEnd+gracePeriod: %s", waitTime, nextRunTime)
	} ***REMOVED*** if reportGracePeriodUnmet {
		logger.Infof("report %s not past grace period yet, ignoring until %s (%s)", report.Name, nextRunTime, waitTime)
		op.enqueueReportAfter(report, waitTime)
		return nil
	}

	logger = logger.WithField("generationQuery", report.Spec.GenerationQueryName)
	genQuery, err := op.informers.Metering().V1alpha1().ReportGenerationQueries().Lister().ReportGenerationQueries(report.Namespace).Get(report.Spec.GenerationQueryName)
	if err != nil {
		logger.WithError(err).Errorf("failed to get report generation query")
		return err
	}

	reportDataSourceLister := op.informers.Metering().V1alpha1().ReportDataSources().Lister()
	reportGenerationQueryLister := op.informers.Metering().V1alpha1().ReportGenerationQueries().Lister()
	depsStatus, err := reporting.GetGenerationQueryDependenciesStatus(
		reporting.NewReportGenerationQueryListerGetter(reportGenerationQueryLister),
		reporting.NewReportDataSourceListerGetter(reportDataSourceLister),
		genQuery,
	)
	if err != nil {
		return fmt.Errorf("unable to run Report %s, ReportGenerationQuery %s, failed to get dependencies: %v", report.Name, genQuery.Name, err)
	}
	_, err = op.validateDependencyStatus(depsStatus)
	if err != nil {
		return fmt.Errorf("unable to run Report %s, ReportGenerationQuery %s, failed to validate dependencies: %v", report.Name, genQuery.Name, err)
	}

	logger.Debug("updating report status to started")
	// update status
	report.Status.Phase = cbTypes.ReportPhaseStarted
	report, err = op.meteringClient.MeteringV1alpha1().Reports(report.Namespace).Update(report)
	if err != nil {
		logger.WithError(err).Errorf("failed to update report status to started for %q", report.Name)
		return err
	}

	logger.Debugf("dropping table %s", tableName)
	err = hive.ExecuteDropTable(op.hiveQueryer, tableName, true)
	if err != nil {
		return err
	}

	columns := generateHiveColumns(genQuery)
	err = op.createTableForStorage(logger, report, cbTypes.SchemeGroupVersion.WithKind("Report"), report.Spec.Output, tableName, columns)
	if err != nil {
		logger.WithError(err).Error("error creating report table for Report")
		return err
	}

	report.Status.TableName = tableName
	report, err = op.meteringClient.MeteringV1alpha1().Reports(report.Namespace).Update(report)
	if err != nil {
		logger.WithError(err).Errorf("unable to update scheduledReport status with tableName")
		return err
	}

	genReportTotalCounter.Inc()
	generateReportStart := op.clock.Now()
	err = op.generateReport(
		logger,
		report,
		"report",
		report.Name,
		tableName,
		report.Spec.ReportingStart.Time,
		report.Spec.ReportingEnd.Time,
		genQuery,
		true,
	)
	generateReportDuration := op.clock.Since(generateReportStart)
	genReportDurationObserver.Observe(float64(generateReportDuration.Seconds()))
	if err != nil {
		genReportFailedCounter.Inc()
		op.setReportError(logger, report, err, "report execution failed")
		return err
	}

	// update status
	report.Status.Phase = cbTypes.ReportPhaseFinished
	_, err = op.meteringClient.MeteringV1alpha1().Reports(report.Namespace).Update(report)
	if err != nil {
		logger.WithError(err).Warnf("failed to update report status to ***REMOVED***nished for %q", report.Name)
	} ***REMOVED*** {
		logger.Infof("***REMOVED***nished report %q", report.Name)
	}
	return nil
}

func (op *Reporting) setReportError(logger log.FieldLogger, report *cbTypes.Report, err error, errMsg string) {
	logger.WithField("report", report.Name).WithError(err).Errorf(errMsg)
	report.Status.Phase = cbTypes.ReportPhaseError
	report.Status.Output = err.Error()
	_, err = op.meteringClient.MeteringV1alpha1().Reports(report.Namespace).Update(report)
	if err != nil {
		logger.WithError(err).Errorf("unable to update report status to error")
	}
}
