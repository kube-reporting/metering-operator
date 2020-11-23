package operator

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/robfig/cron"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"

	metering "github.com/kube-reporting/metering-operator/pkg/apis/metering/v1"
	meteringUtil "github.com/kube-reporting/metering-operator/pkg/apis/metering/v1/util"
	"github.com/kube-reporting/metering-operator/pkg/hive"
	"github.com/kube-reporting/metering-operator/pkg/operator/reporting"
	"github.com/kube-reporting/metering-operator/pkg/operator/reportingutil"
	"github.com/kube-reporting/metering-operator/pkg/util/slice"
)

const (
	reportFinalizer = metering.GroupName + "/report"
)

var (
	reportPrometheusMetricLabels = []string{"report", "namespace", "reportquery", "table_name"}

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

func (op *defaultReportingOperator) runReportWorker() {
	logger := op.logger.WithField("component", "reportWorker")
	logger.Infof("Report worker started")
	const maxRequeues = 5
	for op.processResource(logger, op.syncReport, "Report", op.reportQueue, maxRequeues) {
	}
}

func (op *defaultReportingOperator) syncReport(logger log.FieldLogger, key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		logger.WithError(err).Errorf("invalid resource key :%s", key)
		return nil
	}

	logger = logger.WithFields(log.Fields{"report": name, "namespace": namespace})
	report, err := op.reportLister.Reports(namespace).Get(name)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Infof("report %s/%s does not exist anymore, stopping and removing any running jobs for Report", namespace, name)
			return nil
		}
		return err
	}

	if report.Spec.Expiration != nil {
		logger = logger.WithFields(log.Fields{"expiration": report.Spec.Expiration.Duration})
	}

	// TODO: we may want to move some of the validation checks that are
	// in op.handleExpiredReport and op.handleReport such that we're not
	// performing an expensive operation like creating a deep copy if we
	// know that we're not going to be potentially modifying anything.
	sr := report.DeepCopy()

	err = op.handleExpiredReport(logger, sr, time.Now())
	if err != nil {
		return err
	}

	if report.DeletionTimestamp != nil {
		_, err = op.removeReportFinalizer(sr)
		return err
	}

	return op.handleReport(logger, sr)
}

type reportSchedule interface {
	// Return the next activation time, later than the given time.
	// Next is invoked initially, and then each time the job runs..
	Next(time.Time) time.Time
}

func getSchedule(reportSched *metering.ReportSchedule) (reportSchedule, error) {
	var cronSpec string
	switch reportSched.Period {
	case metering.ReportPeriodCron:
		if reportSched.Cron == nil || reportSched.Cron.Expression == "" {
			return nil, fmt.Errorf("spec.schedule.cron.expression must be specified")
		}
		return cron.ParseStandard(reportSched.Cron.Expression)
	case metering.ReportPeriodHourly:
		sched := reportSched.Hourly
		if sched == nil {
			sched = &metering.ReportScheduleHourly{}
		}
		if err := validateMinute(sched.Minute); err != nil {
			return nil, err
		}
		if err := validateSecond(sched.Second); err != nil {
			return nil, err
		}
		cronSpec = fmt.Sprintf("%d %d * * * *", sched.Second, sched.Minute)
	case metering.ReportPeriodDaily:
		sched := reportSched.Daily
		if sched == nil {
			sched = &metering.ReportScheduleDaily{}
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
	case metering.ReportPeriodWeekly:
		sched := reportSched.Weekly
		if sched == nil {
			sched = &metering.ReportScheduleWeekly{}
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
	case metering.ReportPeriodMonthly:
		sched := reportSched.Monthly
		if sched == nil {
			sched = &metering.ReportScheduleMonthly{}
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
		return nil, fmt.Errorf("invalid Report.spec.schedule.period: %s", reportSched.Period)
	}
	return cron.Parse(cronSpec)
}

func (op *defaultReportingOperator) handleReport(logger log.FieldLogger, report *metering.Report) error {
	if op.cfg.EnableFinalizers && reportNeedsFinalizer(report) {
		var err error
		report, err = op.addReportFinalizer(report)
		if err != nil {
			return err
		}
	}

	return op.runReport(logger, report)
}

type reportPeriod struct {
	periodEnd   time.Time
	periodStart time.Time
}

// isReportFinished is responsible for determing whether the @report Report
// custom resource is currently processing is reporting a Finished status by
// investigating that object's .Status field. If the Report has finished running,
// return true, else return false. Note: at a minimum, we define "Finished" here
// as a Report that has been processing in the queue before (i.e. has non-empty
// conditions in the status array) and contains the "Finished" and "True" conditions
// in the status. If a Report matches those conditions, and we're processing a run-once
// Report, exit early. Else, for scheduled Reports, investigate the spec/ReportingEnd and
// status.LastReportTime for futher analysis.
func isReportFinished(logger log.FieldLogger, report *metering.Report) bool {
	runningCond := meteringUtil.GetReportCondition(report.Status, metering.ReportRunning)

	if runningCond == nil {
		logger.Infof("validating new report")
		return false
	}
	if runningCond.Reason == meteringUtil.ReportFinishedReason && runningCond.Status != v1.ConditionTrue {
		// Check if we're processing a run-once report that has already
		// finished running. If true, log that we're not re-processing
		// run-once reports after they're previously finished and exit early.
		if report.Spec.Schedule == nil {
			logger.Infof("Report %s is a previously finished run-once report, not re-processing", report.Name)
			return true
		}

		// Check if the report's reportingEnd is unset or after the
		// lastReportTime. If either of those conditions are met, then
		// we can make the assumption that the report has been updated
		// since it last finished and this should be recognized as a
		// candidate for re-processing again.
		if report.Spec.ReportingEnd == nil {
			logger.Infof("previously finished report's spec.reportingEnd is unset: beginning processing of report")
			return false
		}
		if report.Status.LastReportTime != nil && report.Spec.ReportingEnd.Time.After(report.Status.LastReportTime.Time) {
			logger.Infof("previously finished report's spec.reportingEnd (%s) is now after lastReportTime (%s): beginning processing of report", report.Spec.ReportingEnd.Time, report.Status.LastReportTime.Time)
			return false
		}

		// return without processing because the report is complete
		logger.Infof("Report %s is already finished: %s", report.Name, runningCond.Message)
		return true
	}

	return false
}

// validateReport takes a Report structure and checks if it contains valid fields
func validateReport(
	report *metering.Report,
	queryGetter reporting.ReportQueryGetter,
	depResolver DependencyResolver,
	handler *reporting.UninitialiedDependendenciesHandler,
) (*metering.ReportQuery, *reporting.DependencyResolutionResult, error) {
	// Validate the ReportQuery is set
	if report.Spec.QueryName == "" {
		return nil, nil, errors.New("must set spec.query")
	}

	// Validate the reportingStart and reportingEnd make sense and are set when
	// required
	if report.Spec.ReportingStart != nil && report.Spec.ReportingEnd != nil && (report.Spec.ReportingStart.Time.After(report.Spec.ReportingEnd.Time) || report.Spec.ReportingStart.Time.Equal(report.Spec.ReportingEnd.Time)) {
		return nil, nil, fmt.Errorf("spec.reportingEnd (%s) must be after spec.reportingStart (%s)", report.Spec.ReportingEnd.Time, report.Spec.ReportingStart.Time)
	}
	if report.Spec.ReportingEnd == nil && report.Spec.RunImmediately {
		return nil, nil, errors.New("spec.reportingEnd must be set if report.spec.runImmediately is true")
	}

	// Validate the ReportQuery that the Report used exists
	query, err := GetReportQueryForReport(report, queryGetter)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, nil, fmt.Errorf("ReportQuery (%s) does not exist", report.Spec.QueryName)
		}
		return nil, nil, fmt.Errorf("failed to get report report query")
	}

	// Validate the dependencies of this Report's query exist
	dependencyResult, err := depResolver.ResolveDependencies(
		query.Namespace,
		query.Spec.Inputs,
		report.Spec.Inputs)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to resolve ReportQuery dependencies %s: %v", query.Name, err)
	}
	err = reporting.ValidateQueryDependencies(dependencyResult.Dependencies, handler)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to validate ReportQuery dependencies %s: %v", query.Name, err)
	}

	return query, dependencyResult, nil
}

