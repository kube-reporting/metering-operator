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

package unstructured

import (
	gojson "encoding/json"
	"fmt"
	"io"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"
)

// NestedFieldCopy returns a deep copy of the value of a nested ***REMOVED***eld.
// Returns false if the value is missing.
// No error is returned for a nil ***REMOVED***eld.
func NestedFieldCopy(obj map[string]interface{}, ***REMOVED***elds ...string) (interface{}, bool, error) {
	val, found, err := NestedFieldNoCopy(obj, ***REMOVED***elds...)
	if !found || err != nil {
		return nil, found, err
	}
	return runtime.DeepCopyJSONValue(val), true, nil
}

// NestedFieldNoCopy returns a reference to a nested ***REMOVED***eld.
// Returns false if value is not found and an error if unable
// to traverse obj.
func NestedFieldNoCopy(obj map[string]interface{}, ***REMOVED***elds ...string) (interface{}, bool, error) {
	var val interface{} = obj

	for i, ***REMOVED***eld := range ***REMOVED***elds {
		if m, ok := val.(map[string]interface{}); ok {
			val, ok = m[***REMOVED***eld]
			if !ok {
				return nil, false, nil
			}
		} ***REMOVED*** {
			return nil, false, fmt.Errorf("%v accessor error: %v is of the type %T, expected map[string]interface{}", jsonPath(***REMOVED***elds[:i+1]), val, val)
		}
	}
	return val, true, nil
}

// NestedString returns the string value of a nested ***REMOVED***eld.
// Returns false if value is not found and an error if not a string.
func NestedString(obj map[string]interface{}, ***REMOVED***elds ...string) (string, bool, error) {
	val, found, err := NestedFieldNoCopy(obj, ***REMOVED***elds...)
	if !found || err != nil {
		return "", found, err
	}
	s, ok := val.(string)
	if !ok {
		return "", false, fmt.Errorf("%v accessor error: %v is of the type %T, expected string", jsonPath(***REMOVED***elds), val, val)
	}
	return s, true, nil
}

// NestedBool returns the bool value of a nested ***REMOVED***eld.
// Returns false if value is not found and an error if not a bool.
func NestedBool(obj map[string]interface{}, ***REMOVED***elds ...string) (bool, bool, error) {
	val, found, err := NestedFieldNoCopy(obj, ***REMOVED***elds...)
	if !found || err != nil {
		return false, found, err
	}
	b, ok := val.(bool)
	if !ok {
		return false, false, fmt.Errorf("%v accessor error: %v is of the type %T, expected bool", jsonPath(***REMOVED***elds), val, val)
	}
	return b, true, nil
}

// NestedFloat64 returns the float64 value of a nested ***REMOVED***eld.
// Returns false if value is not found and an error if not a float64.
func NestedFloat64(obj map[string]interface{}, ***REMOVED***elds ...string) (float64, bool, error) {
	val, found, err := NestedFieldNoCopy(obj, ***REMOVED***elds...)
	if !found || err != nil {
		return 0, found, err
	}
	f, ok := val.(float64)
	if !ok {
		return 0, false, fmt.Errorf("%v accessor error: %v is of the type %T, expected float64", jsonPath(***REMOVED***elds), val, val)
	}
	return f, true, nil
}

// NestedInt64 returns the int64 value of a nested ***REMOVED***eld.
// Returns false if value is not found and an error if not an int64.
func NestedInt64(obj map[string]interface{}, ***REMOVED***elds ...string) (int64, bool, error) {
	val, found, err := NestedFieldNoCopy(obj, ***REMOVED***elds...)
	if !found || err != nil {
		return 0, found, err
	}
	i, ok := val.(int64)
	if !ok {
		return 0, false, fmt.Errorf("%v accessor error: %v is of the type %T, expected int64", jsonPath(***REMOVED***elds), val, val)
	}
	return i, true, nil
}

