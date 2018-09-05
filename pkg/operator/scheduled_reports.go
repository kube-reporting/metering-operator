package operator

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/rob***REMOVED***g/cron"
	log "github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"

	cbTypes "github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1"
	cbutil "github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1/util"
	"github.com/operator-framework/operator-metering/pkg/hive"
	"github.com/operator-framework/operator-metering/pkg/operator/reporting"
)

func (op *Reporting) runScheduledReportWorker() {
	logger := op.logger.WithField("component", "scheduledReportWorker")
	logger.Infof("ScheduledReport worker started")
	for op.processScheduledReport(logger) {

	}
}

func (op *Reporting) processScheduledReport(logger log.FieldLogger) bool {
	obj, quit := op.queues.scheduledReportQueue.Get()
	if quit {
		logger.Infof("queue is shutting down, exiting ScheduledReport worker")
		return false
	}
	defer op.queues.scheduledReportQueue.Done(obj)

	logger = logger.WithFields(newLogIdenti***REMOVED***er(op.rand))
	if key, ok := op.getKeyFromQueueObj(logger, "ScheduledReport", obj, op.queues.scheduledReportQueue); ok {
		err := op.syncScheduledReport(logger, key)
		op.handleErr(logger, err, "ScheduledReport", obj, op.queues.scheduledReportQueue)
	}
	return true
}

func (op *Reporting) syncScheduledReport(logger log.FieldLogger, key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		logger.WithError(err).Errorf("invalid resource key :%s", key)
		return nil
	}

	logger = logger.WithField("scheduledReport", name)
	scheduledReport, err := op.informers.Metering().V1alpha1().ScheduledReports().Lister().ScheduledReports(namespace).Get(name)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Infof("ScheduledReport %s does not exist anymore, stopping and removing any running jobs for ScheduledReport", name)
			if job, exists := op.scheduledReportRunner.RemoveJob(name); exists {
				job.stop(true)
				logger.Infof("stopped running jobs for ScheduledReport")
			}
			return nil
		}
		return err
	}

	logger.Infof("syncing scheduledReport %s", scheduledReport.GetName())
	err = op.handleScheduledReport(logger, scheduledReport)
	if err != nil {
		logger.WithError(err).Errorf("error syncing scheduledReport %s", scheduledReport.GetName())
		return err
	}
	logger.Infof("successfully synced scheduledReport %s", scheduledReport.GetName())
	return nil
}

func (op *Reporting) handleScheduledReportDeleted(obj interface{}) {
	report, ok := obj.(*cbTypes.ScheduledReport)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			op.logger.Errorf("Couldn't get object from tombstone %#v", obj)
			return
		}
		report, ok = tombstone.Obj.(*cbTypes.ScheduledReport)
		if !ok {
			op.logger.Errorf("Tombstone contained object that is not a ScheduledReport %#v", obj)
			return
		}
	}
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(report)
	if err != nil {
		op.logger.WithField("scheduledReport", report.Name).WithError(err).Errorf("couldn't get key for object: %#v", report)
		return
	}
	op.queues.scheduledReportQueue.Add(key)
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
	reportSchedule, err := getSchedule(scheduledReport.Spec.Schedule)
	if err != nil {
		return err
	}
	job := newScheduledReportJob(op, scheduledReport, reportSchedule)
	op.scheduledReportRunner.AddJob(job)

	return nil
}

type scheduledReportJob struct {
	operator *Reporting
	report   *cbTypes.ScheduledReport
	schedule reportSchedule
	once     sync.Once
	stopCh   chan struct{}
	doneCh   chan struct{}
}

func newScheduledReportJob(operator *Reporting, report *cbTypes.ScheduledReport, schedule reportSchedule) *scheduledReportJob {
	return &scheduledReportJob{
		operator: operator,
		report:   report,
		schedule: schedule,
		stopCh:   make(chan struct{}),
		doneCh:   make(chan struct{}),
	}
}

func (job *scheduledReportJob) stop(dropTable bool) {
	logger := job.operator.logger.WithField("scheduledReport", job.report.Name)
	job.once.Do(func() {
		logger.Info("stopping scheduledReport job")
		close(job.stopCh)
		// wait for start() to exit
		logger.Info("waiting for scheduledReport job to ***REMOVED***nish")
		<-job.doneCh
		if dropTable {
			tableName := scheduledReportTableName(job.report.Name)
			logger.Infof("deleting scheduledReport table %s", tableName)
			err := hive.ExecuteDropTable(job.operator.hiveQueryer, tableName, true)
			if err != nil {
				job.operator.logger.WithError(err).Error("unable to drop table")
			}
			job.operator.logger.Infof("successfully deleted table %s", tableName)
		}
	})
}

