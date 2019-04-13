package operator

import (
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/retry"

	cbTypes "github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1"
	"github.com/operator-framework/operator-metering/pkg/aws"
	cbInterfaces "github.com/operator-framework/operator-metering/pkg/generated/clientset/versioned/typed/metering/v1alpha1"
	"github.com/operator-framework/operator-metering/pkg/hive"
	"github.com/operator-framework/operator-metering/pkg/operator/prestostore"
	"github.com/operator-framework/operator-metering/pkg/operator/reporting"
	"github.com/operator-framework/operator-metering/pkg/operator/reportingutil"
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
	if op.cfg.EnableFinalizers && reportDataSourceNeedsFinalizer(dataSource) {
		var err error
		dataSource, err = op.addReportDataSourceFinalizer(dataSource)
		if err != nil {
			return err
		}
	}

	var err error
	switch {
	case dataSource.Spec.Promsum != nil:
		err = op.handlePrometheusMetricsDataSource(logger, dataSource)
	case dataSource.Spec.AWSBilling != nil:
		err = op.handleAWSBillingDataSource(logger, dataSource)
	case dataSource.Spec.PrestoTable != nil:
		err = op.handlePrestoTableDataSource(logger, dataSource)
	case dataSource.Spec.GenerationQueryView != nil:
		err = op.handleGenerationQueryViewDataSource(logger, dataSource)
	default:
		err = fmt.Errorf("ReportDataSource %s: improperly configured missing promsum or awsBilling configuration", dataSource.Name)
	}
	return err

}

