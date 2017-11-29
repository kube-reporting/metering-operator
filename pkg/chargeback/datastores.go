package chargeback

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/cache"

	cbTypes "github.com/coreos-inc/kube-chargeback/pkg/apis/chargeback/v1alpha1"
	"github.com/coreos-inc/kube-chargeback/pkg/aws"
	"github.com/coreos-inc/kube-chargeback/pkg/hive"
)

func (c *Chargeback) runReportDataStoreWorker() {
	logger := c.logger.WithField("component", "reportDataStoreWorker")
	for c.processReportDataStore(logger) {

	}
}

func (c *Chargeback) processReportDataStore(logger log.FieldLogger) bool {
	key, quit := c.informers.reportDataStoreQueue.Get()
	if quit {
		return false
	}
	defer c.informers.reportDataStoreQueue.Done(key)

	logger = logger.WithFields(newLogIdentifier())
	err := c.syncReportDataStore(logger, key.(string))
	c.handleErr(logger, err, "ReportDataStore", key, c.informers.reportDataStoreQueue)
	return true
}

func (c *Chargeback) syncReportDataStore(logger log.FieldLogger, key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		logger.WithError(err).Errorf("invalid resource key :%s", key)
		return nil
	}

	logger = logger.WithField("datastore", name)
	reportDataStore, err := c.informers.reportDataStoreLister.ReportDataStores(namespace).Get(name)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Infof("ReportDataStore %s does not exist anymore", key)
			return nil
		}
		return err
	}

	logger.Infof("syncing reportDataStore %s", reportDataStore.GetName())
	err = c.handleReportDataStore(logger, reportDataStore)
	if err != nil {
		logger.WithError(err).Errorf("error syncing reportDataStore %s", reportDataStore.GetName())
		return err
	}
	logger.Infof("successfully synced reportDataStore %s", reportDataStore.GetName())
	return nil
}

func (c *Chargeback) handleReportDataStore(logger log.FieldLogger, dataStore *cbTypes.ReportDataStore) error {
	dataStore = dataStore.DeepCopy()
	if dataStore.TableName == "" {
		logger.Infof("new dataStore discovered")
	} else {
		logger.Infof("existing dataStore discovered, tableName: %s", dataStore.TableName)
		return nil
	}

	switch {
	case dataStore.Spec.Promsum != nil:
		return c.handlePromsumDataStore(logger, dataStore)
	case dataStore.Spec.AWSBilling != nil:
		return c.handleAWSBillingDataStore(logger, dataStore)
	default:
		return fmt.Errorf("datastore %s: improperly configured missing promsum or awsBilling configuration", dataStore.Name)
	}
}

func (c *Chargeback) handlePromsumDataStore(logger log.FieldLogger, dataStore *cbTypes.ReportDataStore) error {
	storage := dataStore.Spec.Promsum.Storage
	tableName := dataStoreTableName(dataStore.Name)
	switch {
	case storage == nil || storage.Local != nil:
		logger.Debugf("creating local table %s", tableName)
		// store the data locally
		err := hive.CreateLocalPromsumTable(c.hiveQueryer, tableName)
		if err != nil {
			return err
		}
	case storage.S3 != nil:
		// store the data in S3
		logger.Debugf("creating table %s backed by s3 bucket %s at prefix %s", tableName, storage.S3.Bucket, storage.S3.Prefix)
		err := hive.CreatePromsumTable(c.hiveQueryer, tableName, storage.S3.Bucket, storage.S3.Prefix)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("storage incorrectly configured on datastore %s", dataStore.Name)
	}
	logger.Debugf("successfully created table %s", tableName)

	return c.updateDataStoreTableName(logger, dataStore, tableName)
}

func (c *Chargeback) handleAWSBillingDataStore(logger log.FieldLogger, dataStore *cbTypes.ReportDataStore) error {
	source := dataStore.Spec.AWSBilling.Source
	if source == nil {
		return fmt.Errorf("datastore %q: improperly configured datastore, source is empty", dataStore.Name)
	}

	manifestRetriever, err := aws.NewManifestRetriever(source.Bucket, source.Prefix)
	if err != nil {
		return err
	}

	manifests, err := manifestRetriever.RetrieveManifests()
	if err != nil {
		return err
	}

	if len(manifests) == 0 {
		logger.Infof("datastore %q has no report manifests in it's bucket, the first report has likely not been generated yet", dataStore.Name)
		return nil
	}

	tableName := dataStoreTableName(dataStore.Name)
	logger.Debugf("creating AWS Billing DataSource table %s pointing to s3 bucket %s at prefix %s", tableName, source.Bucket, source.Prefix)
	err = hive.CreateAWSUsageTable(c.hiveQueryer, tableName, source.Bucket, source.Prefix, manifests)
	if err != nil {
		return err
	}
	logger.Debugf("successfully created AWS Billing DataSource table %s pointing to s3 bucket %s at prefix %s", tableName, source.Bucket, source.Prefix)

	logger.Debugf("updating table %s partitions", tableName)
	err = hive.UpdateAWSUsageTable(c.hiveQueryer, tableName, source.Bucket, source.Prefix, manifests)
	if err != nil {
		return err
	}
	logger.Debugf("successfully updated table %s partitions", tableName)

	return c.updateDataStoreTableName(logger, dataStore, tableName)
}

func (c *Chargeback) updateDataStoreTableName(logger log.FieldLogger, dataStore *cbTypes.ReportDataStore, tableName string) error {
	dataStore.TableName = tableName
	_, err := c.chargebackClient.ChargebackV1alpha1().ReportDataStores(dataStore.Namespace).Update(dataStore)
	if err != nil {
		logger.WithError(err).Errorf("failed to update ReportDataStore table name for %q", dataStore.Name)
		return err
	}
	return nil
}
