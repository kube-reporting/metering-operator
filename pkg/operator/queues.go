package operator

import (
	"time"

	_ "github.com/prestodb/presto-go-client/presto"
	log "github.com/sirupsen/logrus"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	cbTypes "github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1"
	cbInformers "github.com/operator-framework/operator-metering/pkg/generated/informers/externalversions"
	_ "github.com/operator-framework/operator-metering/pkg/util/reflector/prometheus" // for prometheus metric registration
	_ "github.com/operator-framework/operator-metering/pkg/util/workqueue/prometheus" // for prometheus metric registration
)

type queues struct {
	queueList                  []workqueue.RateLimitingInterface
	reportQueue                workqueue.RateLimitingInterface
	scheduledReportQueue       workqueue.RateLimitingInterface
	reportDataSourceQueue      workqueue.RateLimitingInterface
	reportGenerationQueryQueue workqueue.RateLimitingInterface
	prestoTableQueue           workqueue.RateLimitingInterface
}

func (op *Reporting) setupInformers() {
	op.informers = cbInformers.NewFilteredSharedInformerFactory(op.meteringClient, defaultResyncPeriod, op.cfg.Namespace, nil)
	inf := op.informers.Metering().V1alpha1()
	// hacks to ensure these informers are created before we call
	// op.informers.Start()
	inf.PrestoTables().Informer()
	inf.StorageLocations().Informer()
	inf.ReportDataSources().Informer()
	inf.ReportGenerationQueries().Informer()
	inf.ReportPrometheusQueries().Informer()
	inf.Reports().Informer()
	inf.ScheduledReports().Informer()
}

func (op *Reporting) setupQueues() {
	reportQueue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "reports")
	scheduledReportQueue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "scheduledreports")
	reportDataSourceQueue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "reportdatasources")
	reportGenerationQueryQueue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "reportgenerationqueries")
	prestoTableQueue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "prestotables")

	op.queues = queues{
		queueList: []workqueue.RateLimitingInterface{
			reportQueue,
			scheduledReportQueue,
			reportDataSourceQueue,
			reportGenerationQueryQueue,
			prestoTableQueue,
		},
		reportQueue:                reportQueue,
		scheduledReportQueue:       scheduledReportQueue,
		reportDataSourceQueue:      reportDataSourceQueue,
		reportGenerationQueryQueue: reportGenerationQueryQueue,
		prestoTableQueue:           prestoTableQueue,
	}
}

func (op *Reporting) setupEventHandlers() {
	op.informers.Metering().V1alpha1().Reports().Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    op.addReport,
		UpdateFunc: op.updateReport,
		DeleteFunc: op.deleteReport,
	})

	op.informers.Metering().V1alpha1().ScheduledReports().Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    op.addScheduledReport,
		UpdateFunc: op.updateScheduledReport,
		DeleteFunc: op.deleteScheduledReport,
	})

	op.informers.Metering().V1alpha1().ReportDataSources().Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    op.addReportDataSource,
		UpdateFunc: op.updateReportDataSource,
		DeleteFunc: op.deleteReportDataSource,
	})

	op.informers.Metering().V1alpha1().ReportGenerationQueries().Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    op.addReportGenerationQuery,
		UpdateFunc: op.updateReportGenerationQuery,
	})

	op.informers.Metering().V1alpha1().PrestoTables().Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    op.addPrestoTable,
		UpdateFunc: op.updatePrestoTable,
		DeleteFunc: op.deletePrestoTable,
	})
}

func (qs queues) ShutdownQueues() {
	for _, queue := range qs.queueList {
		queue.ShutDown()
	}
}

func (op *Reporting) addReport(obj interface{}) {
	report := obj.(*cbTypes.Report)
	if report.DeletionTimestamp != nil {
		op.deleteReport(report)
		return
	}

	op.logger.Infof("adding Report %s", report.Name)
	op.enqueueReport(report)
}

func (op *Reporting) updateReport(_, cur interface{}) {
	curReport := cur.(*cbTypes.Report)
	if curReport.DeletionTimestamp != nil {
		op.deleteReport(curReport)
		return
	}
	op.logger.Infof("updating Report %s", curReport.Name)
	op.enqueueReport(curReport)
}

func (op *Reporting) deleteReport(obj interface{}) {
	report, ok := obj.(*cbTypes.Report)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			op.logger.WithField("report", report.Name).Errorf("Couldn't get object from tombstone %#v", obj)
			return
		}
		report, ok = tombstone.Obj.(*cbTypes.Report)
		if !ok {
			op.logger.WithField("report", report.Name).Errorf("Tombstone contained object that is not a Report %#v", obj)
			return
		}
	}
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(report)
	if err != nil {
		op.logger.WithField("report", report.Name).WithError(err).Errorf("couldn't get key for object: %#v", report)
		return
	}
	op.queues.reportQueue.Add(key)
}

func (op *Reporting) enqueueReport(report *cbTypes.Report) {
	key, err := cache.MetaNamespaceKeyFunc(report)
	if err != nil {
		op.logger.WithField("report", report.Name).WithError(err).Errorf("couldn't get key for object: %#v", report)
		return
	}
	op.queues.reportQueue.Add(key)
}

