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
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"strings"

	"github.com/imdario/mergo"
	"k8s.io/klog"

	restclient "k8s.io/client-go/rest"
	clientauth "k8s.io/client-go/tools/auth"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

var (
	// ClusterDefaults has the same behavior as the old EnvVar and DefaultCluster ***REMOVED***elds
	// DEPRECATED will be replaced
	ClusterDefaults = clientcmdapi.Cluster{Server: getDefaultServer()}
	// DefaultClientCon***REMOVED***g represents the legacy behavior of this package for defaulting
	// DEPRECATED will be replace
	DefaultClientCon***REMOVED***g = DirectClientCon***REMOVED***g{*clientcmdapi.NewCon***REMOVED***g(), "", &Con***REMOVED***gOverrides{
		ClusterDefaults: ClusterDefaults,
	}, nil, NewDefaultClientCon***REMOVED***gLoadingRules(), promptedCredentials{}}
)

// getDefaultServer returns a default setting for DefaultClientCon***REMOVED***g
// DEPRECATED
func getDefaultServer() string {
	if server := os.Getenv("KUBERNETES_MASTER"); len(server) > 0 {
		return server
	}
	return "http://localhost:8080"
}

// ClientCon***REMOVED***g is used to make it easy to get an api server client
type ClientCon***REMOVED***g interface {
	// RawCon***REMOVED***g returns the merged result of all overrides
	RawCon***REMOVED***g() (clientcmdapi.Con***REMOVED***g, error)
	// ClientCon***REMOVED***g returns a complete client con***REMOVED***g
	ClientCon***REMOVED***g() (*restclient.Con***REMOVED***g, error)
	// Namespace returns the namespace resulting from the merged
	// result of all overrides and a boolean indicating if it was
	// overridden
	Namespace() (string, bool, error)
	// Con***REMOVED***gAccess returns the rules for loading/persisting the con***REMOVED***g.
	Con***REMOVED***gAccess() Con***REMOVED***gAccess
}

type PersistAuthProviderCon***REMOVED***gForUser func(user string) restclient.AuthProviderCon***REMOVED***gPersister

type promptedCredentials struct {
	username string
	password string
}

// DirectClientCon***REMOVED***g is a ClientCon***REMOVED***g interface that is backed by a clientcmdapi.Con***REMOVED***g, options overrides, and an optional fallbackReader for auth information
type DirectClientCon***REMOVED***g struct {
	con***REMOVED***g         clientcmdapi.Con***REMOVED***g
	contextName    string
	overrides      *Con***REMOVED***gOverrides
	fallbackReader io.Reader
	con***REMOVED***gAccess   Con***REMOVED***gAccess
	// promptedCredentials store the credentials input by the user
	promptedCredentials promptedCredentials
}

// NewDefaultClientCon***REMOVED***g creates a DirectClientCon***REMOVED***g using the con***REMOVED***g.CurrentContext as the context name
func NewDefaultClientCon***REMOVED***g(con***REMOVED***g clientcmdapi.Con***REMOVED***g, overrides *Con***REMOVED***gOverrides) ClientCon***REMOVED***g {
	return &DirectClientCon***REMOVED***g{con***REMOVED***g, con***REMOVED***g.CurrentContext, overrides, nil, NewDefaultClientCon***REMOVED***gLoadingRules(), promptedCredentials{}}
}

// NewNonInteractiveClientCon***REMOVED***g creates a DirectClientCon***REMOVED***g using the passed context name and does not have a fallback reader for auth information
func NewNonInteractiveClientCon***REMOVED***g(con***REMOVED***g clientcmdapi.Con***REMOVED***g, contextName string, overrides *Con***REMOVED***gOverrides, con***REMOVED***gAccess Con***REMOVED***gAccess) ClientCon***REMOVED***g {
	return &DirectClientCon***REMOVED***g{con***REMOVED***g, contextName, overrides, nil, con***REMOVED***gAccess, promptedCredentials{}}
}

