package operator

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rob***REMOVED***g/cron"
	log "github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"

	cbTypes "github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1"
	cbutil "github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1/util"
	"github.com/operator-framework/operator-metering/pkg/operator/reporting"
	"github.com/operator-framework/operator-metering/pkg/util/slice"
)

const (
	scheduledReportFinalizer = cbTypes.GroupName + "/scheduledreport"
)

var (
	scheduledReportPrometheusMetricLabels = []string{"scheduledreport", "reportgenerationquery", "table_name"}

	generateScheduledReportTotalCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: prometheusMetricNamespace,
			Name:      "generate_scheduledreports_total",
			Help:      "Duration to generate a Report.",
		},
		scheduledReportPrometheusMetricLabels,
	)

	generateScheduledReportFailedCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: prometheusMetricNamespace,
			Name:      "generate_scheduledreports_failed_total",
			Help:      "Duration to generate a Report.",
		},
		scheduledReportPrometheusMetricLabels,
	)

	generateScheduledReportDurationHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: prometheusMetricNamespace,
			Name:      "generate_scheduledreport_duration_seconds",
			Help:      "Duration to generate a ScheduledReport.",
			Buckets:   []float64{60.0, 300.0, 600.0},
		},
		scheduledReportPrometheusMetricLabels,
	)
)

func init() {
	prometheus.MustRegister(generateScheduledReportFailedCounter)
	prometheus.MustRegister(generateScheduledReportTotalCounter)
	prometheus.MustRegister(generateScheduledReportDurationHistogram)
}

func (op *Reporting) runScheduledReportWorker() {
	logger := op.logger.WithField("component", "scheduledReportWorker")
	logger.Infof("ScheduledReport worker started")
	for op.processResource(logger, op.syncScheduledReport, "ScheduledReport", op.queues.scheduledReportQueue) {
	}
}

func (op *Reporting) syncScheduledReport(logger log.FieldLogger, key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		logger.WithError(err).Errorf("invalid resource key :%s", key)
		return nil
	}

	logger = logger.WithField("ScheduledReport", name)
	scheduledReport, err := op.informers.Metering().V1alpha1().ScheduledReports().Lister().ScheduledReports(namespace).Get(name)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Infof("ScheduledReport %s does not exist anymore, stopping and removing any running jobs for ScheduledReport", name)
			if exists := op.scheduledReportRunner.RemoveJob(name); exists {
				logger.Infof("stopped running jobs for ScheduledReport")
			}
			return nil
		}
		return err
	}

	if scheduledReport.DeletionTimestamp != nil {
		if exists := op.scheduledReportRunner.RemoveJob(name); exists {
			logger.Infof("stopped running jobs for ScheduledReport")
		}
		_, err = op.removeScheduledReportFinalizer(scheduledReport)
		return err
	}

	return op.handleScheduledReport(logger, scheduledReport)
}

type reportSchedule interface {
	// Return the next activation time, later than the given time.
	// Next is invoked initially, and then each time the job runs..
	Next(time.Time) time.Time
}

