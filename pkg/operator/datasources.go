package operator

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"

	cbTypes "github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1"
	"github.com/operator-framework/operator-metering/pkg/aws"
	"github.com/operator-framework/operator-metering/pkg/hive"
	"github.com/operator-framework/operator-metering/pkg/operator/prestostore"
	"github.com/operator-framework/operator-metering/pkg/operator/reportingutil"
	"github.com/operator-framework/operator-metering/pkg/presto"
	"github.com/operator-framework/operator-metering/pkg/util/slice"
)

const (
	reportDataSourceFinalizer = cbTypes.GroupName + "/reportdatasource"
	partitionUpdateInterval   = 30 * time.Minute
	// allowIncompleteChunks must be true generally if we have a large
	// chunkSize because otherwise we will wait for an entire chunks worth of
	// data before importing metrics into Presto.
	allowIncompleteChunks = true
)

var (
	awsBillingReportDatasourcePartitionsGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "metering",
			Name:      "aws_billing_reportdatasource_partitions",
			Help:      "Current number of partitions in a AWSBilling ReportDataSource table.",
		},
		[]string{"reportdatasource", "table_name"},
	)
)

func init() {
	prometheus.MustRegister(awsBillingReportDatasourcePartitionsGauge)
}

func (op *Reporting) runReportDataSourceWorker() {
	logger := op.logger.WithField("component", "reportDataSourceWorker")
	logger.Infof("ReportDataSource worker started")
	const maxRequeues = 20
	for op.processResource(logger, op.syncReportDataSource, "ReportDataSource", op.reportDataSourceQueue, maxRequeues) {
	}
}

func (op *Reporting) syncReportDataSource(logger log.FieldLogger, key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		logger.WithError(err).Errorf("invalid resource key :%s", key)
		return nil
	}

	logger = logger.WithFields(log.Fields{"reportDataSource": name, "namespace": namespace})
	reportDataSource, err := op.reportDataSourceLister.ReportDataSources(namespace).Get(name)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Infof("ReportDataSource %s does not exist anymore", key)
			return nil
		}
		return err
	}

	if reportDataSource.DeletionTimestamp != nil {
		logger.Infof("ReportDataSource is marked for deletion, performing cleanup")
		_, err = op.removeReportDataSourceFinalizer(reportDataSource)
		return err
	}

	// Deep-copy otherwise we are mutating our cache
	ds := reportDataSource.DeepCopy()
	return op.handleReportDataSource(logger, ds)
}

func (op *Reporting) handleReportDataSource(logger log.FieldLogger, dataSource *cbTypes.ReportDataSource) error {
	var err error
	switch {
	case dataSource.Spec.Promsum != nil:
		err = op.handlePrometheusMetricsDataSource(logger, dataSource)
	case dataSource.Spec.AWSBilling != nil:
		err = op.handleAWSBillingDataSource(logger, dataSource)
	default:
		err = fmt.Errorf("ReportDataSource %s: improperly con***REMOVED***gured missing promsum or awsBilling con***REMOVED***guration", dataSource.Name)
	}
	return err

}

