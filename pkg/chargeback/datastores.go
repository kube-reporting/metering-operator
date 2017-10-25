package chargeback

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/cache"

	cbTypes "github.com/coreos-inc/kube-chargeback/pkg/apis/chargeback/v1alpha1"
	"github.com/coreos-inc/kube-chargeback/pkg/aws"
	"github.com/coreos-inc/kube-chargeback/pkg/hive"
	"github.com/coreos-inc/kube-chargeback/pkg/presto"
)

func (c *Chargeback) runReportDataStoreWorker() {
	for c.processReportDataStore() {

	}
}

func (c *Chargeback) processReportDataStore() bool {
	key, quit := c.informers.reportDataStoreQueue.Get()
	if quit {
		return false
	}
	defer c.informers.reportDataStoreQueue.Done(key)

	err := c.syncReportDataStore(key.(string))
	c.handleErr(err, "ReportDataStore", key, c.informers.reportDataStoreQueue)
	return true
}

func (c *Chargeback) syncReportDataStore(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		c.logger.WithError(err).Errorf("invalid resource key :%s", key)
		return nil
	}

	reportDataStore, err := c.informers.reportDataStoreLister.ReportDataStores(namespace).Get(name)
	if err != nil {
		if apierrors.IsNotFound(err) {
			c.logger.Infof("ReportDataStore %s does not exist anymore", key)
			return nil
		}
		return err
	}

	c.logger.Infof("syncing reportDataStore %s", reportDataStore.GetName())
	err = c.handleReportDataStore(reportDataStore)
	if err != nil {
		c.logger.WithError(err).Errorf("error syncing reportDataStore %s", reportDataStore.GetName())
		return err
	}
	c.logger.Infof("successfully synced reportDataStore %s", reportDataStore.GetName())
	return nil
}

func (c *Chargeback) handleReportDataStore(dataStore *cbTypes.ReportDataStore) error {
	dataStore = dataStore.DeepCopy()

	logger := c.logger.WithFields(log.Fields{
		"name": dataStore.Name,
	})

	if dataStore.TableName == "" {
		logger.Infof("new dataStore discovered")
	} ***REMOVED*** {
		logger.Infof("existing dataStore discovered, tableName: %s", dataStore.TableName)
		return nil
	}

	switch {
	case dataStore.Spec.Promsum != nil:
		return c.handlePromsumDataStore(logger, dataStore)
	case dataStore.Spec.AWSBilling != nil:
		return c.handleAWSBillingDataStore(logger, dataStore)
	default:
		return fmt.Errorf("datastore %s: improperly con***REMOVED***gured missing promsum or awsBilling con***REMOVED***guration", dataStore.Name)
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

		// There's currently a strange issue where selects in presto will fail
		// unless an insert has been made ***REMOVED***rst. Don't ask me why.

		// After creating foobar via the hive cli:

		//presto:default> select * from foobar;

		//Query 20171025_185138_00002_cu5nq, FAILED, 1 node
		//Splits: 16 total, 0 done (0.00%)
		//0:05 [0 rows, 0B] [0 rows/s, 0B/s]

		//Query 20171025_185138_00002_cu5nq failed: Partition location does not exist: ***REMOVED***le:/user/hive/warehouse/foobar
		_, err = presto.ExecuteSelect(c.prestoConn, fmt.Sprintf("INSERT INTO %s VALUES ('',0.0,null,0.0,map(ARRAY[],ARRAY[]))", tableName))
		if err != nil {
			return err
		}
		_, err = presto.ExecuteSelect(c.prestoConn, fmt.Sprintf("DELETE FROM %s", tableName))
		if err != nil {
			return err
		}
	case storage.S3 != nil:
		// store the data in S3
		logger.Debugf("creating table %s backed by s3 bucket %s at pre***REMOVED***x %s", tableName, storage.S3.Bucket, storage.S3.Pre***REMOVED***x)
		err := hive.CreatePromsumTable(c.hiveQueryer, tableName, storage.S3.Bucket, storage.S3.Pre***REMOVED***x)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("storage incorrectly con***REMOVED***gured on datastore %s", dataStore.Name)
	}
	logger.Debugf("successfully created table %s", tableName)

	return c.updateDataStoreTableName(logger, dataStore, tableName)
}

func (c *Chargeback) handleAWSBillingDataStore(logger log.FieldLogger, dataStore *cbTypes.ReportDataStore) error {
	source := dataStore.Spec.AWSBilling.Source
	if source == nil {
		return fmt.Errorf("datastore %q: improperly con***REMOVED***gured datastore, source is empty", dataStore.Name)
	}

	manifestRetriever, err := aws.NewManifestRetriever(source.Bucket, source.Pre***REMOVED***x)
	if err != nil {
		return err
	}

	manifests, err := manifestRetriever.RetrieveManifests()
	if err != nil {
		return err
	}

	if len(manifests) == 0 {
		logger.Infof("datastore %q has no report manifests in it's bucket, the ***REMOVED***rst report has likely not been generated yet", dataStore.Name)
		return nil
	}

	tableName := dataStoreTableName(dataStore.Name)
	logger.Debugf("creating AWS Billing DataSource table %s pointing to s3 bucket %s at pre***REMOVED***x %s", tableName, source.Bucket, source.Pre***REMOVED***x)
	err = hive.CreateAWSUsageTable(c.hiveQueryer, tableName, source.Bucket, source.Pre***REMOVED***x, manifests)
	if err != nil {
		return err
	}
	logger.Debugf("successfully created AWS Billing DataSource table %s pointing to s3 bucket %s at pre***REMOVED***x %s", tableName, source.Bucket, source.Pre***REMOVED***x)

	logger.Debugf("updating table %s partitions", tableName)
	err = hive.UpdateAWSUsageTable(c.hiveQueryer, tableName, source.Bucket, source.Pre***REMOVED***x, manifests)
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