func (op *Reporting) handlePrometheusMetricsDataSource(logger log.FieldLogger, dataSource *cbTypes.ReportDataSource) error {
	if dataSource.Spec.Promsum == nil {
		return fmt.Errorf("%s is not a Promsum ReportDataSource", dataSource.Name)
	}

	var prestoTable *cbTypes.PrestoTable
	if dataSource.Status.TableRef.Name != "" {
		var err error
		prestoTable, err = op.prestoTableLister.PrestoTables(dataSource.Namespace).Get(dataSource.Status.TableRef.Name)
		if err != nil {
			return fmt.Errorf("unable to get PrestoTable %s for ReportDataSource %s, %s", dataSource.Status.TableRef, dataSource.Name, err)
		}
		logger.Infof("existing Prometheus ReportDataSource discovered, tableName: %s", prestoTable.Status.TableName)
	} else {
		logger.Infof("new Prometheus ReportDataSource %s discovered", dataSource.Name)
		tableName := reportingutil.DataSourceTableName(dataSource.Namespace, dataSource.Name)
		hiveStorage, err := op.getHiveStorage(dataSource.Spec.Promsum.Storage, dataSource.Namespace)
		if err != nil {
			return fmt.Errorf("storage incorrectly configured for ReportDataSource %s, err: %v", dataSource.Name, err)
		}
		if hiveStorage.Status.Hive.DatabaseName == "" {
			return fmt.Errorf("StorageLocation %s Hive database %s does not exist yet", hiveStorage.Name, hiveStorage.Spec.Hive.DatabaseName)
		}
		params := hive.TableParameters{
			Database:      hiveStorage.Status.Hive.DatabaseName,
			Name:          tableName,
			Columns:       prestostore.PromsumHiveTableColumns,
			PartitionedBy: prestostore.PromsumHivePartitionColumns,
		}
		if hiveStorage.Spec.Hive.DefaultTableProperties != nil {
			params.SerdeFormat = hiveStorage.Spec.Hive.DefaultTableProperties.SerdeFormat
			params.FileFormat = hiveStorage.Spec.Hive.DefaultTableProperties.FileFormat
			params.SerdeRowProperties = hiveStorage.Spec.Hive.DefaultTableProperties.SerdeRowProperties
		}

		logger.Infof("creating table %s", tableName)
		hiveTable, err := op.createHiveTableCR(dataSource, cbTypes.ReportDataSourceGVK, params, false, nil)
		if err != nil {
			return fmt.Errorf("error creating table for ReportDataSource %s: %s", dataSource.Name, err)
		}
		hiveTable, err = op.waitForHiveTable(hiveTable.Namespace, hiveTable.Name, time.Second, 30*time.Second)
		if err != nil {
			return fmt.Errorf("error creating table for ReportDataSource %s: %s", dataSource.Name, err)
		}
		prestoTable, err = op.waitForPrestoTable(hiveTable.Namespace, hiveTable.Name, time.Second, 30*time.Second)
		if err != nil {
			return fmt.Errorf("error creating table for ReportDataSource %s: %s", dataSource.Name, err)
		}
		logger.Infof("created table %s", tableName)

		dsClient := op.meteringClient.MeteringV1alpha1().ReportDataSources(dataSource.Namespace)
		dataSource, err = updateReportDataSource(dsClient, dataSource.Name, func(newDS *cbTypes.ReportDataSource) {
			newDS.Status.TableRef = v1.LocalObjectReference{Name: hiveTable.Name}
		})
		if err != nil {
			logger.WithError(err).Errorf("failed to update ReportDataSource TableName field %q", tableName)
			return err
		}

		if err := op.queueDependentReportsForDataSource(dataSource); err != nil {
			logger.WithError(err).Errorf("error queuing Report dependents of ReportDataSource %s", dataSource.Name)
		}
		if err := op.queueDependentReportDataSourcesForDataSource(dataSource); err != nil {
			logger.WithError(err).Errorf("error queuing Report dependents of ReportDataSource %s", dataSource.Name)
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
		"tableName":        prestoTable.Status.TableName,
	})

	importerCfg := op.newPromImporterCfg(dataSource, reportPromQuery, prestoTable)

	// wrap in a closure to handle lock and unlock of the mutex
	importer, err := func() (*prestostore.PrometheusImporter, error) {
		op.importersMu.Lock()
		defer op.importersMu.Unlock()
		importer, exists := op.importers[dataSource.Name]
		if exists {
			dataSourceLogger.Debugf("ReportDataSource %s already has an importer, updating configuration", dataSource.Name)
			importer.UpdateConfig(importerCfg)
			return importer, nil
		}
		// don't already have an importer, so create a new one
		importer, err := op.newPromImporter(dataSourceLogger, dataSource, reportPromQuery, prestoTable, importerCfg)
		if err != nil {
			return nil, err
		}
		op.importers[dataSource.Name] = importer
		return importer, nil
	}()
	if err != nil {
		return err
	}

	importStatus := dataSource.Status.PrometheusMetricImportStatus
	if importStatus == nil {
		importStatus = &cbTypes.PrometheusMetricImportStatus{}
	}

	// record the lastImportTime
	importStatus.LastImportTime = &metav1.Time{Time: op.clock.Now().UTC()}

	// run the import
	results, err := importer.ImportFromLastTimestamp(context.Background(), allowIncompleteChunks)
	if err != nil {
		return fmt.Errorf("ImportFromLastTimestamp errored: %v", err)
	}

	// Default to importing at the configured import interval.
	importDelay := op.getQueryIntervalForReportDataSource(dataSource)

	if len(results.ProcessedTimeRanges) == 0 {
		logger.Warnf("no time ranges processed for ReportDataSource %s", dataSource.Name)
	} else {
		// This is the last timeRange we processed, and we use the End time on
		// this to determine what time range the importer attempted to import
		// up until, for tracking our process
		firstTimeRange := results.ProcessedTimeRanges[0]
		lastTimeRange := results.ProcessedTimeRanges[len(results.ProcessedTimeRanges)-1]

		// Update the timestamp which records the first timestamp we attempted
		// to query from.
		if importStatus.ImportDataStartTime == nil || firstTimeRange.Start.Before(importStatus.ImportDataStartTime.Time) {
			importStatus.ImportDataStartTime = &metav1.Time{Time: firstTimeRange.Start}
		}
		// Update the timestamp which records the latest we've attempted to query
		// up until.
		if importStatus.ImportDataEndTime == nil || importStatus.ImportDataEndTime.Time.Before(lastTimeRange.End) {
			importStatus.ImportDataEndTime = &metav1.Time{Time: lastTimeRange.End}
		}

		// The data we collected is farther back than 1.5 their chunkSize, so requeue sooner
		// since we're backlogged. We use 1.5 because being behind 1 full chunk
		// is typical, but we shouldn't be 2 full chunks after catching up.
		backlogDetectionDuration := time.Duration(1.5*importerCfg.ChunkSize.Seconds()) * time.Second
		backlogDuration := op.clock.Now().Sub(importStatus.ImportDataEndTime.Time)
		if backlogDuration > backlogDetectionDuration {
			// import delay has jitter so that processing backlogged
			// ReportDataSources happens in a more randomized order to allow
			// all of them to get processed when the queue is blocked.
			importDelay = wait.Jitter(5*time.Second, 2)
			logger.Warnf("Prometheus metrics import backlog detected: imported data for Prometheus ReportDataSource %s newest imported metric timestamp %s is %s away, queuing to reprocess in %s", dataSource.Name, importStatus.ImportDataEndTime.Time, backlogDuration, importDelay)
		}

		if len(results.Metrics) != 0 {
			// These are the first and last metric from the import, which we use to
			// determine the data we've actually imported, versus what we've asked
			// for.
			firstMetric := results.Metrics[0]
			lastMetric := results.Metrics[len(results.Metrics)-1]

			// if there is no existing timestamp then this must be the first import
			// and we should set the earliestImportedMetricTime
			if importStatus.EarliestImportedMetricTime == nil {
				importStatus.EarliestImportedMetricTime = &metav1.Time{Time: firstMetric.Timestamp}
			} else if importStatus.EarliestImportedMetricTime.After(firstMetric.Timestamp) {
				dataSourceLogger.Errorf("detected time new metric import has older data than previously imported, data is likely duplicated.")
				// TODO(chance): Look at adding an error to the status.
				return nil // strop processing this ReportDataSource
			}

			if importStatus.NewestImportedMetricTime == nil || lastMetric.Timestamp.After(importStatus.NewestImportedMetricTime.Time) {
				importStatus.NewestImportedMetricTime = &metav1.Time{Time: lastMetric.Timestamp}
			}

			if err := op.queueDependentReportsForDataSource(dataSource); err != nil {
				logger.WithError(err).Errorf("error queuing Report dependents of ReportDataSource %s", dataSource.Name)
			}
			if err := op.queueDependentReportDataSourcesForDataSource(dataSource); err != nil {
				logger.WithError(err).Errorf("error queuing Report dependents of ReportDataSource %s", dataSource.Name)
			}
		}

		// Update the status to indicate where we are in the metric import process
		dsClient := op.meteringClient.MeteringV1alpha1().ReportDataSources(dataSource.Namespace)
		dataSource, err = updateReportDataSource(dsClient, dataSource.Name, func(newDS *cbTypes.ReportDataSource) {
			newDS.Status.PrometheusMetricImportStatus = importStatus
		})
		if err != nil {
			return fmt.Errorf("unable to update ReportDataSource %s PrometheusMetricImportStatus: %v", dataSource.Name, err)
		}
	}

	nextImport := op.clock.Now().Add(importDelay).UTC()
	logger.Infof("queuing Prometheus ReportDataSource %s to import data again in %s at %s", dataSource.Name, importDelay, nextImport)
	op.enqueueReportDataSourceAfter(dataSource, importDelay)
	return nil
}

