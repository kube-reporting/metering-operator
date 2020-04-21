package operator

import (
	"fmt"

	metering "github.com/kubernetes-reporting/metering-operator/pkg/apis/metering/v1"
	cbListers "github.com/kubernetes-reporting/metering-operator/pkg/generated/listers/metering/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func (op *Reporting) getDefaultStorageLocation(lister cbListers.StorageLocationLister, namespace string) (*metering.StorageLocation, error) {
	storageLocations, err := lister.StorageLocations(namespace).List(labels.Everything())
	if err != nil {
		return nil, err
	}

	var defaultStorageLocations []*metering.StorageLocation

	for _, storageLocation := range storageLocations {
		if storageLocation.Annotations[metering.IsDefaultStorageLocationAnnotation] == "true" {
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

func (op *Reporting) getStorage(storage *metering.StorageLocationRef, namespace string) (*metering.StorageLocation, error) {
	// Nothing specified, try to use default storage location
	if storage == nil || storage.StorageLocationName == "" {
		storageLocation, err := op.getDefaultStorageLocation(op.storageLocationLister, namespace)
		if err != nil {
			return storageLocation, err
		}
		if storageLocation == nil {
			return storageLocation, fmt.Errorf("storage spec or storageLocationName not set and namespace %s has no default StorageLocation", namespace)
		}
		return storageLocation, nil
	} else if storage.StorageLocationName != "" { // Specific storage location specified
		return op.storageLocationLister.StorageLocations(namespace).Get(storage.StorageLocationName)
	}
	return nil, fmt.Errorf("no default storageLocation and storageLocationName is empty")
}

func (op *Reporting) getHiveStorage(storageRef *metering.StorageLocationRef, namespace string) (*metering.StorageLocation, error) {
	storageLocation, err := op.getStorage(storageRef, namespace)
	if err != nil {
		return nil, err
	}

	if storageLocation.Spec.Hive == nil {
		return nil, fmt.Errorf("incorrect storage configuration, has no Hive storage configuration")
	}
	return storageLocation, nil
}
