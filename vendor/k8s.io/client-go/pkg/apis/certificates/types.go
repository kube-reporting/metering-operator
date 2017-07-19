/*
Copyright 2016 The Kubernetes Authors.

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

package certi***REMOVED***cates

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +genclient=true
// +nonNamespaced=true

// Describes a certi***REMOVED***cate signing request
type Certi***REMOVED***cateSigningRequest struct {
	metav1.TypeMeta
	// +optional
	metav1.ObjectMeta

	// The certi***REMOVED***cate request itself and any additional information.
	// +optional
	Spec Certi***REMOVED***cateSigningRequestSpec

	// Derived information about the request.
	// +optional
	Status Certi***REMOVED***cateSigningRequestStatus
}

// This information is immutable after the request is created. Only the Request
// and Usages ***REMOVED***elds can be set on creation, other ***REMOVED***elds are derived by
// Kubernetes and cannot be modi***REMOVED***ed by users.
type Certi***REMOVED***cateSigningRequestSpec struct {
	// Base64-encoded PKCS#10 CSR data
	Request []byte

	// usages speci***REMOVED***es a set of usage contexts the key will be
	// valid for.
	// See: https://tools.ietf.org/html/rfc5280#section-4.2.1.3
	//      https://tools.ietf.org/html/rfc5280#section-4.2.1.12
	Usages []KeyUsage

	// Information about the requesting user.
	// See user.Info interface for details.
	// +optional
	Username string
	// UID information about the requesting user.
	// See user.Info interface for details.
	// +optional
	UID string
	// Group information about the requesting user.
	// See user.Info interface for details.
	// +optional
	Groups []string
	// Extra information about the requesting user.
	// See user.Info interface for details.
	// +optional
	Extra map[string]ExtraValue
}

// ExtraValue masks the value so protobuf can generate
type ExtraValue []string

type Certi***REMOVED***cateSigningRequestStatus struct {
	// Conditions applied to the request, such as approval or denial.
	// +optional
	Conditions []Certi***REMOVED***cateSigningRequestCondition

	// If request was approved, the controller will place the issued certi***REMOVED***cate here.
	// +optional
	Certi***REMOVED***cate []byte
}

type RequestConditionType string

// These are the possible conditions for a certi***REMOVED***cate request.
const (
	Certi***REMOVED***cateApproved RequestConditionType = "Approved"
	Certi***REMOVED***cateDenied   RequestConditionType = "Denied"
)

type Certi***REMOVED***cateSigningRequestCondition struct {
	// request approval state, currently Approved or Denied.
	Type RequestConditionType
	// brief reason for the request state
	// +optional
	Reason string
	// human readable message with details about the request state
	// +optional
	Message string
	// timestamp for the last update to this condition
	// +optional
	LastUpdateTime metav1.Time
}

type Certi***REMOVED***cateSigningRequestList struct {
	metav1.TypeMeta
	// +optional
	metav1.ListMeta

	// +optional
	Items []Certi***REMOVED***cateSigningRequest
}

// KeyUsages speci***REMOVED***es valid usage contexts for keys.
// See: https://tools.ietf.org/html/rfc5280#section-4.2.1.3
//      https://tools.ietf.org/html/rfc5280#section-4.2.1.12
type KeyUsage string

const (
	UsageSigning            KeyUsage = "signing"
	UsageDigitalSignature   KeyUsage = "digital signature"
	UsageContentCommittment KeyUsage = "content committment"
	UsageKeyEncipherment    KeyUsage = "key encipherment"
	UsageKeyAgreement       KeyUsage = "key agreement"
	UsageDataEncipherment   KeyUsage = "data encipherment"
	UsageCertSign           KeyUsage = "cert sign"
	UsageCRLSign            KeyUsage = "crl sign"
	UsageEncipherOnly       KeyUsage = "encipher only"
	UsageDecipherOnly       KeyUsage = "decipher only"
	UsageAny                KeyUsage = "any"
	UsageServerAuth         KeyUsage = "server auth"
	UsageClientAuth         KeyUsage = "client auth"
	UsageCodeSigning        KeyUsage = "code signing"
	UsageEmailProtection    KeyUsage = "email protection"
	UsageSMIME              KeyUsage = "s/mime"
	UsageIPsecEndSystem     KeyUsage = "ipsec end system"
	UsageIPsecTunnel        KeyUsage = "ipsec tunnel"
	UsageIPsecUser          KeyUsage = "ipsec user"
	UsageTimestamping       KeyUsage = "timestamping"
	UsageOCSPSigning        KeyUsage = "ocsp signing"
	UsageMicrosoftSGC       KeyUsage = "microsoft sgc"
	UsageNetscapSGC         KeyUsage = "netscape sgc"
)
