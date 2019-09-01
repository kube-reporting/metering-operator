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

package v1beta1

import (
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
)

func addDefaultingFuncs(scheme *runtime.Scheme) error {
	scheme.AddTypeDefaultingFunc(&CustomResourceDe***REMOVED***nition{}, func(obj interface{}) { SetDefaults_CustomResourceDe***REMOVED***nition(obj.(*CustomResourceDe***REMOVED***nition)) })
	// TODO ***REMOVED***gure out why I can't seem to get my defaulter generated
	// return RegisterDefaults(scheme)
	return nil
}

func SetDefaults_CustomResourceDe***REMOVED***nition(obj *CustomResourceDe***REMOVED***nition) {
	SetDefaults_CustomResourceDe***REMOVED***nitionSpec(&obj.Spec)
	if len(obj.Status.StoredVersions) == 0 {
		for _, v := range obj.Spec.Versions {
			if v.Storage {
				obj.Status.StoredVersions = append(obj.Status.StoredVersions, v.Name)
				break
			}
		}
	}
}

func SetDefaults_CustomResourceDe***REMOVED***nitionSpec(obj *CustomResourceDe***REMOVED***nitionSpec) {
	if len(obj.Scope) == 0 {
		obj.Scope = NamespaceScoped
	}
	if len(obj.Names.Singular) == 0 {
		obj.Names.Singular = strings.ToLower(obj.Names.Kind)
	}
	if len(obj.Names.ListKind) == 0 && len(obj.Names.Kind) > 0 {
		obj.Names.ListKind = obj.Names.Kind + "List"
	}
	// If there is no list of versions, create on using deprecated Version ***REMOVED***eld.
	if len(obj.Versions) == 0 && len(obj.Version) != 0 {
		obj.Versions = []CustomResourceDe***REMOVED***nitionVersion{{
			Name:    obj.Version,
			Storage: true,
			Served:  true,
		}}
	}
	// For backward compatibility set the version ***REMOVED***eld to the ***REMOVED***rst item in versions list.
	if len(obj.Version) == 0 && len(obj.Versions) != 0 {
		obj.Version = obj.Versions[0].Name
	}
	if obj.Conversion == nil {
		obj.Conversion = &CustomResourceConversion{
			Strategy: NoneConverter,
		}
	}
}

// hasPerVersionColumns returns true if a CRD uses per-version columns.
func hasPerVersionColumns(versions []CustomResourceDe***REMOVED***nitionVersion) bool {
	for _, v := range versions {
		if len(v.AdditionalPrinterColumns) > 0 {
			return true
		}
	}
	return false
}
