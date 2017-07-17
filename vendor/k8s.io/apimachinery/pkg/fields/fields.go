/*
Copyright 2015 The Kubernetes Authors.

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

package ***REMOVED***elds

import (
	"sort"
	"strings"
)

// Fields allows you to present ***REMOVED***elds independently from their storage.
type Fields interface {
	// Has returns whether the provided ***REMOVED***eld exists.
	Has(***REMOVED***eld string) (exists bool)

	// Get returns the value for the provided ***REMOVED***eld.
	Get(***REMOVED***eld string) (value string)
}

// Set is a map of ***REMOVED***eld:value. It implements Fields.
type Set map[string]string

// String returns all ***REMOVED***elds listed as a human readable string.
// Conveniently, exactly the format that ParseSelector takes.
func (ls Set) String() string {
	selector := make([]string, 0, len(ls))
	for key, value := range ls {
		selector = append(selector, key+"="+value)
	}
	// Sort for determinism.
	sort.StringSlice(selector).Sort()
	return strings.Join(selector, ",")
}

// Has returns whether the provided ***REMOVED***eld exists in the map.
func (ls Set) Has(***REMOVED***eld string) bool {
	_, exists := ls[***REMOVED***eld]
	return exists
}

// Get returns the value in the map for the provided ***REMOVED***eld.
func (ls Set) Get(***REMOVED***eld string) string {
	return ls[***REMOVED***eld]
}

// AsSelector converts ***REMOVED***elds into a selectors.
func (ls Set) AsSelector() Selector {
	return SelectorFromSet(ls)
}