// NewInteractiveClientCon***REMOVED***g creates a DirectClientCon***REMOVED***g using the passed context name and a reader in case auth information is not provided via ***REMOVED***les or flags
func NewInteractiveClientCon***REMOVED***g(con***REMOVED***g clientcmdapi.Con***REMOVED***g, contextName string, overrides *Con***REMOVED***gOverrides, fallbackReader io.Reader, con***REMOVED***gAccess Con***REMOVED***gAccess) ClientCon***REMOVED***g {
	return &DirectClientCon***REMOVED***g{con***REMOVED***g, contextName, overrides, fallbackReader, con***REMOVED***gAccess, promptedCredentials{}}
}

// NewClientCon***REMOVED***gFromBytes takes your kubecon***REMOVED***g and gives you back a ClientCon***REMOVED***g
func NewClientCon***REMOVED***gFromBytes(con***REMOVED***gBytes []byte) (ClientCon***REMOVED***g, error) {
	con***REMOVED***g, err := Load(con***REMOVED***gBytes)
	if err != nil {
		return nil, err
	}

	return &DirectClientCon***REMOVED***g{*con***REMOVED***g, "", &Con***REMOVED***gOverrides{}, nil, nil, promptedCredentials{}}, nil
}

// RESTCon***REMOVED***gFromKubeCon***REMOVED***g is a convenience method to give back a restcon***REMOVED***g from your kubecon***REMOVED***g bytes.
// For programmatic access, this is what you want 80% of the time
func RESTCon***REMOVED***gFromKubeCon***REMOVED***g(con***REMOVED***gBytes []byte) (*restclient.Con***REMOVED***g, error) {
	clientCon***REMOVED***g, err := NewClientCon***REMOVED***gFromBytes(con***REMOVED***gBytes)
	if err != nil {
		return nil, err
	}
	return clientCon***REMOVED***g.ClientCon***REMOVED***g()
}

func (con***REMOVED***g *DirectClientCon***REMOVED***g) RawCon***REMOVED***g() (clientcmdapi.Con***REMOVED***g, error) {
	return con***REMOVED***g.con***REMOVED***g, nil
}

// ClientCon***REMOVED***g implements ClientCon***REMOVED***g
func (con***REMOVED***g *DirectClientCon***REMOVED***g) ClientCon***REMOVED***g() (*restclient.Con***REMOVED***g, error) {
	// check that getAuthInfo, getContext, and getCluster do not return an error.
	// Do this before checking if the current con***REMOVED***g is usable in the event that an
	// AuthInfo, Context, or Cluster con***REMOVED***g with user-de***REMOVED***ned names are not found.
	// This provides a user with the immediate cause for error if one is found
	con***REMOVED***gAuthInfo, err := con***REMOVED***g.getAuthInfo()
	if err != nil {
		return nil, err
	}

	_, err = con***REMOVED***g.getContext()
	if err != nil {
		return nil, err
	}

	con***REMOVED***gClusterInfo, err := con***REMOVED***g.getCluster()
	if err != nil {
		return nil, err
	}

	if err := con***REMOVED***g.Con***REMOVED***rmUsable(); err != nil {
		return nil, err
	}

	clientCon***REMOVED***g := &restclient.Con***REMOVED***g{}
	clientCon***REMOVED***g.Host = con***REMOVED***gClusterInfo.Server

	if len(con***REMOVED***g.overrides.Timeout) > 0 {
		timeout, err := ParseTimeout(con***REMOVED***g.overrides.Timeout)
		if err != nil {
			return nil, err
		}
		clientCon***REMOVED***g.Timeout = timeout
	}

	if u, err := url.ParseRequestURI(clientCon***REMOVED***g.Host); err == nil && u.Opaque == "" && len(u.Path) > 1 {
		u.RawQuery = ""
		u.Fragment = ""
		clientCon***REMOVED***g.Host = u.String()
	}
	if len(con***REMOVED***gAuthInfo.Impersonate) > 0 {
		clientCon***REMOVED***g.Impersonate = restclient.ImpersonationCon***REMOVED***g{
			UserName: con***REMOVED***gAuthInfo.Impersonate,
			Groups:   con***REMOVED***gAuthInfo.ImpersonateGroups,
			Extra:    con***REMOVED***gAuthInfo.ImpersonateUserExtra,
		}
	}

	// only try to read the auth information if we are secure
	if restclient.IsCon***REMOVED***gTransportTLS(*clientCon***REMOVED***g) {
		var err error
		var persister restclient.AuthProviderCon***REMOVED***gPersister
		if con***REMOVED***g.con***REMOVED***gAccess != nil {
			authInfoName, _ := con***REMOVED***g.getAuthInfoName()
			persister = PersisterForUser(con***REMOVED***g.con***REMOVED***gAccess, authInfoName)
		}
		userAuthPartialCon***REMOVED***g, err := con***REMOVED***g.getUserIdenti***REMOVED***cationPartialCon***REMOVED***g(con***REMOVED***gAuthInfo, con***REMOVED***g.fallbackReader, persister)
		if err != nil {
			return nil, err
		}
		mergo.MergeWithOverwrite(clientCon***REMOVED***g, userAuthPartialCon***REMOVED***g)

		serverAuthPartialCon***REMOVED***g, err := getServerIdenti***REMOVED***cationPartialCon***REMOVED***g(con***REMOVED***gAuthInfo, con***REMOVED***gClusterInfo)
		if err != nil {
			return nil, err
		}
		mergo.MergeWithOverwrite(clientCon***REMOVED***g, serverAuthPartialCon***REMOVED***g)
	}

	return clientCon***REMOVED***g, nil
}

