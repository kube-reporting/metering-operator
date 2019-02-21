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

package api

import (
	"k8s.io/apimachinery/pkg/runtime"
)

// Where possible, json tags match the cli argument names.
// Top level con***REMOVED***g objects and all values required for proper functioning are not "omitempty".  Any truly optional piece of con***REMOVED***g is allowed to be omitted.

// Con***REMOVED***g holds the information needed to build connect to remote kubernetes clusters as a given user
// IMPORTANT if you add ***REMOVED***elds to this struct, please update IsCon***REMOVED***gEmpty()
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
	Clusters map[string]*Cluster `json:"clusters"`
	// AuthInfos is a map of referencable names to user con***REMOVED***gs
	AuthInfos map[string]*AuthInfo `json:"users"`
	// Contexts is a map of referencable names to context con***REMOVED***gs
	Contexts map[string]*Context `json:"contexts"`
	// CurrentContext is the name of the context that you would like to use by default
	CurrentContext string `json:"current-context"`
	// Extensions holds additional information. This is useful for extenders so that reads and writes don't clobber unknown ***REMOVED***elds
	// +optional
	Extensions map[string]runtime.Object `json:"extensions,omitempty"`
}

// IMPORTANT if you add ***REMOVED***elds to this struct, please update IsCon***REMOVED***gEmpty()
type Preferences struct {
	// +optional
	Colors bool `json:"colors,omitempty"`
	// Extensions holds additional information. This is useful for extenders so that reads and writes don't clobber unknown ***REMOVED***elds
	// +optional
	Extensions map[string]runtime.Object `json:"extensions,omitempty"`
}

// Cluster contains information about how to communicate with a kubernetes cluster
type Cluster struct {
	// LocationOfOrigin indicates where this object came from.  It is used for round tripping con***REMOVED***g post-merge, but never serialized.
	LocationOfOrigin string
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
	Extensions map[string]runtime.Object `json:"extensions,omitempty"`
}

// AuthInfo contains information that describes identity information.  This is use to tell the kubernetes cluster who you are.
type AuthInfo struct {
	// LocationOfOrigin indicates where this object came from.  It is used for round tripping con***REMOVED***g post-merge, but never serialized.
	LocationOfOrigin string
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
	// Impersonate is the username to act-as.
	// +optional
	Impersonate string `json:"act-as,omitempty"`
	// ImpersonateGroups is the groups to imperonate.
	// +optional
	ImpersonateGroups []string `json:"act-as-groups,omitempty"`
	// ImpersonateUserExtra contains additional information for impersonated user.
	// +optional
	ImpersonateUserExtra map[string][]string `json:"act-as-user-extra,omitempty"`
	// Username is the username for basic authentication to the kubernetes cluster.
	// +optional
	Username string `json:"username,omitempty"`
	// Password is the password for basic authentication to the kubernetes cluster.
	// +optional
	Password string `json:"password,omitempty"`
	// AuthProvider speci***REMOVED***es a custom authentication plugin for the kubernetes cluster.
	// +optional
	AuthProvider *AuthProviderCon***REMOVED***g `json:"auth-provider,omitempty"`
	// Exec speci***REMOVED***es a custom exec-based authentication plugin for the kubernetes cluster.
	// +optional
	Exec *ExecCon***REMOVED***g `json:"exec,omitempty"`
	// Extensions holds additional information. This is useful for extenders so that reads and writes don't clobber unknown ***REMOVED***elds
	// +optional
	Extensions map[string]runtime.Object `json:"extensions,omitempty"`
}

// Context is a tuple of references to a cluster (how do I communicate with a kubernetes cluster), a user (how do I identify myself), and a namespace (what subset of resources do I want to work with)
type Context struct {
	// LocationOfOrigin indicates where this object came from.  It is used for round tripping con***REMOVED***g post-merge, but never serialized.
	LocationOfOrigin string
	// Cluster is the name of the cluster for this context
	Cluster string `json:"cluster"`
	// AuthInfo is the name of the authInfo for this context
	AuthInfo string `json:"user"`
	// Namespace is the default namespace to use on unspeci***REMOVED***ed requests
	// +optional
	Namespace string `json:"namespace,omitempty"`
	// Extensions holds additional information. This is useful for extenders so that reads and writes don't clobber unknown ***REMOVED***elds
	// +optional
	Extensions map[string]runtime.Object `json:"extensions,omitempty"`
}

// AuthProviderCon***REMOVED***g holds the con***REMOVED***guration for a speci***REMOVED***ed auth provider.
type AuthProviderCon***REMOVED***g struct {
	Name string `json:"name"`
	// +optional
	Con***REMOVED***g map[string]string `json:"con***REMOVED***g,omitempty"`
}

// ExecCon***REMOVED***g speci***REMOVED***es a command to provide client credentials. The command is exec'd
// and outputs structured stdout holding credentials.
//
// See the client.authentiction.k8s.io API group for speci***REMOVED***cations of the exact input
// and output format
type ExecCon***REMOVED***g struct {
	// Command to execute.
	Command string `json:"command"`
	// Arguments to pass to the command when executing it.
	// +optional
	Args []string `json:"args"`
	// Env de***REMOVED***nes additional environment variables to expose to the process. These
	// are unioned with the host's environment, as well as variables client-go uses
	// to pass argument to the plugin.
	// +optional
	Env []ExecEnvVar `json:"env"`

	// Preferred input version of the ExecInfo. The returned ExecCredentials MUST use
	// the same encoding version as the input.
	APIVersion string `json:"apiVersion,omitempty"`
}

// ExecEnvVar is used for setting environment variables when executing an exec-based
// credential plugin.
type ExecEnvVar struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// NewCon***REMOVED***g is a convenience function that returns a new Con***REMOVED***g object with non-nil maps
func NewCon***REMOVED***g() *Con***REMOVED***g {
	return &Con***REMOVED***g{
		Preferences: *NewPreferences(),
		Clusters:    make(map[string]*Cluster),
		AuthInfos:   make(map[string]*AuthInfo),
		Contexts:    make(map[string]*Context),
		Extensions:  make(map[string]runtime.Object),
	}
}

// NewContext is a convenience function that returns a new Context
// object with non-nil maps
func NewContext() *Context {
	return &Context{Extensions: make(map[string]runtime.Object)}
}

// NewCluster is a convenience function that returns a new Cluster
// object with non-nil maps
func NewCluster() *Cluster {
	return &Cluster{Extensions: make(map[string]runtime.Object)}
}

// NewAuthInfo is a convenience function that returns a new AuthInfo
// object with non-nil maps
func NewAuthInfo() *AuthInfo {
	return &AuthInfo{
		Extensions:           make(map[string]runtime.Object),
		ImpersonateUserExtra: make(map[string][]string),
	}
}

// NewPreferences is a convenience function that returns a new
// Preferences object with non-nil maps
func NewPreferences() *Preferences {
	return &Preferences{Extensions: make(map[string]runtime.Object)}
}
