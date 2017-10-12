package chargeback

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/cache"

	cbTypes "github.com/coreos-inc/kube-chargeback/pkg/apis/chargeback/v1alpha1"
	"github.com/coreos-inc/kube-chargeback/pkg/hive"
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

	if dataStore.Spec.Promsum == nil {
		log.Infof("datastore %q: skipping, not promsum datastore", dataStore.Name)
		return nil
	}

	storage := dataStore.Spec.Promsum.Storage
	if storage == nil {
		return fmt.Errorf("datastore %q: improperly con***REMOVED***gured datastore, storage is empty", dataStore.Name)
	}
	if storage.S3 == nil {
		return fmt.Errorf("datastore %q: unsupported storage type (must be s3)", dataStore.Name)
	}

	replacer := strings.NewReplacer("-", "_", ".", "_")
	tableName := fmt.Sprintf("datastore_%s", replacer.Replace(dataStore.Name))

	logger.Debugf("creating table %s pointing to s3 bucket %s at pre***REMOVED***x %s", tableName, storage.S3.Bucket, storage.S3.Pre***REMOVED***x)
	if err := hive.CreatePromsumTable(c.hiveConn, tableName, storage.S3.Bucket, storage.S3.Pre***REMOVED***x); err != nil {
		return err
	}
	dataStore.TableName = tableName

	_, err := c.chargebackClient.ChargebackV1alpha1().ReportDataStores(dataStore.Namespace).Update(dataStore)
	if err != nil {
		logger.WithError(err).Errorf("failed to update ReportDataStore table name for %q", dataStore.Name)
		return err
	}

	return nil
}