// clientauth.Info object contain both user identi***REMOVED***cation and server identi***REMOVED***cation.  We want different precedence orders for
// both, so we have to split the objects and merge them separately
// we want this order of precedence for the server identi***REMOVED***cation
// 1.  con***REMOVED***gClusterInfo (the ***REMOVED***nal result of command line flags and merged .kubecon***REMOVED***g ***REMOVED***les)
// 2.  con***REMOVED***gAuthInfo.auth-path (this ***REMOVED***le can contain information that conflicts with #1, and we want #1 to win the priority)
// 3.  load the ~/.kubernetes_auth ***REMOVED***le as a default
func getServerIdenti***REMOVED***cationPartialCon***REMOVED***g(con***REMOVED***gAuthInfo clientcmdapi.AuthInfo, con***REMOVED***gClusterInfo clientcmdapi.Cluster) (*restclient.Con***REMOVED***g, error) {
	mergedCon***REMOVED***g := &restclient.Con***REMOVED***g{}

	// con***REMOVED***gClusterInfo holds the information identify the server provided by .kubecon***REMOVED***g
	con***REMOVED***gClientCon***REMOVED***g := &restclient.Con***REMOVED***g{}
	con***REMOVED***gClientCon***REMOVED***g.CAFile = con***REMOVED***gClusterInfo.Certi***REMOVED***cateAuthority
	con***REMOVED***gClientCon***REMOVED***g.CAData = con***REMOVED***gClusterInfo.Certi***REMOVED***cateAuthorityData
	con***REMOVED***gClientCon***REMOVED***g.Insecure = con***REMOVED***gClusterInfo.InsecureSkipTLSVerify
	mergo.MergeWithOverwrite(mergedCon***REMOVED***g, con***REMOVED***gClientCon***REMOVED***g)

	return mergedCon***REMOVED***g, nil
}