func (op *Reporting) enqueueReportRateLimited(report *cbTypes.Report) {
	key, err := cache.MetaNamespaceKeyFunc(report)
	if err != nil {
		op.logger.WithField("report", report.Name).WithError(err).Errorf("couldn't get key for object: %#v", report)
		return
	}
	op.queues.reportQueue.AddRateLimited(key)
}

func (op *Reporting) enqueueReportAfter(report *cbTypes.Report, duration time.Duration) {
	key, err := cache.MetaNamespaceKeyFunc(report)
	if err != nil {
		op.logger.WithField("report", report.Name).WithError(err).Errorf("couldn't get key for object: %#v", report)
		return
	}
	op.queues.reportQueue.AddAfter(key, duration)
}

func (op *Reporting) addScheduledReport(obj interface{}) {
	report := obj.(*cbTypes.ScheduledReport)
	if report.DeletionTimestamp != nil {
		op.deleteScheduledReport(report)
		return
	}
	op.logger.Infof("adding ScheduledReport %s", report.Name)
	op.enqueueScheduledReport(report)
}

func (op *Reporting) updateScheduledReport(_, cur interface{}) {
	curScheduledReport := cur.(*cbTypes.ScheduledReport)
	if curScheduledReport.DeletionTimestamp != nil {
		op.deleteScheduledReport(curScheduledReport)
		return
	}
	op.logger.Infof("updating ScheduledReport %s", curScheduledReport.Name)
	op.enqueueScheduledReport(curScheduledReport)
}

func (op *Reporting) deleteScheduledReport(obj interface{}) {
	report, ok := obj.(*cbTypes.ScheduledReport)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			op.logger.WithField("scheduledReport", report.Name).Errorf("Couldn't get object from tombstone %#v", obj)
			return
		}
		report, ok = tombstone.Obj.(*cbTypes.ScheduledReport)
		if !ok {
			op.logger.WithField("scheduledReport", report.Name).Errorf("Tombstone contained object that is not a ScheduledReport %#v", obj)
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

func (op *Reporting) enqueueScheduledReport(report *cbTypes.ScheduledReport) {
	key, err := cache.MetaNamespaceKeyFunc(report)
	if err != nil {
		op.logger.WithError(err).Errorf("Couldn't get key for object %#v: %v", report, err)
		return
	}
	op.queues.scheduledReportQueue.Add(key)
}

func (op *Reporting) addReportDataSource(obj interface{}) {
	ds := obj.(*cbTypes.ReportDataSource)
	if ds.DeletionTimestamp != nil {
		op.deleteReportDataSource(ds)
		return
	}
	op.logger.Infof("adding ReportDataSource %s", ds.Name)
	op.enqueueReportDataSource(ds)
}

func (op *Reporting) updateReportDataSource(_, cur interface{}) {
	curReportDataSource := cur.(*cbTypes.ReportDataSource)
	if curReportDataSource.DeletionTimestamp != nil {
		op.deleteReportDataSource(curReportDataSource)
		return
	}
	op.logger.Infof("updating ReportDataSource %s", curReportDataSource.Name)
	op.enqueueReportDataSource(curReportDataSource)
}

func (op *Reporting) deleteReportDataSource(obj interface{}) {
	dataSource, ok := obj.(*cbTypes.ReportDataSource)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			op.logger.WithField("reportDataSource", dataSource.Name).Errorf("Couldn't get object from tombstone %#v", obj)
			return
		}
		dataSource, ok = tombstone.Obj.(*cbTypes.ReportDataSource)
		if !ok {
			op.logger.WithField("reportDataSource", dataSource.Name).Errorf("Tombstone contained object that is not a ReportDataSource %#v", obj)
			return
		}
	}
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(dataSource)
	if err != nil {
		op.logger.WithField("reportDataSource", dataSource.Name).WithError(err).Errorf("couldn't get key for object: %#v", dataSource)
		return
	}
	op.queues.reportDataSourceQueue.Add(key)
}

func (op *Reporting) enqueueReportDataSource(ds *cbTypes.ReportDataSource) {
	key, err := cache.MetaNamespaceKeyFunc(ds)
	if err != nil {
		op.logger.WithField("reportDataSource", ds.Name).WithError(err).Errorf("couldn't get key for object: %#v", ds)
		return
	}
	op.queues.reportDataSourceQueue.Add(key)
}

func (op *Reporting) addReportGenerationQuery(obj interface{}) {
	report := obj.(*cbTypes.ReportGenerationQuery)
	op.logger.Infof("adding ReportGenerationQuery %s", report.Name)
	op.enqueueReportGenerationQuery(report)
}

func (op *Reporting) updateReportGenerationQuery(_, cur interface{}) {
	curReportGenerationQuery := cur.(*cbTypes.ReportGenerationQuery)
	op.logger.Infof("updating ReportGenerationQuery %s", curReportGenerationQuery.Name)
	op.enqueueReportGenerationQuery(curReportGenerationQuery)
}

