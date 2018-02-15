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

package runtime

import (
	"fmt"
	"reflect"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

type notRegisteredErr struct {
	gvk    schema.GroupVersionKind
	target GroupVersioner
	t      reflect.Type
}

func NewNotRegisteredErrForKind(gvk schema.GroupVersionKind) error {
	return &notRegisteredErr{gvk: gvk}
}

func NewNotRegisteredErrForType(t reflect.Type) error {
	return &notRegisteredErr{t: t}
}

func NewNotRegisteredErrForTarget(t reflect.Type, target GroupVersioner) error {
	return &notRegisteredErr{t: t, target: target}
}

func (k *notRegisteredErr) Error() string {
	if k.t != nil && k.target != nil {
		return fmt.Sprintf("%v is not suitable for converting to %q", k.t, k.target)
	}
	if k.t != nil {
		return fmt.Sprintf("no kind is registered for the type %v", k.t)
	}
	if len(k.gvk.Kind) == 0 {
		return fmt.Sprintf("no version %q has been registered", k.gvk.GroupVersion())
	}
	if k.gvk.Version == APIVersionInternal {
		return fmt.Sprintf("no kind %q is registered for the internal version of group %q", k.gvk.Kind, k.gvk.Group)
	}

	return fmt.Sprintf("no kind %q is registered for version %q", k.gvk.Kind, k.gvk.GroupVersion())
}

// IsNotRegisteredError returns true if the error indicates the provided
// object or input data is not registered.
func IsNotRegisteredError(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(*notRegisteredErr)
	return ok
}

type missingKindErr struct {
	data string
}

func NewMissingKindErr(data string) error {
	return &missingKindErr{data}
}

func (k *missingKindErr) Error() string {
	return fmt.Sprintf("Object 'Kind' is missing in '%s'", k.data)
}

// IsMissingKind returns true if the error indicates that the provided object
// is missing a 'Kind' ***REMOVED***eld.
func IsMissingKind(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(*missingKindErr)
	return ok
}

type missingVersionErr struct {
	data string
}

func NewMissingVersionErr(data string) error {
	return &missingVersionErr{data}
}

func (k *missingVersionErr) Error() string {
	return fmt.Sprintf("Object 'apiVersion' is missing in '%s'", k.data)
}

// IsMissingVersion returns true if the error indicates that the provided object
// is missing a 'Version' ***REMOVED***eld.
func IsMissingVersion(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(*missingVersionErr)
	return ok
}