func getSchedule(reportSched cbTypes.ScheduledReportSchedule) (reportSchedule, error) {
	var cronSpec string
	switch reportSched.Period {
	case cbTypes.ScheduledReportPeriodCron:
		return cron.ParseStandard(reportSched.Cron.Expression)
	case cbTypes.ScheduledReportPeriodHourly:
		sched := reportSched.Hourly
		if sched == nil {
			sched = &cbTypes.ScheduledReportScheduleHourly{}
		}
		if err := validateMinute(sched.Minute); err != nil {
			return nil, err
		}
		if err := validateSecond(sched.Second); err != nil {
			return nil, err
		}
		cronSpec = fmt.Sprintf("%d %d * * * *", sched.Second, sched.Minute)
	case cbTypes.ScheduledReportPeriodDaily:
		sched := reportSched.Daily
		if sched == nil {
			sched = &cbTypes.ScheduledReportScheduleDaily{}
		}
		if err := validateHour(sched.Hour); err != nil {
			return nil, err
		}
		if err := validateMinute(sched.Minute); err != nil {
			return nil, err
		}
		if err := validateSecond(sched.Second); err != nil {
			return nil, err
		}
		cronSpec = fmt.Sprintf("%d %d %d * * *", sched.Second, sched.Minute, sched.Hour)
	case cbTypes.ScheduledReportPeriodWeekly:
		sched := reportSched.Weekly
		if sched == nil {
			sched = &cbTypes.ScheduledReportScheduleWeekly{}
		}
		dow := 0
		if sched.DayOfWeek != nil {
			var err error
			dow, err = convertDayOfWeek(*sched.DayOfWeek)
			if err != nil {
				return nil, err
			}
		}
		if err := validateHour(sched.Hour); err != nil {
			return nil, err
		}
		if err := validateMinute(sched.Minute); err != nil {
			return nil, err
		}
		if err := validateSecond(sched.Second); err != nil {
			return nil, err
		}
		cronSpec = fmt.Sprintf("%d %d %d * * %d", sched.Second, sched.Minute, sched.Hour, dow)
	case cbTypes.ScheduledReportPeriodMonthly:
		sched := reportSched.Monthly
		if sched == nil {
			sched = &cbTypes.ScheduledReportScheduleMonthly{}
		}
		dom := int64(1)
		if sched.DayOfMonth != nil {
			dom = *sched.DayOfMonth
		}
		if err := validateDayOfMonth(dom); err != nil {
			return nil, err
		}
		if err := validateHour(sched.Hour); err != nil {
			return nil, err
		}
		if err := validateMinute(sched.Minute); err != nil {
			return nil, err
		}
		if err := validateSecond(sched.Second); err != nil {
			return nil, err
		}
		cronSpec = fmt.Sprintf("%d %d %d %d * *", sched.Second, sched.Minute, sched.Hour, dom)
	default:
		return nil, fmt.Errorf("invalid ScheduledReport.spec.schedule.period: %s", reportSched.Period)
	}
	return cron.Parse(cronSpec)
}

func (op *Reporting) handleScheduledReport(logger log.FieldLogger, scheduledReport *cbTypes.ScheduledReport) error {
	scheduledReport = scheduledReport.DeepCopy()

	if op.cfg.EnableFinalizers && scheduledReportNeedsFinalizer(scheduledReport) {
		var err error
		scheduledReport, err = op.addScheduledReportFinalizer(scheduledReport)
		if err != nil {
			return err
		}
	}

	reportSchedule, err := getSchedule(scheduledReport.Spec.Schedule)
	if err != nil {
		return err
	}
	job := newScheduledReportJob(op, scheduledReport.Name, scheduledReport.Namespace, reportSchedule)
	op.scheduledReportRunner.AddJob(job)

	return nil
}

type scheduledReportJob struct {
	operator        *Reporting
	reportName      string
	reportNamespace string
	schedule        reportSchedule
	stopCh          chan struct{}
	doneCh          chan struct{}
}

func newScheduledReportJob(operator *Reporting, reportName, reportNamespace string, schedule reportSchedule) *scheduledReportJob {
	return &scheduledReportJob{
		operator:        operator,
		reportName:      reportName,
		reportNamespace: reportNamespace,
		schedule:        schedule,
		stopCh:          make(chan struct{}),
		doneCh:          make(chan struct{}),
	}
}

func (job *scheduledReportJob) stop() {
	logger := job.operator.logger.WithField("ScheduledReport", job.reportName)
	logger.Info("stopping ScheduledReport job")
	close(job.stopCh)
	// wait for start() to exit
	logger.Info("waiting for ScheduledReport job to ***REMOVED***nish")
	<-job.doneCh
}

type reportPeriod struct {
	periodEnd   time.Time
	periodStart time.Time
}

