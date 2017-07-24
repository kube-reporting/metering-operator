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

// Package conversion provides go object versioning.
//
// Speci***REMOVED***cally, conversion provides a way for you to de***REMOVED***ne multiple versions
// of the same object. You may write functions which implement conversion logic,
// but for the ***REMOVED***elds which did not change, copying is automated. This makes it
// easy to modify the structures you use in memory without affecting the format
// you store on disk or respond to in your external API calls.
package conversion // import "k8s.io/apimachinery/pkg/conversion"