// getReportPeriod determines a Report's reporting period based off the report parameter's fields.
// Returns a pointer to a reportPeriod structure if no error was encountered, else panic or return an error.
func getReportPeriod(now time.Time, logger log.FieldLogger, report *metering.Report) (*reportPeriod, error) {
	var reportPeriod *reportPeriod

	// check if the report's schedule spec is set
	if report.Spec.Schedule != nil {
		reportSchedule, err := getSchedule(report.Spec.Schedule)
		if err != nil {
			return nil, err
		}

		if report.Status.LastReportTime != nil {
			reportPeriod = getNextReportPeriod(reportSchedule, report.Spec.Schedule.Period, report.Status.LastReportTime.Time)
		} else {
			if report.Spec.ReportingStart != nil {
				logger.Infof("no last report time for report, using spec.reportingStart %s as starting point", report.Spec.ReportingStart.Time)
				reportPeriod = getNextReportPeriod(reportSchedule, report.Spec.Schedule.Period, report.Spec.ReportingStart.Time)
			} else if report.Status.NextReportTime != nil {
				logger.Infof("no last report time for report, using status.nextReportTime %s as starting point", report.Status.NextReportTime.Time)
				reportPeriod = getNextReportPeriod(reportSchedule, report.Spec.Schedule.Period, report.Status.NextReportTime.Time)
			} else {
				// the current period, [now, nextScheduledTime]
				currentPeriod := getNextReportPeriod(reportSchedule, report.Spec.Schedule.Period, now)
				// the next full report period from [nextScheduledTime, nextScheduledTime+1]
				reportPeriod = getNextReportPeriod(reportSchedule, report.Spec.Schedule.Period, currentPeriod.periodEnd)
				report.Status.NextReportTime = &metav1.Time{Time: reportPeriod.periodStart}
			}
		}
	} else {
		var err error
		// if there's the Spec.Schedule field is unset, then the report must be a run-once report
		reportPeriod, err = getRunOnceReportPeriod(report)
		if err != nil {
			return nil, err
		}
	}

	if reportPeriod.periodStart.After(reportPeriod.periodEnd) {
		panic("periodStart should never come after periodEnd")
	}

	if report.Spec.ReportingEnd != nil && reportPeriod.periodEnd.After(report.Spec.ReportingEnd.Time) {
		logger.Debugf("calculated Report periodEnd %s goes beyond spec.reportingEnd %s, setting periodEnd to reportingEnd", reportPeriod.periodEnd, report.Spec.ReportingEnd.Time)
		// we need to truncate the reportPeriod to align with the reportingEnd
		reportPeriod.periodEnd = report.Spec.ReportingEnd.Time
	}

	return reportPeriod, nil
}