func (op *Reporting) handlePrometheusMetricsDataSource(logger log.FieldLogger, dataSource *cbTypes.ReportDataSource) error {
	if dataSource.Spec.Promsum == nil {
		return fmt.Errorf("%s is not a Promsum ReportDataSource", dataSource.Name)
	}

	if op.cfg.EnableFinalizers && reportDataSourceNeedsFinalizer(dataSource) {
		var err error
		dataSource, err = op.addReportDataSourceFinalizer(dataSource)
		if err != nil {
			return err
		}
	}

	if dataSource.Status.TableName != "" {
		logger.Infof("existing Prometheus ReportDataSource discovered, tableName: %s", dataSource.Status.TableName)
	} ***REMOVED*** {
		logger.Infof("new Prometheus ReportDataSource discovered")
		storage := dataSource.Spec.Promsum.Storage
		tableName := reportingutil.DataSourceTableName(dataSource.Namespace, dataSource.Name)
		logger.Infof("creating table %s", tableName)
		err := op.createTableForStorage(logger, dataSource, cbTypes.SchemeGroupVersion.WithKind("ReportDataSource"), storage, tableName, prestostore.PromsumHiveTableColumns, prestostore.PromsumHivePartitionColumns)
		if err != nil {
			return err
		}
		logger.Infof("created table %s", tableName)

		dataSource, err = op.updateDataSourceTableName(logger, dataSource, tableName)
		if err != nil {
			logger.WithError(err).Errorf("failed to update ReportDataSource TableName ***REMOVED***eld %q", tableName)
			return err
		}

		// Queue queries that depend on this when the tables created, as they
		// may be pending initialization.
		if err := op.queueDependentReportGenerationQueriesForDataSource(dataSource); err != nil {
			logger.WithError(err).Errorf("error queuing ReportGenerationQuery dependents of ReportDataSource %s", dataSource.Name)
		}

		// instead of immediately importing, return early after creating the
		// table, to allow other tables to be created if a bunch of
		// ReportDataSources are created at once. 2-5 seconds is good enough
		// since we'll be blocked by other ReportDataSources when redelivered.
		op.enqueueReportDataSourceAfter(dataSource, wait.Jitter(2*time.Second, 2.5))
		return nil
	}

	if op.cfg.DisablePromsum {
		logger.Infof("Periodic Prometheus ReportDataSource importing disabled")
		return nil
	}

	queryName := dataSource.Spec.Promsum.Query

	reportPromQuery, err := op.reportPrometheusQueryLister.ReportPrometheusQueries(dataSource.Namespace).Get(queryName)
	if err != nil {
		return fmt.Errorf("unable to get ReportPrometheusQuery %s for ReportDataSource %s, %s", queryName, dataSource.Name, err)
	}

	dataSourceLogger := logger.WithFields(log.Fields{
		"queryName":        queryName,
		"reportDataSource": dataSource.Name,
		"tableName":        dataSource.Status.TableName,
	})

	importerCfg := op.newPromImporterCfg(dataSource, reportPromQuery)

	// wrap in a closure to handle lock and unlock of the mutex
	importer, err := func() (*prestostore.PrometheusImporter, error) {
		op.importersMu.Lock()
		defer op.importersMu.Unlock()
		importer, exists := op.importers[dataSource.Name]
		if exists {
			dataSourceLogger.Debugf("ReportDataSource %s already has an importer, updating con***REMOVED***guration", dataSource.Name)
			importer.UpdateCon***REMOVED***g(importerCfg)
			return importer, nil
		}
		// don't already have an importer, so create a new one
		importer, err := op.newPromImporter(dataSourceLogger, dataSource, reportPromQuery, importerCfg)
		if err != nil {
			return nil, err
		}
		op.importers[dataSource.Name] = importer
		return importer, nil
	}()
	if err != nil {
		return err
	}

	importTime := op.clock.Now().UTC()
	results, err := importer.ImportFromLastTimestamp(context.Background(), allowIncompleteChunks)
	if err != nil {
		return fmt.Errorf("ImportFromLastTimestamp errored: %v", err)
	}
	numResultsImported := len(results.ProcessedTimeRanges)

	// default to importing at the con***REMOVED***gured import interval
	importDelay := op.getQueryIntervalForReportDataSource(dataSource)

	var (
		earliestImportedMetricTime,
		newestImportedMetricTime,
		importDataEndTime *metav1.Time
	)
	if dataSource.Status.PrometheusMetricImportStatus != nil {
		if dataSource.Status.PrometheusMetricImportStatus.EarliestImportedMetricTime != nil {
			earliestImportedMetricTime = dataSource.Status.PrometheusMetricImportStatus.EarliestImportedMetricTime
		}
		if dataSource.Status.PrometheusMetricImportStatus.NewestImportedMetricTime != nil {
			newestImportedMetricTime = dataSource.Status.PrometheusMetricImportStatus.NewestImportedMetricTime
		}
		if dataSource.Status.PrometheusMetricImportStatus.ImportDataEndTime != nil {
			importDataEndTime = dataSource.Status.PrometheusMetricImportStatus.ImportDataEndTime
		}
	}

	// determine if we need to adjust our next import and update the status
	// information if we've imported new metrics.
	if numResultsImported != 0 {
		// This is the last timeRange we processed, and we use the End time on
		// this to determine what time range the importer attempted to import
		// up until, for tracking our process
		lastTimeRange := results.ProcessedTimeRanges[len(results.ProcessedTimeRanges)-1]

		// These are the ***REMOVED***rst and last metric from the import, which we use to
		// determine the data we've actually imported, versus what we've asked
		// for.
		***REMOVED***rstMetric := results.Metrics[0]
		lastMetric := results.Metrics[len(results.Metrics)-1]

		// if there is no existing timestamp then this must be the ***REMOVED***rst import
		// and we should set the earliestImportedMetricTime
		if earliestImportedMetricTime == nil {
			earliestImportedMetricTime = &metav1.Time{***REMOVED***rstMetric.Timestamp}
		} ***REMOVED*** if earliestImportedMetricTime.After(***REMOVED***rstMetric.Timestamp) {
			dataSourceLogger.Errorf("detected time new metric import has older data than previously imported, data is likely duplicated.")
			// TODO(chance): Look at adding an error to the status.
			return nil // strop processing this ReportDataSource
		}

		if newestImportedMetricTime == nil || newestImportedMetricTime.Time.Before(lastMetric.Timestamp) {
			newestImportedMetricTime = &metav1.Time{lastMetric.Timestamp}
		}

		// Update the timestamp which records the latest we've attempted to query
		// up until.
		if importDataEndTime == nil || importDataEndTime.Time.Before(lastTimeRange.End) {
			importDataEndTime = &metav1.Time{lastTimeRange.End}
		}
		// the data we collected is farther back than 1.5 their chunkSize, so requeue sooner
		// since we're backlogged. We use 1.5 because being behind 1 full chunk
		// is typical, but we shouldn't be 2 full chunks after catching up
		backlogDetectionDuration := time.Duration(1.5*importerCfg.ChunkSize.Seconds()) * time.Second
		backlogDuration := op.clock.Now().Sub(newestImportedMetricTime.Time)
		if backlogDuration > backlogDetectionDuration {
			// import delay has jitter so that processing backlogged
			// ReportDataSources happens in a more randomized order to allow
			// all of them to get processed when the queue is blocked.
			importDelay = wait.Jitter(5*time.Second, 2)
			logger.Warnf("Prometheus metrics import backlog detected: imported data for Prometheus ReportDataSource %s newest imported metric timestamp %s is %s away, queuing to reprocess in %s", dataSource.Name, newestImportedMetricTime.Time, backlogDuration, importDelay)
		}
	}

	// Update the status to indicate where we are in the metric import process
	dataSource.Status.PrometheusMetricImportStatus = &cbTypes.PrometheusMetricImportStatus{
		EarliestImportedMetricTime: earliestImportedMetricTime,
		NewestImportedMetricTime:   newestImportedMetricTime,
		ImportDataEndTime:          importDataEndTime,
		LastImportTime:             &metav1.Time{importTime},
	}
	dataSource, err = op.meteringClient.MeteringV1alpha1().ReportDataSources(dataSource.Namespace).Update(dataSource)
	if err != nil {
		return fmt.Errorf("unable to update ReportDataSource %s PrometheusMetricImportStatus: %v", dataSource.Name, err)
	}

	nextImport := op.clock.Now().Add(importDelay).UTC()
	logger.Infof("queuing Prometheus ReportDataSource %s to import data again in %s at %s", dataSource.Name, importDelay, nextImport)
	op.enqueueReportDataSourceAfter(dataSource, importDelay)
	op.queueDependentsOfDataSource(dataSource)
	return nil
}

