package chargeback

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"

	cbTypes "github.com/coreos-inc/kube-chargeback/pkg/apis/chargeback/v1alpha1"
	"github.com/coreos-inc/kube-chargeback/pkg/aws"
	cbListers "github.com/coreos-inc/kube-chargeback/pkg/generated/listers/chargeback/v1alpha1"
	"github.com/coreos-inc/kube-chargeback/pkg/hive"
)

func (c *Chargeback) runReportDataStoreWorker() {
	logger := c.logger.WithField("component", "reportDataStoreWorker")
	logger.Infof("ReportDataStore worker started")
	for c.processReportDataStore(logger) {

	}
}

func (c *Chargeback) processReportDataStore(logger log.FieldLogger) bool {
	key, quit := c.informers.reportDataStoreQueue.Get()
	if quit {
		return false
	}
	defer c.informers.reportDataStoreQueue.Done(key)

	logger = logger.WithFields(newLogIdenti***REMOVED***er())
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

	var storageSpec cbTypes.StorageLocationSpec
	// Nothing speci***REMOVED***ed, try to use default storage location
	if storage == nil || (storage.StorageSpec == nil && storage.StorageLocationName == "") {
		logger.Info("reportDataStore does not have a storageSpec or storageLocationName set, using default storage location")
		storageLocation, err := c.getDefaultStorageLocation(c.informers.storageLocationLister)
		if err != nil {
			return err
		}
		if storageLocation == nil {
			return fmt.Errorf("invalid promsum DataStore, no storageSpec or storageLocationName and cluster has no default StorageLocation")
		}

		storageSpec = storageLocation.Spec
	} ***REMOVED*** if storage.StorageLocationName != "" { // Speci***REMOVED***c storage location speci***REMOVED***ed
		logger.Infof("reportDataStore con***REMOVED***gured to use StorageLocation %s", storage.StorageLocationName)
		storageLocation, err := c.informers.storageLocationLister.StorageLocations(c.namespace).Get(storage.StorageLocationName)
		if err != nil {
			return err
		}
		storageSpec = storageLocation.Spec
	} ***REMOVED*** if storage.StorageSpec != nil { // Storage location is inlined in the datastore
		storageSpec = *storage.StorageSpec
	}

	if storageSpec.Local != nil {
		logger.Debugf("creating local table %s", tableName)
		err := hive.CreateLocalPromsumTable(c.hiveQueryer, tableName)
		if err != nil {
			return err
		}
	} ***REMOVED*** if storageSpec.S3 != nil {
		logger.Debugf("creating table %s backed by s3 bucket %s at pre***REMOVED***x %s", tableName, storageSpec.S3.Bucket, storageSpec.S3.Pre***REMOVED***x)
		err := hive.CreatePromsumTable(c.hiveQueryer, tableName, storageSpec.S3.Bucket, storageSpec.S3.Pre***REMOVED***x)
		if err != nil {
			return err
		}
		return nil
	} ***REMOVED*** {
		return fmt.Errorf("storage incorrectly con***REMOVED***gured on datastore %s", dataStore.Name)
	}

	logger.Debugf("successfully created table %s", tableName)

	return c.updateDataStoreTableName(logger, dataStore, tableName)
}

func (c *Chargeback) getDefaultStorageLocation(lister cbListers.StorageLocationLister) (*cbTypes.StorageLocation, error) {
	storageLocations, err := c.informers.storageLocationLister.StorageLocations(c.namespace).List(labels.Everything())
	if err != nil {
		return nil, err
	}

	var defaultStorageLocations []*cbTypes.StorageLocation

	for _, storageLocation := range storageLocations {
		if storageLocation.Annotations[cbTypes.IsDefaultStorageLocationAnnotation] == "true" {
			defaultStorageLocations = append(defaultStorageLocations, storageLocation)
		}
	}

	if len(defaultStorageLocations) == 0 {
		return nil, nil
	}

	if len(defaultStorageLocations) > 1 {
		c.logger.Infof("getDefaultStorageLocation %s default storageLocations found", len(defaultStorageLocations))
		return nil, fmt.Errorf("%d defaultStorageLocations were found", len(defaultStorageLocations))
	}

	return defaultStorageLocations[0], nil

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
		logger.Warnf("datastore %q has no report manifests in it's bucket, the ***REMOVED***rst report has likely not been generated yet", dataStore.Name)
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
