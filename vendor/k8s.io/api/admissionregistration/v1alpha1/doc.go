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

// +k8s:deepcopy-gen=package,register
// +k8s:openapi-gen=true

// Package v1alpha1 is the v1alpha1 version of the API.
// AdmissionCon***REMOVED***guration and AdmissionPluginCon***REMOVED***guration are legacy static admission plugin con***REMOVED***guration
// InitializerCon***REMOVED***guration and ExternalAdmissionHookCon***REMOVED***guration is for the
// new dynamic admission controller con***REMOVED***guration.
// +groupName=admissionregistration.k8s.io
package v1alpha1 // import "k8s.io/api/admissionregistration/v1alpha1"
