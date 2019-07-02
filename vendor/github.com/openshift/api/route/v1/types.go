package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// A route allows developers to expose services through an HTTP(S) aware load balancing and proxy
// layer via a public DNS entry. The route may further specify TLS options and a certi***REMOVED***cate, or
// specify a public CNAME that the router should also accept for HTTP and HTTPS traf***REMOVED***c. An
// administrator typically con***REMOVED***gures their router to be visible outside the cluster ***REMOVED***rewall, and
// may also add additional security, caching, or traf***REMOVED***c controls on the service content. Routers
// usually talk directly to the service endpoints.
//
// Once a route is created, the `host` ***REMOVED***eld may not be changed. Generally, routers use the oldest
// route with a given host when resolving conflicts.
//
// Routers are subject to additional customization and may support additional controls via the
// annotations ***REMOVED***eld.
//
// Because administrators may con***REMOVED***gure multiple routers, the route status ***REMOVED***eld is used to
// return information to clients about the names and states of the route under each router.
// If a client chooses a duplicate name, for instance, the route status conditions are used
// to indicate the route cannot be chosen.
type Route struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object metadata.
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// spec is the desired state of the route
	Spec RouteSpec `json:"spec" protobuf:"bytes,2,opt,name=spec"`
	// status is the current state of the route
	Status RouteStatus `json:"status" protobuf:"bytes,3,opt,name=status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RouteList is a collection of Routes.
type RouteList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object metadata.
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// items is a list of routes
	Items []Route `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// RouteSpec describes the hostname or path the route exposes, any security information,
// and one to four backends (services) the route points to. Requests are distributed
// among the backends depending on the weights assigned to each backend. When using
// roundrobin scheduling the portion of requests that go to each backend is the backend
// weight divided by the sum of all of the backend weights. When the backend has more than
// one endpoint the requests that end up on the backend are roundrobin distributed among
// the endpoints. Weights are between 0 and 256 with default 1. Weight 0 causes no requests
// to the backend. If all weights are zero the route will be considered to have no backends
// and return a standard 503 response.
//
// The `tls` ***REMOVED***eld is optional and allows speci***REMOVED***c certi***REMOVED***cates or behavior for the
// route. Routers typically con***REMOVED***gure a default certi***REMOVED***cate on a wildcard domain to
// terminate routes without explicit certi***REMOVED***cates, but custom hostnames usually must
// choose passthrough (send traf***REMOVED***c directly to the backend via the TLS Server-Name-
// Indication ***REMOVED***eld) or provide a certi***REMOVED***cate.
type RouteSpec struct {
	// host is an alias/DNS that points to the service. Optional.
	// If not speci***REMOVED***ed a route name will typically be automatically
	// chosen.
	// Must follow DNS952 subdomain conventions.
	Host string `json:"host" protobuf:"bytes,1,opt,name=host"`
	// Path that the router watches for, to route traf***REMOVED***c for to the service. Optional
	Path string `json:"path,omitempty" protobuf:"bytes,2,opt,name=path"`

	// to is an object the route should use as the primary backend. Only the Service kind
	// is allowed, and it will be defaulted to Service. If the weight ***REMOVED***eld (0-256 default 1)
	// is set to zero, no traf***REMOVED***c will be sent to this backend.
	To RouteTargetReference `json:"to" protobuf:"bytes,3,opt,name=to"`

	// alternateBackends allows up to 3 additional backends to be assigned to the route.
	// Only the Service kind is allowed, and it will be defaulted to Service.
	// Use the weight ***REMOVED***eld in RouteTargetReference object to specify relative preference.
	AlternateBackends []RouteTargetReference `json:"alternateBackends,omitempty" protobuf:"bytes,4,rep,name=alternateBackends"`

	// If speci***REMOVED***ed, the port to be used by the router. Most routers will use all
	// endpoints exposed by the service by default - set this value to instruct routers
	// which port to use.
	Port *RoutePort `json:"port,omitempty" protobuf:"bytes,5,opt,name=port"`

	// The tls ***REMOVED***eld provides the ability to con***REMOVED***gure certi***REMOVED***cates and termination for the route.
	TLS *TLSCon***REMOVED***g `json:"tls,omitempty" protobuf:"bytes,6,opt,name=tls"`

	// Wildcard policy if any for the route.
	// Currently only 'Subdomain' or 'None' is allowed.
	WildcardPolicy WildcardPolicyType `json:"wildcardPolicy,omitempty" protobuf:"bytes,7,opt,name=wildcardPolicy"`
}

