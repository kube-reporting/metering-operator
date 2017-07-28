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

package apiextensions

// SetCRDCondition sets the status condition.  It either overwrites the existing one or
// creates a new one
func SetCRDCondition(crd *CustomResourceDe***REMOVED***nition, newCondition CustomResourceDe***REMOVED***nitionCondition) {
	existingCondition := FindCRDCondition(crd, newCondition.Type)
	if existingCondition == nil {
		crd.Status.Conditions = append(crd.Status.Conditions, newCondition)
		return
	}

	if existingCondition.Status != newCondition.Status {
		existingCondition.Status = newCondition.Status
		existingCondition.LastTransitionTime = newCondition.LastTransitionTime
	}

	existingCondition.Reason = newCondition.Reason
	existingCondition.Message = newCondition.Message
}

// RemoveCRDCondition removes the status condition.
func RemoveCRDCondition(crd *CustomResourceDe***REMOVED***nition, conditionType CustomResourceDe***REMOVED***nitionConditionType) {
	newConditions := []CustomResourceDe***REMOVED***nitionCondition{}
	for _, condition := range crd.Status.Conditions {
		if condition.Type != conditionType {
			newConditions = append(newConditions, condition)
		}
	}
	crd.Status.Conditions = newConditions
}

// FindCRDCondition returns the condition you're looking for or nil
func FindCRDCondition(crd *CustomResourceDe***REMOVED***nition, conditionType CustomResourceDe***REMOVED***nitionConditionType) *CustomResourceDe***REMOVED***nitionCondition {
	for i := range crd.Status.Conditions {
		if crd.Status.Conditions[i].Type == conditionType {
			return &crd.Status.Conditions[i]
		}
	}

	return nil
}

// IsCRDConditionTrue indicates if the condition is present and strictly true
func IsCRDConditionTrue(crd *CustomResourceDe***REMOVED***nition, conditionType CustomResourceDe***REMOVED***nitionConditionType) bool {
	return IsCRDConditionPresentAndEqual(crd, conditionType, ConditionTrue)
}

// IsCRDConditionFalse indicates if the condition is present and false true
func IsCRDConditionFalse(crd *CustomResourceDe***REMOVED***nition, conditionType CustomResourceDe***REMOVED***nitionConditionType) bool {
	return IsCRDConditionPresentAndEqual(crd, conditionType, ConditionFalse)
}

// IsCRDConditionPresentAndEqual indicates if the condition is present and equal to the arg
func IsCRDConditionPresentAndEqual(crd *CustomResourceDe***REMOVED***nition, conditionType CustomResourceDe***REMOVED***nitionConditionType, status ConditionStatus) bool {
	for _, condition := range crd.Status.Conditions {
		if condition.Type == conditionType {
			return condition.Status == status
		}
	}
	return false
}

// IsCRDConditionEquivalent returns true if the lhs and rhs are equivalent except for times
func IsCRDConditionEquivalent(lhs, rhs *CustomResourceDe***REMOVED***nitionCondition) bool {
	if lhs == nil && rhs == nil {
		return true
	}
	if lhs == nil || rhs == nil {
		return false
	}

	return lhs.Message == rhs.Message && lhs.Reason == rhs.Reason && lhs.Status == rhs.Status && lhs.Type == rhs.Type
}

// CRDHasFinalizer returns true if the ***REMOVED***nalizer is in the list
func CRDHasFinalizer(crd *CustomResourceDe***REMOVED***nition, needle string) bool {
	for _, ***REMOVED***nalizer := range crd.Finalizers {
		if ***REMOVED***nalizer == needle {
			return true
		}
	}

	return false
}

// CRDRemoveFinalizer removes the ***REMOVED***nalizer if present
func CRDRemoveFinalizer(crd *CustomResourceDe***REMOVED***nition, needle string) {
	newFinalizers := []string{}
	for _, ***REMOVED***nalizer := range crd.Finalizers {
		if ***REMOVED***nalizer != needle {
			newFinalizers = append(newFinalizers, ***REMOVED***nalizer)
		}
	}
	crd.Finalizers = newFinalizers
}