func (op *Reporting) handleAWSBillingDataSource(logger log.FieldLogger, dataSource *cbTypes.ReportDataSource) error {
	source := dataSource.Spec.AWSBilling.Source
	if source == nil {
		return fmt.Errorf("ReportDataSource %q: improperly con***REMOVED***gured datasource, source is empty", dataSource.Name)
	}

	if dataSource.Status.TableName != "" {
		logger.Infof("existing AWSBilling ReportDataSource discovered, tableName: %s", dataSource.Status.TableName)
	} ***REMOVED*** {
		logger.Infof("new AWSBilling ReportDataSource discovered")
	}

	manifestRetriever := aws.NewManifestRetriever(source.Region, source.Bucket, source.Pre***REMOVED***x)

	manifests, err := manifestRetriever.RetrieveManifests()
	if err != nil {
		return err
	}

	if len(manifests) == 0 {
		logger.Warnf("ReportDataSource %q has no report manifests in it's bucket, the ***REMOVED***rst report has likely not been generated yet", dataSource.Name)
		return nil
	}

	if dataSource.Status.TableName == "" {
		tableName := reportingutil.DataSourceTableName(dataSource.Namespace, dataSource.Name)
		logger.Debugf("creating AWS Billing DataSource table %s pointing to s3 bucket %s at pre***REMOVED***x %s", tableName, source.Bucket, source.Pre***REMOVED***x)
		err = op.createAWSUsageTable(logger, dataSource, tableName, source.Bucket, source.Pre***REMOVED***x, manifests)
		if err != nil {
			return err
		}

		logger.Debugf("successfully created AWS Billing DataSource table %s pointing to s3 bucket %s at pre***REMOVED***x %s", tableName, source.Bucket, source.Pre***REMOVED***x)
		dataSource, err = op.updateDataSourceTableName(logger, dataSource, tableName)
		if err != nil {
			return err
		}
	}

	gauge := awsBillingReportDatasourcePartitionsGauge.WithLabelValues(dataSource.Name, dataSource.Status.TableName)
	prestoTableResourceName := reportingutil.PrestoTableResourceNameFromKind("ReportDataSource", dataSource.Namespace, dataSource.Name)
	prestoTable, err := op.prestoTableLister.PrestoTables(dataSource.Namespace).Get(prestoTableResourceName)
	if err != nil {
		// if not found, try for the uncached copy
		if apierrors.IsNotFound(err) {
			prestoTable, err = op.meteringClient.MeteringV1alpha1().PrestoTables(dataSource.Namespace).Get(prestoTableResourceName, metav1.GetOptions{})
			if err != nil {
				return err
			}
		} ***REMOVED*** {
			return err
		}
	}

	err = op.updateAWSBillingPartitions(logger, gauge, source, prestoTable, manifests)
	if err != nil {
		return fmt.Errorf("error updating AWS billing partitions for ReportDataSource %s: %v", dataSource.Name, err)
	}

	nextUpdate := op.clock.Now().Add(partitionUpdateInterval).UTC()

	logger.Infof("queuing AWSBilling ReportDataSource %s to update partitions again in %s at %s", dataSource.Name, partitionUpdateInterval, nextUpdate)
	op.enqueueReportDataSourceAfter(dataSource, partitionUpdateInterval)

	op.queueDependentsOfDataSource(dataSource)
	return nil
}

