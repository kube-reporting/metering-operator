package chargeback

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/robfig/cron"
	log "github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"

	cbTypes "github.com/operator-framework/operator-metering/pkg/apis/chargeback/v1alpha1"
	cbutil "github.com/operator-framework/operator-metering/pkg/apis/chargeback/v1alpha1/util"
	"github.com/operator-framework/operator-metering/pkg/hive"
)

func (c *Chargeback) runScheduledReportWorker() {
	logger := c.logger.WithField("component", "scheduledReportWorker")
	logger.Infof("ScheduledReport worker started")
	for c.processScheduledReport(logger) {

	}
}

func (c *Chargeback) processScheduledReport(logger log.FieldLogger) bool {
	if c.queues.scheduledReportQueue.ShuttingDown() {
		logger.Infof("queue is shutting down")
	}
	obj, quit := c.queues.scheduledReportQueue.Get()
	if quit {
		logger.Infof("queue is shutting down, exiting worker")
		return false
	}
	defer c.queues.scheduledReportQueue.Done(obj)

	logger = logger.WithFields(c.newLogIdentifier())
	if key, ok := c.getKeyFromQueueObj(logger, "ScheduledReport", obj, c.queues.scheduledReportQueue); ok {
		err := c.syncScheduledReport(logger, key)
		c.handleErr(logger, err, "ScheduledReport", obj, c.queues.scheduledReportQueue)
	}
	return true
}

func (c *Chargeback) syncScheduledReport(logger log.FieldLogger, key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		logger.WithError(err).Errorf("invalid resource key :%s", key)
		return nil
	}

	logger = logger.WithField("scheduledReport", name)
	scheduledReport, err := c.informers.Chargeback().V1alpha1().ScheduledReports().Lister().ScheduledReports(namespace).Get(name)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Infof("ScheduledReport %s does not exist anymore, stopping and removing any running jobs for ScheduledReport", name)
			if job, exists := c.scheduledReportRunner.RemoveJob(name); exists {
				job.stop(true)
				logger.Infof("stopped running jobs for ScheduledReport")
			}
			return nil
		}
		return err
	}

	logger.Infof("syncing scheduledReport %s", scheduledReport.GetName())
	err = c.handleScheduledReport(logger, scheduledReport)
	if err != nil {
		logger.WithError(err).Errorf("error syncing scheduledReport %s", scheduledReport.GetName())
		return err
	}
	logger.Infof("successfully synced scheduledReport %s", scheduledReport.GetName())
	return nil
}

func (c *Chargeback) handleScheduledReportDeleted(obj interface{}) {
	report, ok := obj.(*cbTypes.ScheduledReport)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			c.logger.Errorf("Couldn't get object from tombstone %#v", obj)
			return
		}
		report, ok = tombstone.Obj.(*cbTypes.ScheduledReport)
		if !ok {
			c.logger.Errorf("Tombstone contained object that is not a ScheduledReport %#v", obj)
			return
		}
	}
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(report)
	if err != nil {
		c.logger.WithField("scheduledReport", report.Name).WithError(err).Errorf("couldn't get key for object: %#v", report)
		return
	}
	c.queues.scheduledReportQueue.Add(key)
}

type reportSchedule interface {
	// Return the next activation time, later than the given time.
	// Next is invoked initially, and then each time the job runs..
	Next(time.Time) time.Time
}

func getSchedule(reportSched cbTypes.ScheduledReportSchedule) (reportSchedule, error) {
	var cronSpec string
	switch reportSched.Period {
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
		cronSpec = fmt.Sprintf("%d %d %d * * %d", sched.Second, sched.Minute, sched.Second, dow)
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
		cronSpec = fmt.Sprintf("%d %d %d %d * *", sched.Second, sched.Minute, sched.Second, dom)
	default:
		return nil, fmt.Errorf("invalid ScheduledReport.spec.schedule.period: %s", reportSched.Period)
	}
	return cron.Parse(cronSpec)
}

func (c *Chargeback) handleScheduledReport(logger log.FieldLogger, scheduledReport *cbTypes.ScheduledReport) error {
	scheduledReport = scheduledReport.DeepCopy()
	reportSchedule, err := getSchedule(scheduledReport.Spec.Schedule)
	if err != nil {
		return err
	}
	job := newScheduledReportJob(c, scheduledReport, reportSchedule)
	c.scheduledReportRunner.AddJob(job)

	return nil
}

type scheduledReportJob struct {
	chargeback *Chargeback
	report     *cbTypes.ScheduledReport
	schedule   reportSchedule
	once       sync.Once
	stopCh     chan struct{}
	doneCh     chan struct{}
}

func newScheduledReportJob(chargeback *Chargeback, report *cbTypes.ScheduledReport, schedule reportSchedule) *scheduledReportJob {
	return &scheduledReportJob{
		chargeback: chargeback,
		report:     report,
		schedule:   schedule,
		stopCh:     make(chan struct{}),
		doneCh:     make(chan struct{}),
	}
}