// NestedStringSlice returns a copy of []string value of a nested ***REMOVED***eld.
// Returns false if value is not found and an error if not a []interface{} or contains non-string items in the slice.
func NestedStringSlice(obj map[string]interface{}, ***REMOVED***elds ...string) ([]string, bool, error) {
	val, found, err := NestedFieldNoCopy(obj, ***REMOVED***elds...)
	if !found || err != nil {
		return nil, found, err
	}
	m, ok := val.([]interface{})
	if !ok {
		return nil, false, fmt.Errorf("%v accessor error: %v is of the type %T, expected []interface{}", jsonPath(***REMOVED***elds), val, val)
	}
	strSlice := make([]string, 0, len(m))
	for _, v := range m {
		if str, ok := v.(string); ok {
			strSlice = append(strSlice, str)
		} ***REMOVED*** {
			return nil, false, fmt.Errorf("%v accessor error: contains non-string key in the slice: %v is of the type %T, expected string", jsonPath(***REMOVED***elds), v, v)
		}
	}
	return strSlice, true, nil
}

// NestedSlice returns a deep copy of []interface{} value of a nested ***REMOVED***eld.
// Returns false if value is not found and an error if not a []interface{}.
func NestedSlice(obj map[string]interface{}, ***REMOVED***elds ...string) ([]interface{}, bool, error) {
	val, found, err := NestedFieldNoCopy(obj, ***REMOVED***elds...)
	if !found || err != nil {
		return nil, found, err
	}
	_, ok := val.([]interface{})
	if !ok {
		return nil, false, fmt.Errorf("%v accessor error: %v is of the type %T, expected []interface{}", jsonPath(***REMOVED***elds), val, val)
	}
	return runtime.DeepCopyJSONValue(val).([]interface{}), true, nil
}

// NestedStringMap returns a copy of map[string]string value of a nested ***REMOVED***eld.
// Returns false if value is not found and an error if not a map[string]interface{} or contains non-string values in the map.
func NestedStringMap(obj map[string]interface{}, ***REMOVED***elds ...string) (map[string]string, bool, error) {
	m, found, err := nestedMapNoCopy(obj, ***REMOVED***elds...)
	if !found || err != nil {
		return nil, found, err
	}
	strMap := make(map[string]string, len(m))
	for k, v := range m {
		if str, ok := v.(string); ok {
			strMap[k] = str
		} ***REMOVED*** {
			return nil, false, fmt.Errorf("%v accessor error: contains non-string key in the map: %v is of the type %T, expected string", jsonPath(***REMOVED***elds), v, v)
		}
	}
	return strMap, true, nil
}

// NestedMap returns a deep copy of map[string]interface{} value of a nested ***REMOVED***eld.
// Returns false if value is not found and an error if not a map[string]interface{}.
func NestedMap(obj map[string]interface{}, ***REMOVED***elds ...string) (map[string]interface{}, bool, error) {
	m, found, err := nestedMapNoCopy(obj, ***REMOVED***elds...)
	if !found || err != nil {
		return nil, found, err
	}
	return runtime.DeepCopyJSON(m), true, nil
}

// nestedMapNoCopy returns a map[string]interface{} value of a nested ***REMOVED***eld.
// Returns false if value is not found and an error if not a map[string]interface{}.
func nestedMapNoCopy(obj map[string]interface{}, ***REMOVED***elds ...string) (map[string]interface{}, bool, error) {
	val, found, err := NestedFieldNoCopy(obj, ***REMOVED***elds...)
	if !found || err != nil {
		return nil, found, err
	}
	m, ok := val.(map[string]interface{})
	if !ok {
		return nil, false, fmt.Errorf("%v accessor error: %v is of the type %T, expected map[string]interface{}", jsonPath(***REMOVED***elds), val, val)
	}
	return m, true, nil
}

// SetNestedField sets the value of a nested ***REMOVED***eld to a deep copy of the value provided.
// Returns an error if value cannot be set because one of the nesting levels is not a map[string]interface{}.
func SetNestedField(obj map[string]interface{}, value interface{}, ***REMOVED***elds ...string) error {
	return setNestedFieldNoCopy(obj, runtime.DeepCopyJSONValue(value), ***REMOVED***elds...)
}

func setNestedFieldNoCopy(obj map[string]interface{}, value interface{}, ***REMOVED***elds ...string) error {
	m := obj

	for i, ***REMOVED***eld := range ***REMOVED***elds[:len(***REMOVED***elds)-1] {
		if val, ok := m[***REMOVED***eld]; ok {
			if valMap, ok := val.(map[string]interface{}); ok {
				m = valMap
			} ***REMOVED*** {
				return fmt.Errorf("value cannot be set because %v is not a map[string]interface{}", jsonPath(***REMOVED***elds[:i+1]))
			}
		} ***REMOVED*** {
			newVal := make(map[string]interface{})
			m[***REMOVED***eld] = newVal
			m = newVal
		}
	}
	m[***REMOVED***elds[len(***REMOVED***elds)-1]] = value
	return nil
}