func (op *Reporting) updateAWSBillingPartitions(logger log.FieldLogger, partitionsGauge prometheus.Gauge, source *cbTypes.S3Bucket, prestoTable *cbTypes.PrestoTable, manifests []*aws.Manifest) error {
	logger.Infof("updating partitions for presto table %s", prestoTable.Name)
	// Fetch the billing manifests
	if len(manifests) == 0 {
		logger.Warnf("PrestoTable %q has no report manifests in its bucket, the ***REMOVED***rst report has likely not been generated yet", prestoTable.Name)
		return nil
	}

	// Compare the manifests list and existing partitions, deleting stale
	// partitions and creating missing partitions
	currentPartitions := prestoTable.Status.Partitions
	desiredPartitions, err := getDesiredPartitions(source.Bucket, manifests)
	if err != nil {
		return err
	}

	changes := getPartitionChanges(currentPartitions, desiredPartitions)

	currentPartitionsList := make([]string, len(currentPartitions))
	desiredPartitionsList := make([]string, len(desiredPartitions))
	toRemovePartitionsList := make([]string, len(changes.toRemovePartitions))
	toAddPartitionsList := make([]string, len(changes.toAddPartitions))
	toUpdatePartitionsList := make([]string, len(changes.toUpdatePartitions))

	for i, p := range currentPartitions {
		currentPartitionsList[i] = fmt.Sprintf("%#v", p)
	}
	for i, p := range desiredPartitions {
		desiredPartitionsList[i] = fmt.Sprintf("%#v", p)
	}
	for i, p := range changes.toRemovePartitions {
		toRemovePartitionsList[i] = fmt.Sprintf("%#v", p)
	}
	for i, p := range changes.toAddPartitions {
		toAddPartitionsList[i] = fmt.Sprintf("%#v", p)
	}
	for i, p := range changes.toUpdatePartitions {
		toUpdatePartitionsList[i] = fmt.Sprintf("%#v", p)
	}

	logger.Debugf("current partitions: %s", strings.Join(currentPartitionsList, ", "))
	logger.Debugf("desired partitions: %s", strings.Join(desiredPartitionsList, ", "))
	logger.Debugf("partitions to remove: [%s]", strings.Join(toRemovePartitionsList, ", "))
	logger.Debugf("partitions to add: [%s]", strings.Join(toAddPartitionsList, ", "))
	logger.Debugf("partitions to update: [%s]", strings.Join(toUpdatePartitionsList, ", "))

	var toRemove []cbTypes.TablePartition = append(changes.toRemovePartitions, changes.toUpdatePartitions...)
	var toAdd []cbTypes.TablePartition = append(changes.toAddPartitions, changes.toUpdatePartitions...)
	// We do removals then additions so that updates are supported as a combination of remove + add partition

	tableName := prestoTable.Status.Parameters.Name
	for _, p := range toRemove {
		start := p.PartitionSpec["start"]
		end := p.PartitionSpec["end"]
		logger.Warnf("Deleting partition from presto table %q with range %s-%s", tableName, start, end)
		err = op.awsTablePartitionManager.DropPartition(tableName, start, end)
		if err != nil {
			logger.WithError(err).Errorf("failed to drop partition in table %s for range %s-%s", tableName, start, end)
			return err
		}
		logger.Debugf("partition successfully deleted from presto table %q with range %s-%s", tableName, start, end)
	}

	for _, p := range toAdd {
		start := p.PartitionSpec["start"]
		end := p.PartitionSpec["end"]
		// This partition doesn't exist in hive. Create it.
		logger.Debugf("Adding partition to presto table %q with range %s-%s", tableName, start, end)
		err = op.awsTablePartitionManager.AddPartition(tableName, start, end, p.Location)
		if err != nil {
			logger.WithError(err).Errorf("failed to add partition in table %s for range %s-%s at location %s", prestoTable.Status.Parameters.Name, p.PartitionSpec["start"], p.PartitionSpec["end"], p.Location)
			return err
		}
		logger.Debugf("partition successfully added to presto table %q with range %s-%s", tableName, start, end)
	}

	prestoTable.Status.Partitions = desiredPartitions

	numPartitions := len(desiredPartitionsList)
	partitionsGauge.Set(float64(numPartitions))

	_, err = op.meteringClient.MeteringV1alpha1().PrestoTables(prestoTable.Namespace).Update(prestoTable)
	if err != nil {
		logger.WithError(err).Errorf("failed to update PrestoTable CR partitions for %q", prestoTable.Name)
		return err
	}

	logger.Infof("***REMOVED***nished updating partitions for prestoTable %q", prestoTable.Name)
	return nil
}