func (op *defaultReportingOperator) handleExpiredReport(logger log.FieldLogger, report *metering.Report, now time.Time) error {
	// check if the Report is being deleted already
	if report.DeletionTimestamp != nil {
		logger.Warnf("report was already marked for deletion")
		return nil
	}
	// check if the Report is past its retention time
	if reportExpired := isReportExpired(op.logger, report, now); !reportExpired {
		logger.Debugf("the %s Report in the %s namespace has not yet reached the expiration date", report.Name, report.Namespace)
		return nil
	}
	// check if the Report is used by any other Report or ReportQuery, if not delete it
	if reportIsNotInput := isReportNotUsedAsInput(logger, report, op); reportIsNotInput {
		newDeletionTimestamp := metav1.Now()
		report.SetDeletionTimestamp(&newDeletionTimestamp)
		err := op.meteringClient.MeteringV1().Reports(report.Namespace).
			Delete(context.TODO(), report.Name, metav1.DeleteOptions{})
		if apierrors.IsNotFound(err) {
			logger.Infof("report: %s, not deleted because it was not found at time delete attempted", report.Name)
			return nil
		}
		if err != nil {
			return fmt.Errorf("failed to delete the expired report: %s. err: %s", report.Name, err)
		}
		logger.Infof("deleted Report: %s in the Namespace: %s because it reached the expiration time",
			report.Name, report.Namespace)
		op.eventRecorder.Event(report, v1.EventTypeNormal, "ExpiredReportHasBeenDeleted",
			fmt.Sprintf("Deleted the %s Report as the configured expiration date has passed", report.Name))

		return nil
	}
	// if here, warn about dependency before return
	logger.Warnf("report: %s, would be deleted because expired, but is depended on", report.Name)
	op.eventRecorder.Event(report, v1.EventTypeWarning, "ExpiredReportHasDependencies",
		fmt.Sprintf("Skipping the deletion of the %s Report as other resources are dependent on it, "+
			"despite reaching the desired expiration date.", report.Name))

	return nil
}

