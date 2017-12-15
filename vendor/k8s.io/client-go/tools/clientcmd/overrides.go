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

package clientcmd

import (
	"strconv"
	"strings"

	"github.com/spf13/pflag"

	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// Con***REMOVED***gOverrides holds values that should override whatever information is pulled from the actual Con***REMOVED***g object.  You can't
// simply use an actual Con***REMOVED***g object, because Con***REMOVED***gs hold maps, but overrides are restricted to "at most one"
type Con***REMOVED***gOverrides struct {
	AuthInfo clientcmdapi.AuthInfo
	// ClusterDefaults are applied before the con***REMOVED***gured cluster info is loaded.
	ClusterDefaults clientcmdapi.Cluster
	ClusterInfo     clientcmdapi.Cluster
	Context         clientcmdapi.Context
	CurrentContext  string
	Timeout         string
}

// Con***REMOVED***gOverrideFlags holds the flag names to be used for binding command line flags. Notice that this structure tightly
// corresponds to Con***REMOVED***gOverrides
type Con***REMOVED***gOverrideFlags struct {
	AuthOverrideFlags    AuthOverrideFlags
	ClusterOverrideFlags ClusterOverrideFlags
	ContextOverrideFlags ContextOverrideFlags
	CurrentContext       FlagInfo
	Timeout              FlagInfo
}

// AuthOverrideFlags holds the flag names to be used for binding command line flags for AuthInfo objects
type AuthOverrideFlags struct {
	ClientCerti***REMOVED***cate FlagInfo
	ClientKey         FlagInfo
	Token             FlagInfo
	Impersonate       FlagInfo
	ImpersonateGroups FlagInfo
	Username          FlagInfo
	Password          FlagInfo
}

// ContextOverrideFlags holds the flag names to be used for binding command line flags for Cluster objects
type ContextOverrideFlags struct {
	ClusterName  FlagInfo
	AuthInfoName FlagInfo
	Namespace    FlagInfo
}

// ClusterOverride holds the flag names to be used for binding command line flags for Cluster objects
type ClusterOverrideFlags struct {
	APIServer             FlagInfo
	APIVersion            FlagInfo
	Certi***REMOVED***cateAuthority  FlagInfo
	InsecureSkipTLSVerify FlagInfo
}

// FlagInfo contains information about how to register a flag.  This struct is useful if you want to provide a way for an extender to
// get back a set of recommended flag names, descriptions, and defaults, but allow for customization by an extender.  This makes for
// coherent extension, without full prescription
type FlagInfo struct {
	// LongName is the long string for a flag.  If this is empty, then the flag will not be bound
	LongName string
	// ShortName is the single character for a flag.  If this is empty, then there will be no short flag
	ShortName string
	// Default is the default value for the flag
	Default string
	// Description is the description for the flag
	Description string
}

// AddSecretAnnotation add secret flag to Annotation.
func (f FlagInfo) AddSecretAnnotation(flags *pflag.FlagSet) FlagInfo {
	flags.SetAnnotation(f.LongName, "classi***REMOVED***ed", []string{"true"})
	return f
}

// BindStringFlag binds the flag based on the provided info.  If LongName == "", nothing is registered
func (f FlagInfo) BindStringFlag(flags *pflag.FlagSet, target *string) FlagInfo {
	// you can't register a flag without a long name
	if len(f.LongName) > 0 {
		flags.StringVarP(target, f.LongName, f.ShortName, f.Default, f.Description)
	}
	return f
}

// BindTransformingStringFlag binds the flag based on the provided info.  If LongName == "", nothing is registered
func (f FlagInfo) BindTransformingStringFlag(flags *pflag.FlagSet, target *string, transformer func(string) (string, error)) FlagInfo {
	// you can't register a flag without a long name
	if len(f.LongName) > 0 {
		flags.VarP(newTransformingStringValue(f.Default, target, transformer), f.LongName, f.ShortName, f.Description)
	}
	return f
}

// BindStringSliceFlag binds the flag based on the provided info.  If LongName == "", nothing is registered
func (f FlagInfo) BindStringArrayFlag(flags *pflag.FlagSet, target *[]string) FlagInfo {
	// you can't register a flag without a long name
	if len(f.LongName) > 0 {
		sliceVal := []string{}
		if len(f.Default) > 0 {
			sliceVal = []string{f.Default}
		}
		flags.StringArrayVarP(target, f.LongName, f.ShortName, sliceVal, f.Description)
	}
	return f
}

// BindBoolFlag binds the flag based on the provided info.  If LongName == "", nothing is registered
func (f FlagInfo) BindBoolFlag(flags *pflag.FlagSet, target *bool) FlagInfo {
	// you can't register a flag without a long name
	if len(f.LongName) > 0 {
		// try to parse Default as a bool.  If it fails, assume false
		boolVal, err := strconv.ParseBool(f.Default)
		if err != nil {
			boolVal = false
		}

		flags.BoolVarP(target, f.LongName, f.ShortName, boolVal, f.Description)
	}
	return f
}

const (
	FlagClusterName      = "cluster"
	FlagAuthInfoName     = "user"
	FlagContext          = "context"
	FlagNamespace        = "namespace"
	FlagAPIServer        = "server"
	FlagInsecure         = "insecure-skip-tls-verify"
	FlagCertFile         = "client-certi***REMOVED***cate"
	FlagKeyFile          = "client-key"
	FlagCAFile           = "certi***REMOVED***cate-authority"
	FlagEmbedCerts       = "embed-certs"
	FlagBearerToken      = "token"
	FlagImpersonate      = "as"
	FlagImpersonateGroup = "as-group"
	FlagUsername         = "username"
	FlagPassword         = "password"
	FlagTimeout          = "request-timeout"
)

// RecommendedCon***REMOVED***gOverrideFlags is a convenience method to return recommended flag names pre***REMOVED***xed with a string of your choosing
func RecommendedCon***REMOVED***gOverrideFlags(pre***REMOVED***x string) Con***REMOVED***gOverrideFlags {
	return Con***REMOVED***gOverrideFlags{
		AuthOverrideFlags:    RecommendedAuthOverrideFlags(pre***REMOVED***x),
		ClusterOverrideFlags: RecommendedClusterOverrideFlags(pre***REMOVED***x),
		ContextOverrideFlags: RecommendedContextOverrideFlags(pre***REMOVED***x),

		CurrentContext: FlagInfo{pre***REMOVED***x + FlagContext, "", "", "The name of the kubecon***REMOVED***g context to use"},
		Timeout:        FlagInfo{pre***REMOVED***x + FlagTimeout, "", "0", "The length of time to wait before giving up on a single server request. Non-zero values should contain a corresponding time unit (e.g. 1s, 2m, 3h). A value of zero means don't timeout requests."},
	}
}

// RecommendedAuthOverrideFlags is a convenience method to return recommended flag names pre***REMOVED***xed with a string of your choosing
func RecommendedAuthOverrideFlags(pre***REMOVED***x string) AuthOverrideFlags {
	return AuthOverrideFlags{
		ClientCerti***REMOVED***cate: FlagInfo{pre***REMOVED***x + FlagCertFile, "", "", "Path to a client certi***REMOVED***cate ***REMOVED***le for TLS"},
		ClientKey:         FlagInfo{pre***REMOVED***x + FlagKeyFile, "", "", "Path to a client key ***REMOVED***le for TLS"},
		Token:             FlagInfo{pre***REMOVED***x + FlagBearerToken, "", "", "Bearer token for authentication to the API server"},
		Impersonate:       FlagInfo{pre***REMOVED***x + FlagImpersonate, "", "", "Username to impersonate for the operation"},
		ImpersonateGroups: FlagInfo{pre***REMOVED***x + FlagImpersonateGroup, "", "", "Group to impersonate for the operation, this flag can be repeated to specify multiple groups."},
		Username:          FlagInfo{pre***REMOVED***x + FlagUsername, "", "", "Username for basic authentication to the API server"},
		Password:          FlagInfo{pre***REMOVED***x + FlagPassword, "", "", "Password for basic authentication to the API server"},
	}
}

// RecommendedClusterOverrideFlags is a convenience method to return recommended flag names pre***REMOVED***xed with a string of your choosing
func RecommendedClusterOverrideFlags(pre***REMOVED***x string) ClusterOverrideFlags {
	return ClusterOverrideFlags{
		APIServer:             FlagInfo{pre***REMOVED***x + FlagAPIServer, "", "", "The address and port of the Kubernetes API server"},
		Certi***REMOVED***cateAuthority:  FlagInfo{pre***REMOVED***x + FlagCAFile, "", "", "Path to a cert ***REMOVED***le for the certi***REMOVED***cate authority"},
		InsecureSkipTLSVerify: FlagInfo{pre***REMOVED***x + FlagInsecure, "", "false", "If true, the server's certi***REMOVED***cate will not be checked for validity. This will make your HTTPS connections insecure"},
	}
}

// RecommendedContextOverrideFlags is a convenience method to return recommended flag names pre***REMOVED***xed with a string of your choosing
func RecommendedContextOverrideFlags(pre***REMOVED***x string) ContextOverrideFlags {
	return ContextOverrideFlags{
		ClusterName:  FlagInfo{pre***REMOVED***x + FlagClusterName, "", "", "The name of the kubecon***REMOVED***g cluster to use"},
		AuthInfoName: FlagInfo{pre***REMOVED***x + FlagAuthInfoName, "", "", "The name of the kubecon***REMOVED***g user to use"},
		Namespace:    FlagInfo{pre***REMOVED***x + FlagNamespace, "n", "", "If present, the namespace scope for this CLI request"},
	}
}

// BindOverrideFlags is a convenience method to bind the speci***REMOVED***ed flags to their associated variables
func BindOverrideFlags(overrides *Con***REMOVED***gOverrides, flags *pflag.FlagSet, flagNames Con***REMOVED***gOverrideFlags) {
	BindAuthInfoFlags(&overrides.AuthInfo, flags, flagNames.AuthOverrideFlags)
	BindClusterFlags(&overrides.ClusterInfo, flags, flagNames.ClusterOverrideFlags)
	BindContextFlags(&overrides.Context, flags, flagNames.ContextOverrideFlags)
	flagNames.CurrentContext.BindStringFlag(flags, &overrides.CurrentContext)
	flagNames.Timeout.BindStringFlag(flags, &overrides.Timeout)
}

// BindAuthInfoFlags is a convenience method to bind the speci***REMOVED***ed flags to their associated variables
func BindAuthInfoFlags(authInfo *clientcmdapi.AuthInfo, flags *pflag.FlagSet, flagNames AuthOverrideFlags) {
	flagNames.ClientCerti***REMOVED***cate.BindStringFlag(flags, &authInfo.ClientCerti***REMOVED***cate).AddSecretAnnotation(flags)
	flagNames.ClientKey.BindStringFlag(flags, &authInfo.ClientKey).AddSecretAnnotation(flags)
	flagNames.Token.BindStringFlag(flags, &authInfo.Token).AddSecretAnnotation(flags)
	flagNames.Impersonate.BindStringFlag(flags, &authInfo.Impersonate).AddSecretAnnotation(flags)
	flagNames.ImpersonateGroups.BindStringArrayFlag(flags, &authInfo.ImpersonateGroups).AddSecretAnnotation(flags)
	flagNames.Username.BindStringFlag(flags, &authInfo.Username).AddSecretAnnotation(flags)
	flagNames.Password.BindStringFlag(flags, &authInfo.Password).AddSecretAnnotation(flags)
}

// BindClusterFlags is a convenience method to bind the speci***REMOVED***ed flags to their associated variables
func BindClusterFlags(clusterInfo *clientcmdapi.Cluster, flags *pflag.FlagSet, flagNames ClusterOverrideFlags) {
	flagNames.APIServer.BindStringFlag(flags, &clusterInfo.Server)
	flagNames.Certi***REMOVED***cateAuthority.BindStringFlag(flags, &clusterInfo.Certi***REMOVED***cateAuthority)
	flagNames.InsecureSkipTLSVerify.BindBoolFlag(flags, &clusterInfo.InsecureSkipTLSVerify)
}

// BindFlags is a convenience method to bind the speci***REMOVED***ed flags to their associated variables
func BindContextFlags(contextInfo *clientcmdapi.Context, flags *pflag.FlagSet, flagNames ContextOverrideFlags) {
	flagNames.ClusterName.BindStringFlag(flags, &contextInfo.Cluster)
	flagNames.AuthInfoName.BindStringFlag(flags, &contextInfo.AuthInfo)
	flagNames.Namespace.BindTransformingStringFlag(flags, &contextInfo.Namespace, RemoveNamespacesPre***REMOVED***x)
}

// RemoveNamespacesPre***REMOVED***x is a transformer that strips "ns/", "namespace/" and "namespaces/" pre***REMOVED***xes case-insensitively
func RemoveNamespacesPre***REMOVED***x(value string) (string, error) {
	for _, pre***REMOVED***x := range []string{"namespaces/", "namespace/", "ns/"} {
		if len(value) > len(pre***REMOVED***x) && strings.EqualFold(value[0:len(pre***REMOVED***x)], pre***REMOVED***x) {
			value = value[len(pre***REMOVED***x):]
			break
		}
	}
	return value, nil
}