// start runs a scheduledReportJob according to it's con***REMOVED***gured schedule. It
// returns nothing because it should never stop unless Metering is shutting
// down or the scheduledReport for this job has been deleted.
func (job *scheduledReportJob) start(logger log.FieldLogger) error {
	// Close doneCh at the end so that stop() can determine when start() has exited.
	defer func() {
		close(job.doneCh)
	}()

	for {
		report, err := job.operator.meteringClient.MeteringV1alpha1().ScheduledReports(job.reportNamespace).Get(job.reportName, metav1.GetOptions{})
		if err != nil {
			logger.WithError(err).Errorf("unable to get scheduledReport")
			return err
		}
		report = report.DeepCopy()

		tableName := scheduledReportTableName(report.Name)
		metricLabels := prometheus.Labels{
			"scheduledreport":       report.Name,
			"reportgenerationquery": report.Spec.GenerationQueryName,
			"table_name":            tableName,
		}

		genReportTotalCounter := generateScheduledReportTotalCounter.With(metricLabels)
		genReportFailedCounter := generateScheduledReportFailedCounter.With(metricLabels)
		genReportDurationObserver := generateScheduledReportDurationHistogram.With(metricLabels)

		msg := fmt.Sprintf("Validating generationQuery %s", report.Spec.GenerationQueryName)
		runningCondition := cbutil.NewScheduledReportCondition(cbTypes.ScheduledReportRunning, v1.ConditionTrue, cbutil.ValidatingScheduledReportReason, msg)
		cbutil.SetScheduledReportCondition(&report.Status, *runningCondition)

		report, err = job.operator.meteringClient.MeteringV1alpha1().ScheduledReports(report.Namespace).Update(report)
		if err != nil {
			logger.WithError(err).Errorf("unable to update ScheduledReport status")
			return err
		}

		genQuery, err := job.operator.informers.Metering().V1alpha1().ReportGenerationQueries().Lister().ReportGenerationQueries(report.Namespace).Get(report.Spec.GenerationQueryName)
		if err != nil {
			logger.WithError(err).Errorf("failed to get report generation query")
			return err
		}

		reportGenerationQueryLister := job.operator.informers.Metering().V1alpha1().ReportGenerationQueries().Lister()
		reportDataSourceLister := job.operator.informers.Metering().V1alpha1().ReportDataSources().Lister()

		depsStatus, err := reporting.GetGenerationQueryDependenciesStatus(
			reporting.NewReportGenerationQueryListerGetter(reportGenerationQueryLister),
			reporting.NewReportDataSourceListerGetter(reportDataSourceLister),
			genQuery,
		)
		if err != nil {
			logger.Errorf("failed to get dependencies for ScheduledReport %s, err: %v", job.reportName, err)
		}

		_, err = job.operator.validateDependencyStatus(depsStatus)
		if err != nil {
			logger.Errorf("failed to validate dependencies for ScheduledReport %s, err: %v", job.reportName, err)
			return err
		}

		columns := generateHiveColumns(genQuery)
		err = job.operator.createTableForStorage(logger, report, cbTypes.SchemeGroupVersion.WithKind("ScheduledReport"), report.Spec.Output, tableName, columns)
		if err != nil {
			logger.WithError(err).Error("error creating report table for scheduledReport")
			return err
		}

		report.Status.TableName = tableName
		report, err = job.operator.meteringClient.MeteringV1alpha1().ScheduledReports(report.Namespace).Update(report)
		if err != nil {
			logger.WithError(err).Errorf("unable to update ScheduledReport status with tableName")
			return err
		}

		now := job.operator.clock.Now().UTC()
		var lastScheduled time.Time
		lastReportTime := report.Status.LastReportTime
		if lastReportTime != nil {
			logger = logger.WithField("lastReportTime", lastReportTime.Time)
			logger.Infof("last report time was %s", lastReportTime.Time)
			lastScheduled = lastReportTime.Time
		} ***REMOVED*** {
			lastScheduled = now
		}

		reportPeriod := getNextReportPeriod(job.schedule, report.Spec.Schedule.Period, lastScheduled)

		loggerWithFields := logger.WithFields(log.Fields{
			"periodStart":       reportPeriod.periodStart,
			"periodEnd":         reportPeriod.periodEnd,
			"period":            report.Spec.Schedule.Period,
			"overwriteExisting": report.Spec.OverwriteExistingData,
		})

		var gracePeriod time.Duration
		if report.Spec.GracePeriod != nil {
			gracePeriod = report.Spec.GracePeriod.Duration
		} ***REMOVED*** {
			gracePeriod = job.operator.getDefaultReportGracePeriod()
			loggerWithFields.Debugf("ScheduledReport has no gracePeriod con***REMOVED***gured, falling back to defaultGracePeriod: %s", gracePeriod)
		}

		var waitTime time.Duration
		nextRunTime := reportPeriod.periodEnd.Add(gracePeriod)
		reportGracePeriodUnmet := nextRunTime.After(now)
		if reportGracePeriodUnmet {
			waitTime = nextRunTime.Sub(now)
		}

		waitMsg := fmt.Sprintf("next scheduled report period is [%s to %s] with gracePeriod: %s. next run time is %s", reportPeriod.periodStart, reportPeriod.periodEnd, gracePeriod, nextRunTime)
		loggerWithFields.Infof(waitMsg+". waiting %s", waitTime)

		runningCondition = cbutil.NewScheduledReportCondition(cbTypes.ScheduledReportRunning, v1.ConditionTrue, cbutil.ReportPeriodWaitingReason, waitMsg)
		cbutil.SetScheduledReportCondition(&report.Status, *runningCondition)

		report, err = job.operator.meteringClient.MeteringV1alpha1().ScheduledReports(report.Namespace).Update(report)
		if err != nil {
			loggerWithFields.WithError(err).Errorf("unable to update ScheduledReport status")
			return err
		}

		select {
		case <-job.stopCh:
			loggerWithFields.Info("got stop signal, stopping ScheduledReport job")
			return nil
		case <-job.operator.clock.After(waitTime):
			runningMsg := fmt.Sprintf("reached end of last reporting period [%s to %s]", reportPeriod.periodStart, reportPeriod.periodEnd)
			runningCondition := cbutil.NewScheduledReportCondition(cbTypes.ScheduledReportRunning, v1.ConditionTrue, cbutil.ScheduledReason, runningMsg)
			cbutil.SetScheduledReportCondition(&report.Status, *runningCondition)

			report, err = job.operator.meteringClient.MeteringV1alpha1().ScheduledReports(report.Namespace).Update(report)
			if err != nil {
				loggerWithFields.WithError(err).Errorf("unable to update ScheduledReport status")
				return err
			}
			genReportTotalCounter.Inc()
			generateReportStart := job.operator.clock.Now()
			err = job.operator.generateReport(
				loggerWithFields,
				report,
				"scheduledreport",
				report.Name,
				tableName,
				reportPeriod.periodStart,
				reportPeriod.periodEnd,
				genQuery,
				report.Spec.OverwriteExistingData,
			)
			generateReportDuration := job.operator.clock.Since(generateReportStart)
			genReportDurationObserver.Observe(float64(generateReportDuration.Seconds()))

			if err != nil {
				genReportFailedCounter.Inc()
				// update the status to Failed with message containing the
				// error
				errMsg := fmt.Sprintf("error occurred while generating report: %s", err)
				failureCondition := cbutil.NewScheduledReportCondition(cbTypes.ScheduledReportFailure, v1.ConditionTrue, cbutil.GenerateReportErrorReason, errMsg)
				cbutil.RemoveScheduledReportCondition(&report.Status, cbTypes.ScheduledReportRunning)
				cbutil.SetScheduledReportCondition(&report.Status, *failureCondition)

				_, updateErr := job.operator.meteringClient.MeteringV1alpha1().ScheduledReports(report.Namespace).Update(report)
				if updateErr != nil {
					loggerWithFields.WithError(updateErr).Errorf("unable to update ScheduledReport status")
					return updateErr
				}
				loggerWithFields.WithError(err).Errorf("error occurred while generating report")
				return err
			}

			// We generated a report successfully, remove the failure condition
			cbutil.RemoveScheduledReportCondition(&report.Status, cbTypes.ScheduledReportFailure)
			report.Status.LastReportTime = &metav1.Time{Time: reportPeriod.periodEnd}
			_, err = job.operator.meteringClient.MeteringV1alpha1().ScheduledReports(report.Namespace).Update(report)
			if err != nil {
				loggerWithFields.WithError(err).Errorf("unable to update ScheduledReport status")
				return err
			}
		}
	}
}