// runReport takes a report, and generates reporting data
// according the report's schedule. If the next scheduled reporting period
// hasn't elapsed, runReport will requeue the resource for a time when
// the period has elapsed.
func (op *defaultReportingOperator) runReport(logger log.FieldLogger, report *metering.Report) error {
	// check if the report we're currently processing is considered
	// "expired". If true, exit early and requeue that object so
	// op.syncReport calls the proper handler for this resource.
	if reportExpired := isReportExpired(logger, report, time.Now()); reportExpired {
		logger.Infof("requeueing report that has reached its expiration date during the op.runReport method")
		op.enqueueReport(report)
		return nil
	}

	// check if the report was previously finished; store result in bool
	if reportFinished := isReportFinished(logger, report); reportFinished {
		return nil
	}

	queryGetter := reporting.NewReportQueryListerGetter(op.reportQueryLister)

	// validate that Report contains valid Spec fields
	reportQuery, dependencyResult, err := validateReport(report, queryGetter, op.dependencyResolver, op.uninitialiedDependendenciesHandler())
	if err != nil {
		return op.setReportStatusInvalidReport(report, err.Error())
	}

	now := op.clock.Now().UTC()

	// get the report's reporting period
	reportPeriod, err := getReportPeriod(now, logger, report)
	if err != nil {
		return err
	}

	logger = logger.WithFields(log.Fields{
		"periodStart":       reportPeriod.periodStart,
		"periodEnd":         reportPeriod.periodEnd,
		"overwriteExisting": report.Spec.OverwriteExistingData,
	})

	// create the table before we check to see if the report has dependencies
	// that are missing data
	var prestoTable *metering.PrestoTable
	// if tableName isn't set, this report is still new and we should make sure
	// no tables exist already in case of a previously failed cleanup.
	if report.Status.TableRef.Name != "" {
		prestoTable, err = op.prestoTableLister.PrestoTables(report.Namespace).Get(report.Status.TableRef.Name)
		if err != nil {
			return fmt.Errorf("unable to get PrestoTable %s for Report %s, %s", report.Status.TableRef, report.Name, err)
		}
		tableName, err := reportingutil.FullyQualifiedTableName(prestoTable)
		if err != nil {
			return err
		}
		logger.Infof("Report %s table already exists, tableName: %s", report.Name, tableName)
	} else {
		tableName := reportingutil.ReportTableName(report.Namespace, report.Name)
		hiveStorage, err := op.getHiveStorage(report.Spec.Output, report.Namespace)
		if err != nil {
			return fmt.Errorf("storage incorrectly configured for Report %s, err: %v", report.Name, err)
		}
		if hiveStorage.Status.Hive.DatabaseName == "" {
			op.enqueueStorageLocation(hiveStorage)
			return fmt.Errorf("StorageLocation %s Hive database %s does not exist yet", hiveStorage.Name, hiveStorage.Spec.Hive.DatabaseName)
		}

		cols, err := reportingutil.PrestoColumnsToHiveColumns(reportingutil.GeneratePrestoColumns(reportQuery))
		if err != nil {
			return fmt.Errorf("unable to convert Presto columns to Hive columns: %s", err)
		}

		params := hive.TableParameters{
			Database: hiveStorage.Status.Hive.DatabaseName,
			Name:     tableName,
			Columns:  cols,
		}
		if hiveStorage.Spec.Hive.DefaultTableProperties != nil {
			params.RowFormat = hiveStorage.Spec.Hive.DefaultTableProperties.RowFormat
			params.FileFormat = hiveStorage.Spec.Hive.DefaultTableProperties.FileFormat
		}

		logger.Infof("creating Hive table %s in database %s", tableName, hiveStorage.Status.Hive.DatabaseName)
		hiveTable, err := op.createHiveTableCR(report, metering.ReportGVK, params, false, nil)
		if err != nil {
			return fmt.Errorf("error creating table for Report %s: %s", report.Name, err)
		}
		hiveTable, err = op.waitForHiveTable(hiveTable.Namespace, hiveTable.Name, time.Second, 20*time.Second)
		if err != nil {
			return fmt.Errorf("error creating table for Report %s: %s", report.Name, err)
		}
		prestoTable, err = op.waitForPrestoTable(hiveTable.Namespace, hiveTable.Name, time.Second, 20*time.Second)
		if err != nil {
			return fmt.Errorf("error creating table for Report %s: %s", report.Name, err)
		}

		logger.Infof("created Hive table %s in database %s", tableName, hiveStorage.Status.Hive.DatabaseName)

		tableName, err = reportingutil.FullyQualifiedTableName(prestoTable)
		if err != nil {
			return err
		}
		dataSourceName := fmt.Sprintf("report-%s", report.Name)

		logger.Infof("creating PrestoTable ReportDataSource %s pointing at report table %s", dataSourceName, tableName)
		ownerRef := metav1.NewControllerRef(prestoTable, metering.PrestoTableGVK)
		newReportDataSource := &metering.ReportDataSource{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ReportDataSource",
				APIVersion: metering.ReportDataSourceGVK.GroupVersion().String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      dataSourceName,
				Namespace: prestoTable.Namespace,
				Labels:    prestoTable.ObjectMeta.Labels,
				OwnerReferences: []metav1.OwnerReference{
					*ownerRef,
				},
			},
			Spec: metering.ReportDataSourceSpec{
				PrestoTable: &metering.PrestoTableDataSource{
					TableRef: v1.LocalObjectReference{
						Name: prestoTable.Name,
					},
				},
			},
		}
		_, err = op.meteringClient.MeteringV1().ReportDataSources(report.Namespace).Create(context.TODO(), newReportDataSource, metav1.CreateOptions{})
		if err != nil {
			if apierrors.IsAlreadyExists(err) {
				logger.Infof("ReportDataSource %s already exists", dataSourceName)
			} else {
				return fmt.Errorf("error creating PrestoTable ReportDataSource %s: %s", dataSourceName, err)
			}
		}
		logger.Infof("created PrestoTable ReportDataSource %s", dataSourceName)

		report.Status.TableRef = v1.LocalObjectReference{Name: hiveTable.Name}
		report, err = op.meteringClient.MeteringV1().Reports(report.Namespace).Update(context.TODO(), report, metav1.UpdateOptions{})
		if err != nil {
			logger.WithError(err).Errorf("unable to update Report status with tableName")
			return err
		}

		// queue dependents so that they're aware the table now exists
		if err := op.queueDependentReportQueriesForReport(report); err != nil {
			logger.WithError(err).Errorf("error queuing ReportQuery dependents of Report %s", report.Name)
		}
		if err := op.queueDependentReportsForReport(report); err != nil {
			logger.WithError(err).Errorf("error queuing Report dependents of Report %s", report.Name)
		}
	}

	runningCond := meteringUtil.GetReportCondition(report.Status, metering.ReportRunning)

	var (
		runningMsg    string
		runningReason string
	)
	if report.Spec.RunImmediately {
		runningReason = meteringUtil.RunImmediatelyReason
		runningMsg = fmt.Sprintf("Report %s scheduled: runImmediately=true and reporting period [%s to %s].", report.Name, reportPeriod.periodStart, reportPeriod.periodEnd)
	} else {
		// Check if it's time to generate the report
		if reportPeriod.periodEnd.After(now) {
			waitTime := reportPeriod.periodEnd.Sub(now)
			waitMsg := fmt.Sprintf("Next scheduled report period is [%s to %s]. next run time is %s.", reportPeriod.periodStart, reportPeriod.periodEnd, reportPeriod.periodEnd)
			logger.Infof(waitMsg+". waiting %s", waitTime)

			if runningCond := meteringUtil.GetReportCondition(report.Status, metering.ReportRunning); runningCond != nil && runningCond.Status == v1.ConditionTrue && runningCond.Reason == meteringUtil.ReportingPeriodWaitingReason {
				op.enqueueReportAfter(report, waitTime)
				return nil
			}

			var err error
			report, err = op.updateReportStatus(report, meteringUtil.NewReportCondition(metering.ReportRunning, v1.ConditionFalse, meteringUtil.ReportingPeriodWaitingReason, waitMsg))
			if err != nil {
				return err
			}

			// we requeue this for later when the period we need to report on next
			// has elapsed
			op.enqueueReportAfter(report, waitTime)
			return nil
		}

		runningReason = meteringUtil.ScheduledReason
		runningMsg = fmt.Sprintf("Report %s scheduled: reached end of reporting period [%s to %s].", report.Name, reportPeriod.periodStart, reportPeriod.periodEnd)

		var unmetDataStartDataSourceDependendencies, unmetDataEndDataSourceDependendencies, unstartedDataSourceDependencies []string
		// Validate all ReportDataSources that the Report depends on have indicated
		// they have data available that covers the current reportPeriod.
		for _, dataSource := range dependencyResult.Dependencies.ReportDataSources {
			if dataSource.Spec.PrometheusMetricsImporter != nil {
				// queue the dataSource and store the list of reports so we can
				// add information to the Report's status on what's currently
				// not ready
				queue := false
				if dataSource.Status.PrometheusMetricsImportStatus == nil {
					unstartedDataSourceDependencies = append(unmetDataStartDataSourceDependendencies, dataSource.Name)
					queue = true
				} else {
					// reportPeriod lower bound not covered
					if dataSource.Status.PrometheusMetricsImportStatus.ImportDataStartTime == nil || reportPeriod.periodStart.Before(dataSource.Status.PrometheusMetricsImportStatus.ImportDataStartTime.Time) {
						queue = true
						unmetDataStartDataSourceDependendencies = append(unmetDataStartDataSourceDependendencies, dataSource.Name)
					}
					// reportPeriod upper bound is not covered
					if dataSource.Status.PrometheusMetricsImportStatus.ImportDataEndTime == nil || reportPeriod.periodEnd.After(dataSource.Status.PrometheusMetricsImportStatus.ImportDataEndTime.Time) {
						queue = true
						unmetDataEndDataSourceDependendencies = append(unmetDataEndDataSourceDependendencies, dataSource.Name)
					}
				}
				if queue {
					op.enqueueReportDataSource(dataSource)
				}
			}
		}

		// Validate all sub-reports that the Report depends on have reported on the
		// current reportPeriod
		var unmetReportDependendencies []string
		for _, subReport := range dependencyResult.Dependencies.Reports {
			if subReport.Status.LastReportTime != nil && subReport.Status.LastReportTime.Time.Before(reportPeriod.periodEnd) {
				op.enqueueReport(subReport)
				unmetReportDependendencies = append(unmetReportDependendencies, subReport.Name)
			}
		}

		if len(unstartedDataSourceDependencies) != 0 || len(unmetDataStartDataSourceDependendencies) != 0 || len(unmetDataEndDataSourceDependendencies) != 0 || len(unmetReportDependendencies) != 0 {
			unmetMsg := "The following Report dependencies do not have data currently available for the current reportPeriod being processed:"
			if len(unstartedDataSourceDependencies) != 0 || len(unmetDataStartDataSourceDependendencies) != 0 || len(unmetDataEndDataSourceDependendencies) != 0 {
				var msgs []string
				if len(unstartedDataSourceDependencies) != 0 {
					// sort so the message is reproducible
					sort.Strings(unstartedDataSourceDependencies)
					msgs = append(msgs, fmt.Sprintf("no data: [%s]", strings.Join(unstartedDataSourceDependencies, ", ")))
				}
				if len(unmetDataStartDataSourceDependendencies) != 0 {
					// sort so the message is reproducible
					sort.Strings(unmetDataStartDataSourceDependendencies)
					msgs = append(msgs, fmt.Sprintf("periodStart %s is before importDataStartTime of [%s]", reportPeriod.periodStart, strings.Join(unmetDataStartDataSourceDependendencies, ", ")))
				}
				if len(unmetDataEndDataSourceDependendencies) != 0 {
					// sort so the message is reproducible
					sort.Strings(unmetDataEndDataSourceDependendencies)
					msgs = append(msgs, fmt.Sprintf("periodEnd %s is after importDataEndTime of [%s]", reportPeriod.periodEnd, strings.Join(unmetDataEndDataSourceDependendencies, ", ")))
				}
				unmetMsg += fmt.Sprintf(" ReportDataSources: %s", strings.Join(msgs, ", "))
			}
			if len(unmetReportDependendencies) != 0 {
				// sort so the message is reproducible
				sort.Strings(unmetReportDependendencies)
				unmetMsg += fmt.Sprintf(" Reports: lastReportTime not prior to periodEnd %s: [%s]", reportPeriod.periodEnd, strings.Join(unmetReportDependendencies, ", "))
			}

			// If the previous condition is unmet dependencies, check if the
			// message changes, and only update if it does
			if runningCond != nil && runningCond.Status == v1.ConditionFalse && runningCond.Reason == meteringUtil.ReportingPeriodUnmetDependenciesReason && runningCond.Message == unmetMsg {
				logger.Debugf("Report %s already has Running condition=false with reason=%s and unchanged message, skipping update", report.Name, meteringUtil.ReportingPeriodUnmetDependenciesReason)
				return nil
			}
			logger.Warnf(unmetMsg)
			_, err := op.updateReportStatus(report, meteringUtil.NewReportCondition(metering.ReportRunning, v1.ConditionFalse, meteringUtil.ReportingPeriodUnmetDependenciesReason, unmetMsg))
			return err
		}
	}
	logger.Infof(runningMsg + " Running now.")

	report, err = op.updateReportStatus(report, meteringUtil.NewReportCondition(metering.ReportRunning, v1.ConditionTrue, runningReason, runningMsg))
	if err != nil {
		return err
	}

	prestoTables, err := op.prestoTableLister.PrestoTables(report.Namespace).List(labels.Everything())
	if err != nil {
		return err
	}

	reports, err := op.reportLister.Reports(report.Namespace).List(labels.Everything())
	if err != nil {
		return err
	}

	datasources, err := op.reportDataSourceLister.ReportDataSources(report.Namespace).List(labels.Everything())
	if err != nil {
		return err
	}

	queries, err := op.reportQueryLister.ReportQueries(report.Namespace).List(labels.Everything())
	if err != nil {
		return err
	}

	requiredInputs := reportingutil.ConvertInputDefinitionsIntoInputList(reportQuery.Spec.Inputs)
	queryCtx := &reporting.ReportQueryTemplateContext{
		Namespace:         report.Namespace,
		Query:             reportQuery.Spec.Query,
		RequiredInputs:    requiredInputs,
		Reports:           reports,
		ReportQueries:     queries,
		ReportDataSources: datasources,
		PrestoTables:      prestoTables,
	}
	tmplCtx := reporting.TemplateContext{
		Report: reporting.ReportTemplateInfo{
			ReportingStart: &reportPeriod.periodStart,
			ReportingEnd:   &reportPeriod.periodEnd,
			Inputs:         dependencyResult.InputValues,
		},
	}

	// Render the query template
	query, err := reporting.RenderQuery(queryCtx, tmplCtx)
	if err != nil {
		return err
	}

	tableName, err := reportingutil.FullyQualifiedTableName(prestoTable)
	if err != nil {
		return err
	}

	metricLabels := prometheus.Labels{
		"report":      report.Name,
		"namespace":   report.Namespace,
		"reportquery": report.Spec.QueryName,
		"table_name":  tableName,
	}

	genReportTotalCounter := generateReportTotalCounter.With(metricLabels)
	genReportFailedCounter := generateReportFailedCounter.With(metricLabels)
	genReportDurationObserver := generateReportDurationHistogram.With(metricLabels)

	logger.Infof("generating Report %s using query %s and periodStart: %s, periodEnd: %s", report.Name, reportQuery.Name, reportPeriod.periodStart, reportPeriod.periodEnd)

	genReportTotalCounter.Inc()
	generateReportStart := op.clock.Now()
	err = op.reportGenerator.GenerateReport(tableName, query, report.Spec.OverwriteExistingData)
	generateReportDuration := op.clock.Since(generateReportStart)
	genReportDurationObserver.Observe(float64(generateReportDuration.Seconds()))
	if err != nil {
		genReportFailedCounter.Inc()
		// update the status to Failed with message containing the
		// error
		errMsg := fmt.Sprintf("error occurred while generating report: %s", err)
		_, updateErr := op.updateReportStatus(report, meteringUtil.NewReportCondition(metering.ReportRunning, v1.ConditionFalse, meteringUtil.GenerateReportFailedReason, errMsg))
		if updateErr != nil {
			logger.WithError(updateErr).Errorf("unable to update Report status")
			return updateErr
		}
		return fmt.Errorf("failed to generateReport for Report %s, err: %v", report.Name, err)
	}

	logger.Infof("successfully generated Report %s using query %s and periodStart: %s, periodEnd: %s", report.Name, reportQuery.Name, reportPeriod.periodStart, reportPeriod.periodEnd)

	// Update the LastReportTime on the report status
	report.Status.LastReportTime = &metav1.Time{Time: reportPeriod.periodEnd}

	// check if we've reached the configured ReportingEnd, and if so, update
	// the status to indicate the report has finished
	if report.Spec.ReportingEnd != nil && report.Status.LastReportTime.Time.Equal(report.Spec.ReportingEnd.Time) {
		msg := fmt.Sprintf("Report has finished reporting. Report has reached the configured spec.reportingEnd: %s", report.Spec.ReportingEnd.Time)
		runningCond := meteringUtil.NewReportCondition(metering.ReportRunning, v1.ConditionFalse, meteringUtil.ReportFinishedReason, msg)

		if condErr := meteringUtil.SetReportCondition(&report.Status, *runningCond); condErr != nil {
			return condErr
		}

		logger.Infof(msg)
	} else if report.Spec.Schedule != nil {
		// determine the next reportTime, if it's not a run-once report and then
		// queue the report for that time
		reportSchedule, err := getSchedule(report.Spec.Schedule)
		if err != nil {
			return err
		}

		nextReportPeriod := getNextReportPeriod(reportSchedule, report.Spec.Schedule.Period, report.Status.LastReportTime.Time)

		// update the NextReportTime on the report status
		report.Status.NextReportTime = &metav1.Time{Time: nextReportPeriod.periodEnd}

		// calculate the time to reprocess after queuing
		now = op.clock.Now().UTC()
		nextRunTime := nextReportPeriod.periodEnd
		waitTime := nextRunTime.Sub(now)

		waitMsg := fmt.Sprintf("Next scheduled report period is [%s to %s]. next run time is %s.", reportPeriod.periodStart, reportPeriod.periodEnd, nextRunTime)
		runningCond := meteringUtil.NewReportCondition(metering.ReportRunning, v1.ConditionFalse, meteringUtil.ReportingPeriodWaitingReason, waitMsg)

		if condErr := meteringUtil.SetReportCondition(&report.Status, *runningCond); condErr != nil {
			return condErr
		}

		logger.Infof(waitMsg+". waiting %s", waitTime)
		op.enqueueReportAfter(report, waitTime)
	}

	// Update the status
	report, err = op.meteringClient.MeteringV1().Reports(report.Namespace).Update(context.TODO(), report, metav1.UpdateOptions{})
	if err != nil {
		logger.WithError(err).Errorf("unable to update Report status")
		return err
	}

	if err := op.queueDependentReportQueriesForReport(report); err != nil {
		logger.WithError(err).Errorf("error queuing ReportQuery dependents of Report %s", report.Name)
	}
	if err := op.queueDependentReportsForReport(report); err != nil {
		logger.WithError(err).Errorf("error queuing Report dependents of Report %s", report.Name)
	}
	return nil
}