// SetNestedStringSlice sets the string slice value of a nested ***REMOVED***eld.
// Returns an error if value cannot be set because one of the nesting levels is not a map[string]interface{}.
func SetNestedStringSlice(obj map[string]interface{}, value []string, ***REMOVED***elds ...string) error {
	m := make([]interface{}, 0, len(value)) // convert []string into []interface{}
	for _, v := range value {
		m = append(m, v)
	}
	return setNestedFieldNoCopy(obj, m, ***REMOVED***elds...)
}

// SetNestedSlice sets the slice value of a nested ***REMOVED***eld.
// Returns an error if value cannot be set because one of the nesting levels is not a map[string]interface{}.
func SetNestedSlice(obj map[string]interface{}, value []interface{}, ***REMOVED***elds ...string) error {
	return SetNestedField(obj, value, ***REMOVED***elds...)
}

// SetNestedStringMap sets the map[string]string value of a nested ***REMOVED***eld.
// Returns an error if value cannot be set because one of the nesting levels is not a map[string]interface{}.
func SetNestedStringMap(obj map[string]interface{}, value map[string]string, ***REMOVED***elds ...string) error {
	m := make(map[string]interface{}, len(value)) // convert map[string]string into map[string]interface{}
	for k, v := range value {
		m[k] = v
	}
	return setNestedFieldNoCopy(obj, m, ***REMOVED***elds...)
}

// SetNestedMap sets the map[string]interface{} value of a nested ***REMOVED***eld.
// Returns an error if value cannot be set because one of the nesting levels is not a map[string]interface{}.
func SetNestedMap(obj map[string]interface{}, value map[string]interface{}, ***REMOVED***elds ...string) error {
	return SetNestedField(obj, value, ***REMOVED***elds...)
}

// RemoveNestedField removes the nested ***REMOVED***eld from the obj.
func RemoveNestedField(obj map[string]interface{}, ***REMOVED***elds ...string) {
	m := obj
	for _, ***REMOVED***eld := range ***REMOVED***elds[:len(***REMOVED***elds)-1] {
		if x, ok := m[***REMOVED***eld].(map[string]interface{}); ok {
			m = x
		} ***REMOVED*** {
			return
		}
	}
	delete(m, ***REMOVED***elds[len(***REMOVED***elds)-1])
}

func getNestedString(obj map[string]interface{}, ***REMOVED***elds ...string) string {
	val, found, err := NestedString(obj, ***REMOVED***elds...)
	if !found || err != nil {
		return ""
	}
	return val
}

func jsonPath(***REMOVED***elds []string) string {
	return "." + strings.Join(***REMOVED***elds, ".")
}

func extractOwnerReference(v map[string]interface{}) metav1.OwnerReference {
	// though this ***REMOVED***eld is a *bool, but when decoded from JSON, it's
	// unmarshalled as bool.
	var controllerPtr *bool
	if controller, found, err := NestedBool(v, "controller"); err == nil && found {
		controllerPtr = &controller
	}
	var blockOwnerDeletionPtr *bool
	if blockOwnerDeletion, found, err := NestedBool(v, "blockOwnerDeletion"); err == nil && found {
		blockOwnerDeletionPtr = &blockOwnerDeletion
	}
	return metav1.OwnerReference{
		Kind:               getNestedString(v, "kind"),
		Name:               getNestedString(v, "name"),
		APIVersion:         getNestedString(v, "apiVersion"),
		UID:                types.UID(getNestedString(v, "uid")),
		Controller:         controllerPtr,
		BlockOwnerDeletion: blockOwnerDeletionPtr,
	}
}

// UnstructuredJSONScheme is capable of converting JSON data into the Unstructured
// type, which can be used for generic access to objects without a prede***REMOVED***ned scheme.
// TODO: move into serializer/json.
var UnstructuredJSONScheme runtime.Codec = unstructuredJSONScheme{}

type unstructuredJSONScheme struct{}