// RouteTargetReference speci***REMOVED***es the target that resolve into endpoints. Only the 'Service'
// kind is allowed. Use 'weight' ***REMOVED***eld to emphasize one over others.
type RouteTargetReference struct {
	// The kind of target that the route is referring to. Currently, only 'Service' is allowed
	Kind string `json:"kind" protobuf:"bytes,1,opt,name=kind"`

	// name of the service/target that is being referred to. e.g. name of the service
	Name string `json:"name" protobuf:"bytes,2,opt,name=name"`

	// weight as an integer between 0 and 256, default 1, that speci***REMOVED***es the target's relative weight
	// against other target reference objects. 0 suppresses requests to this backend.
	Weight *int32 `json:"weight" protobuf:"varint,3,opt,name=weight"`
}

// RoutePort de***REMOVED***nes a port mapping from a router to an endpoint in the service endpoints.
type RoutePort struct {
	// The target port on pods selected by the service this route points to.
	// If this is a string, it will be looked up as a named port in the target
	// endpoints port list. Required
	TargetPort intstr.IntOrString `json:"targetPort" protobuf:"bytes,1,opt,name=targetPort"`
}

// RouteStatus provides relevant info about the status of a route, including which routers
// acknowledge it.
type RouteStatus struct {
	// ingress describes the places where the route may be exposed. The list of
	// ingress points may contain duplicate Host or RouterName values. Routes
	// are considered live once they are `Ready`
	Ingress []RouteIngress `json:"ingress" protobuf:"bytes,1,rep,name=ingress"`
}

// RouteIngress holds information about the places where a route is exposed.
type RouteIngress struct {
	// Host is the host string under which the route is exposed; this value is required
	Host string `json:"host,omitempty" protobuf:"bytes,1,opt,name=host"`
	// Name is a name chosen by the router to identify itself; this value is required
	RouterName string `json:"routerName,omitempty" protobuf:"bytes,2,opt,name=routerName"`
	// Conditions is the state of the route, may be empty.
	Conditions []RouteIngressCondition `json:"conditions,omitempty" protobuf:"bytes,3,rep,name=conditions"`
	// Wildcard policy is the wildcard policy that was allowed where this route is exposed.
	WildcardPolicy WildcardPolicyType `json:"wildcardPolicy,omitempty" protobuf:"bytes,4,opt,name=wildcardPolicy"`
	// CanonicalHostname is the external host name for the router that can be used as a CNAME
	// for the host requested for this route. This value is optional and may not be set in all cases.
	RouterCanonicalHostname string `json:"routerCanonicalHostname,omitempty" protobuf:"bytes,5,opt,name=routerCanonicalHostname"`
}

// RouteIngressConditionType is a valid value for RouteCondition
type RouteIngressConditionType string

// These are valid conditions of pod.
const (
	// RouteAdmitted means the route is able to service requests for the provided Host
	RouteAdmitted RouteIngressConditionType = "Admitted"
	// TODO: add other route condition types
)

// RouteIngressCondition contains details for the current condition of this route on a particular
// router.
type RouteIngressCondition struct {
	// Type is the type of the condition.
	// Currently only Ready.
	Type RouteIngressConditionType `json:"type" protobuf:"bytes,1,opt,name=type,casttype=RouteIngressConditionType"`
	// Status is the status of the condition.
	// Can be True, False, Unknown.
	Status corev1.ConditionStatus `json:"status" protobuf:"bytes,2,opt,name=status,casttype=k8s.io/api/core/v1.ConditionStatus"`
	// (brief) reason for the condition's last transition, and is usually a machine and human
	// readable constant
	Reason string `json:"reason,omitempty" protobuf:"bytes,3,opt,name=reason"`
	// Human readable message indicating details about last transition.
	Message string `json:"message,omitempty" protobuf:"bytes,4,opt,name=message"`
	// RFC 3339 date and time when this condition last transitioned
	LastTransitionTime *metav1.Time `json:"lastTransitionTime,omitempty" protobuf:"bytes,5,opt,name=lastTransitionTime"`
}

// RouterShard has information of a routing shard and is used to
// generate host names and routing table entries when a routing shard is
// allocated for a speci***REMOVED***c route.
// Caveat: This is WIP and will likely undergo modi***REMOVED***cations when sharding
//         support is added.
type RouterShard struct {
	// shardName uniquely identi***REMOVED***es a router shard in the "set" of
	// routers used for routing traf***REMOVED***c to the services.
	ShardName string `json:"shardName" protobuf:"bytes,1,opt,name=shardName"`

	// dnsSuf***REMOVED***x for the shard ala: shard-1.v3.openshift.com
	DNSSuf***REMOVED***x string `json:"dnsSuf***REMOVED***x" protobuf:"bytes,2,opt,name=dnsSuf***REMOVED***x"`
}