func (op *Reporting) handleAWSBillingDataSource(logger log.FieldLogger, dataSource *cbTypes.ReportDataSource) error {
	source := dataSource.Spec.AWSBilling.Source
	if source == nil {
		return fmt.Errorf("ReportDataSource %q: improperly configured datasource, source is empty", dataSource.Name)
	}

	logger.Debugf("querying bucket %#v for AWS Billing manifests for ReportDataSource %s", source, dataSource.Name)
	manifestRetriever := aws.NewManifestRetriever(source.Region, source.Bucket, source.Prefix)
	manifests, err := manifestRetriever.RetrieveManifests()
	if err != nil {
		return err
	}

	if len(manifests) == 0 {
		logger.Warnf("ReportDataSource %q has no report manifests in it's bucket, the first report has likely not been generated yet", dataSource.Name)
		return nil
	}

	var hiveTable *cbTypes.HiveTable
	if dataSource.Status.TableRef.Name == "" {
		logger.Infof("new AWSBilling ReportDataSource discovered")
		tableName := reportingutil.DataSourceTableName(dataSource.Namespace, dataSource.Name)
		logger.Debugf("creating AWS Billing DataSource table %s pointing to s3 bucket %s at prefix %s", tableName, source.Bucket, source.Prefix)
		hiveTable, err = op.createAWSUsageHiveTableCR(logger, dataSource, tableName, source.Bucket, source.Prefix, manifests)
		if err != nil {
			return err
		}

		logger.Debugf("successfully created AWS Billing DataSource table %s pointing to s3 bucket %s at prefix %s", tableName, source.Bucket, source.Prefix)
		dsClient := op.meteringClient.MeteringV1alpha1().ReportDataSources(dataSource.Namespace)
		dataSource, err = updateReportDataSource(dsClient, dataSource.Name, func(newDS *cbTypes.ReportDataSource) {
			newDS.Status.TableRef = v1.LocalObjectReference{Name: hiveTable.Name}
		})
		if err != nil {
			return err
		}
	} else {
		hiveTableResourceName := reportingutil.TableResourceNameFromKind("ReportDataSource", dataSource.Namespace, dataSource.Name)
		hiveTable, err = op.hiveTableLister.HiveTables(dataSource.Namespace).Get(hiveTableResourceName)
		if err != nil {
			// if not found, try for the uncached copy
			if apierrors.IsNotFound(err) {
				hiveTable, err = op.meteringClient.MeteringV1alpha1().HiveTables(dataSource.Namespace).Get(hiveTableResourceName, metav1.GetOptions{})
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}
		logger.Infof("existing AWSBilling ReportDataSource discovered, tableName: %s", hiveTable.Status.TableName)
	}

	err = op.updateAWSBillingPartitions(logger, dataSource, source, hiveTable, manifests)
	if err != nil {
		return fmt.Errorf("error updating AWS billing partitions for ReportDataSource %s: %v", dataSource.Name, err)
	}

	nextUpdate := op.clock.Now().Add(partitionUpdateInterval).UTC()

	logger.Infof("queuing AWSBilling ReportDataSource %s to update partitions again in %s at %s", dataSource.Name, partitionUpdateInterval, nextUpdate)
	op.enqueueReportDataSourceAfter(dataSource, partitionUpdateInterval)

	if err := op.queueDependentReportsForDataSource(dataSource); err != nil {
		logger.WithError(err).Errorf("error queuing Report dependents of ReportDataSource %s", dataSource.Name)
	}
	if err := op.queueDependentReportDataSourcesForDataSource(dataSource); err != nil {
		logger.WithError(err).Errorf("error queuing Report dependents of ReportDataSource %s", dataSource.Name)
	}
	return nil
}

func (op *Reporting) handlePrestoTableDataSource(logger log.FieldLogger, dataSource *cbTypes.ReportDataSource) error {
	if dataSource.Spec.PrestoTable == nil {
		return fmt.Errorf("%s is not a PrestoTable ReportDataSource", dataSource.Name)
	}
	if dataSource.Spec.PrestoTable.TableRef.Name == "" {
		return fmt.Errorf("invalid PrestoTable ReportDataSource %s, spec.prestoTable.tableRef.name must be set", dataSource.Name)
	}

	var prestoTable *cbTypes.PrestoTable
	if dataSource.Status.TableRef.Name != "" {
		var err error
		prestoTable, err = op.prestoTableLister.PrestoTables(dataSource.Namespace).Get(dataSource.Status.TableRef.Name)
		if err != nil {
			return fmt.Errorf("unable to get PrestoTable %s for ReportDataSource %s, %s", dataSource.Status.TableRef, dataSource.Name, err)
		}
		logger.Infof("existing PrestoTable ReportDataSource discovered, tableName: %s", prestoTable.Status.TableName)
	} else {
		logger.Infof("new PrestoTable ReportDataSource discovered, tableName: %s", dataSource.Spec.PrestoTable.TableRef.Name)
		var err error
		prestoTable, err = op.waitForPrestoTable(dataSource.Namespace, dataSource.Spec.PrestoTable.TableRef.Name, time.Second, 10*time.Second)
		if err != nil {
			return fmt.Errorf("error creating table for ReportDataSource %s: %s", dataSource.Name, err)
		}

		dsClient := op.meteringClient.MeteringV1alpha1().ReportDataSources(dataSource.Namespace)
		dataSource, err = updateReportDataSource(dsClient, dataSource.Name, func(newDS *cbTypes.ReportDataSource) {
			newDS.Status.TableRef = v1.LocalObjectReference{Name: prestoTable.Name}
		})
		if err != nil {
			logger.WithError(err).Errorf("failed to update ReportDataSource status.tableRef field %q", prestoTable.Name)
			return err
		}
	}

	return nil
}

func (op *Reporting) handleGenerationQueryViewDataSource(logger log.FieldLogger, dataSource *cbTypes.ReportDataSource) error {
	if dataSource.Spec.GenerationQueryView == nil {
		return fmt.Errorf("%s is not a GenerationQueryView ReportDataSource", dataSource.Name)
	}
	if dataSource.Spec.GenerationQueryView.QueryName == "" {
		return fmt.Errorf("invalid GenerationQueryView ReportDataSource %s, spec.generationQueryView.queryName must be set", dataSource.Name)
	}

	generationQuery, err := op.reportGenerationQueryLister.ReportGenerationQueries(dataSource.Namespace).Get(dataSource.Spec.GenerationQueryView.QueryName)
	if err != nil {
		return fmt.Errorf("unable to get ReportGenerationQuery %s for GenerationQueryView ReportDataSource %s: %s", dataSource.Spec.GenerationQueryView.QueryName, dataSource.Name, err)
	}

	var viewName string
	createView := false
	if dataSource.Status.TableRef.Name == "" {
		logger.Infof("new ReportDataSource discovered")
		viewName = reportingutil.DataSourceTableName(dataSource.Namespace, dataSource.Name)
		createView = true
	} else {
		prestoTable, err := op.prestoTableLister.PrestoTables(dataSource.Namespace).Get(dataSource.Status.TableRef.Name)
		if err != nil {
			return fmt.Errorf("unable to get PrestoTable %s for ReportDataSource %s, %s", dataSource.Status.TableRef, dataSource.Name, err)
		}
		logger.Infof("existing ReportDataSource discovered, viewName: %s", prestoTable.Status.TableName)
		viewName = prestoTable.Status.TableName
	}

	dependencyResult, err := op.dependencyResolver.ResolveDependencies(generationQuery.Namespace, generationQuery.Spec.Inputs, nil)
	if err != nil {
		return err
	}

	err = reporting.ValidateGenerationQueryDependencies(dependencyResult.Dependencies, op.uninitialiedDependendenciesHandler())
	if err != nil {
		if reporting.IsUninitializedDependencyError(err) {
			logger.Warnf("unable to validate ReportGenerationQuery %s, has uninitialized dependencies: %v", generationQuery.Name, err)
			// We do not return an error because we do not need to requeue this
			// query. Instead we can wait until this queries uninitialized
			// dependencies become initialized. After they're initialized they
			// will queue anything that depends on them, including this query.
			return nil
		} else if reporting.IsInvalidDependencyError(err) {
			logger.WithError(err).Errorf("unable to validate ReportGenerationQuery %s, has invalid dependencies, dropping off queue", generationQuery.Name)
			// Invalid dependency means it will not resolve itself, so do not
			// return an error since we do not want to be requeued unless the
			// resource is modified, or it's dependencies are modified.
			return nil
		} else {
			// The error occurred when getting the dependencies or for an
			// unknown reason so we want to retry up to a limit. This most
			// commonly occurs when fetching a dependency from the API fails,
			// or if there is a cyclic dependency.
			return fmt.Errorf("unable to get or validate ReportGenerationQuery dependencies %s: %v", generationQuery.Name, err)
		}
	}

	if createView {
		hiveStorage, err := op.getHiveStorage(nil, dataSource.Namespace)
		if err != nil {
			return fmt.Errorf("storage incorrectly configured for ReportDataSource %s, err: %v", dataSource.Name, err)
		}
		if hiveStorage.Status.Hive.DatabaseName == "" {
			return fmt.Errorf("StorageLocation %s Hive database %s does not exist yet", hiveStorage.Name, hiveStorage.Spec.Hive.DatabaseName)
		}
		prestoTables, err := op.prestoTableLister.PrestoTables(dataSource.Namespace).List(labels.Everything())
		if err != nil {
			return err
		}

		queryCtx := &reporting.ReportQueryTemplateContext{
			Namespace:               dataSource.Namespace,
			ReportQuery:             generationQuery,
			Reports:                 dependencyResult.Dependencies.Reports,
			ReportGenerationQueries: dependencyResult.Dependencies.ReportGenerationQueries,
			ReportDataSources:       dependencyResult.Dependencies.ReportDataSources,
			PrestoTables:            prestoTables,
		}
		renderedQuery, err := reporting.RenderQuery(queryCtx, reporting.TemplateContext{
			Report: reporting.ReportTemplateInfo{
				Inputs: dependencyResult.InputValues,
			},
		})
		if err != nil {
			return err
		}

		columns, err := reportingutil.GeneratePrestoColumns(generationQuery)
		if err != nil {
			return fmt.Errorf("error converting ReportGenerationQuery columns to Presto columns for ReportGenerationQuery %s: %v", generationQuery.Name, err)
		}
		logger.Infof("creating view %s", viewName)
		prestoTable, err := op.createPrestoTableCR(dataSource, cbTypes.ReportGenerationQueryGVK, "hive", hiveStorage.Status.Hive.DatabaseName, viewName, columns, false, true, renderedQuery)
		if err != nil {
			return fmt.Errorf("error creating view %s for ReportDataSource %s: %v", viewName, dataSource.Name, err)
		}
		prestoTable, err = op.waitForPrestoTable(prestoTable.Namespace, prestoTable.Name, time.Second, 10*time.Second)
		if err != nil {
			return fmt.Errorf("error creating table for ReportDataSource %s: %s", dataSource.Name, err)
		}

		logger.Infof("created view %s", viewName)

		dsClient := op.meteringClient.MeteringV1alpha1().ReportDataSources(dataSource.Namespace)
		dataSource, err = updateReportDataSource(dsClient, dataSource.Name, func(newDS *cbTypes.ReportDataSource) {
			newDS.Status.TableRef.Name = prestoTable.Name
		})
		if err != nil {
			logger.WithError(err).Errorf("failed to update ReportDataSource tableRef field to %q", prestoTable.Name)
			return err
		}
	}

	if err := op.queueDependentReportsForDataSource(dataSource); err != nil {
		logger.WithError(err).Errorf("error queuing Report dependents of ReportDataSource %s", dataSource.Name)
	}
	if err := op.queueDependentReportDataSourcesForDataSource(dataSource); err != nil {
		logger.WithError(err).Errorf("error queuing Report dependents of ReportDataSource %s", dataSource.Name)
	}

	return nil
}

func (op *Reporting) addReportDataSourceFinalizer(ds *cbTypes.ReportDataSource) (*cbTypes.ReportDataSource, error) {
	ds.Finalizers = append(ds.Finalizers, reportDataSourceFinalizer)
	newReportDataSource, err := op.meteringClient.MeteringV1alpha1().ReportDataSources(ds.Namespace).Update(ds)
	logger := op.logger.WithFields(log.Fields{"reportDataSource": ds.Name, "namespace": ds.Namespace})
	if err != nil {
		logger.WithError(err).Errorf("error adding %s finalizer to ReportDataSource: %s/%s", reportDataSourceFinalizer, ds.Namespace, ds.Name)
		return nil, err
	}
	logger.Infof("added %s finalizer to ReportDataSource: %s/%s", reportDataSourceFinalizer, ds.Namespace, ds.Name)
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
		logger.WithError(err).Errorf("error removing %s finalizer from ReportDataSource: %s/%s", reportDataSourceFinalizer, ds.Namespace, ds.Name)
		return nil, err
	}
	logger.Infof("removed %s finalizer from ReportDataSource: %s/%s", reportDataSourceFinalizer, ds.Namespace, ds.Name)
	return newReportDataSource, nil
}

