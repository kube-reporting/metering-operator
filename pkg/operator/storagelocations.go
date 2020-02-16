package operator

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"

	metering "github.com/operator-framework/operator-metering/pkg/apis/metering/v1"
	// "github.com/operator-framework/operator-metering/pkg/hive"
	"github.com/operator-framework/operator-metering/pkg/operator/reportingutil"
	"github.com/operator-framework/operator-metering/pkg/util/slice"
)

const (
	storageLocationFinalizer = metering.GroupName + "/storagelocation"
)

func (op *Reporting) runStorageLocationWorker() {
	logger := op.logger.WithField("component", "storageLocationWorker")
	logger.Infof("StorageLocation worker started")
	const maxRequeues = -1
	for op.processResource(logger, op.syncStorageLocation, "StorageLocation", op.storageLocationQueue, maxRequeues) {
	}
}

func (op *Reporting) syncStorageLocation(logger log.FieldLogger, key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		logger.WithError(err).Errorf("invalid resource key :%s", key)
		return nil
	}

	logger = logger.WithFields(log.Fields{"storageLocation": name, "namespace": namespace})

	storageLocationLister := op.storageLocationLister
	storageLocation, err := storageLocationLister.StorageLocations(namespace).Get(name)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Infof("StorageLocation %s does not exist anymore", key)
			return nil
		}
		return err
	}
	storage := storageLocation.DeepCopy()

	if storage.DeletionTimestamp != nil {
		logger.Infof("StorageLocation is marked for deletion, performing cleanup")
		err := op.deleteStorage(storage)
		if err != nil {
			return err
		}

		_, err = op.removeStorageLocationFinalizer(storage)
		return err
	}

	logger.Infof("syncing StorageLocation %s", storage.GetName())
	err = op.handleStorageLocation(logger, storage)
	if err != nil {
		logger.WithError(err).Errorf("error syncing StorageLocation %s", storage.GetName())
		return err
	}
	logger.Infof("successfully synced StorageLocation %s", storage.GetName())
	return nil
}

func (op *Reporting) handleStorageLocation(logger log.FieldLogger, storageLocation *metering.StorageLocation) error {
	if op.cfg.EnableFinalizers && storageLocationNeedsFinalizer(storageLocation) {
		var err error
		storageLocation, err = op.addStorageLocationFinalizer(storageLocation)
		if err != nil {
			return err
		}
	}

	var needsUpdate bool

	switch {
	case storageLocation.Spec.Hive != nil:
		if storageLocation.Spec.Hive.UnmanagedDatabase {
			logger.Infof("StorageLocation %s is unmanaged", storageLocation.Name)
			if storageLocation.Status.Hive.DatabaseName != storageLocation.Spec.Hive.DatabaseName {
				storageLocation.Status.Hive.DatabaseName = storageLocation.Spec.Hive.DatabaseName
				needsUpdate = true
			}
			if storageLocation.Status.Hive.Location != storageLocation.Spec.Hive.Location {
				storageLocation.Status.Hive.Location = storageLocation.Spec.Hive.Location
				needsUpdate = true
			}
		} ***REMOVED*** {
			if storageLocation.Spec.Hive.DatabaseName == "" {
				return fmt.Errorf("spec.hive.databaseName is required if spec.hive is set")
			}
			if !reportingutil.IsValidSQLIdenti***REMOVED***er(storageLocation.Spec.Hive.DatabaseName) {
				return fmt.Errorf("spec.hive.databaseName %s is invalid, must contain only alpha numeric values, underscores, and start with a letter or underscore", storageLocation.Spec.Hive.DatabaseName)
			}
			if storageLocation.Status.Hive.DatabaseName == "" {
				logger.Infof("Using the useLocalStorage option")
				catalog := "hive"
				schema := "metering"
				err := op.prestoTableManager.CreateSchema(catalog, schema)
				if err != nil {
					return fmt.Errorf("failed to create the %s.%s Presto schema: %v", catalog, schema, err)
				}
				logger.Infof("Create the %s.%s Presto schema", catalog, schema)
				storageLocation.Status.Hive.DatabaseName = storageLocation.Spec.Hive.DatabaseName
				storageLocation.Status.Hive.Location = fmt.Sprintf("%s.%s", catalog, schema)
				needsUpdate = true
			}
		}

	default:
		return fmt.Errorf("only Hive storage is supported currently")
	}

	if needsUpdate {
		var err error
		storageLocation, err = op.meteringClient.MeteringV1().StorageLocations(storageLocation.Namespace).Update(storageLocation)
		if err != nil {
			return fmt.Errorf("unable to update StorageLocation %s status: %s", storageLocation.Name, err)
		}

		if err = op.queueDependentsOfStorageLocation(storageLocation); err != nil {
			logger.WithError(err).Errorf("error queuing dependents of StorageLocation %s", storageLocation.Name)
		}
	}
	return nil
}

