package chargeback

import (
	"fmt"

	cbTypes "github.com/coreos-inc/kube-chargeback/pkg/apis/chargeback/v1alpha1"
	cbListers "github.com/coreos-inc/kube-chargeback/pkg/generated/listers/chargeback/v1alpha1"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/labels"
)

func (c *Chargeback) getDefaultStorageLocation(lister cbListers.StorageLocationLister) (*cbTypes.StorageLocation, error) {
	storageLocations, err := lister.StorageLocations(c.cfg.Namespace).List(labels.Everything())
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

func (c *Chargeback) getStorageSpec(logger log.FieldLogger, storage *cbTypes.StorageLocationRef, kind string) (cbTypes.StorageLocationSpec, error) {
	storageLister := c.informers.Chargeback().V1alpha1().StorageLocations().Lister()
	var storageSpec cbTypes.StorageLocationSpec
	// Nothing speci***REMOVED***ed, try to use default storage location
	if storage == nil || (storage.StorageSpec == nil && storage.StorageLocationName == "") {
		logger.Infof("%s storage does not have a spec or storageLocationName set, using default storage location", kind)
		storageLocation, err := c.getDefaultStorageLocation(storageLister)
		if err != nil {
			return storageSpec, err
		}
		if storageLocation == nil {
			return storageSpec, fmt.Errorf("invalid %s, storage spec or storageLocationName not set and cluster has no default StorageLocation", kind)
		}

		storageSpec = storageLocation.Spec
	} ***REMOVED*** if storage.StorageLocationName != "" { // Speci***REMOVED***c storage location speci***REMOVED***ed
		logger.Infof("%s con***REMOVED***gured to use StorageLocation %s", kind, storage.StorageLocationName)
		storageLocation, err := storageLister.StorageLocations(c.cfg.Namespace).Get(storage.StorageLocationName)
		if err != nil {
			return storageSpec, err
		}
		storageSpec = storageLocation.Spec
	} ***REMOVED*** if storage.StorageSpec != nil { // Storage location is inlined in the datastore
		storageSpec = *storage.StorageSpec
	}
	return storageSpec, nil
}
