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

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// CustomResourceDe***REMOVED***nitionSpec describes how a user wants their resource to appear
type CustomResourceDe***REMOVED***nitionSpec struct {
	// Group is the group this resource belongs in
	Group string `json:"group" protobuf:"bytes,1,opt,name=group"`
	// Version is the version this resource belongs in
	Version string `json:"version" protobuf:"bytes,2,opt,name=version"`
	// Names are the names used to describe this custom resource
	Names CustomResourceDe***REMOVED***nitionNames `json:"names" protobuf:"bytes,3,opt,name=names"`

	// Scope indicates whether this resource is cluster or namespace scoped.  Default is namespaced
	Scope ResourceScope `json:"scope" protobuf:"bytes,4,opt,name=scope,casttype=ResourceScope"`
}

// CustomResourceDe***REMOVED***nitionNames indicates the names to serve this CustomResourceDe***REMOVED***nition
type CustomResourceDe***REMOVED***nitionNames struct {
	// Plural is the plural name of the resource to serve.  It must match the name of the CustomResourceDe***REMOVED***nition-registration
	// too: plural.group and it must be all lowercase.
	Plural string `json:"plural" protobuf:"bytes,1,opt,name=plural"`
	// Singular is the singular name of the resource.  It must be all lowercase  Defaults to lowercased <kind>
	Singular string `json:"singular,omitempty" protobuf:"bytes,2,opt,name=singular"`
	// ShortNames are short names for the resource.  It must be all lowercase.
	ShortNames []string `json:"shortNames,omitempty" protobuf:"bytes,3,opt,name=shortNames"`
	// Kind is the serialized kind of the resource.  It is normally CamelCase and singular.
	Kind string `json:"kind" protobuf:"bytes,4,opt,name=kind"`
	// ListKind is the serialized kind of the list for this resource.  Defaults to <kind>List.
	ListKind string `json:"listKind,omitempty" protobuf:"bytes,5,opt,name=listKind"`
}

// ResourceScope is an enum de***REMOVED***ning the different scopes availabe to a custom resource
type ResourceScope string

const (
	ClusterScoped   ResourceScope = "Cluster"
	NamespaceScoped ResourceScope = "Namespaced"
)

type ConditionStatus string

// These are valid condition statuses. "ConditionTrue" means a resource is in the condition.
// "ConditionFalse" means a resource is not in the condition. "ConditionUnknown" means kubernetes
// can't decide if a resource is in the condition or not. In the future, we could add other
// intermediate conditions, e.g. ConditionDegraded.
const (
	ConditionTrue    ConditionStatus = "True"
	ConditionFalse   ConditionStatus = "False"
	ConditionUnknown ConditionStatus = "Unknown"
)

// CustomResourceDe***REMOVED***nitionConditionType is a valid value for CustomResourceDe***REMOVED***nitionCondition.Type
type CustomResourceDe***REMOVED***nitionConditionType string

const (
	// Established means that the resource has become active. A resource is established when all names are
	// accepted without a conflict for the ***REMOVED***rst time. A resource stays established until deleted, even during
	// a later NamesAccepted due to changed names. Note that not all names can be changed.
	Established CustomResourceDe***REMOVED***nitionConditionType = "Established"
	// NamesAccepted means the names chosen for this CustomResourceDe***REMOVED***nition do not conflict with others in
	// the group and are therefore accepted.
	NamesAccepted CustomResourceDe***REMOVED***nitionConditionType = "NamesAccepted"
	// Terminating means that the CustomResourceDe***REMOVED***nition has been deleted and is cleaning up.
	Terminating CustomResourceDe***REMOVED***nitionConditionType = "Terminating"
)

// CustomResourceDe***REMOVED***nitionCondition contains details for the current condition of this pod.
type CustomResourceDe***REMOVED***nitionCondition struct {
	// Type is the type of the condition.
	Type CustomResourceDe***REMOVED***nitionConditionType `json:"type" protobuf:"bytes,1,opt,name=type,casttype=CustomResourceDe***REMOVED***nitionConditionType"`
	// Status is the status of the condition.
	// Can be True, False, Unknown.
	Status ConditionStatus `json:"status" protobuf:"bytes,2,opt,name=status,casttype=ConditionStatus"`
	// Last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty" protobuf:"bytes,3,opt,name=lastTransitionTime"`
	// Unique, one-word, CamelCase reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty" protobuf:"bytes,4,opt,name=reason"`
	// Human-readable message indicating details about last transition.
	// +optional
	Message string `json:"message,omitempty" protobuf:"bytes,5,opt,name=message"`
}

// CustomResourceDe***REMOVED***nitionStatus indicates the state of the CustomResourceDe***REMOVED***nition
type CustomResourceDe***REMOVED***nitionStatus struct {
	// Conditions indicate state for particular aspects of a CustomResourceDe***REMOVED***nition
	Conditions []CustomResourceDe***REMOVED***nitionCondition `json:"conditions" protobuf:"bytes,1,opt,name=conditions"`

	// AcceptedNames are the names that are actually being used to serve discovery
	// They may be different than the names in spec.
	AcceptedNames CustomResourceDe***REMOVED***nitionNames `json:"acceptedNames" protobuf:"bytes,2,opt,name=acceptedNames"`
}

// CustomResourceCleanupFinalizer is the name of the ***REMOVED***nalizer which will delete instances of
// a CustomResourceDe***REMOVED***nition
const CustomResourceCleanupFinalizer = "customresourcecleanup.apiextensions.k8s.io"

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CustomResourceDe***REMOVED***nition represents a resource that should be exposed on the API server.  Its name MUST be in the format
// <.spec.name>.<.spec.group>.
type CustomResourceDe***REMOVED***nition struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Spec describes how the user wants the resources to appear
	Spec CustomResourceDe***REMOVED***nitionSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	// Status indicates the actual state of the CustomResourceDe***REMOVED***nition
	Status CustomResourceDe***REMOVED***nitionStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CustomResourceDe***REMOVED***nitionList is a list of CustomResourceDe***REMOVED***nition objects.
type CustomResourceDe***REMOVED***nitionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Items individual CustomResourceDe***REMOVED***nitions
	Items []CustomResourceDe***REMOVED***nition `json:"items" protobuf:"bytes,2,rep,name=items"`
}