// TLSCon***REMOVED***g de***REMOVED***nes con***REMOVED***g used to secure a route and provide termination
type TLSCon***REMOVED***g struct {
	// termination indicates termination type.
	Termination TLSTerminationType `json:"termination" protobuf:"bytes,1,opt,name=termination,casttype=TLSTerminationType"`

	// certi***REMOVED***cate provides certi***REMOVED***cate contents
	Certi***REMOVED***cate string `json:"certi***REMOVED***cate,omitempty" protobuf:"bytes,2,opt,name=certi***REMOVED***cate"`

	// key provides key ***REMOVED***le contents
	Key string `json:"key,omitempty" protobuf:"bytes,3,opt,name=key"`

	// caCerti***REMOVED***cate provides the cert authority certi***REMOVED***cate contents
	CACerti***REMOVED***cate string `json:"caCerti***REMOVED***cate,omitempty" protobuf:"bytes,4,opt,name=caCerti***REMOVED***cate"`

	// destinationCACerti***REMOVED***cate provides the contents of the ca certi***REMOVED***cate of the ***REMOVED***nal destination.  When using reencrypt
	// termination this ***REMOVED***le should be provided in order to have routers use it for health checks on the secure connection.
	// If this ***REMOVED***eld is not speci***REMOVED***ed, the router may provide its own destination CA and perform hostname validation using
	// the short service name (service.namespace.svc), which allows infrastructure generated certi***REMOVED***cates to automatically
	// verify.
	DestinationCACerti***REMOVED***cate string `json:"destinationCACerti***REMOVED***cate,omitempty" protobuf:"bytes,5,opt,name=destinationCACerti***REMOVED***cate"`

	// insecureEdgeTerminationPolicy indicates the desired behavior for insecure connections to a route. While
	// each router may make its own decisions on which ports to expose, this is normally port 80.
	//
	// * Allow - traf***REMOVED***c is sent to the server on the insecure port (default)
	// * Disable - no traf***REMOVED***c is allowed on the insecure port.
	// * Redirect - clients are redirected to the secure port.
	InsecureEdgeTerminationPolicy InsecureEdgeTerminationPolicyType `json:"insecureEdgeTerminationPolicy,omitempty" protobuf:"bytes,6,opt,name=insecureEdgeTerminationPolicy,casttype=InsecureEdgeTerminationPolicyType"`
}

// TLSTerminationType dictates where the secure communication will stop
// TODO: Reconsider this type in v2
type TLSTerminationType string

// InsecureEdgeTerminationPolicyType dictates the behavior of insecure
// connections to an edge-terminated route.
type InsecureEdgeTerminationPolicyType string

const (
	// TLSTerminationEdge terminate encryption at the edge router.
	TLSTerminationEdge TLSTerminationType = "edge"
	// TLSTerminationPassthrough terminate encryption at the destination, the destination is responsible for decrypting traf***REMOVED***c
	TLSTerminationPassthrough TLSTerminationType = "passthrough"
	// TLSTerminationReencrypt terminate encryption at the edge router and re-encrypt it with a new certi***REMOVED***cate supplied by the destination
	TLSTerminationReencrypt TLSTerminationType = "reencrypt"

	// InsecureEdgeTerminationPolicyNone disables insecure connections for an edge-terminated route.
	InsecureEdgeTerminationPolicyNone InsecureEdgeTerminationPolicyType = "None"
	// InsecureEdgeTerminationPolicyAllow allows insecure connections for an edge-terminated route.
	InsecureEdgeTerminationPolicyAllow InsecureEdgeTerminationPolicyType = "Allow"
	// InsecureEdgeTerminationPolicyRedirect redirects insecure connections for an edge-terminated route.
	// As an example, for routers that support HTTP and HTTPS, the
	// insecure HTTP connections will be redirected to use HTTPS.
	InsecureEdgeTerminationPolicyRedirect InsecureEdgeTerminationPolicyType = "Redirect"
)

// WildcardPolicyType indicates the type of wildcard support needed by routes.
type WildcardPolicyType string

const (
	// WildcardPolicyNone indicates no wildcard support is needed.
	WildcardPolicyNone WildcardPolicyType = "None"

	// WildcardPolicySubdomain indicates the host needs wildcard support for the subdomain.
	// Example: For host = "www.acme.test", indicates that the router
	//          should support requests for *.acme.test
	//          Note that this will not match acme.test only *.acme.test
	WildcardPolicySubdomain WildcardPolicyType = "Subdomain"
)