func (op *Reporting) addStorageLocationFinalizer(storageLocation *metering.StorageLocation) (*metering.StorageLocation, error) {
	storageLocation.Finalizers = append(storageLocation.Finalizers, storageLocationFinalizer)
	newStorageLocation, err := op.meteringClient.MeteringV1().StorageLocations(storageLocation.Namespace).Update(storageLocation)
	logger := op.logger.WithFields(log.Fields{"storageLocation": storageLocation.Name, "namespace": storageLocation.Namespace})
	if err != nil {
		logger.WithError(err).Errorf("error adding %s ***REMOVED***nalizer to StorageLocation: %s/%s", storageLocationFinalizer, storageLocation.Namespace, storageLocation.Name)
		return nil, err
	}
	logger.Infof("added %s ***REMOVED***nalizer to StorageLocation: %s/%s", storageLocationFinalizer, storageLocation.Namespace, storageLocation.Name)
	return newStorageLocation, nil
}

func (op *Reporting) removeStorageLocationFinalizer(storageLocation *metering.StorageLocation) (*metering.StorageLocation, error) {
	if !slice.ContainsString(storageLocation.ObjectMeta.Finalizers, storageLocationFinalizer, nil) {
		return storageLocation, nil
	}
	storageLocation.Finalizers = slice.RemoveString(storageLocation.Finalizers, storageLocationFinalizer, nil)
	newStorageLocation, err := op.meteringClient.MeteringV1().StorageLocations(storageLocation.Namespace).Update(storageLocation)
	logger := op.logger.WithFields(log.Fields{"storageLocation": storageLocation.Name, "namespace": storageLocation.Namespace})
	if err != nil {
		logger.WithError(err).Errorf("error removing %s ***REMOVED***nalizer from StorageLocation: %s/%s", storageLocationFinalizer, storageLocation.Namespace, storageLocation.Name)
		return nil, err
	}
	logger.Infof("removed %s ***REMOVED***nalizer from StorageLocation: %s/%s", storageLocationFinalizer, storageLocation.Namespace, storageLocation.Name)
	return newStorageLocation, nil
}

func storageLocationNeedsFinalizer(storageLocation *metering.StorageLocation) bool {
	return storageLocation.ObjectMeta.DeletionTimestamp == nil && !slice.ContainsString(storageLocation.ObjectMeta.Finalizers, storageLocationFinalizer, nil)
}

func (op *Reporting) deleteStorage(storageLocation *metering.StorageLocation) error {
	if storageLocation.Spec.Hive != nil {
		if !storageLocation.Spec.Hive.UnmanagedDatabase && storageLocation.Status.Hive.DatabaseName != "" {
			return op.hiveDatabaseManager.DropDatabase(storageLocation.Status.Hive.DatabaseName, true, false)
		}
	}
	return nil
}

func (op *Reporting) queueDependentsOfStorageLocation(storageLocation *metering.StorageLocation) error {
	reports, err := op.reportLister.Reports(storageLocation.Namespace).List(labels.Everything())
	if err != nil {
		return err
	}
	datasources, err := op.reportDataSourceLister.ReportDataSources(storageLocation.Namespace).List(labels.Everything())
	if err != nil {
		return err
	}
	var errs []string
	for _, datasource := range datasources {
		if datasource.Status.TableRef.Name != "" {
			continue
		}
		switch {
		case datasource.Spec.PrometheusMetricsImporter != nil:
			storage, err := op.getStorage(datasource.Spec.PrometheusMetricsImporter.Storage, datasource.Namespace)
			if err != nil {
				errs = append(errs, err.Error())
				continue
			}
			if storage.Name == storageLocation.Name {
				op.enqueueReportDataSource(datasource)
			}
		case datasource.Spec.AWSBilling != nil && datasource.Spec.AWSBilling.DatabaseName == "":
			storage, err := op.getStorage(nil, datasource.Namespace)
			if err != nil {
				errs = append(errs, err.Error())
				continue
			}
			if storage.Name == storageLocation.Name {
				op.enqueueReportDataSource(datasource)
			}
		}
	}
	for _, report := range reports {
		if report.Status.TableRef.Name != "" {
			continue
		}
		storage, err := op.getStorage(report.Spec.Output, report.Namespace)
		if err != nil {
			errs = append(errs, err.Error())
			continue
		}
		if storage.Name == storageLocation.Name {
			op.enqueueReport(report)
		}
	}

	if len(errs) != 0 {
		return fmt.Errorf("got errors queuing dependents of StorageLocation %s: %s", storageLocation.Name, strings.Join(errs, ", "))
	}

	return nil
}