// clientauth.Info object contain both user identi***REMOVED***cation and server identi***REMOVED***cation.  We want different precedence orders for
// both, so we have to split the objects and merge them separately
// we want this order of precedence for user identi***REMOVED***cation
// 1.  con***REMOVED***gAuthInfo minus auth-path (the ***REMOVED***nal result of command line flags and merged .kubecon***REMOVED***g ***REMOVED***les)
// 2.  con***REMOVED***gAuthInfo.auth-path (this ***REMOVED***le can contain information that conflicts with #1, and we want #1 to win the priority)
// 3.  if there is not enough information to identify the user, load try the ~/.kubernetes_auth ***REMOVED***le
// 4.  if there is not enough information to identify the user, prompt if possible
func (con***REMOVED***g *DirectClientCon***REMOVED***g) getUserIdenti***REMOVED***cationPartialCon***REMOVED***g(con***REMOVED***gAuthInfo clientcmdapi.AuthInfo, fallbackReader io.Reader, persistAuthCon***REMOVED***g restclient.AuthProviderCon***REMOVED***gPersister) (*restclient.Con***REMOVED***g, error) {
	mergedCon***REMOVED***g := &restclient.Con***REMOVED***g{}

	// blindly overwrite existing values based on precedence
	if len(con***REMOVED***gAuthInfo.Token) > 0 {
		mergedCon***REMOVED***g.BearerToken = con***REMOVED***gAuthInfo.Token
	} ***REMOVED*** if len(con***REMOVED***gAuthInfo.TokenFile) > 0 {
		tokenBytes, err := ioutil.ReadFile(con***REMOVED***gAuthInfo.TokenFile)
		if err != nil {
			return nil, err
		}
		mergedCon***REMOVED***g.BearerToken = string(tokenBytes)
		mergedCon***REMOVED***g.BearerTokenFile = con***REMOVED***gAuthInfo.TokenFile
	}
	if len(con***REMOVED***gAuthInfo.Impersonate) > 0 {
		mergedCon***REMOVED***g.Impersonate = restclient.ImpersonationCon***REMOVED***g{
			UserName: con***REMOVED***gAuthInfo.Impersonate,
			Groups:   con***REMOVED***gAuthInfo.ImpersonateGroups,
			Extra:    con***REMOVED***gAuthInfo.ImpersonateUserExtra,
		}
	}
	if len(con***REMOVED***gAuthInfo.ClientCerti***REMOVED***cate) > 0 || len(con***REMOVED***gAuthInfo.ClientCerti***REMOVED***cateData) > 0 {
		mergedCon***REMOVED***g.CertFile = con***REMOVED***gAuthInfo.ClientCerti***REMOVED***cate
		mergedCon***REMOVED***g.CertData = con***REMOVED***gAuthInfo.ClientCerti***REMOVED***cateData
		mergedCon***REMOVED***g.KeyFile = con***REMOVED***gAuthInfo.ClientKey
		mergedCon***REMOVED***g.KeyData = con***REMOVED***gAuthInfo.ClientKeyData
	}
	if len(con***REMOVED***gAuthInfo.Username) > 0 || len(con***REMOVED***gAuthInfo.Password) > 0 {
		mergedCon***REMOVED***g.Username = con***REMOVED***gAuthInfo.Username
		mergedCon***REMOVED***g.Password = con***REMOVED***gAuthInfo.Password
	}
	if con***REMOVED***gAuthInfo.AuthProvider != nil {
		mergedCon***REMOVED***g.AuthProvider = con***REMOVED***gAuthInfo.AuthProvider
		mergedCon***REMOVED***g.AuthCon***REMOVED***gPersister = persistAuthCon***REMOVED***g
	}
	if con***REMOVED***gAuthInfo.Exec != nil {
		mergedCon***REMOVED***g.ExecProvider = con***REMOVED***gAuthInfo.Exec
	}

	// if there still isn't enough information to authenticate the user, try prompting
	if !canIdentifyUser(*mergedCon***REMOVED***g) && (fallbackReader != nil) {
		if len(con***REMOVED***g.promptedCredentials.username) > 0 && len(con***REMOVED***g.promptedCredentials.password) > 0 {
			mergedCon***REMOVED***g.Username = con***REMOVED***g.promptedCredentials.username
			mergedCon***REMOVED***g.Password = con***REMOVED***g.promptedCredentials.password
			return mergedCon***REMOVED***g, nil
		}
		prompter := NewPromptingAuthLoader(fallbackReader)
		promptedAuthInfo, err := prompter.Prompt()
		if err != nil {
			return nil, err
		}
		promptedCon***REMOVED***g := makeUserIdenti***REMOVED***cationCon***REMOVED***g(*promptedAuthInfo)
		previouslyMergedCon***REMOVED***g := mergedCon***REMOVED***g
		mergedCon***REMOVED***g = &restclient.Con***REMOVED***g{}
		mergo.MergeWithOverwrite(mergedCon***REMOVED***g, promptedCon***REMOVED***g)
		mergo.MergeWithOverwrite(mergedCon***REMOVED***g, previouslyMergedCon***REMOVED***g)
		con***REMOVED***g.promptedCredentials.username = mergedCon***REMOVED***g.Username
		con***REMOVED***g.promptedCredentials.password = mergedCon***REMOVED***g.Password
	}

	return mergedCon***REMOVED***g, nil
}