type reportPeriod struct {
	periodEnd   time.Time
	periodStart time.Time
}

// start runs a scheduledReportJob according to it's con***REMOVED***gured schedule. It
// returns nothing because it should never stop unless Metering is shutting
// down or the scheduledReport for this job has been deleted.
func (job *scheduledReportJob) start(logger log.FieldLogger) {
	// Close doneCh at the end so that stop() can determine when start() has exited.
	defer func() {
		close(job.doneCh)
	}()

	for {
		report, err := job.operator.meteringClient.MeteringV1alpha1().ScheduledReports(job.report.Namespace).Get(job.report.Name, metav1.GetOptions{})
		if err != nil {
			logger.WithError(err).Errorf("unable to get scheduledReport")
			return
		}
		report = report.DeepCopy()

		msg := fmt.Sprintf("Validating generationQuery %s", job.report.Spec.GenerationQueryName)
		runningCondition := cbutil.NewScheduledReportCondition(cbTypes.ScheduledReportRunning, v1.ConditionTrue, cbutil.ValidatingScheduledReportReason, msg)
		cbutil.SetScheduledReportCondition(&report.Status, *runningCondition)

		report, err = job.operator.meteringClient.MeteringV1alpha1().ScheduledReports(job.report.Namespace).Update(report)
		if err != nil {
			logger.WithError(err).Errorf("unable to update scheduledReport status")
			return
		}

		genQuery, err := job.operator.informers.Metering().V1alpha1().ReportGenerationQueries().Lister().ReportGenerationQueries(job.report.Namespace).Get(job.report.Spec.GenerationQueryName)
		if err != nil {
			logger.WithError(err).Errorf("failed to get report generation query")
			return
		}

		reportGenerationQueryLister := job.operator.informers.Metering().V1alpha1().ReportGenerationQueries().Lister()
		reportDataSourceLister := job.operator.informers.Metering().V1alpha1().ReportDataSources().Lister()

		depsStatus, err := reporting.GetGenerationQueryDependenciesStatus(
			reporting.NewReportGenerationQueryListerGetter(reportGenerationQueryLister),
			reporting.NewReportDataSourceListerGetter(reportDataSourceLister),
			genQuery,
		)
		if err != nil {
			logger.Errorf("failed to get dependencies for ScheduledReport %s, err: %v", job.report.Name, err)
			return
		}

		_, err = job.operator.validateDependencyStatus(depsStatus)
		if err != nil {
			logger.Errorf("failed to validate dependencies for ScheduledReport %s, err: %v", job.report.Name, err)
			return
		}

		tableName := scheduledReportTableName(job.report.Name)
		columns := generateHiveColumns(genQuery)
		err = job.operator.createTableForStorage(logger, job.report, "scheduledreport", job.report.Name, job.report.Spec.Output, tableName, columns)
		if err != nil {
			logger.WithError(err).Error("error creating report table for scheduledReport")
			return
		}

		report.Status.TableName = tableName
		report, err = job.operator.meteringClient.MeteringV1alpha1().ScheduledReports(job.report.Namespace).Update(report)
		if err != nil {
			logger.WithError(err).Errorf("unable to update scheduledReport status with tableName")
			return
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

		reportPeriod := getNextReportPeriod(job.schedule, job.report.Spec.Schedule.Period, lastScheduled)

		loggerWithFields := logger.WithFields(log.Fields{
			"periodStart":       reportPeriod.periodStart,
			"periodEnd":         reportPeriod.periodEnd,
			"period":            job.report.Spec.Schedule.Period,
			"overwriteExisting": job.report.Spec.OverwriteExistingData,
		})

		var gracePeriod time.Duration
		if job.report.Spec.GracePeriod != nil {
			gracePeriod = job.report.Spec.GracePeriod.Duration
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

		report, err = job.operator.meteringClient.MeteringV1alpha1().ScheduledReports(job.report.Namespace).Update(report)
		if err != nil {
			loggerWithFields.WithError(err).Errorf("unable to update scheduledReport status")
			return
		}

		select {
		case <-job.stopCh:
			loggerWithFields.Info("got stop signal, stopping scheduledReport job")
			return
		case <-job.operator.clock.After(waitTime):
			runningMsg := fmt.Sprintf("reached end of last reporting period [%s to %s]", reportPeriod.periodStart, reportPeriod.periodEnd)
			runningCondition := cbutil.NewScheduledReportCondition(cbTypes.ScheduledReportRunning, v1.ConditionTrue, cbutil.ScheduledReason, runningMsg)
			cbutil.SetScheduledReportCondition(&report.Status, *runningCondition)

			report, err = job.operator.meteringClient.MeteringV1alpha1().ScheduledReports(job.report.Namespace).Update(report)
			if err != nil {
				loggerWithFields.WithError(err).Errorf("unable to update scheduledReport status")
				return
			}

			err = job.operator.generateReport(
				loggerWithFields,
				job.report,
				"scheduledreport",
				job.report.Name,
				tableName,
				reportPeriod.periodStart,
				reportPeriod.periodEnd,
				genQuery,
				job.report.Spec.OverwriteExistingData,
			)

			if err != nil {
				// update the status to Failed with message containing the
				// error
				errMsg := fmt.Sprintf("error occurred while generating report: %s", err)
				failureCondition := cbutil.NewScheduledReportCondition(cbTypes.ScheduledReportFailure, v1.ConditionTrue, cbutil.GenerateReportErrorReason, errMsg)
				cbutil.RemoveScheduledReportCondition(&report.Status, cbTypes.ScheduledReportRunning)
				cbutil.SetScheduledReportCondition(&report.Status, *failureCondition)

				_, updateErr := job.operator.meteringClient.MeteringV1alpha1().ScheduledReports(job.report.Namespace).Update(report)
				if updateErr != nil {
					loggerWithFields.WithError(updateErr).Errorf("unable to update scheduledReport status")
				}
				loggerWithFields.WithError(err).Errorf("error occurred while generating report")
				return
			}

			// We generated a report successfully, remove the failure condition
			cbutil.RemoveScheduledReportCondition(&report.Status, cbTypes.ScheduledReportFailure)
			report.Status.LastReportTime = &metav1.Time{Time: reportPeriod.periodEnd}
			_, err = job.operator.meteringClient.MeteringV1alpha1().ScheduledReports(job.report.Namespace).Update(report)
			if err != nil {
				loggerWithFields.WithError(err).Errorf("unable to update scheduledReport status")
				return
			}
		}
	}
}

type scheduledReportRunner struct {
	reportsMu sync.Mutex
	reports   map[string]*scheduledReportJob
	jobsChan  chan *scheduledReportJob
	wg        sync.WaitGroup
	operator  *Reporting
}

func newScheduledReportRunner(operator *Reporting) *scheduledReportRunner {
	return &scheduledReportRunner{
		reports:  make(map[string]*scheduledReportJob),
		jobsChan: make(chan *scheduledReportJob),
		operator: operator,
	}
}

func (runner *scheduledReportRunner) Run(stop <-chan struct{}) {
	defer runner.wg.Wait()
	for {
		select {
		case <-stop:
			return
		case job := <-runner.jobsChan:
			runner.wg.Add(1)
			go func() {
				runner.handleJob(stop, job)
				runner.wg.Done()
			}()
		}
	}
}

func (runner *scheduledReportRunner) AddJob(job *scheduledReportJob) {
	runner.jobsChan <- job
}

func (runner *scheduledReportRunner) RemoveJob(name string) (*scheduledReportJob, bool) {
	runner.reportsMu.Lock()
	defer runner.reportsMu.Unlock()
	job, exists := runner.reports[name]
	if exists {
		delete(runner.reports, name)
	}
	return job, exists
}

func (runner *scheduledReportRunner) handleJob(stop <-chan struct{}, job *scheduledReportJob) {
	logger := runner.operator.logger.WithField("scheduledReport", job.report.Name)
	runner.reportsMu.Lock()
	_, exists := runner.reports[job.report.Name]
	if exists {
		runner.reportsMu.Unlock()
		logger.Info("scheduled report is already being ran, updates to scheduled report not currently supported")
		return
	}

	runner.reports[job.report.Name] = job
	runner.reportsMu.Unlock()

	logger.Info("starting scheduledReport job")
	defer runner.RemoveJob(job.report.Name)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		job.start(logger)
		wg.Done()
	}()
	go func() {
		// when stop is closed, stop the running job
		<-stop
		job.stop(false)
	}()
	wg.Wait()
	logger.Info("scheduledReport job stopped")

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