func (job *scheduledReportJob) stop(dropTable bool) {
	logger := job.chargeback.logger.WithField("scheduledReport", job.report.Name)
	job.once.Do(func() {
		logger.Info("stopping scheduledReport job")
		close(job.stopCh)
		// wait for start() to exit
		logger.Info("waiting for scheduledReport job to finish")
		<-job.doneCh
		if dropTable {
			tableName := scheduledReportTableName(job.report.Name)
			logger.Infof("deleting scheduledReport table %s", tableName)
			err := hive.ExecuteDropTable(job.chargeback.hiveQueryer, tableName, true)
			if err != nil {
				job.chargeback.logger.WithError(err).Error("unable to drop table")
			}
			job.chargeback.logger.Infof("successfully deleted table %s", tableName)
		}
	})
}

type reportPeriod struct {
	periodEnd   time.Time
	periodStart time.Time
}

// start runs a scheduledReportJob according to it's configured schedule. It
// returns nothing because it should never stop unless Chargeback is shutting
// down or the scheduledReport for this job has been deleted.
func (job *scheduledReportJob) start(logger log.FieldLogger) {
	// Close doneCh at the end so that stop() can determine when start() has exited.
	defer func() {
		close(job.doneCh)
	}()

	for {
		report, err := job.chargeback.chargebackClient.ChargebackV1alpha1().ScheduledReports(job.report.Namespace).Get(job.report.Name, metav1.GetOptions{})
		if err != nil {
			logger.WithError(err).Errorf("unable to get scheduledReport")
			return
		}
		report = report.DeepCopy()

		msg := fmt.Sprintf("Validating generationQuery %s", job.report.Spec.GenerationQueryName)
		runningCondition := cbutil.NewScheduledReportCondition(cbTypes.ScheduledReportRunning, v1.ConditionTrue, cbutil.ValidatingScheduledReportReason, msg)
		cbutil.SetScheduledReportCondition(&report.Status, *runningCondition)

		report, err = job.chargeback.chargebackClient.ChargebackV1alpha1().ScheduledReports(job.report.Namespace).Update(report)
		if err != nil {
			logger.WithError(err).Errorf("unable to update scheduledReport status")
			return
		}

		genQuery, err := job.chargeback.informers.Chargeback().V1alpha1().ReportGenerationQueries().Lister().ReportGenerationQueries(job.report.Namespace).Get(job.report.Spec.GenerationQueryName)
		if err != nil {
			logger.WithError(err).Errorf("failed to get report generation query")
			return
		}

		if valid, err := job.chargeback.validateGenerationQuery(logger, genQuery, true); err != nil {
			logger.WithError(err).Errorf("invalid report generation query for scheduled report %s", job.report.Name)
			return
		} else if !valid {
			logger.Warnf("cannot generate report, it has uninitialized dependencies")
			return
		}

		tableName := scheduledReportTableName(job.report.Name)
		columns := generateHiveColumns(genQuery)
		err = job.chargeback.createTableForStorage(logger, job.report, "scheduledreport", job.report.Name, job.report.Spec.Output, tableName, columns, false)
		if err != nil {
			logger.WithError(err).Error("error creating report table for scheduledReport")
			return
		}

		now := job.chargeback.clock.Now()
		var lastScheduled time.Time
		lastReportTime := report.Status.LastReportTime
		if lastReportTime != nil {
			logger = logger.WithField("lastReportTime", lastReportTime.Time)
			logger.Infof("last report time was %s", lastReportTime.Time)
			lastScheduled = lastReportTime.Time
		} else {
			lastScheduled = now
		}

		reportPeriod, err := getNextReportPeriod(job.schedule, job.report.Spec.Schedule.Period, lastScheduled)
		if err != nil {
			logger.WithError(err).Error("to get next report period for scheduledReport")
			return
		}

		loggerWithFields := logger.WithFields(log.Fields{
			"periodStart": reportPeriod.periodStart,
			"periodEnd":   reportPeriod.periodEnd,
			"period":      job.report.Spec.Schedule.Period,
		})

		var waitTime, gracePeriod time.Duration
		if job.report.Spec.GracePeriod != nil {
			gracePeriod = job.report.Spec.GracePeriod.Duration
		}

		nextRunTime := reportPeriod.periodEnd.Add(gracePeriod)
		if nextRunTime.After(now) {
			waitTime = nextRunTime.Sub(now)
		}

		waitMsg := fmt.Sprintf("next scheduled report period is [%s to %s] and has %s until next report period start and will run at %s (gracePeriod: %s)", reportPeriod.periodStart, reportPeriod.periodEnd, waitTime, nextRunTime, gracePeriod)
		loggerWithFields.Info(waitMsg)

		runningCondition = cbutil.NewScheduledReportCondition(cbTypes.ScheduledReportRunning, v1.ConditionTrue, cbutil.ReportPeriodNotFinishedReason, waitMsg)
		cbutil.SetScheduledReportCondition(&report.Status, *runningCondition)

		report, err = job.chargeback.chargebackClient.ChargebackV1alpha1().ScheduledReports(job.report.Namespace).Update(report)
		if err != nil {
			loggerWithFields.WithError(err).Errorf("unable to update scheduledReport status")
			return
		}

		select {
		case <-job.stopCh:
			loggerWithFields.Info("got stop signal, stopping scheduledReport job")
			return
		case <-job.chargeback.clock.After(waitTime):
			runningMsg := fmt.Sprintf("reached end of last reporting period [%s to %s]", reportPeriod.periodStart, reportPeriod.periodEnd)
			runningCondition := cbutil.NewScheduledReportCondition(cbTypes.ScheduledReportRunning, v1.ConditionTrue, cbutil.ScheduledReason, runningMsg)
			cbutil.SetScheduledReportCondition(&report.Status, *runningCondition)

			report, err = job.chargeback.chargebackClient.ChargebackV1alpha1().ScheduledReports(job.report.Namespace).Update(report)
			if err != nil {
				loggerWithFields.WithError(err).Errorf("unable to update scheduledReport status")
				return
			}

			_, err = job.chargeback.generateReport(
				loggerWithFields,
				job.report,
				"scheduledreport",
				job.report.Name,
				tableName,
				reportPeriod.periodStart,
				reportPeriod.periodEnd,
				job.report.Spec.Output,
				genQuery,
				false,
			)

			if err != nil {
				// update the status to Failed with message containing the
				// error
				errMsg := fmt.Sprintf("error occurred while generating report: %s", err)
				failureCondition := cbutil.NewScheduledReportCondition(cbTypes.ScheduledReportFailure, v1.ConditionTrue, cbutil.GenerateReportErrorReason, errMsg)
				cbutil.RemoveScheduledReportCondition(&report.Status, cbTypes.ScheduledReportRunning)
				cbutil.SetScheduledReportCondition(&report.Status, *failureCondition)

				_, updateErr := job.chargeback.chargebackClient.ChargebackV1alpha1().ScheduledReports(job.report.Namespace).Update(report)
				if updateErr != nil {
					loggerWithFields.WithError(updateErr).Errorf("unable to update scheduledReport status")
				}
				loggerWithFields.WithError(err).Errorf("error occurred while generating report")
				return
			}

			// We generated a report successfully, remove the failure condition
			cbutil.RemoveScheduledReportCondition(&report.Status, cbTypes.ScheduledReportFailure)
			report.Status.LastReportTime = &metav1.Time{Time: reportPeriod.periodEnd}
			_, err = job.chargeback.chargebackClient.ChargebackV1alpha1().ScheduledReports(job.report.Namespace).Update(report)
			if err != nil {
				loggerWithFields.WithError(err).Errorf("unable to update scheduledReport status")
				return
			}
		}
	}
}

