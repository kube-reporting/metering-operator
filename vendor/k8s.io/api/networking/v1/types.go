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

package v1

import (
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NetworkPolicy describes what network traf***REMOVED***c is allowed for a set of Pods
type NetworkPolicy struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Speci***REMOVED***cation of the desired behavior for this NetworkPolicy.
	// +optional
	Spec NetworkPolicySpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
}

// Policy Type string describes the NetworkPolicy type
// This type is beta-level in 1.8
type PolicyType string

const (
	// PolicyTypeIngress is a NetworkPolicy that affects ingress traf***REMOVED***c on selected pods
	PolicyTypeIngress PolicyType = "Ingress"
	// PolicyTypeEgress is a NetworkPolicy that affects egress traf***REMOVED***c on selected pods
	PolicyTypeEgress PolicyType = "Egress"
)

// NetworkPolicySpec provides the speci***REMOVED***cation of a NetworkPolicy
type NetworkPolicySpec struct {
	// Selects the pods to which this NetworkPolicy object applies. The array of
	// ingress rules is applied to any pods selected by this ***REMOVED***eld. Multiple network
	// policies can select the same set of pods. In this case, the ingress rules for
	// each are combined additively. This ***REMOVED***eld is NOT optional and follows standard
	// label selector semantics. An empty podSelector matches all pods in this
	// namespace.
	PodSelector metav1.LabelSelector `json:"podSelector" protobuf:"bytes,1,opt,name=podSelector"`

	// List of ingress rules to be applied to the selected pods. Traf***REMOVED***c is allowed to
	// a pod if there are no NetworkPolicies selecting the pod
	// (and cluster policy otherwise allows the traf***REMOVED***c), OR if the traf***REMOVED***c source is
	// the pod's local node, OR if the traf***REMOVED***c matches at least one ingress rule
	// across all of the NetworkPolicy objects whose podSelector matches the pod. If
	// this ***REMOVED***eld is empty then this NetworkPolicy does not allow any traf***REMOVED***c (and serves
	// solely to ensure that the pods it selects are isolated by default)
	// +optional
	Ingress []NetworkPolicyIngressRule `json:"ingress,omitempty" protobuf:"bytes,2,rep,name=ingress"`

	// List of egress rules to be applied to the selected pods. Outgoing traf***REMOVED***c is
	// allowed if there are no NetworkPolicies selecting the pod (and cluster policy
	// otherwise allows the traf***REMOVED***c), OR if the traf***REMOVED***c matches at least one egress rule
	// across all of the NetworkPolicy objects whose podSelector matches the pod. If
	// this ***REMOVED***eld is empty then this NetworkPolicy limits all outgoing traf***REMOVED***c (and serves
	// solely to ensure that the pods it selects are isolated by default).
	// This ***REMOVED***eld is beta-level in 1.8
	// +optional
	Egress []NetworkPolicyEgressRule `json:"egress,omitempty" protobuf:"bytes,3,rep,name=egress"`

	// List of rule types that the NetworkPolicy relates to.
	// Valid options are Ingress, Egress, or Ingress,Egress.
	// If this ***REMOVED***eld is not speci***REMOVED***ed, it will default based on the existence of Ingress or Egress rules;
	// policies that contain an Egress section are assumed to affect Egress, and all policies
	// (whether or not they contain an Ingress section) are assumed to affect Ingress.
	// If you want to write an egress-only policy, you must explicitly specify policyTypes [ "Egress" ].
	// Likewise, if you want to write a policy that speci***REMOVED***es that no egress is allowed,
	// you must specify a policyTypes value that include "Egress" (since such a policy would not include
	// an Egress section and would otherwise default to just [ "Ingress" ]).
	// This ***REMOVED***eld is beta-level in 1.8
	// +optional
	PolicyTypes []PolicyType `json:"policyTypes,omitempty" protobuf:"bytes,4,rep,name=policyTypes,casttype=PolicyType"`
}

// NetworkPolicyIngressRule describes a particular set of traf***REMOVED***c that is allowed to the pods
// matched by a NetworkPolicySpec's podSelector. The traf***REMOVED***c must match both ports and from.
type NetworkPolicyIngressRule struct {
	// List of ports which should be made accessible on the pods selected for this
	// rule. Each item in this list is combined using a logical OR. If this ***REMOVED***eld is
	// empty or missing, this rule matches all ports (traf***REMOVED***c not restricted by port).
	// If this ***REMOVED***eld is present and contains at least one item, then this rule allows
	// traf***REMOVED***c only if the traf***REMOVED***c matches at least one port in the list.
	// +optional
	Ports []NetworkPolicyPort `json:"ports,omitempty" protobuf:"bytes,1,rep,name=ports"`

	// List of sources which should be able to access the pods selected for this rule.
	// Items in this list are combined using a logical OR operation. If this ***REMOVED***eld is
	// empty or missing, this rule matches all sources (traf***REMOVED***c not restricted by
	// source). If this ***REMOVED***eld is present and contains at least on item, this rule
	// allows traf***REMOVED***c only if the traf***REMOVED***c matches at least one item in the from list.
	// +optional
	From []NetworkPolicyPeer `json:"from,omitempty" protobuf:"bytes,2,rep,name=from"`
}

