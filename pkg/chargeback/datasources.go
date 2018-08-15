package chargeback

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

func (c *Metering) runReportDataSourceWorker() {
	logger := c.logger.WithField("component", "reportDataSourceWorker")
	logger.Infof("ReportDataSource worker started")
	for c.processReportDataSource(logger) {

	}
}

func (c *Metering) processReportDataSource(logger log.FieldLogger) bool {
	obj, quit := c.queues.reportDataSourceQueue.Get()
	if quit {
		logger.Infof("queue is shutting down, exiting ReportDataSource worker")
		return false
	}
	defer c.queues.reportDataSourceQueue.Done(obj)

	logger = logger.WithFields(newLogIdenti***REMOVED***er(c.rand))
	if key, ok := c.getKeyFromQueueObj(logger, "ReportDataSource", obj, c.queues.reportDataSourceQueue); ok {
		err := c.syncReportDataSource(logger, key)
		c.handleErr(logger, err, "ReportDataSource", key, c.queues.reportDataSourceQueue)
	}
	return true
}

func (c *Metering) syncReportDataSource(logger log.FieldLogger, key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		logger.WithError(err).Errorf("invalid resource key :%s", key)
		return nil
	}

	logger = logger.WithField("datasource", name)
	reportDataSource, err := c.informers.Metering().V1alpha1().ReportDataSources().Lister().ReportDataSources(namespace).Get(name)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Infof("ReportDataSource %s does not exist anymore, deleting data associated with it", key)
			c.prometheusImporterDeletedDataSourceQueue <- name
			c.deleteReportDataSourceTable(name)
			return nil
		}
		return err
	}

	logger.Infof("syncing reportDataSource %s", reportDataSource.GetName())
	err = c.handleReportDataSource(logger, reportDataSource)
	if err != nil {
		logger.WithError(err).Errorf("error syncing reportDataSource %s", reportDataSource.GetName())
		return err
	}
	logger.Infof("successfully synced reportDataSource %s", reportDataSource.GetName())
	return nil
}

func (c *Metering) handleReportDataSourceDeleted(obj interface{}) {
	dataSource, ok := obj.(*cbTypes.ReportDataSource)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			c.logger.Errorf("Couldn't get object from tombstone %#v", obj)
			return
		}
		dataSource, ok = tombstone.Obj.(*cbTypes.ReportDataSource)
		if !ok {
			c.logger.Errorf("Tombstone contained object that is not a ReportDataSource %#v", obj)
			return
		}
	}
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(dataSource)
	if err != nil {
		c.logger.WithField("reportDataSource", dataSource.Name).WithError(err).Errorf("couldn't get key for object: %#v", dataSource)
		return
	}
	c.queues.reportDataSourceQueue.Add(key)
}

func (c *Metering) handleReportDataSource(logger log.FieldLogger, dataSource *cbTypes.ReportDataSource) error {
	dataSource = dataSource.DeepCopy()
	if dataSource.TableName == "" {
		logger.Infof("new dataSource discovered")
	} ***REMOVED*** {
		logger.Infof("existing dataSource discovered, tableName: %s", dataSource.TableName)
	}

	switch {
	case dataSource.Spec.Promsum != nil:
		return c.handlePrometheusMetricsDataSource(logger, dataSource)
	case dataSource.Spec.AWSBilling != nil:
		return c.handleAWSBillingDataSource(logger, dataSource)
	default:
		return fmt.Errorf("datasource %s: improperly con***REMOVED***gured missing promsum or awsBilling con***REMOVED***guration", dataSource.Name)
	}
}

func (c *Metering) handlePrometheusMetricsDataSource(logger log.FieldLogger, dataSource *cbTypes.ReportDataSource) error {
	if dataSource.TableName == "" {
		storage := dataSource.Spec.Promsum.Storage
		tableName := dataSourceTableName(dataSource.Name)
		err := c.createTableForStorage(logger, dataSource, "ReportDataSource", dataSource.Name, storage, tableName, promsumHiveColumns)
		if err != nil {
			return err
		}

		err = c.updateDataSourceTableName(logger, dataSource, tableName)
		if err != nil {
			logger.WithError(err).Errorf("failed to update ReportDataSource TableName ***REMOVED***eld %q", tableName)
			return err
		}
	}

	c.prometheusImporterNewDataSourceQueue <- dataSource

	return nil
}

func (c *Metering) handleAWSBillingDataSource(logger log.FieldLogger, dataSource *cbTypes.ReportDataSource) error {
	source := dataSource.Spec.AWSBilling.Source
	if source == nil {
		return fmt.Errorf("datasource %q: improperly con***REMOVED***gured datasource, source is empty", dataSource.Name)
	}

	manifestRetriever := aws.NewManifestRetriever(source.Region, source.Bucket, source.Pre***REMOVED***x)

	manifests, err := manifestRetriever.RetrieveManifests()
	if err != nil {
		return err
	}

	if len(manifests) == 0 {
		logger.Warnf("datasource %q has no report manifests in it's bucket, the ***REMOVED***rst report has likely not been generated yet", dataSource.Name)
		return nil
	}

	if dataSource.TableName == "" {
		tableName := dataSourceTableName(dataSource.Name)
		logger.Debugf("creating AWS Billing DataSource table %s pointing to s3 bucket %s at pre***REMOVED***x %s", tableName, source.Bucket, source.Pre***REMOVED***x)
		err = c.createAWSUsageTable(logger, dataSource, tableName, source.Bucket, source.Pre***REMOVED***x, manifests)
		if err != nil {
			return err
		}

		logger.Debugf("successfully created AWS Billing DataSource table %s pointing to s3 bucket %s at pre***REMOVED***x %s", tableName, source.Bucket, source.Pre***REMOVED***x)
		err = c.updateDataSourceTableName(logger, dataSource, tableName)
		if err != nil {
			return err
		}
	}

	c.prestoTablePartitionQueue <- dataSource
	return nil
}

func (c *Metering) updateDataSourceTableName(logger log.FieldLogger, dataSource *cbTypes.ReportDataSource, tableName string) error {
	dataSource.TableName = tableName
	_, err := c.chargebackClient.MeteringV1alpha1().ReportDataSources(dataSource.Namespace).Update(dataSource)
	if err != nil {
		logger.WithError(err).Errorf("failed to update ReportDataSource table name for %q", dataSource.Name)
		return err
	}
	return nil
}

func (c *Metering) deleteReportDataSourceTable(name string) {
	tableName := dataSourceTableName(name)
	err := hive.ExecuteDropTable(c.hiveQueryer, tableName, true)
	if err != nil {
		c.logger.WithError(err).Error("unable to drop ReportDataSource table")
	}
	c.logger.Infof("successfully deleted table %s", tableName)
}
