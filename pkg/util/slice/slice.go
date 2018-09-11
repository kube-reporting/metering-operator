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

// Package slice provides utility methods for common operations on slices.
package slice

import (
	"sort"
)

// CopyStrings copies the contents of the speci***REMOVED***ed string slice
// into a new slice.
func CopyStrings(s []string) []string {
	if s == nil {
		return nil
	}
	c := make([]string, len(s))
	copy(c, s)
	return c
}

// SortStrings sorts the speci***REMOVED***ed string slice in place. It returns the same
// slice that was provided in order to facilitate method chaining.
func SortStrings(s []string) []string {
	sort.Strings(s)
	return s
}

// ContainsString checks if a given slice of strings contains the provided string.
// If a modi***REMOVED***er func is provided, it is called with the slice item before the comparation.
func ContainsString(slice []string, s string, modi***REMOVED***er func(s string) string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
		if modi***REMOVED***er != nil && modi***REMOVED***er(item) == s {
			return true
		}
	}
	return false
}

// RemoveString returns a newly created []string that contains all items from slice that
// are not equal to s and modi***REMOVED***er(s) in case modi***REMOVED***er func is provided.
func RemoveString(slice []string, s string, modi***REMOVED***er func(s string) string) []string {
	newSlice := make([]string, 0)
	for _, item := range slice {
		if item == s {
			continue
		}
		if modi***REMOVED***er != nil && modi***REMOVED***er(item) == s {
			continue
		}
		newSlice = append(newSlice, item)
	}
	if len(newSlice) == 0 {
		// Sanitize for unit tests so we don't need to distinguish empty array
		// and nil.
		newSlice = nil
	}
	return newSlice
}