func getDesiredPartitions(bucket string, manifests []*aws.Manifest) ([]cbTypes.TablePartition, error) {
	desiredPartitions := make([]cbTypes.TablePartition, 0)
	// Manifests have a one-to-one correlation with hive currentPartitions
	for _, manifest := range manifests {
		manifestPath := manifest.DataDirectory()
		location, err := hive.S3Location(bucket, manifestPath)
		if err != nil {
			return nil, err
		}

		start := reportingutil.BillingPeriodTimestamp(manifest.BillingPeriod.Start.Time)
		end := reportingutil.BillingPeriodTimestamp(manifest.BillingPeriod.End.Time)
		p := cbTypes.TablePartition{
			Location: location,
			PartitionSpec: presto.PartitionSpec{
				"start": start,
				"end":   end,
			},
		}
		desiredPartitions = append(desiredPartitions, p)
	}
	return desiredPartitions, nil
}

type partitionChanges struct {
	toRemovePartitions []cbTypes.TablePartition
	toAddPartitions    []cbTypes.TablePartition
	toUpdatePartitions []cbTypes.TablePartition
}

func getPartitionChanges(currentPartitions, desiredPartitions []cbTypes.TablePartition) partitionChanges {
	currentPartitionsSet := make(map[string]cbTypes.TablePartition)
	desiredPartitionsSet := make(map[string]cbTypes.TablePartition)

	for _, p := range currentPartitions {
		currentPartitionsSet[fmt.Sprintf("%s_%s", p.PartitionSpec["start"], p.PartitionSpec["end"])] = p
	}
	for _, p := range desiredPartitions {
		desiredPartitionsSet[fmt.Sprintf("%s_%s", p.PartitionSpec["start"], p.PartitionSpec["end"])] = p
	}

	var toRemovePartitions, toAddPartitions, toUpdatePartitions []cbTypes.TablePartition

	for key, partition := range currentPartitionsSet {
		if _, exists := desiredPartitionsSet[key]; !exists {
			toRemovePartitions = append(toRemovePartitions, partition)
		}
	}
	for key, partition := range desiredPartitionsSet {
		if _, exists := currentPartitionsSet[key]; !exists {
			toAddPartitions = append(toAddPartitions, partition)
		}
	}
	for key, existingPartition := range currentPartitionsSet {
		if newPartition, exists := desiredPartitionsSet[key]; exists && (newPartition.Location != existingPartition.Location) {
			// use newPartition so toUpdatePartitions contains the desired partition state
			toUpdatePartitions = append(toUpdatePartitions, newPartition)
		}
	}

	return partitionChanges{
		toRemovePartitions: toRemovePartitions,
		toAddPartitions:    toAddPartitions,
		toUpdatePartitions: toUpdatePartitions,
	}
}