// We check first for Reports depending on this Report, return if any, and then check ReportQueries depending on Report
func isReportNotUsedAsInput(logger log.FieldLogger, report *metering.Report, op *defaultReportingOperator) bool {
	// Consider Reports referencing this report in the namespace of this report
	reports, err := op.reportLister.Reports(report.Namespace).List(labels.Everything())
	if err != nil {
		logger.Errorf("unable to determine list of reports that might be input for report: %s", report.Name)
		return false
	}
	for _, report := range reports {
		// If Report has no spec.inputs it can't be depending on this report.
		if report.Spec.Inputs == nil {
			continue
		}
		if does, depReport := op.reportsDependOnThisReport(logger, report, reports); does {
			logger.Infof(
				"report %s exists that uses report %s as input, "+
					"will not delete though retention period has expired", depReport, report.Name)
			return false
		}
	}
	// Consider ReportQueries referencing this report in the namespace of this report
	reportQueries, err := op.reportQueryLister.ReportQueries(report.Namespace).List(labels.Everything())
	if err != nil {
		logger.Errorf("unable to determine list of ReportQueries that might be input for report: %s", report.Name)
		return false
	}
	for _, reportQuery := range reportQueries {
		// If ReportQuery has no spec.inputs it can't be depending on this report.
		if reportQuery.Spec.Inputs == nil {
			continue
		}

		if does, depQuery := op.reportQueriesDependOnThisReport(logger, report, reportQueries); does {
			logger.Infof("ReportQuery, %s exists that uses report %s as input, "+
				"will not delete though retention period has expired", depQuery, report.Name)
			return false
		}
	}
	return true
}