// NetworkPolicyEgressRule describes a particular set of traf***REMOVED***c that is allowed out of pods
// matched by a NetworkPolicySpec's podSelector. The traf***REMOVED***c must match both ports and to.
// This type is beta-level in 1.8
type NetworkPolicyEgressRule struct {
	// List of destination ports for outgoing traf***REMOVED***c.
	// Each item in this list is combined using a logical OR. If this ***REMOVED***eld is
	// empty or missing, this rule matches all ports (traf***REMOVED***c not restricted by port).
	// If this ***REMOVED***eld is present and contains at least one item, then this rule allows
	// traf***REMOVED***c only if the traf***REMOVED***c matches at least one port in the list.
	// +optional
	Ports []NetworkPolicyPort `json:"ports,omitempty" protobuf:"bytes,1,rep,name=ports"`

	// List of destinations for outgoing traf***REMOVED***c of pods selected for this rule.
	// Items in this list are combined using a logical OR operation. If this ***REMOVED***eld is
	// empty or missing, this rule matches all destinations (traf***REMOVED***c not restricted by
	// destination). If this ***REMOVED***eld is present and contains at least one item, this rule
	// allows traf***REMOVED***c only if the traf***REMOVED***c matches at least one item in the to list.
	// +optional
	To []NetworkPolicyPeer `json:"to,omitempty" protobuf:"bytes,2,rep,name=to"`
}

// NetworkPolicyPort describes a port to allow traf***REMOVED***c on
type NetworkPolicyPort struct {
	// The protocol (TCP or UDP) which traf***REMOVED***c must match. If not speci***REMOVED***ed, this
	// ***REMOVED***eld defaults to TCP.
	// +optional
	Protocol *v1.Protocol `json:"protocol,omitempty" protobuf:"bytes,1,opt,name=protocol,casttype=k8s.io/api/core/v1.Protocol"`

	// The port on the given protocol. This can either be a numerical or named port on
	// a pod. If this ***REMOVED***eld is not provided, this matches all port names and numbers.
	// +optional
	Port *intstr.IntOrString `json:"port,omitempty" protobuf:"bytes,2,opt,name=port"`
}

// IPBlock describes a particular CIDR (Ex. "192.168.1.1/24") that is allowed to the pods
// matched by a NetworkPolicySpec's podSelector. The except entry describes CIDRs that should
// not be included within this rule.
type IPBlock struct {
	// CIDR is a string representing the IP Block
	// Valid examples are "192.168.1.1/24"
	CIDR string `json:"cidr" protobuf:"bytes,1,name=cidr"`
	// Except is a slice of CIDRs that should not be included within an IP Block
	// Valid examples are "192.168.1.1/24"
	// Except values will be rejected if they are outside the CIDR range
	// +optional
	Except []string `json:"except,omitempty" protobuf:"bytes,2,rep,name=except"`
}

// NetworkPolicyPeer describes a peer to allow traf***REMOVED***c from. Exactly one of its ***REMOVED***elds
// must be speci***REMOVED***ed.
type NetworkPolicyPeer struct {
	// This is a label selector which selects Pods in this namespace. This ***REMOVED***eld
	// follows standard label selector semantics. If present but empty, this selector
	// selects all pods in this namespace.
	// +optional
	PodSelector *metav1.LabelSelector `json:"podSelector,omitempty" protobuf:"bytes,1,opt,name=podSelector"`

	// Selects Namespaces using cluster scoped-labels. This matches all pods in all
	// namespaces selected by this label selector. This ***REMOVED***eld follows standard label
	// selector semantics. If present but empty, this selector selects all namespaces.
	// +optional
	NamespaceSelector *metav1.LabelSelector `json:"namespaceSelector,omitempty" protobuf:"bytes,2,opt,name=namespaceSelector"`

	// IPBlock de***REMOVED***nes policy on a particular IPBlock
	// +optional
	IPBlock *IPBlock `json:"ipBlock,omitempty" protobuf:"bytes,3,rep,name=ipBlock"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NetworkPolicyList is a list of NetworkPolicy objects.
type NetworkPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Items is a list of schema objects.
	Items []NetworkPolicy `json:"items" protobuf:"bytes,2,rep,name=items"`
}
