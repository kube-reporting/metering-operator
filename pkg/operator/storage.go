package operator

import (
	"fmt"

	cbTypes "github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1"
	cbListers "github.com/operator-framework/operator-metering/pkg/generated/listers/metering/v1alpha1"
	"github.com/operator-framework/operator-metering/pkg/hive"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/labels"
)

func (op *Reporting) getDefaultStorageLocation(lister cbListers.StorageLocationLister, namespace string) (*cbTypes.StorageLocation, error) {
	storageLocations, err := lister.StorageLocations(namespace).List(labels.Everything())
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
		op.logger.Infof("getDefaultStorageLocation: %d default storageLocations found", len(defaultStorageLocations))
		return nil, fmt.Errorf("%d defaultStorageLocations were found", len(defaultStorageLocations))
	}

	return defaultStorageLocations[0], nil

}

func (op *Reporting) getStorageSpec(logger log.FieldLogger, storage *cbTypes.StorageLocationRef, kind, namespace string) (cbTypes.StorageLocationSpec, error) {
	storageLister := op.storageLocationLister
	var storageSpec cbTypes.StorageLocationSpec
	// Nothing specified, try to use default storage location
	if storage == nil || (storage.StorageSpec == nil && storage.StorageLocationName == "") {
		logger.Debugf("%s storage does not have a spec or storageLocationName set, getting default storage location in namespace %s", kind, namespace)
		storageLocation, err := op.getDefaultStorageLocation(storageLister, namespace)
		if err != nil {
			return storageSpec, err
		}
		if storageLocation == nil {
			return storageSpec, fmt.Errorf("invalid %s, storage spec or storageLocationName not set and namespace %s has no default StorageLocation", kind, namespace)
		}

		storageSpec = storageLocation.Spec
	} else if storage.StorageLocationName != "" { // Specific storage location specified
		logger.Debugf("%s configured to use StorageLocation %s", kind, storage.StorageLocationName)
		storageLocation, err := storageLister.StorageLocations(namespace).Get(storage.StorageLocationName)
		if err != nil {
			return storageSpec, err
		}
		storageSpec = storageLocation.Spec
	} else if storage.StorageSpec != nil { // Storage location is inlined in the datastore
		storageSpec = *storage.StorageSpec
	}
	return storageSpec, nil
}

func (op *Reporting) getHiveTableProperties(logger log.FieldLogger, storage *cbTypes.StorageLocationRef, kind, namespace string) (*hive.TableProperties, error) {
	storageSpec, err := op.getStorageSpec(logger, storage, kind, namespace)
	if err != nil {
		return nil, err
	}
	if storageSpec.Hive != nil {
		props := hive.TableProperties(storageSpec.Hive.TableProperties)
		return &props, nil
	} else {
		return nil, fmt.Errorf("incorrect storage configuration, must configure spec.hive")
	}
}
