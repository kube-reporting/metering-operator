package operator

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/cache"

	cbTypes "github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1"
	"github.com/operator-framework/operator-metering/pkg/aws"
	"github.com/operator-framework/operator-metering/pkg/hive"
)

var (
	promsumHiveColumns = []hive.Column{
		{Name: "amount", Type: "double"},
		{Name: "timestamp", Type: "timestamp"},
		{Name: "timePrecision", Type: "double"},
		{Name: "labels", Type: "map<string, string>"},
	}
)

func (op *Reporting) runReportDataSourceWorker() {
	logger := op.logger.WithField("component", "reportDataSourceWorker")
	logger.Infof("ReportDataSource worker started")
	for op.processReportDataSource(logger) {

	}
}

func (op *Reporting) processReportDataSource(logger log.FieldLogger) bool {
	obj, quit := op.queues.reportDataSourceQueue.Get()
	if quit {
		logger.Infof("queue is shutting down, exiting ReportDataSource worker")
		return false
	}
	defer op.queues.reportDataSourceQueue.Done(obj)

	logger = logger.WithFields(newLogIdentifier(op.rand))
	if key, ok := op.getKeyFromQueueObj(logger, "ReportDataSource", obj, op.queues.reportDataSourceQueue); ok {
		err := op.syncReportDataSource(logger, key)
		op.handleErr(logger, err, "ReportDataSource", key, op.queues.reportDataSourceQueue)
	}
	return true
}

func (op *Reporting) syncReportDataSource(logger log.FieldLogger, key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		logger.WithError(err).Errorf("invalid resource key :%s", key)
		return nil
	}

	logger = logger.WithField("datasource", name)
	reportDataSource, err := op.informers.Metering().V1alpha1().ReportDataSources().Lister().ReportDataSources(namespace).Get(name)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Infof("ReportDataSource %s does not exist anymore, performing cleanup.", key)
			done := make(chan struct{})
			op.stopPrometheusImporterQueue <- &stopPrometheusImporter{
				ReportDataSource: reportDataSource.Name,
				Done:             done,
			}
			// wait for the importer to be stopped
			<-done
		}
		return err
	}

	if reportDataSource.DeletionTimestamp != nil {
		logger.Infof("ReportDataSource is marked for deletion, performing cleanup")
		done := make(chan struct{})
		op.stopPrometheusImporterQueue <- &stopPrometheusImporter{
			ReportDataSource: reportDataSource.Name,
			Done:             done,
		}
		// wait for the importer to be stopped before we delete the table
		<-done
		return op.deleteReportDataSourceTable(reportDataSource)
	}

	logger.Infof("syncing reportDataSource %s", reportDataSource.GetName())
	err = op.handleReportDataSource(logger, reportDataSource)
	if err != nil {
		logger.WithError(err).Errorf("error syncing reportDataSource %s", reportDataSource.GetName())
		return err
	}
	logger.Infof("successfully synced reportDataSource %s", reportDataSource.GetName())
	return nil
}

func (op *Reporting) handleReportDataSource(logger log.FieldLogger, dataSource *cbTypes.ReportDataSource) error {
	dataSource = dataSource.DeepCopy()
	if dataSource.TableName == "" {
		logger.Infof("new dataSource discovered")
	} else {
		logger.Infof("existing dataSource discovered, tableName: %s", dataSource.TableName)
	}

	switch {
	case dataSource.Spec.Promsum != nil:
		return op.handlePrometheusMetricsDataSource(logger, dataSource)
	case dataSource.Spec.AWSBilling != nil:
		return op.handleAWSBillingDataSource(logger, dataSource)
	default:
		return fmt.Errorf("datasource %s: improperly configured missing promsum or awsBilling configuration", dataSource.Name)
	}
}

func (op *Reporting) handlePrometheusMetricsDataSource(logger log.FieldLogger, dataSource *cbTypes.ReportDataSource) error {
	if dataSource.TableName == "" {
		storage := dataSource.Spec.Promsum.Storage
		tableName := dataSourceTableName(dataSource.Name)
		err := op.createTableForStorage(logger, dataSource, "ReportDataSource", dataSource.Name, storage, tableName, promsumHiveColumns)
		if err != nil {
			return err
		}

		err = op.updateDataSourceTableName(logger, dataSource, tableName)
		if err != nil {
			logger.WithError(err).Errorf("failed to update ReportDataSource TableName field %q", tableName)
			return err
		}
	}

	op.prometheusImporterNewDataSourceQueue <- dataSource

	return nil
}

func (op *Reporting) handleAWSBillingDataSource(logger log.FieldLogger, dataSource *cbTypes.ReportDataSource) error {
	source := dataSource.Spec.AWSBilling.Source
	if source == nil {
		return fmt.Errorf("datasource %q: improperly configured datasource, source is empty", dataSource.Name)
	}

	manifestRetriever := aws.NewManifestRetriever(source.Region, source.Bucket, source.Prefix)

	manifests, err := manifestRetriever.RetrieveManifests()
	if err != nil {
		return err
	}

	if len(manifests) == 0 {
		logger.Warnf("datasource %q has no report manifests in it's bucket, the first report has likely not been generated yet", dataSource.Name)
		return nil
	}

	if dataSource.TableName == "" {
		tableName := dataSourceTableName(dataSource.Name)
		logger.Debugf("creating AWS Billing DataSource table %s pointing to s3 bucket %s at prefix %s", tableName, source.Bucket, source.Prefix)
		err = op.createAWSUsageTable(logger, dataSource, tableName, source.Bucket, source.Prefix, manifests)
		if err != nil {
			return err
		}

		logger.Debugf("successfully created AWS Billing DataSource table %s pointing to s3 bucket %s at prefix %s", tableName, source.Bucket, source.Prefix)
		err = op.updateDataSourceTableName(logger, dataSource, tableName)
		if err != nil {
			return err
		}
	}

	op.prestoTablePartitionQueue <- dataSource
	return nil
}

func (op *Reporting) updateDataSourceTableName(logger log.FieldLogger, dataSource *cbTypes.ReportDataSource, tableName string) error {
	dataSource.TableName = tableName
	_, err := op.meteringClient.MeteringV1alpha1().ReportDataSources(dataSource.Namespace).Update(dataSource)
	if err != nil {
		logger.WithError(err).Errorf("failed to update ReportDataSource table name for %q", dataSource.Name)
		return err
	}
	return nil
}

func (op *Reporting) deleteReportDataSourceTable(reportDataSource *cbTypes.ReportDataSource) error {
	tableName := reportDataSource.TableName
	err := hive.ExecuteDropTable(op.hiveQueryer, tableName, true)
	logger := op.logger.WithFields(log.Fields{"reportDataSource": reportDataSource.Name, "tableName": tableName})
	if err != nil {
		logger.WithError(err).Error("unable to drop ReportDataSource table")
		return err
	}
	logger.Infof("successfully deleted table %s", tableName)
	return nil
}
