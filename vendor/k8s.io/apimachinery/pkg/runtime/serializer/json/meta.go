/*
Copyright 2014 The Kubernetes Authors.

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

package json

import (
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

// MetaFactory is used to store and retrieve the version and kind
// information for JSON objects in a serializer.
type MetaFactory interface {
	// Interpret should return the version and kind of the wire-format of
	// the object.
	Interpret(data []byte) (*schema.GroupVersionKind, error)
}

// DefaultMetaFactory is a default factory for versioning objects in JSON. The object
// in memory and in the default JSON serialization will use the "kind" and "apiVersion"
// ***REMOVED***elds.
var DefaultMetaFactory = SimpleMetaFactory{}

// SimpleMetaFactory provides default methods for retrieving the type and version of objects
// that are identi***REMOVED***ed with an "apiVersion" and "kind" ***REMOVED***elds in their JSON
// serialization. It may be parameterized with the names of the ***REMOVED***elds in memory, or an
// optional list of base structs to search for those ***REMOVED***elds in memory.
type SimpleMetaFactory struct {
}

// Interpret will return the APIVersion and Kind of the JSON wire-format
// encoding of an object, or an error.
func (SimpleMetaFactory) Interpret(data []byte) (*schema.GroupVersionKind, error) {
	***REMOVED***ndKind := struct {
		// +optional
		APIVersion string `json:"apiVersion,omitempty"`
		// +optional
		Kind string `json:"kind,omitempty"`
	}{}
	if err := json.Unmarshal(data, &***REMOVED***ndKind); err != nil {
		return nil, fmt.Errorf("couldn't get version/kind; json parse error: %v", err)
	}
	gv, err := schema.ParseGroupVersion(***REMOVED***ndKind.APIVersion)
	if err != nil {
		return nil, err
	}
	return &schema.GroupVersionKind{Group: gv.Group, Version: gv.Version, Kind: ***REMOVED***ndKind.Kind}, nil
}