type scheduledReportRunner struct {
	reportsMu sync.Mutex
	reports   map[string]*scheduledReportJob
	jobsChan  chan *scheduledReportRunnerAddJob
	wg        sync.WaitGroup
	operator  *Reporting
}

func newScheduledReportRunner(operator *Reporting) *scheduledReportRunner {
	return &scheduledReportRunner{
		reports:  make(map[string]*scheduledReportJob),
		jobsChan: make(chan *scheduledReportRunnerAddJob),
		operator: operator,
	}
}

func (runner *scheduledReportRunner) Run(stop <-chan struct{}) {
	defer runner.wg.Wait()
	for {
		select {
		case <-stop:
			return
		case addJob := <-runner.jobsChan:
			runner.wg.Add(1)
			go func() {
				addJob.err <- runner.handleJob(stop, addJob.job)
				runner.wg.Done()
			}()
		}
	}
}

type scheduledReportRunnerAddJob struct {
	job *scheduledReportJob
	err chan error
}

func (runner *scheduledReportRunner) AddJob(job *scheduledReportJob) error {
	errCh := make(chan error)
	defer close(errCh)
	runner.jobsChan <- &scheduledReportRunnerAddJob{
		job: job,
		err: errCh,
	}
	return <-errCh
}

func (runner *scheduledReportRunner) RemoveJob(name string) bool {
	runner.reportsMu.Lock()
	defer runner.reportsMu.Unlock()
	job, exists := runner.reports[name]
	if exists {
		job.stop()
		delete(runner.reports, name)
	}
	return exists
}