// makeUserIdenti***REMOVED***cationFieldsCon***REMOVED***g returns a client.Con***REMOVED***g capable of being merged using mergo for only user identi***REMOVED***cation information
func makeUserIdenti***REMOVED***cationCon***REMOVED***g(info clientauth.Info) *restclient.Con***REMOVED***g {
	con***REMOVED***g := &restclient.Con***REMOVED***g{}
	con***REMOVED***g.Username = info.User
	con***REMOVED***g.Password = info.Password
	con***REMOVED***g.CertFile = info.CertFile
	con***REMOVED***g.KeyFile = info.KeyFile
	con***REMOVED***g.BearerToken = info.BearerToken
	return con***REMOVED***g
}

// makeUserIdenti***REMOVED***cationFieldsCon***REMOVED***g returns a client.Con***REMOVED***g capable of being merged using mergo for only server identi***REMOVED***cation information
func makeServerIdenti***REMOVED***cationCon***REMOVED***g(info clientauth.Info) restclient.Con***REMOVED***g {
	con***REMOVED***g := restclient.Con***REMOVED***g{}
	con***REMOVED***g.CAFile = info.CAFile
	if info.Insecure != nil {
		con***REMOVED***g.Insecure = *info.Insecure
	}
	return con***REMOVED***g
}

func canIdentifyUser(con***REMOVED***g restclient.Con***REMOVED***g) bool {
	return len(con***REMOVED***g.Username) > 0 ||
		(len(con***REMOVED***g.CertFile) > 0 || len(con***REMOVED***g.CertData) > 0) ||
		len(con***REMOVED***g.BearerToken) > 0 ||
		con***REMOVED***g.AuthProvider != nil ||
		con***REMOVED***g.ExecProvider != nil
}

// Namespace implements ClientCon***REMOVED***g
func (con***REMOVED***g *DirectClientCon***REMOVED***g) Namespace() (string, bool, error) {
	if con***REMOVED***g.overrides != nil && con***REMOVED***g.overrides.Context.Namespace != "" {
		// In the event we have an empty con***REMOVED***g but we do have a namespace override, we should return
		// the namespace override instead of having con***REMOVED***g.Con***REMOVED***rmUsable() return an error. This allows
		// things like in-cluster clients to execute `kubectl get pods --namespace=foo` and have the
		// --namespace flag honored instead of being ignored.
		return con***REMOVED***g.overrides.Context.Namespace, true, nil
	}

	if err := con***REMOVED***g.Con***REMOVED***rmUsable(); err != nil {
		return "", false, err
	}

	con***REMOVED***gContext, err := con***REMOVED***g.getContext()
	if err != nil {
		return "", false, err
	}

	if len(con***REMOVED***gContext.Namespace) == 0 {
		return "default", false, nil
	}

	return con***REMOVED***gContext.Namespace, false, nil
}

// Con***REMOVED***gAccess implements ClientCon***REMOVED***g
func (con***REMOVED***g *DirectClientCon***REMOVED***g) Con***REMOVED***gAccess() Con***REMOVED***gAccess {
	return con***REMOVED***g.con***REMOVED***gAccess
}