type scheduledReportRunner struct {
	reportsMu  sync.Mutex
	reports    map[string]*scheduledReportJob
	jobsChan   chan *scheduledReportJob
	wg         sync.WaitGroup
	chargeback *Chargeback
}

func newScheduledReportRunner(chargeback *Chargeback) *scheduledReportRunner {
	return &scheduledReportRunner{
		reports:    make(map[string]*scheduledReportJob),
		jobsChan:   make(chan *scheduledReportJob),
		chargeback: chargeback,
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
	logger := runner.chargeback.logger.WithField("scheduledReport", job.report.Name)
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

func getNextReportPeriod(schedule reportSchedule, period cbTypes.ScheduledReportPeriod, lastScheduled time.Time) (reportPeriod, error) {
	periodEnd := schedule.Next(lastScheduled)
	periodStart, err := getPreviousReportDay(periodEnd, period)
	if err != nil {
		return reportPeriod{}, err
	}
	return reportPeriod{
		periodEnd:   periodEnd.Truncate(time.Millisecond),
		periodStart: periodStart.Truncate(time.Millisecond),
	}, nil
}

func getPreviousReportDay(next time.Time, period cbTypes.ScheduledReportPeriod) (time.Time, error) {
	switch period {
	case cbTypes.ScheduledReportPeriodHourly:
		return next.Add(-time.Hour), nil
	case cbTypes.ScheduledReportPeriodDaily:
		return next.AddDate(0, 0, -1), nil
	case cbTypes.ScheduledReportPeriodWeekly:
		return next.AddDate(0, 0, -7), nil
	case cbTypes.ScheduledReportPeriodMonthly:
		return next.AddDate(0, -1, 0), nil
	default:
		return time.Time{}, fmt.Errorf("unknown report period: %s", period)
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