func (runner *scheduledReportRunner) handleJob(stop <-chan struct{}, job *scheduledReportJob) error {
	logger := runner.operator.logger.WithField("ScheduledReport", job.reportName)
	runner.reportsMu.Lock()
	_, exists := runner.reports[job.reportName]
	if exists {
		runner.reportsMu.Unlock()
		logger.Info("scheduled report is already being ran, updates to scheduled report not currently supported")
		return nil
	}

	runner.reports[job.reportName] = job
	runner.reportsMu.Unlock()

	logger.Info("starting ScheduledReport job")
	var wg sync.WaitGroup
	errCh := make(chan error)
	wg.Add(1)

	defer func() {
		runner.RemoveJob(job.reportName)
		wg.Wait()
		close(errCh)
		logger.Info("ScheduledReport job stopped")
	}()

	go func() {
		select {
		case errCh <- job.start(logger):
		case <-stop:
			errCh <- fmt.Errorf("ScheduledReport job got shutdown signal")
		}
	}()

	return <-errCh
}

func getNextReportPeriod(schedule reportSchedule, period cbTypes.ScheduledReportPeriod, lastScheduled time.Time) reportPeriod {
	periodStart := lastScheduled
	periodEnd := schedule.Next(periodStart)
	return reportPeriod{
		periodEnd:   periodEnd.Truncate(time.Millisecond).UTC(),
		periodStart: periodStart.Truncate(time.Millisecond).UTC(),
	}
}

func convertDayOfWeek(dow string) (int, error) {
	switch strings.ToLower(dow) {
	case "sun", "sunday":
		return 0, nil
	case "mon", "monday":
		return 1, nil
	case "tue", "tues", "tuesday":
		return 2, nil
	case "wed", "weds", "wednesday":
		return 3, nil
	case "thur", "thurs", "thursday":
		return 4, nil
	case "fri", "friday":
		return 5, nil
	case "sat", "saturday":
		return 6, nil
	}
	return 0, fmt.Errorf("invalid day of week: %s", dow)
}

func (op *Reporting) addScheduledReportFinalizer(report *cbTypes.ScheduledReport) (*cbTypes.ScheduledReport, error) {
	report.Finalizers = append(report.Finalizers, scheduledReportFinalizer)
	newScheduledReport, err := op.meteringClient.MeteringV1alpha1().ScheduledReports(report.Namespace).Update(report)
	logger := op.logger.WithField("ScheduledReport", report.Name)
	if err != nil {
		logger.WithError(err).Errorf("error adding %s ***REMOVED***nalizer to ScheduledReport: %s/%s", scheduledReportFinalizer, report.Namespace, report.Name)
		return nil, err
	}
	logger.Infof("added %s ***REMOVED***nalizer to ScheduledReport: %s/%s", scheduledReportFinalizer, report.Namespace, report.Name)
	return newScheduledReport, nil
}

func (op *Reporting) removeScheduledReportFinalizer(report *cbTypes.ScheduledReport) (*cbTypes.ScheduledReport, error) {
	if !slice.ContainsString(report.ObjectMeta.Finalizers, scheduledReportFinalizer, nil) {
		return report, nil
	}
	report.Finalizers = slice.RemoveString(report.Finalizers, scheduledReportFinalizer, nil)
	newScheduledReport, err := op.meteringClient.MeteringV1alpha1().ScheduledReports(report.Namespace).Update(report)
	logger := op.logger.WithField("ScheduledReport", report.Name)
	if err != nil {
		logger.WithError(err).Errorf("error removing %s ***REMOVED***nalizer from ScheduledReport: %s/%s", scheduledReportFinalizer, report.Namespace, report.Name)
		return nil, err
	}
	logger.Infof("removed %s ***REMOVED***nalizer from ScheduledReport: %s/%s", scheduledReportFinalizer, report.Namespace, report.Name)
	return newScheduledReport, nil
}

func scheduledReportNeedsFinalizer(report *cbTypes.ScheduledReport) bool {
	return report.ObjectMeta.DeletionTimestamp == nil && !slice.ContainsString(report.ObjectMeta.Finalizers, scheduledReportFinalizer, nil)
}
