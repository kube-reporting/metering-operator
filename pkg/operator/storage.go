package operator

import (
	"fmt"

	metering "github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1"
	cbListers "github.com/operator-framework/operator-metering/pkg/generated/listers/metering/v1alpha1"
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
	// Nothing speci***REMOVED***ed, try to use default storage location
	if storage == nil || storage.StorageLocationName == "" {
		storageLocation, err := op.getDefaultStorageLocation(op.storageLocationLister, namespace)
		if err != nil {
			return storageLocation, err
		}
		if storageLocation == nil {
			return storageLocation, fmt.Errorf("storage spec or storageLocationName not set and namespace %s has no default StorageLocation", namespace)
		}
		return storageLocation, nil
	} ***REMOVED*** if storage.StorageLocationName != "" { // Speci***REMOVED***c storage location speci***REMOVED***ed
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
		return nil, fmt.Errorf("incorrect storage con***REMOVED***guration, has no Hive storage con***REMOVED***guration")
	}
	return storageLocation, nil
}