func (op *Reporting) updateDataSourceTableName(logger log.FieldLogger, dataSource *cbTypes.ReportDataSource, tableName string) (*cbTypes.ReportDataSource, error) {
	dataSource.Status.TableName = tableName
	ds, err := op.meteringClient.MeteringV1alpha1().ReportDataSources(dataSource.Namespace).Update(dataSource)
	if err != nil {
		logger.WithError(err).Errorf("failed to update ReportDataSource table name for %q", dataSource.Name)
		return nil, err
	}
	return ds, nil
}

func (op *Reporting) addReportDataSourceFinalizer(ds *cbTypes.ReportDataSource) (*cbTypes.ReportDataSource, error) {
	ds.Finalizers = append(ds.Finalizers, reportDataSourceFinalizer)
	newReportDataSource, err := op.meteringClient.MeteringV1alpha1().ReportDataSources(ds.Namespace).Update(ds)
	logger := op.logger.WithFields(log.Fields{"reportDataSource": ds.Name, "namespace": ds.Namespace})
	if err != nil {
		logger.WithError(err).Errorf("error adding %s ***REMOVED***nalizer to ReportDataSource: %s/%s", reportDataSourceFinalizer, ds.Namespace, ds.Name)
		return nil, err
	}
	logger.Infof("added %s ***REMOVED***nalizer to ReportDataSource: %s/%s", reportDataSourceFinalizer, ds.Namespace, ds.Name)
	return newReportDataSource, nil
}

