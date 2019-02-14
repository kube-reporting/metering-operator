/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this ***REMOVED***le except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the speci***REMOVED***c language governing permissions and
limitations under the License.
*/

package resourcelock

import (
	"encoding/json"
	"errors"
	"fmt"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
)

// TODO: This is almost a exact replica of Endpoints lock.
// going forwards as we self host more and more components
// and use Con***REMOVED***gMaps as the means to pass that con***REMOVED***guration
// data we will likely move to deprecate the Endpoints lock.

type Con***REMOVED***gMapLock struct {
	// Con***REMOVED***gMapMeta should contain a Name and a Namespace of a
	// Con***REMOVED***gMapMeta object that the LeaderElector will attempt to lead.
	Con***REMOVED***gMapMeta metav1.ObjectMeta
	Client        corev1client.Con***REMOVED***gMapsGetter
	LockCon***REMOVED***g    ResourceLockCon***REMOVED***g
	cm            *v1.Con***REMOVED***gMap
}

// Get returns the election record from a Con***REMOVED***gMap Annotation
func (cml *Con***REMOVED***gMapLock) Get() (*LeaderElectionRecord, error) {
	var record LeaderElectionRecord
	var err error
	cml.cm, err = cml.Client.Con***REMOVED***gMaps(cml.Con***REMOVED***gMapMeta.Namespace).Get(cml.Con***REMOVED***gMapMeta.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	if cml.cm.Annotations == nil {
		cml.cm.Annotations = make(map[string]string)
	}
	if recordBytes, found := cml.cm.Annotations[LeaderElectionRecordAnnotationKey]; found {
		if err := json.Unmarshal([]byte(recordBytes), &record); err != nil {
			return nil, err
		}
	}
	return &record, nil
}

// Create attempts to create a LeaderElectionRecord annotation
func (cml *Con***REMOVED***gMapLock) Create(ler LeaderElectionRecord) error {
	recordBytes, err := json.Marshal(ler)
	if err != nil {
		return err
	}
	cml.cm, err = cml.Client.Con***REMOVED***gMaps(cml.Con***REMOVED***gMapMeta.Namespace).Create(&v1.Con***REMOVED***gMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cml.Con***REMOVED***gMapMeta.Name,
			Namespace: cml.Con***REMOVED***gMapMeta.Namespace,
			Annotations: map[string]string{
				LeaderElectionRecordAnnotationKey: string(recordBytes),
			},
		},
	})
	return err
}

// Update will update an existing annotation on a given resource.
func (cml *Con***REMOVED***gMapLock) Update(ler LeaderElectionRecord) error {
	if cml.cm == nil {
		return errors.New("con***REMOVED***gmap not initialized, call get or create ***REMOVED***rst")
	}
	recordBytes, err := json.Marshal(ler)
	if err != nil {
		return err
	}
	cml.cm.Annotations[LeaderElectionRecordAnnotationKey] = string(recordBytes)
	cml.cm, err = cml.Client.Con***REMOVED***gMaps(cml.Con***REMOVED***gMapMeta.Namespace).Update(cml.cm)
	return err
}

// RecordEvent in leader election while adding meta-data
func (cml *Con***REMOVED***gMapLock) RecordEvent(s string) {
	events := fmt.Sprintf("%v %v", cml.LockCon***REMOVED***g.Identity, s)
	cml.LockCon***REMOVED***g.EventRecorder.Eventf(&v1.Con***REMOVED***gMap{ObjectMeta: cml.cm.ObjectMeta}, v1.EventTypeNormal, "LeaderElection", events)
}

// Describe is used to convert details on current resource lock
// into a string
func (cml *Con***REMOVED***gMapLock) Describe() string {
	return fmt.Sprintf("%v/%v", cml.Con***REMOVED***gMapMeta.Namespace, cml.Con***REMOVED***gMapMeta.Name)
}

// returns the Identity of the lock
func (cml *Con***REMOVED***gMapLock) Identity() string {
	return cml.LockCon***REMOVED***g.Identity
}