func (s unstructuredJSONScheme) Decode(data []byte, _ *schema.GroupVersionKind, obj runtime.Object) (runtime.Object, *schema.GroupVersionKind, error) {
	var err error
	if obj != nil {
		err = s.decodeInto(data, obj)
	} ***REMOVED*** {
		obj, err = s.decode(data)
	}

	if err != nil {
		return nil, nil, err
	}

	gvk := obj.GetObjectKind().GroupVersionKind()
	if len(gvk.Kind) == 0 {
		return nil, &gvk, runtime.NewMissingKindErr(string(data))
	}

	return obj, &gvk, nil
}

func (unstructuredJSONScheme) Encode(obj runtime.Object, w io.Writer) error {
	switch t := obj.(type) {
	case *Unstructured:
		return json.NewEncoder(w).Encode(t.Object)
	case *UnstructuredList:
		items := make([]interface{}, 0, len(t.Items))
		for _, i := range t.Items {
			items = append(items, i.Object)
		}
		listObj := make(map[string]interface{}, len(t.Object)+1)
		for k, v := range t.Object { // Make a shallow copy
			listObj[k] = v
		}
		listObj["items"] = items
		return json.NewEncoder(w).Encode(listObj)
	case *runtime.Unknown:
		// TODO: Unstructured needs to deal with ContentType.
		_, err := w.Write(t.Raw)
		return err
	default:
		return json.NewEncoder(w).Encode(t)
	}
}

func (s unstructuredJSONScheme) decode(data []byte) (runtime.Object, error) {
	type detector struct {
		Items gojson.RawMessage
	}
	var det detector
	if err := json.Unmarshal(data, &det); err != nil {
		return nil, err
	}

	if det.Items != nil {
		list := &UnstructuredList{}
		err := s.decodeToList(data, list)
		return list, err
	}

	// No Items ***REMOVED***eld, so it wasn't a list.
	unstruct := &Unstructured{}
	err := s.decodeToUnstructured(data, unstruct)
	return unstruct, err
}

func (s unstructuredJSONScheme) decodeInto(data []byte, obj runtime.Object) error {
	switch x := obj.(type) {
	case *Unstructured:
		return s.decodeToUnstructured(data, x)
	case *UnstructuredList:
		return s.decodeToList(data, x)
	case *runtime.VersionedObjects:
		o, err := s.decode(data)
		if err == nil {
			x.Objects = []runtime.Object{o}
		}
		return err
	default:
		return json.Unmarshal(data, x)
	}
}

func (unstructuredJSONScheme) decodeToUnstructured(data []byte, unstruct *Unstructured) error {
	m := make(map[string]interface{})
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	unstruct.Object = m

	return nil
}

func (s unstructuredJSONScheme) decodeToList(data []byte, list *UnstructuredList) error {
	type decodeList struct {
		Items []gojson.RawMessage
	}

	var dList decodeList
	if err := json.Unmarshal(data, &dList); err != nil {
		return err
	}

	if err := json.Unmarshal(data, &list.Object); err != nil {
		return err
	}

	// For typed lists, e.g., a PodList, API server doesn't set each item's
	// APIVersion and Kind. We need to set it.
	listAPIVersion := list.GetAPIVersion()
	listKind := list.GetKind()
	itemKind := strings.TrimSuf***REMOVED***x(listKind, "List")

	delete(list.Object, "items")
	list.Items = make([]Unstructured, 0, len(dList.Items))
	for _, i := range dList.Items {
		unstruct := &Unstructured{}
		if err := s.decodeToUnstructured([]byte(i), unstruct); err != nil {
			return err
		}
		// This is hacky. Set the item's Kind and APIVersion to those inferred
		// from the List.
		if len(unstruct.GetKind()) == 0 && len(unstruct.GetAPIVersion()) == 0 {
			unstruct.SetKind(itemKind)
			unstruct.SetAPIVersion(listAPIVersion)
		}
		list.Items = append(list.Items, *unstruct)
	}
	return nil
}

type JSONFallbackEncoder struct {
	runtime.Encoder
}

func (c JSONFallbackEncoder) Encode(obj runtime.Object, w io.Writer) error {
	err := c.Encoder.Encode(obj, w)
	if runtime.IsNotRegisteredError(err) {
		switch obj.(type) {
		case *Unstructured, *UnstructuredList:
			return UnstructuredJSONScheme.Encode(obj, w)
		}
	}
	return err
}