func isReportExpired(logger log.FieldLogger, report *metering.Report, currentTime time.Time) bool {
	if report.Spec.Expiration != nil {
		currentTimeNano := currentTime.UTC().Truncate(time.Millisecond).UnixNano()
		reportCreationNano := report.ObjectMeta.CreationTimestamp.UTC().Truncate(time.Millisecond).UnixNano()
		retentionAmountNano := report.Spec.Expiration.Nanoseconds()
		if currentTimeNano > (reportCreationNano + retentionAmountNano) {
			return true
		}
		logger.Infof("curr: %d, creat: %d, retain: %d", currentTimeNano, reportCreationNano, retentionAmountNano)
	}
	return false
}

func getRunOnceReportPeriod(report *metering.Report) (*reportPeriod, error) {
	if report.Spec.ReportingEnd == nil || report.Spec.ReportingStart == nil {
		return nil, fmt.Errorf("run-once reports must have both ReportingEnd and ReportingStart")
	}
	reportPeriod := &reportPeriod{
		periodStart: report.Spec.ReportingStart.UTC(),
		periodEnd:   report.Spec.ReportingEnd.UTC(),
	}
	return reportPeriod, nil
}

func getNextReportPeriod(schedule reportSchedule, period metering.ReportPeriod, lastScheduled time.Time) *reportPeriod {
	periodStart := lastScheduled.UTC()
	periodEnd := schedule.Next(periodStart)
	return &reportPeriod{
		periodStart: periodStart.Truncate(time.Millisecond).UTC(),
		periodEnd:   periodEnd.Truncate(time.Millisecond).UTC(),
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

func (op *defaultReportingOperator) addReportFinalizer(report *metering.Report) (*metering.Report, error) {
	report.Finalizers = append(report.Finalizers, reportFinalizer)
	newReport, err := op.meteringClient.MeteringV1().Reports(report.Namespace).Update(context.TODO(), report, metav1.UpdateOptions{})
	logger := op.logger.WithFields(log.Fields{"report": report.Name, "namespace": report.Namespace})
	if err != nil {
		logger.WithError(err).Errorf("error adding %s finalizer to Report: %s/%s", reportFinalizer, report.Namespace, report.Name)
		return nil, err
	}
	logger.Infof("added %s finalizer to Report: %s/%s", reportFinalizer, report.Namespace, report.Name)
	return newReport, nil
}

func (op *defaultReportingOperator) removeReportFinalizer(report *metering.Report) (*metering.Report, error) {
	if !slice.ContainsString(report.ObjectMeta.Finalizers, reportFinalizer, nil) {
		return report, nil
	}
	report.Finalizers = slice.RemoveString(report.Finalizers, reportFinalizer, nil)
	newReport, err := op.meteringClient.MeteringV1().Reports(report.Namespace).Update(context.TODO(), report, metav1.UpdateOptions{})
	logger := op.logger.WithFields(log.Fields{"report": report.Name, "namespace": report.Namespace})
	if err != nil {
		logger.WithError(err).Errorf("error removing %s finalizer from Report: %s/%s", reportFinalizer, report.Namespace, report.Name)
		return nil, err
	}
	logger.Infof("removed %s finalizer from Report: %s/%s", reportFinalizer, report.Namespace, report.Name)
	return newReport, nil
}

func reportNeedsFinalizer(report *metering.Report) bool {
	return report.ObjectMeta.DeletionTimestamp == nil && !slice.ContainsString(report.ObjectMeta.Finalizers, reportFinalizer, nil)
}

func (op *defaultReportingOperator) updateReportStatus(report *metering.Report, cond *metering.ReportCondition) (*metering.Report, error) {
	if condErr := meteringUtil.SetReportCondition(&report.Status, *cond); condErr != nil {
		return report, condErr
	}
	return op.meteringClient.MeteringV1().Reports(report.Namespace).Update(context.TODO(), report, metav1.UpdateOptions{})
}

func (op *defaultReportingOperator) setReportStatusInvalidReport(report *metering.Report, msg string) error {
	logger := op.logger.WithFields(log.Fields{"report": report.Name, "namespace": report.Namespace})
	// don't update unless the validation error changes
	if runningCond := meteringUtil.GetReportCondition(report.Status, metering.ReportRunning); runningCond != nil && runningCond.Status == v1.ConditionFalse && runningCond.Reason == meteringUtil.InvalidReportReason && runningCond.Message == msg {
		logger.Debugf("Report %s failed validation last reconcile, skipping updating status", report.Name)
		return nil
	}

	logger.Warnf("Report %s failed validation: %s", report.Name, msg)
	cond := meteringUtil.NewReportCondition(metering.ReportRunning, v1.ConditionFalse, meteringUtil.InvalidReportReason, msg)
	_, err := op.updateReportStatus(report, cond)
	return err
}

// GetReportQueryForReport returns the ReportQuery that was used in the Report parameter
func GetReportQueryForReport(report *metering.Report, queryGetter reporting.ReportQueryGetter) (*metering.ReportQuery, error) {
	return queryGetter.GetReportQuery(report.Namespace, report.Spec.QueryName)
}

func (op *defaultReportingOperator) getReportDependencies(report *metering.Report) (*reporting.ReportQueryDependencies, error) {
	return op.getQueryDependencies(report.Namespace, report.Spec.QueryName, report.Spec.Inputs)
}

func (op *defaultReportingOperator) queueDependentReportsForReport(report *metering.Report) error {
	// Look for all reports in the namespace
	reports, err := op.reportLister.Reports(report.Namespace).List(labels.Everything())
	if err != nil {
		return err
	}

	// for each report in the namespace, find ones that depend on the report
	// passed into the function.
	for _, otherReport := range reports {
		deps, err := op.getReportDependencies(otherReport)
		if err != nil {
			return err
		}
		// If this otherReport has a dependency on the passed in report, queue
		// it
		for _, dep := range deps.Reports {
			if dep.Name == report.Name {
				op.enqueueReport(otherReport)
				break
			}
		}
	}
	return nil
}

// for each report in the namespace, find ones that depend on the report passed into the function
func (op *defaultReportingOperator) reportsDependOnThisReport(logger log.FieldLogger, report *metering.Report, reports []*metering.Report) (bool, string) {
	var depReportName string
	for _, otherReport := range reports {
		deps, err := op.getReportDependencies(otherReport)
		if err != nil {
			logger.Errorf("was unable to get dependencies for report: %s, while looking for dependent reports", report.Name)
			return true, depReportName
		}
		// If a Report has a dependency on the passed in report, we're done
		for _, dep := range deps.Reports {
			if dep.Name == report.Name {
				return true, dep.Name
			}
		}
	}
	return false, depReportName
}

// for each ReportQuery in the namespace, find ones that depend on the report passed into the function
func (op *defaultReportingOperator) reportQueriesDependOnThisReport(logger log.FieldLogger, report *metering.Report, reportQueries []*metering.ReportQuery) (bool, string) {
	var depQueryName string
	// for each report in the namespace, find queries that depend on the report passed into the function.
	for _, reportQuery := range reportQueries {
		deps, err := op.getQueryDependencies(reportQuery.Namespace, reportQuery.Name, nil)
		if err != nil {
			logger.Errorf("was unable to get dependencies for report: %s, while looking for dependent queries")
			return true, ""
		}
		// If this ReportQuery has a dependency on the passed in report, we're done
		for _, dep := range deps.Reports {
			if dep.Name == report.Name {
				return true, dep.Name
			}
		}
	}
	return false, depQueryName
}

// queueDependentReportQueriesForReport will queue all
// ReportQueries in the namespace which have a dependency on the
// report
func (op *defaultReportingOperator) queueDependentReportQueriesForReport(report *metering.Report) error {
	queryLister := op.meteringClient.MeteringV1().ReportQueries(report.Namespace)
	queries, err := queryLister.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, query := range queries.Items {
		// For every query in the namespace, lookup it's dependencies, and if
		// it has a dependency on the passed in Report, requeue it
		deps, err := op.getQueryDependencies(query.Namespace, query.Name, nil)
		if err != nil {
			return err
		}
		for _, dependency := range deps.Reports {
			if dependency.Name == report.Name {
				// this query depends on the Report passed in
				op.enqueueReportQuery(query)
				break
			}
		}
	}
	return nil
}