func (op *Reporting) enqueueReportGenerationQuery(query *cbTypes.ReportGenerationQuery) {
	key, err := cache.MetaNamespaceKeyFunc(query)
	if err != nil {
		op.logger.WithField("reportGenerationQuery", query.Name).WithError(err).Errorf("couldn't get key for object: %#v", query)
		return
	}
	op.queues.reportGenerationQueryQueue.Add(key)
}

func (op *Reporting) addPrestoTable(obj interface{}) {
	table := obj.(*cbTypes.PrestoTable)
	if table.DeletionTimestamp != nil {
		op.deletePrestoTable(table)
		return
	}
	op.logger.Infof("adding PrestoTable %s", table.Name)
	op.enqueuePrestoTable(table)
}

func (op *Reporting) updatePrestoTable(_, cur interface{}) {
	curPrestoTable := cur.(*cbTypes.PrestoTable)
	if curPrestoTable.DeletionTimestamp != nil {
		op.deletePrestoTable(curPrestoTable)
		return
	}
	op.logger.Infof("updating PrestoTable %s", curPrestoTable.Name)
	op.enqueuePrestoTable(curPrestoTable)
}

func (op *Reporting) deletePrestoTable(obj interface{}) {
	prestoTable, ok := obj.(*cbTypes.PrestoTable)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			op.logger.WithField("prestoTable", prestoTable.Name).Errorf("Couldn't get object from tombstone %#v", obj)
			return
		}
		prestoTable, ok = tombstone.Obj.(*cbTypes.PrestoTable)
		if !ok {
			op.logger.WithField("prestoTable", prestoTable.Name).Errorf("Tombstone contained object that is not a PrestoTable %#v", obj)
			return
		}
	}
	// when finalizers aren't enabled, it's pretty likely by the time our
	// worker get the event from the queue that the resource will no longer
	// exist in our store, so we eagerly drop the table upon seeing the delete
	// event when finalizers are disabled
	if !op.cfg.EnableFinalizers && prestoTable != nil {
		_ = op.dropPrestoTable(prestoTable)
	}
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(prestoTable)
	if err != nil {
		op.logger.WithField("prestoTable", prestoTable.Name).WithError(err).Errorf("couldn't get key for object: %#v", prestoTable)
		return
	}
	op.queues.prestoTableQueue.Add(key)
}

func (op *Reporting) enqueuePrestoTable(table *cbTypes.PrestoTable) {
	key, err := cache.MetaNamespaceKeyFunc(table)
	if err != nil {
		op.logger.WithField("prestoTable", table.Name).WithError(err).Errorf("couldn't get key for object: %#v", table)
		return
	}
	op.queues.prestoTableQueue.Add(key)
}

type workerProcessFunc func(logger log.FieldLogger) bool

func (op *Reporting) processResource(logger log.FieldLogger, handlerFunc syncHandler, objType string, queue workqueue.RateLimitingInterface) bool {
	obj, quit := queue.Get()
	if quit {
		logger.Infof("queue is shutting down, exiting %s worker", objType)
		return false
	}
	defer queue.Done(obj)

	op.runHandler(logger, handlerFunc, objType, obj, queue)
	return true
}

type syncHandler func(logger log.FieldLogger, key string) error

func (op *Reporting) runHandler(logger log.FieldLogger, handlerFunc syncHandler, objType string, obj interface{}, queue workqueue.RateLimitingInterface) {
	logger = logger.WithFields(newLogIdentifier(op.rand))
	if key, ok := op.getKeyFromQueueObj(logger, objType, obj, queue); ok {
		logger.Infof("syncing %s %s", objType, key)
		err := handlerFunc(logger, key)
		op.handleErr(logger, err, objType, key, queue)
	}
}

// getKeyFromQueueObj tries to convert the object from the queue into a string,
// and if it isn't, it forgets the key from the queue, and logs an error.
//
// We expect strings to come off the workqueue. These are of the
// form namespace/name. We do this as the delayed nature of the
// workqueue means the items in the informer cache may actually be
// more up to date that when the item was initially put onto the
// workqueue.
func (op *Reporting) getKeyFromQueueObj(logger log.FieldLogger, objType string, obj interface{}, queue workqueue.RateLimitingInterface) (string, bool) {
	if key, ok := obj.(string); ok {
		return key, ok
	}
	queue.Forget(obj)
	logger.WithField(objType, obj).Errorf("expected string in work queue but got %#v", obj)
	return "", false
}

// handleErr checks if an error happened and makes sure we will retry later.
func (op *Reporting) handleErr(logger log.FieldLogger, err error, objType string, obj interface{}, queue workqueue.RateLimitingInterface) {
	logger = logger.WithField(objType, obj)

	if err == nil {
		logger.Infof("successfully synced %s %q", objType, obj)
		queue.Forget(obj)
		return
	}

	// This controller retries 5 times if something goes wrong. After that, it stops trying.
	if queue.NumRequeues(obj) < 5 {
		logger.WithError(err).Errorf("error syncing %s %q, adding back to queue", objType, obj)
		queue.AddRateLimited(obj)
		return
	}

	queue.Forget(obj)
	logger.WithError(err).Infof("error syncing %s %q, dropping out of the queue", objType, obj)
}
