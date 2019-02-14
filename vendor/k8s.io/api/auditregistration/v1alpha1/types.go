/*
Copyright 2018 The Kubernetes Authors.

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

// +k8s:openapi-gen=true

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Level de***REMOVED***nes the amount of information logged during auditing
type Level string

// Valid audit levels
const (
	// LevelNone disables auditing
	LevelNone Level = "None"
	// LevelMetadata provides the basic level of auditing.
	LevelMetadata Level = "Metadata"
	// LevelRequest provides Metadata level of auditing, and additionally
	// logs the request object (does not apply for non-resource requests).
	LevelRequest Level = "Request"
	// LevelRequestResponse provides Request level of auditing, and additionally
	// logs the response object (does not apply for non-resource requests and watches).
	LevelRequestResponse Level = "RequestResponse"
)

// Stage de***REMOVED***nes the stages in request handling during which audit events may be generated.
type Stage string

// Valid audit stages.
const (
	// The stage for events generated after the audit handler receives the request, but before it
	// is delegated down the handler chain.
	StageRequestReceived = "RequestReceived"
	// The stage for events generated after the response headers are sent, but before the response body
	// is sent. This stage is only generated for long-running requests (e.g. watch).
	StageResponseStarted = "ResponseStarted"
	// The stage for events generated after the response body has been completed, and no more bytes
	// will be sent.
	StageResponseComplete = "ResponseComplete"
	// The stage for events generated when a panic occurred.
	StagePanic = "Panic"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AuditSink represents a cluster level audit sink
type AuditSink struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Spec de***REMOVED***nes the audit con***REMOVED***guration spec
	Spec AuditSinkSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
}

// AuditSinkSpec holds the spec for the audit sink
type AuditSinkSpec struct {
	// Policy de***REMOVED***nes the policy for selecting which events should be sent to the webhook
	// required
	Policy Policy `json:"policy" protobuf:"bytes,1,opt,name=policy"`

	// Webhook to send events
	// required
	Webhook Webhook `json:"webhook" protobuf:"bytes,2,opt,name=webhook"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AuditSinkList is a list of AuditSink items.
type AuditSinkList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// List of audit con***REMOVED***gurations.
	Items []AuditSink `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// Policy de***REMOVED***nes the con***REMOVED***guration of how audit events are logged
type Policy struct {
	// The Level that all requests are recorded at.
	// available options: None, Metadata, Request, RequestResponse
	// required
	Level Level `json:"level" protobuf:"bytes,1,opt,name=level"`

	// Stages is a list of stages for which events are created.
	// +optional
	Stages []Stage `json:"stages" protobuf:"bytes,2,opt,name=stages"`
}

// Webhook holds the con***REMOVED***guration of the webhook
type Webhook struct {
	// Throttle holds the options for throttling the webhook
	// +optional
	Throttle *WebhookThrottleCon***REMOVED***g `json:"throttle,omitempty" protobuf:"bytes,1,opt,name=throttle"`

	// ClientCon***REMOVED***g holds the connection parameters for the webhook
	// required
	ClientCon***REMOVED***g WebhookClientCon***REMOVED***g `json:"clientCon***REMOVED***g" protobuf:"bytes,2,opt,name=clientCon***REMOVED***g"`
}

// WebhookThrottleCon***REMOVED***g holds the con***REMOVED***guration for throttling events
type WebhookThrottleCon***REMOVED***g struct {
	// ThrottleQPS maximum number of batches per second
	// default 10 QPS
	// +optional
	QPS *int64 `json:"qps,omitempty" protobuf:"bytes,1,opt,name=qps"`

	// ThrottleBurst is the maximum number of events sent at the same moment
	// default 15 QPS
	// +optional
	Burst *int64 `json:"burst,omitempty" protobuf:"bytes,2,opt,name=burst"`
}

// WebhookClientCon***REMOVED***g contains the information to make a connection with the webhook
type WebhookClientCon***REMOVED***g struct {
	// `url` gives the location of the webhook, in standard URL form
	// (`scheme://host:port/path`). Exactly one of `url` or `service`
	// must be speci***REMOVED***ed.
	//
	// The `host` should not refer to a service running in the cluster; use
	// the `service` ***REMOVED***eld instead. The host might be resolved via external
	// DNS in some apiservers (e.g., `kube-apiserver` cannot resolve
	// in-cluster DNS as that would be a layering violation). `host` may
	// also be an IP address.
	//
	// Please note that using `localhost` or `127.0.0.1` as a `host` is
	// risky unless you take great care to run this webhook on all hosts
	// which run an apiserver which might need to make calls to this
	// webhook. Such installs are likely to be non-portable, i.e., not easy
	// to turn up in a new cluster.
	//
	// The scheme must be "https"; the URL must begin with "https://".
	//
	// A path is optional, and if present may be any string permissible in
	// a URL. You may use the path to pass an arbitrary string to the
	// webhook, for example, a cluster identi***REMOVED***er.
	//
	// Attempting to use a user or basic auth e.g. "user:password@" is not
	// allowed. Fragments ("#...") and query parameters ("?...") are not
	// allowed, either.
	//
	// +optional
	URL *string `json:"url,omitempty" protobuf:"bytes,1,opt,name=url"`

	// `service` is a reference to the service for this webhook. Either
	// `service` or `url` must be speci***REMOVED***ed.
	//
	// If the webhook is running within the cluster, then you should use `service`.
	//
	// Port 443 will be used if it is open, otherwise it is an error.
	//
	// +optional
	Service *ServiceReference `json:"service,omitempty" protobuf:"bytes,2,opt,name=service"`

	// `caBundle` is a PEM encoded CA bundle which will be used to validate the webhook's server certi***REMOVED***cate.
	// If unspeci***REMOVED***ed, system trust roots on the apiserver are used.
	// +optional
	CABundle []byte `json:"caBundle,omitempty" protobuf:"bytes,3,opt,name=caBundle"`
}

// ServiceReference holds a reference to Service.legacy.k8s.io
type ServiceReference struct {
	// `namespace` is the namespace of the service.
	// Required
	Namespace string `json:"namespace" protobuf:"bytes,1,opt,name=namespace"`

	// `name` is the name of the service.
	// Required
	Name string `json:"name" protobuf:"bytes,2,opt,name=name"`

	// `path` is an optional URL path which will be sent in any request to
	// this service.
	// +optional
	Path *string `json:"path,omitempty" protobuf:"bytes,3,opt,name=path"`
}
