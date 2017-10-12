package chargeback

import (
	"fmt"
	"strings"

	cbTypes "github.com/coreos-inc/kube-chargeback/pkg/apis/chargeback/v1alpha1"
	"github.com/coreos-inc/kube-chargeback/pkg/hive"
	log "github.com/sirupsen/logrus"
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
	indexer := c.informers.reportDataStoreInformer.GetIndexer()
	obj, exists, err := indexer.GetByKey(key)
	if err != nil {
		c.logger.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		c.logger.Infof("ReportDataStore %s does not exist anymore", key)
	} else {
		reportDataStore := obj.(*cbTypes.ReportDataStore)
		c.logger.Infof("syncing reportDataStore %s", reportDataStore.GetName())
		err = c.handleReportDataStore(reportDataStore)
		if err != nil {
			c.logger.WithError(err).Errorf("error syncing reportDataStore %s", reportDataStore.GetName())
		}
		c.logger.Infof("successfully synced reportDataStore %s", reportDataStore.GetName())
	}
	return nil
}

func (c *Chargeback) handleReportDataStore(dataStore *cbTypes.ReportDataStore) error {
	dataStore = dataStore.DeepCopy()

	logger := c.logger.WithFields(log.Fields{
		"name": dataStore.Name,
	})

	if dataStore.TableName == "" {
		logger.Infof("new dataStore discovered")
	} else {
		logger.Infof("existing dataStore discovered, tableName: %s", dataStore.TableName)
		return nil
	}

	replacer := strings.NewReplacer("-", "_", ".", "_")
	tableName := fmt.Sprintf("datastore_%s", replacer.Replace(dataStore.Name))
	bucket, prefix := dataStore.Spec.Storage.Bucket, dataStore.Spec.Storage.Prefix

	logger.Debugf("creating table %s pointing to s3 bucket %s at prefix %s", tableName, bucket, prefix)
	if err := hive.CreatePromsumTable(c.hiveConn, tableName, bucket, prefix); err != nil {
		return err
	}
	dataStore.TableName = tableName

	_, err := c.chargebackClient.ChargebackV1alpha1().ReportDataStores(c.namespace).Update(dataStore)
	if err != nil {
		logger.WithError(err).Errorf("failed to update ReportDataStore table name for %q", dataStore.Name)
		return err
	}

	return nil
}