func reportDataSourceNeedsFinalizer(ds *cbTypes.ReportDataSource) bool {
	return ds.ObjectMeta.DeletionTimestamp == nil && !slice.ContainsString(ds.ObjectMeta.Finalizers, reportDataSourceFinalizer, nil)
}

func (op *Reporting) getGenerationQueryDependencies(namespace, name string, inputVals []cbTypes.ReportGenerationQueryInputValue) (*reporting.ReportGenerationQueryDependencies, error) {
	queryGetter := reporting.NewReportGenerationQueryListerGetter(op.reportGenerationQueryLister)
	genQuery, err := queryGetter.GetReportGenerationQuery(namespace, name)
	if err != nil {
		return nil, err
	}
	result, err := op.dependencyResolver.ResolveDependencies(genQuery.Namespace, genQuery.Spec.Inputs, inputVals)
	if err != nil {
		return nil, err
	}
	return result.Dependencies, nil
}

func (op *Reporting) queueDependentReportDataSourcesForDataSource(dataSource *cbTypes.ReportDataSource) error {
	// Look at reportDataSources in the namespace of this dataSource
	reportDataSources, err := op.reportDataSourceLister.ReportDataSources(dataSource.Namespace).List(labels.Everything())
	if err != nil {
		return err
	}

	// For each reportDataSource in the dataSource's namespace, check for
	// reportDataSources that have a dependency on the provided dataSource
	for _, ds := range reportDataSources {
		// Only ReportDataSources that create a view from a
		// ReportGenerationQuery depend on other ReportDataSources.
		if ds.Spec.GenerationQueryView == nil {
			continue
		}

		deps, err := op.getGenerationQueryDependencies(ds.Namespace, ds.Name, ds.Spec.GenerationQueryView.Inputs)
		if err != nil {
			return fmt.Errorf("unable to get dependencies for GenerationQueryView ReportDataSource %s: %s", ds.Name, err)
		}

		// If this reportDataSource has a dependency on the passed in
		// dataSource, queue it
		for _, depDataSource := range deps.ReportDataSources {
			if depDataSource.Name == dataSource.Name {
				op.enqueueReportDataSource(ds)
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

func updateReportDataSource(dsClient cbInterfaces.ReportDataSourceInterface, dsName string, updateFunc func(*cbTypes.ReportDataSource)) (*cbTypes.ReportDataSource, error) {
	var ds *cbTypes.ReportDataSource
	if err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		newDS, err := dsClient.Get(dsName, metav1.GetOptions{})
		if err != nil {
			return err
		}
		updateFunc(newDS)
		ds, err = dsClient.Update(newDS)
		return err
	}); err != nil {
		return nil, err
	}
	return ds, nil
}
