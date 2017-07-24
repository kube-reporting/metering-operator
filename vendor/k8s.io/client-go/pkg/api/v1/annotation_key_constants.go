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

// This ***REMOVED***le should be consistent with pkg/api/annotation_key_constants.go.

package v1

const (
	// ImagePolicyFailedOpenKey is added to pods created by failing open when the image policy
	// webhook backend fails.
	ImagePolicyFailedOpenKey string = "alpha.image-policy.k8s.io/failed-open"

	// PodPresetOptOutAnnotationKey represents the annotation key for a pod to exempt itself from pod preset manipulation
	PodPresetOptOutAnnotationKey string = "podpreset.admission.kubernetes.io/exclude"

	// MirrorAnnotationKey represents the annotation key set by kubelets when creating mirror pods
	MirrorPodAnnotationKey string = "kubernetes.io/con***REMOVED***g.mirror"

	// TolerationsAnnotationKey represents the key of tolerations data (json serialized)
	// in the Annotations of a Pod.
	TolerationsAnnotationKey string = "scheduler.alpha.kubernetes.io/tolerations"

	// TaintsAnnotationKey represents the key of taints data (json serialized)
	// in the Annotations of a Node.
	TaintsAnnotationKey string = "scheduler.alpha.kubernetes.io/taints"

	// SeccompPodAnnotationKey represents the key of a seccomp pro***REMOVED***le applied
	// to all containers of a pod.
	SeccompPodAnnotationKey string = "seccomp.security.alpha.kubernetes.io/pod"

	// SeccompContainerAnnotationKeyPre***REMOVED***x represents the key of a seccomp pro***REMOVED***le applied
	// to one container of a pod.
	SeccompContainerAnnotationKeyPre***REMOVED***x string = "container.seccomp.security.alpha.kubernetes.io/"

	// CreatedByAnnotation represents the key used to store the spec(json)
	// used to create the resource.
	CreatedByAnnotation = "kubernetes.io/created-by"

	// PreferAvoidPodsAnnotationKey represents the key of preferAvoidPods data (json serialized)
	// in the Annotations of a Node.
	PreferAvoidPodsAnnotationKey string = "scheduler.alpha.kubernetes.io/preferAvoidPods"

	// SysctlsPodAnnotationKey represents the key of sysctls which are set for the infrastructure
	// container of a pod. The annotation value is a comma separated list of sysctl_name=value
	// key-value pairs. Only a limited set of whitelisted and isolated sysctls is supported by
	// the kubelet. Pods with other sysctls will fail to launch.
	SysctlsPodAnnotationKey string = "security.alpha.kubernetes.io/sysctls"

	// UnsafeSysctlsPodAnnotationKey represents the key of sysctls which are set for the infrastructure
	// container of a pod. The annotation value is a comma separated list of sysctl_name=value
	// key-value pairs. Unsafe sysctls must be explicitly enabled for a kubelet. They are properly
	// namespaced to a pod or a container, but their isolation is usually unclear or weak. Their use
	// is at-your-own-risk. Pods that attempt to set an unsafe sysctl that is not enabled for a kubelet
	// will fail to launch.
	UnsafeSysctlsPodAnnotationKey string = "security.alpha.kubernetes.io/unsafe-sysctls"

	// ObjectTTLAnnotations represents a suggestion for kubelet for how long it can cache
	// an object (e.g. secret, con***REMOVED***g map) before fetching it again from apiserver.
	// This annotation can be attached to node.
	ObjectTTLAnnotationKey string = "node.alpha.kubernetes.io/ttl"

	// Af***REMOVED***nityAnnotationKey represents the key of af***REMOVED***nity data (json serialized)
	// in the Annotations of a Pod.
	// TODO: remove when alpha support for af***REMOVED***nity is removed
	Af***REMOVED***nityAnnotationKey string = "scheduler.alpha.kubernetes.io/af***REMOVED***nity"

	// annotation key pre***REMOVED***x used to identify non-convertible json paths.
	NonConvertibleAnnotationPre***REMOVED***x = "non-convertible.kubernetes.io"

	kubectlPre***REMOVED***x = "kubectl.kubernetes.io/"

	// LastAppliedCon***REMOVED***gAnnotation is the annotation used to store the previous
	// con***REMOVED***guration of a resource for use in a three way diff by UpdateApplyAnnotation.
	LastAppliedCon***REMOVED***gAnnotation = kubectlPre***REMOVED***x + "last-applied-con***REMOVED***guration"

	// AnnotationLoadBalancerSourceRangesKey is the key of the annotation on a service to set allowed ingress ranges on their LoadBalancers
	//
	// It should be a comma-separated list of CIDRs, e.g. `0.0.0.0/0` to
	// allow full access (the default) or `18.0.0.0/8,56.0.0.0/8` to allow
	// access only from the CIDRs currently allocated to MIT & the USPS.
	//
	// Not all cloud providers support this annotation, though AWS & GCE do.
	AnnotationLoadBalancerSourceRangesKey = "service.beta.kubernetes.io/load-balancer-source-ranges"

	// AnnotationValueExternalTraf***REMOVED***cLocal Value of annotation to specify local endpoints behavior.
	AnnotationValueExternalTraf***REMOVED***cLocal = "OnlyLocal"
	// AnnotationValueExternalTraf***REMOVED***cGlobal Value of annotation to specify global (legacy) behavior.
	AnnotationValueExternalTraf***REMOVED***cGlobal = "Global"

	// TODO: The beta annotations have been deprecated, remove them when we release k8s 1.8.

	// BetaAnnotationHealthCheckNodePort Annotation specifying the healthcheck nodePort for the service.
	// If not speci***REMOVED***ed, annotation is created by the service api backend with the allocated nodePort.
	// Will use user-speci***REMOVED***ed nodePort value if speci***REMOVED***ed by the client.
	BetaAnnotationHealthCheckNodePort = "service.beta.kubernetes.io/healthcheck-nodeport"

	// BetaAnnotationExternalTraf***REMOVED***c An annotation that denotes if this Service desires to route
	// external traf***REMOVED***c to local endpoints only. This preserves Source IP and avoids a second hop.
	BetaAnnotationExternalTraf***REMOVED***c = "service.beta.kubernetes.io/external-traf***REMOVED***c"
)
