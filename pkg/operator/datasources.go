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
)

var (
	promsumHiveColumns = []hive.Column{
		{Name: "amount", Type: "double"},
		{Name: "timestamp", Type: "timestamp"},
		{Name: "timePrecision", Type: "double"},
		{Name: "labels", Type: "map<string, string>"},
	}

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
	const maxRequeues = 10
	for op.processResource(logger, op.syncReportDataSource, "ReportDataSource", op.reportDataSourceQueue, maxRequeues) {
	}
}

func (op *Reporting) syncReportDataSource(logger log.FieldLogger, key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		logger.WithError(err).Errorf("invalid resource key :%s", key)
		return nil
	}

	logger = logger.WithField("ReportDataSource", name)
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
		err = fmt.Errorf("ReportDataSource %s: improperly configured missing promsum or awsBilling configuration", dataSource.Name)
	}
	if err != nil {
		return err
	}

	if err := op.queueDependentReportGenerationQueriesForDataSource(dataSource); err != nil {
		logger.WithError(err).Errorf("error queuing ReportGenerationQuery dependents of ReportDataSource %s", dataSource.Name)
	}

	return nil
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
	} else {
		logger.Infof("new Prometheus ReportDataSource discovered")
		storage := dataSource.Spec.Promsum.Storage
		tableName := reportingutil.DataSourceTableName(dataSource.Name)
		err := op.createTableForStorage(logger, dataSource, cbTypes.SchemeGroupVersion.WithKind("ReportDataSource"), storage, tableName, promsumHiveColumns)
		if err != nil {
			return err
		}

		dataSource, err = op.updateDataSourceTableName(logger, dataSource, tableName)
		if err != nil {
			logger.WithError(err).Errorf("failed to update ReportDataSource TableName field %q", tableName)
			return err
		}
	}

	if op.cfg.DisablePromsum {
		logger.Infof("Periodic Prometheus ReportDataSource importing disabled")
		return nil
	}

	dataSourceName := dataSource.Name
	queryName := dataSource.Spec.Promsum.Query
	tableName := reportingutil.DataSourceTableName(dataSourceName)

	reportPromQuery, err := op.reportPrometheusQueryLister.ReportPrometheusQueries(dataSource.Namespace).Get(queryName)
	if err != nil {
		return fmt.Errorf("unable to get ReportPrometheusQuery %s for ReportDataSource %s, %s", queryName, dataSourceName, err)
	}

	dataSourceLogger := logger.WithFields(log.Fields{
		"queryName":        queryName,
		"reportDataSource": dataSourceName,
		"tableName":        tableName,
	})

	// wrap in a closure to handle lock and unlock of the mutex
	importer, err := func() (*prestostore.PrometheusImporter, error) {
		op.importersMu.Lock()
		defer op.importersMu.Unlock()
		importer, exists := op.importers[dataSourceName]
		if exists {
			dataSourceLogger.Debugf("ReportDataSource %s already has an importer, updating configuration", dataSourceName)
			cfg := op.newPromImporterCfg(dataSource, reportPromQuery)
			importer.UpdateConfig(cfg)
			return importer, nil
		}
		// don't already have an importer, so create a new one
		importer, err := op.newPromImporter(dataSourceLogger, dataSource, reportPromQuery)
		if err != nil {
			return nil, err
		}
		op.importers[dataSourceName] = importer
		return importer, nil
	}()
	if err != nil {
		return err
	}

	_, err = importer.ImportFromLastTimestamp(context.Background(), false)
	if err != nil {
		return fmt.Errorf("ImportFromLastTimestamp errored: %v", err)
	}

	importInterval := op.getQueryIntervalForReportDataSource(dataSource)
	nextImport := op.clock.Now().Add(importInterval).UTC()
	logger.Infof("queuing Prometheus ReportDataSource %s to importing data again in %s at %s", dataSourceName, importInterval, nextImport)
	op.enqueueReportDataSourceAfter(dataSource, importInterval)
	return nil
}

func (op *Reporting) handleAWSBillingDataSource(logger log.FieldLogger, dataSource *cbTypes.ReportDataSource) error {
	source := dataSource.Spec.AWSBilling.Source
	if source == nil {
		return fmt.Errorf("ReportDataSource %q: improperly configured datasource, source is empty", dataSource.Name)
	}

	if dataSource.Status.TableName != "" {
		logger.Infof("existing AWSBilling ReportDataSource discovered, tableName: %s", dataSource.Status.TableName)
	} else {
		logger.Infof("new AWSBilling ReportDataSource discovered")
	}

	manifestRetriever := aws.NewManifestRetriever(source.Region, source.Bucket, source.Prefix)

	manifests, err := manifestRetriever.RetrieveManifests()
	if err != nil {
		return err
	}

	if len(manifests) == 0 {
		logger.Warnf("ReportDataSource %q has no report manifests in it's bucket, the first report has likely not been generated yet", dataSource.Name)
		return nil
	}

	if dataSource.Status.TableName == "" {
		tableName := reportingutil.DataSourceTableName(dataSource.Name)
		logger.Debugf("creating AWS Billing DataSource table %s pointing to s3 bucket %s at prefix %s", tableName, source.Bucket, source.Prefix)
		err = op.createAWSUsageTable(logger, dataSource, tableName, source.Bucket, source.Prefix, manifests)
		if err != nil {
			return err
		}

		logger.Debugf("successfully created AWS Billing DataSource table %s pointing to s3 bucket %s at prefix %s", tableName, source.Bucket, source.Prefix)
		dataSource, err = op.updateDataSourceTableName(logger, dataSource, tableName)
		if err != nil {
			return err
		}
	}

	gauge := awsBillingReportDatasourcePartitionsGauge.WithLabelValues(dataSource.Name, dataSource.Status.TableName)
	prestoTableResourceName := reportingutil.PrestoTableResourceNameFromKind("ReportDataSource", dataSource.Name)
	prestoTable, err := op.prestoTableLister.PrestoTables(dataSource.Namespace).Get(prestoTableResourceName)
	if err != nil {
		// if not found, try for the uncached copy
		if apierrors.IsNotFound(err) {
			prestoTable, err = op.meteringClient.MeteringV1alpha1().PrestoTables(dataSource.Namespace).Get(prestoTableResourceName, metav1.GetOptions{})
			if err != nil {
				return err
			}
		} else {
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
	return nil
}

func (op *Reporting) updateAWSBillingPartitions(logger log.FieldLogger, partitionsGauge prometheus.Gauge, source *cbTypes.S3Bucket, prestoTable *cbTypes.PrestoTable, manifests []*aws.Manifest) error {
	logger.Infof("updating partitions for presto table %s", prestoTable.Name)
	// Fetch the billing manifests
	if len(manifests) == 0 {
		logger.Warnf("PrestoTable %q has no report manifests in its bucket, the first report has likely not been generated yet", prestoTable.Name)
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

	logger.Infof("finished updating partitions for prestoTable %q", prestoTable.Name)

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
	logger := op.logger.WithField("ReportDataSource", ds.Name)
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
	logger := op.logger.WithField("ReportDataSource", ds.Name)
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

// queueDependentReportGenerationQueriesForDataSource will queue all ReportGenerationQueries in the namespace which have a dependency on the dataSource
func (op *Reporting) queueDependentReportGenerationQueriesForDataSource(dataSource *cbTypes.ReportDataSource) error {
	queryLister := op.meteringClient.MeteringV1alpha1().ReportGenerationQueries(dataSource.Namespace)
	queries, err := queryLister.List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, query := range queries.Items {
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