func (op *Reporting) removeReportDataSourceFinalizer(ds *cbTypes.ReportDataSource) (*cbTypes.ReportDataSource, error) {
	if !slice.ContainsString(ds.ObjectMeta.Finalizers, reportDataSourceFinalizer, nil) {
		return ds, nil
	}
	ds.Finalizers = slice.RemoveString(ds.Finalizers, reportDataSourceFinalizer, nil)
	newReportDataSource, err := op.meteringClient.MeteringV1alpha1().ReportDataSources(ds.Namespace).Update(ds)
	logger := op.logger.WithFields(log.Fields{"reportDataSource": ds.Name, "namespace": ds.Namespace})
	if err != nil {
		logger.WithError(err).Errorf("error removing %s ***REMOVED***nalizer from ReportDataSource: %s/%s", reportDataSourceFinalizer, ds.Namespace, ds.Name)
		return nil, err
	}
	logger.Infof("removed %s ***REMOVED***nalizer from ReportDataSource: %s/%s", reportDataSourceFinalizer, ds.Namespace, ds.Name)
	return newReportDataSource, nil
}

func reportDataSourceNeedsFinalizer(ds *cbTypes.ReportDataSource) bool {
	return ds.ObjectMeta.DeletionTimestamp == nil && !slice.ContainsString(ds.ObjectMeta.Finalizers, reportDataSourceFinalizer, nil)
}

// queueDependentReportGenerationQueriesForDataSource will queue all ReportGenerationQueries in the namespace which have a dependency on the dataSource
func (op *Reporting) queueDependentReportGenerationQueriesForDataSource(dataSource *cbTypes.ReportDataSource) error {
	queries, err := op.reportGenerationQueryLister.ReportGenerationQueries(dataSource.Namespace).List(labels.Everything())
	if err != nil {
		return err
	}

	for _, query := range queries {
		// look at the list ReportDataSource of dependencies
		for _, dependency := range query.Spec.DataSources {
			if dependency == dataSource.Name {
				// this query depends on the ReportDataSource passed in
				op.enqueueReportGenerationQuery(query)
				break
			}
		}
	}
	return nil
}

func (op *Reporting) queueDependentReportsForDataSource(dataSource *cbTypes.ReportDataSource) error {
	// Look at reports in the namespace of this dataSource
	reports, err := op.reportLister.Reports(dataSource.Namespace).List(labels.Everything())
	if err != nil {
		return err
	}

	// For each report in the dataSource's namespace, check for reports that
	// have a dependency on the provided dataSource
	for _, report := range reports {
		deps, err := op.getReportDependencies(report)
		if err != nil {
			return err
		}

		// If this report has a dependency on the passed in dataSource, queue
		// it
		for _, depDataSource := range deps.ReportDataSources {
			if depDataSource.Name == dataSource.Name {
				op.enqueueReport(report)
				break
			}
		}

	}
	return nil
}

func (op *Reporting) queueDependentsOfDataSource(dataSource *cbTypes.ReportDataSource) {
	logger := op.logger.WithFields(log.Fields{"reportDataSource": dataSource.Name, "namespace": dataSource.Namespace})
	if err := op.queueDependentReportGenerationQueriesForDataSource(dataSource); err != nil {
		logger.WithError(err).Errorf("error queuing ReportGenerationQuery dependents of ReportDataSource %s", dataSource.Name)
	}
	if err := op.queueDependentReportsForDataSource(dataSource); err != nil {
		logger.WithError(err).Errorf("error queuing Report dependents of ReportDataSource %s", dataSource.Name)
	}
}