// Con***REMOVED***rmUsable looks a particular context and determines if that particular part of the con***REMOVED***g is useable.  There might still be errors in the con***REMOVED***g,
// but no errors in the sections requested or referenced.  It does not return early so that it can ***REMOVED***nd as many errors as possible.
func (con***REMOVED***g *DirectClientCon***REMOVED***g) Con***REMOVED***rmUsable() error {
	validationErrors := make([]error, 0)

	var contextName string
	if len(con***REMOVED***g.contextName) != 0 {
		contextName = con***REMOVED***g.contextName
	} ***REMOVED*** {
		contextName = con***REMOVED***g.con***REMOVED***g.CurrentContext
	}

	if len(contextName) > 0 {
		_, exists := con***REMOVED***g.con***REMOVED***g.Contexts[contextName]
		if !exists {
			validationErrors = append(validationErrors, &errContextNotFound{contextName})
		}
	}

	authInfoName, _ := con***REMOVED***g.getAuthInfoName()
	authInfo, _ := con***REMOVED***g.getAuthInfo()
	validationErrors = append(validationErrors, validateAuthInfo(authInfoName, authInfo)...)
	clusterName, _ := con***REMOVED***g.getClusterName()
	cluster, _ := con***REMOVED***g.getCluster()
	validationErrors = append(validationErrors, validateClusterInfo(clusterName, cluster)...)
	// when direct client con***REMOVED***g is speci***REMOVED***ed, and our only error is that no server is de***REMOVED***ned, we should
	// return a standard "no con***REMOVED***g" error
	if len(validationErrors) == 1 && validationErrors[0] == ErrEmptyCluster {
		return newErrCon***REMOVED***gurationInvalid([]error{ErrEmptyCon***REMOVED***g})
	}
	return newErrCon***REMOVED***gurationInvalid(validationErrors)
}

// getContextName returns the default, or user-set context name, and a boolean that indicates
// whether the default context name has been overwritten by a user-set flag, or left as its default value
func (con***REMOVED***g *DirectClientCon***REMOVED***g) getContextName() (string, bool) {
	if len(con***REMOVED***g.overrides.CurrentContext) != 0 {
		return con***REMOVED***g.overrides.CurrentContext, true
	}
	if len(con***REMOVED***g.contextName) != 0 {
		return con***REMOVED***g.contextName, false
	}

	return con***REMOVED***g.con***REMOVED***g.CurrentContext, false
}

// getAuthInfoName returns a string containing the current authinfo name for the current context,
// and a boolean indicating  whether the default authInfo name is overwritten by a user-set flag, or
// left as its default value
func (con***REMOVED***g *DirectClientCon***REMOVED***g) getAuthInfoName() (string, bool) {
	if len(con***REMOVED***g.overrides.Context.AuthInfo) != 0 {
		return con***REMOVED***g.overrides.Context.AuthInfo, true
	}
	context, _ := con***REMOVED***g.getContext()
	return context.AuthInfo, false
}

// getClusterName returns a string containing the default, or user-set cluster name, and a boolean
// indicating whether the default clusterName has been overwritten by a user-set flag, or left as
// its default value
func (con***REMOVED***g *DirectClientCon***REMOVED***g) getClusterName() (string, bool) {
	if len(con***REMOVED***g.overrides.Context.Cluster) != 0 {
		return con***REMOVED***g.overrides.Context.Cluster, true
	}
	context, _ := con***REMOVED***g.getContext()
	return context.Cluster, false
}

// getContext returns the clientcmdapi.Context, or an error if a required context is not found.
func (con***REMOVED***g *DirectClientCon***REMOVED***g) getContext() (clientcmdapi.Context, error) {
	contexts := con***REMOVED***g.con***REMOVED***g.Contexts
	contextName, required := con***REMOVED***g.getContextName()

	mergedContext := clientcmdapi.NewContext()
	if con***REMOVED***gContext, exists := contexts[contextName]; exists {
		mergo.MergeWithOverwrite(mergedContext, con***REMOVED***gContext)
	} ***REMOVED*** if required {
		return clientcmdapi.Context{}, fmt.Errorf("context %q does not exist", contextName)
	}
	mergo.MergeWithOverwrite(mergedContext, con***REMOVED***g.overrides.Context)

	return *mergedContext, nil
}

// getAuthInfo returns the clientcmdapi.AuthInfo, or an error if a required auth info is not found.
func (con***REMOVED***g *DirectClientCon***REMOVED***g) getAuthInfo() (clientcmdapi.AuthInfo, error) {
	authInfos := con***REMOVED***g.con***REMOVED***g.AuthInfos
	authInfoName, required := con***REMOVED***g.getAuthInfoName()

	mergedAuthInfo := clientcmdapi.NewAuthInfo()
	if con***REMOVED***gAuthInfo, exists := authInfos[authInfoName]; exists {
		mergo.MergeWithOverwrite(mergedAuthInfo, con***REMOVED***gAuthInfo)
	} ***REMOVED*** if required {
		return clientcmdapi.AuthInfo{}, fmt.Errorf("auth info %q does not exist", authInfoName)
	}
	mergo.MergeWithOverwrite(mergedAuthInfo, con***REMOVED***g.overrides.AuthInfo)

	return *mergedAuthInfo, nil
}

