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

package v1

import (
	"k8s.io/apimachinery/pkg/runtime"
)

// Where possible, json tags match the cli argument names.
// Top level con***REMOVED***g objects and all values required for proper functioning are not "omitempty".  Any truly optional piece of con***REMOVED***g is allowed to be omitted.

// Con***REMOVED***g holds the information needed to build connect to remote kubernetes clusters as a given user
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Con***REMOVED***g struct {
	// Legacy ***REMOVED***eld from pkg/api/types.go TypeMeta.
	// TODO(jlowdermilk): remove this after eliminating downstream dependencies.
	// +optional
	Kind string `json:"kind,omitempty"`
	// Legacy ***REMOVED***eld from pkg/api/types.go TypeMeta.
	// TODO(jlowdermilk): remove this after eliminating downstream dependencies.
	// +optional
	APIVersion string `json:"apiVersion,omitempty"`
	// Preferences holds general information to be use for cli interactions
	Preferences Preferences `json:"preferences"`
	// Clusters is a map of referencable names to cluster con***REMOVED***gs
	Clusters []NamedCluster `json:"clusters"`
	// AuthInfos is a map of referencable names to user con***REMOVED***gs
	AuthInfos []NamedAuthInfo `json:"users"`
	// Contexts is a map of referencable names to context con***REMOVED***gs
	Contexts []NamedContext `json:"contexts"`
	// CurrentContext is the name of the context that you would like to use by default
	CurrentContext string `json:"current-context"`
	// Extensions holds additional information. This is useful for extenders so that reads and writes don't clobber unknown ***REMOVED***elds
	// +optional
	Extensions []NamedExtension `json:"extensions,omitempty"`
}

type Preferences struct {
	// +optional
	Colors bool `json:"colors,omitempty"`
	// Extensions holds additional information. This is useful for extenders so that reads and writes don't clobber unknown ***REMOVED***elds
	// +optional
	Extensions []NamedExtension `json:"extensions,omitempty"`
}

// Cluster contains information about how to communicate with a kubernetes cluster
type Cluster struct {
	// Server is the address of the kubernetes cluster (https://hostname:port).
	Server string `json:"server"`
	// InsecureSkipTLSVerify skips the validity check for the server's certi***REMOVED***cate. This will make your HTTPS connections insecure.
	// +optional
	InsecureSkipTLSVerify bool `json:"insecure-skip-tls-verify,omitempty"`
	// Certi***REMOVED***cateAuthority is the path to a cert ***REMOVED***le for the certi***REMOVED***cate authority.
	// +optional
	Certi***REMOVED***cateAuthority string `json:"certi***REMOVED***cate-authority,omitempty"`
	// Certi***REMOVED***cateAuthorityData contains PEM-encoded certi***REMOVED***cate authority certi***REMOVED***cates. Overrides Certi***REMOVED***cateAuthority
	// +optional
	Certi***REMOVED***cateAuthorityData []byte `json:"certi***REMOVED***cate-authority-data,omitempty"`
	// Extensions holds additional information. This is useful for extenders so that reads and writes don't clobber unknown ***REMOVED***elds
	// +optional
	Extensions []NamedExtension `json:"extensions,omitempty"`
}

// AuthInfo contains information that describes identity information.  This is use to tell the kubernetes cluster who you are.
type AuthInfo struct {
	// ClientCerti***REMOVED***cate is the path to a client cert ***REMOVED***le for TLS.
	// +optional
	ClientCerti***REMOVED***cate string `json:"client-certi***REMOVED***cate,omitempty"`
	// ClientCerti***REMOVED***cateData contains PEM-encoded data from a client cert ***REMOVED***le for TLS. Overrides ClientCerti***REMOVED***cate
	// +optional
	ClientCerti***REMOVED***cateData []byte `json:"client-certi***REMOVED***cate-data,omitempty"`
	// ClientKey is the path to a client key ***REMOVED***le for TLS.
	// +optional
	ClientKey string `json:"client-key,omitempty"`
	// ClientKeyData contains PEM-encoded data from a client key ***REMOVED***le for TLS. Overrides ClientKey
	// +optional
	ClientKeyData []byte `json:"client-key-data,omitempty"`
	// Token is the bearer token for authentication to the kubernetes cluster.
	// +optional
	Token string `json:"token,omitempty"`
	// TokenFile is a pointer to a ***REMOVED***le that contains a bearer token (as described above).  If both Token and TokenFile are present, Token takes precedence.
	// +optional
	TokenFile string `json:"tokenFile,omitempty"`
	// Impersonate is the username to imperonate.  The name matches the flag.
	// +optional
	Impersonate string `json:"as,omitempty"`
	// ImpersonateGroups is the groups to imperonate.
	// +optional
	ImpersonateGroups []string `json:"as-groups,omitempty"`
	// ImpersonateUserExtra contains additional information for impersonated user.
	// +optional
	ImpersonateUserExtra map[string][]string `json:"as-user-extra,omitempty"`
	// Username is the username for basic authentication to the kubernetes cluster.
	// +optional
	Username string `json:"username,omitempty"`
	// Password is the password for basic authentication to the kubernetes cluster.
	// +optional
	Password string `json:"password,omitempty"`
	// AuthProvider speci***REMOVED***es a custom authentication plugin for the kubernetes cluster.
	// +optional
	AuthProvider *AuthProviderCon***REMOVED***g `json:"auth-provider,omitempty"`
	// Extensions holds additional information. This is useful for extenders so that reads and writes don't clobber unknown ***REMOVED***elds
	// +optional
	Extensions []NamedExtension `json:"extensions,omitempty"`
}

// Context is a tuple of references to a cluster (how do I communicate with a kubernetes cluster), a user (how do I identify myself), and a namespace (what subset of resources do I want to work with)
type Context struct {
	// Cluster is the name of the cluster for this context
	Cluster string `json:"cluster"`
	// AuthInfo is the name of the authInfo for this context
	AuthInfo string `json:"user"`
	// Namespace is the default namespace to use on unspeci***REMOVED***ed requests
	// +optional
	Namespace string `json:"namespace,omitempty"`
	// Extensions holds additional information. This is useful for extenders so that reads and writes don't clobber unknown ***REMOVED***elds
	// +optional
	Extensions []NamedExtension `json:"extensions,omitempty"`
}

// NamedCluster relates nicknames to cluster information
type NamedCluster struct {
	// Name is the nickname for this Cluster
	Name string `json:"name"`
	// Cluster holds the cluster information
	Cluster Cluster `json:"cluster"`
}

// NamedContext relates nicknames to context information
type NamedContext struct {
	// Name is the nickname for this Context
	Name string `json:"name"`
	// Context holds the context information
	Context Context `json:"context"`
}

// NamedAuthInfo relates nicknames to auth information
type NamedAuthInfo struct {
	// Name is the nickname for this AuthInfo
	Name string `json:"name"`
	// AuthInfo holds the auth information
	AuthInfo AuthInfo `json:"user"`
}

// NamedExtension relates nicknames to extension information
type NamedExtension struct {
	// Name is the nickname for this Extension
	Name string `json:"name"`
	// Extension holds the extension information
	Extension runtime.RawExtension `json:"extension"`
}

// AuthProviderCon***REMOVED***g holds the con***REMOVED***guration for a speci***REMOVED***ed auth provider.
type AuthProviderCon***REMOVED***g struct {
	Name   string            `json:"name"`
	Con***REMOVED***g map[string]string `json:"con***REMOVED***g"`
}
