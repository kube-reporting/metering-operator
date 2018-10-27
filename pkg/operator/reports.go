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
	"github.com/operator-framework/operator-metering/pkg/operator/reportingutil"
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
	const maxRequeues = 5
	for op.processResource(logger, op.syncReport, "Report", op.reportQueue, maxRequeues) {
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

	logger = logger.WithField("Report", name)
	report, err := op.reportLister.Reports(namespace).Get(name)
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

	tableName := reportingutil.ReportTableName(report.Name)
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
		if report.Status.TableName == "" {
			// this report hasn't had it's table created yet so we failed
			// before we actually generated.
			logger.Debugf("found existing started report %s, with tableName unset", report.Name)
		} else {
			// If it's started, query the API to get the most up to date resource,
			// as it's possible it's finished, but we haven't gotten it yet.
			newReport, err := op.meteringClient.MeteringV1alpha1().Reports(report.Namespace).Get(report.Name, metav1.GetOptions{})
			if err != nil {
				return err
			}

			if report.UID != newReport.UID {
				return fmt.Errorf("started report has different UUID in API than in cache, skipping processing until next reconcile")
			}

			err = op.reportInformer.Informer().GetIndexer().Update(newReport)
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
		}
	case cbTypes.ReportPhaseFinished, cbTypes.ReportPhaseError:
		logger.Infof("ignoring report %s, status: %s", report.Name, report.Status.Phase)
		return nil
	default:
		logger.Infof("new report discovered")
	}

	now := op.clock.Now()

	var gracePeriod time.Duration
	if report.Spec.GracePeriod != nil {
		gracePeriod = report.Spec.GracePeriod.Duration
	} else {
		gracePeriod = op.getDefaultReportGracePeriod()
		logger.Debugf("Report has no gracePeriod configured, falling back to defaultGracePeriod: %s", gracePeriod)
	}

	var reportingStart, reportingEnd *time.Time
	if report.Spec.ReportingStart != nil {
		reportingStart = &report.Spec.ReportingStart.Time
	}
	if report.Spec.ReportingEnd != nil {
		reportingEnd = &report.Spec.ReportingEnd.Time
	}

	if reportingEnd == nil {
		logger.Infof("report has no reportingEnd: running immediately")
	} else {
		var waitTime time.Duration
		nextRunTime := reportingEnd.Add(gracePeriod)
		reportGracePeriodUnmet := nextRunTime.After(now)
		waitTime = nextRunTime.Sub(now)

		if report.Spec.RunImmediately {
			logger.Infof("report configured to run immediately with %s until periodEnd+gracePeriod: %s", waitTime, nextRunTime)
		} else if reportGracePeriodUnmet {
			logger.Infof("report %s not past grace period yet, ignoring until %s (%s)", report.Name, nextRunTime, waitTime)
			op.enqueueReportAfter(report, waitTime)
			return nil
		}
	}

	logger = logger.WithField("generationQuery", report.Spec.GenerationQueryName)
	genQuery, err := op.reportGenerationQueryLister.ReportGenerationQueries(report.Namespace).Get(report.Spec.GenerationQueryName)
	if err != nil {
		logger.WithError(err).Errorf("failed to get report generation query")
		return err
	}

	queryDependencies, err := reporting.GetAndValidateGenerationQueryDependencies(
		reporting.NewReportGenerationQueryListerGetter(op.reportGenerationQueryLister),
		reporting.NewReportDataSourceListerGetter(op.reportDataSourceLister),
		reporting.NewReportListerGetter(op.reportLister),
		reporting.NewScheduledReportListerGetter(op.scheduledReportLister),
		genQuery,
		op.uninitialiedDependendenciesHandler(),
	)
	if err != nil {
		return fmt.Errorf("unable to run Report %s, ReportGenerationQuery %s, failed to validate dependencies: %v", report.Name, genQuery.Name, err)
	}

	logger.Debug("updating report status to started")
	// update status
	report.Status.Phase = cbTypes.ReportPhaseStarted
	report, err = op.meteringClient.MeteringV1alpha1().Reports(report.Namespace).Update(report)
	if err != nil {
		return fmt.Errorf("failed to update report status to started for %q", report.Name)
	}

	logger.Debugf("dropping table %s", tableName)
	err = hive.ExecuteDropTable(op.hiveQueryer, tableName, true)
	if err != nil {
		return fmt.Errorf("unable to drop table %s before creating for report %s: %v", tableName, report.Name, err)
	}

	columns := reportingutil.GenerateHiveColumns(genQuery)
	err = op.createTableForStorage(logger, report, cbTypes.SchemeGroupVersion.WithKind("Report"), report.Spec.Output, tableName, columns)
	if err != nil {
		return fmt.Errorf("unable to create table %s for report %s: %v", tableName, report.Name, err)
	}

	report.Status.TableName = tableName
	report, err = op.meteringClient.MeteringV1alpha1().Reports(report.Namespace).Update(report)
	if err != nil {
		return fmt.Errorf("failed to update report %s status.tableName to %s: %v", report.Name, tableName, err)
	}

	genReportTotalCounter.Inc()
	generateReportStart := op.clock.Now()
	err = op.reportGenerator.GenerateReport(
		tableName,
		reportingStart,
		reportingEnd,
		genQuery,
		queryDependencies.DynamicReportGenerationQueries,
		report.Spec.Inputs,
		true,
	)
	generateReportDuration := op.clock.Since(generateReportStart)
	genReportDurationObserver.Observe(float64(generateReportDuration.Seconds()))
	if err != nil {
		genReportFailedCounter.Inc()
		op.setReportError(logger, report, err, "report execution failed")
		return fmt.Errorf("failed to generateReport for Report %s, err: %v", report.Name, err)
	}

	// update status
	report.Status.Phase = cbTypes.ReportPhaseFinished
	_, err = op.meteringClient.MeteringV1alpha1().Reports(report.Namespace).Update(report)
	if err != nil {
		logger.WithError(err).Warnf("failed to update report status to finished for %q", report.Name)
	} else {
		logger.Infof("finished report %q", report.Name)
	}

	if err := op.queueDependentReportGenerationQueriesForReport(report); err != nil {
		logger.WithError(err).Errorf("error queuing ReportGenerationQuery dependents of Report %s", report.Name)
	}

	return nil
}

func (op *Reporting) setReportError(logger log.FieldLogger, report *cbTypes.Report, err error, errMsg string, errMsgArgs ...interface{}) {
	logger.WithField("Report", report.Name).WithError(err).Errorf(errMsg, errMsgArgs...)
	report.Status.Phase = cbTypes.ReportPhaseError
	report.Status.Output = err.Error()
	_, err = op.meteringClient.MeteringV1alpha1().Reports(report.Namespace).Update(report)
	if err != nil {
		logger.WithError(err).Errorf("unable to update report status to error")
	}
}

// queueDependentReportGenerationQueriesForReport will queue all ReportGenerationQueries in the namespace which have a dependency on the Report
func (op *Reporting) queueDependentReportGenerationQueriesForReport(report *cbTypes.Report) error {
	queryLister := op.meteringClient.MeteringV1alpha1().ReportGenerationQueries(report.Namespace)
	queries, err := queryLister.List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, query := range queries.Items {
		// look at the list Report of dependencies
		for _, dependency := range query.Spec.Reports {
			if dependency == report.Name {
				// this query depends on the Report passed in
				op.enqueueReportGenerationQuery(query)
				break
			}
		}
	}
	return nil
}