// getCluster returns the clientcmdapi.Cluster, or an error if a required cluster is not found.
func (con***REMOVED***g *DirectClientCon***REMOVED***g) getCluster() (clientcmdapi.Cluster, error) {
	clusterInfos := con***REMOVED***g.con***REMOVED***g.Clusters
	clusterInfoName, required := con***REMOVED***g.getClusterName()

	mergedClusterInfo := clientcmdapi.NewCluster()
	mergo.MergeWithOverwrite(mergedClusterInfo, con***REMOVED***g.overrides.ClusterDefaults)
	if con***REMOVED***gClusterInfo, exists := clusterInfos[clusterInfoName]; exists {
		mergo.MergeWithOverwrite(mergedClusterInfo, con***REMOVED***gClusterInfo)
	} ***REMOVED*** if required {
		return clientcmdapi.Cluster{}, fmt.Errorf("cluster %q does not exist", clusterInfoName)
	}
	mergo.MergeWithOverwrite(mergedClusterInfo, con***REMOVED***g.overrides.ClusterInfo)
	// An override of --insecure-skip-tls-verify=true and no accompanying CA/CA data should clear already-set CA/CA data
	// otherwise, a kubecon***REMOVED***g containing a CA reference would return an error that "CA and insecure-skip-tls-verify couldn't both be set"
	caLen := len(con***REMOVED***g.overrides.ClusterInfo.Certi***REMOVED***cateAuthority)
	caDataLen := len(con***REMOVED***g.overrides.ClusterInfo.Certi***REMOVED***cateAuthorityData)
	if con***REMOVED***g.overrides.ClusterInfo.InsecureSkipTLSVerify && caLen == 0 && caDataLen == 0 {
		mergedClusterInfo.Certi***REMOVED***cateAuthority = ""
		mergedClusterInfo.Certi***REMOVED***cateAuthorityData = nil
	}

	return *mergedClusterInfo, nil
}

// inClusterClientCon***REMOVED***g makes a con***REMOVED***g that will work from within a kubernetes cluster container environment.
// Can take options overrides for flags explicitly provided to the command inside the cluster container.
type inClusterClientCon***REMOVED***g struct {
	overrides               *Con***REMOVED***gOverrides
	inClusterCon***REMOVED***gProvider func() (*restclient.Con***REMOVED***g, error)
}

var _ ClientCon***REMOVED***g = &inClusterClientCon***REMOVED***g{}

func (con***REMOVED***g *inClusterClientCon***REMOVED***g) RawCon***REMOVED***g() (clientcmdapi.Con***REMOVED***g, error) {
	return clientcmdapi.Con***REMOVED***g{}, fmt.Errorf("inCluster environment con***REMOVED***g doesn't support multiple clusters")
}

func (con***REMOVED***g *inClusterClientCon***REMOVED***g) ClientCon***REMOVED***g() (*restclient.Con***REMOVED***g, error) {
	if con***REMOVED***g.inClusterCon***REMOVED***gProvider == nil {
		con***REMOVED***g.inClusterCon***REMOVED***gProvider = restclient.InClusterCon***REMOVED***g
	}

	icc, err := con***REMOVED***g.inClusterCon***REMOVED***gProvider()
	if err != nil {
		return nil, err
	}

	// in-cluster con***REMOVED***gs only takes a host, token, or CA ***REMOVED***le
	// if any of them were individually provided, overwrite anything ***REMOVED***
	if con***REMOVED***g.overrides != nil {
		if server := con***REMOVED***g.overrides.ClusterInfo.Server; len(server) > 0 {
			icc.Host = server
		}
		if token := con***REMOVED***g.overrides.AuthInfo.Token; len(token) > 0 {
			icc.BearerToken = token
		}
		if certi***REMOVED***cateAuthorityFile := con***REMOVED***g.overrides.ClusterInfo.Certi***REMOVED***cateAuthority; len(certi***REMOVED***cateAuthorityFile) > 0 {
			icc.TLSClientCon***REMOVED***g.CAFile = certi***REMOVED***cateAuthorityFile
		}
	}

	return icc, err
}

func (con***REMOVED***g *inClusterClientCon***REMOVED***g) Namespace() (string, bool, error) {
	// This way assumes you've set the POD_NAMESPACE environment variable using the downward API.
	// This check has to be done ***REMOVED***rst for backwards compatibility with the way InClusterCon***REMOVED***g was originally set up
	if ns := os.Getenv("POD_NAMESPACE"); ns != "" {
		return ns, false, nil
	}

	// Fall back to the namespace associated with the service account token, if available
	if data, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
		if ns := strings.TrimSpace(string(data)); len(ns) > 0 {
			return ns, false, nil
		}
	}

	return "default", false, nil
}

func (con***REMOVED***g *inClusterClientCon***REMOVED***g) Con***REMOVED***gAccess() Con***REMOVED***gAccess {
	return NewDefaultClientCon***REMOVED***gLoadingRules()
}

// Possible returns true if loading an inside-kubernetes-cluster is possible.
func (con***REMOVED***g *inClusterClientCon***REMOVED***g) Possible() bool {
	***REMOVED***, err := os.Stat("/var/run/secrets/kubernetes.io/serviceaccount/token")
	return os.Getenv("KUBERNETES_SERVICE_HOST") != "" &&
		os.Getenv("KUBERNETES_SERVICE_PORT") != "" &&
		err == nil && !***REMOVED***.IsDir()
}

// BuildCon***REMOVED***gFromFlags is a helper function that builds con***REMOVED***gs from a master
// url or a kubecon***REMOVED***g ***REMOVED***lepath. These are passed in as command line flags for cluster
// components. Warnings should reflect this usage. If neither masterUrl or kubecon***REMOVED***gPath
// are passed in we fallback to inClusterCon***REMOVED***g. If inClusterCon***REMOVED***g fails, we fallback
// to the default con***REMOVED***g.
func BuildCon***REMOVED***gFromFlags(masterUrl, kubecon***REMOVED***gPath string) (*restclient.Con***REMOVED***g, error) {
	if kubecon***REMOVED***gPath == "" && masterUrl == "" {
		klog.Warningf("Neither --kubecon***REMOVED***g nor --master was speci***REMOVED***ed.  Using the inClusterCon***REMOVED***g.  This might not work.")
		kubecon***REMOVED***g, err := restclient.InClusterCon***REMOVED***g()
		if err == nil {
			return kubecon***REMOVED***g, nil
		}
		klog.Warning("error creating inClusterCon***REMOVED***g, falling back to default con***REMOVED***g: ", err)
	}
	return NewNonInteractiveDeferredLoadingClientCon***REMOVED***g(
		&ClientCon***REMOVED***gLoadingRules{ExplicitPath: kubecon***REMOVED***gPath},
		&Con***REMOVED***gOverrides{ClusterInfo: clientcmdapi.Cluster{Server: masterUrl}}).ClientCon***REMOVED***g()
}

// BuildCon***REMOVED***gFromKubecon***REMOVED***gGetter is a helper function that builds con***REMOVED***gs from a master
// url and a kubecon***REMOVED***gGetter.
func BuildCon***REMOVED***gFromKubecon***REMOVED***gGetter(masterUrl string, kubecon***REMOVED***gGetter Kubecon***REMOVED***gGetter) (*restclient.Con***REMOVED***g, error) {
	// TODO: We do not need a DeferredLoader here. Refactor code and see if we can use DirectClientCon***REMOVED***g here.
	cc := NewNonInteractiveDeferredLoadingClientCon***REMOVED***g(
		&ClientCon***REMOVED***gGetter{kubecon***REMOVED***gGetter: kubecon***REMOVED***gGetter},
		&Con***REMOVED***gOverrides{ClusterInfo: clientcmdapi.Cluster{Server: masterUrl}})
	return cc.ClientCon***REMOVED***g()
}
