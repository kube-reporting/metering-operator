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

package rest

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/golang/glog"

	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

type AuthProvider interface {
	// WrapTransport allows the plugin to create a modi***REMOVED***ed RoundTripper that
	// attaches authorization headers (or other info) to requests.
	WrapTransport(http.RoundTripper) http.RoundTripper
	// Login allows the plugin to initialize its con***REMOVED***guration. It must not
	// require direct user interaction.
	Login() error
}

// Factory generates an AuthProvider plugin.
//  clusterAddress is the address of the current cluster.
//  con***REMOVED***g is the initial con***REMOVED***guration for this plugin.
//  persister allows the plugin to save updated con***REMOVED***guration.
type Factory func(clusterAddress string, con***REMOVED***g map[string]string, persister AuthProviderCon***REMOVED***gPersister) (AuthProvider, error)

// AuthProviderCon***REMOVED***gPersister allows a plugin to persist con***REMOVED***guration info
// for just itself.
type AuthProviderCon***REMOVED***gPersister interface {
	Persist(map[string]string) error
}

// All registered auth provider plugins.
var pluginsLock sync.Mutex
var plugins = make(map[string]Factory)

func RegisterAuthProviderPlugin(name string, plugin Factory) error {
	pluginsLock.Lock()
	defer pluginsLock.Unlock()
	if _, found := plugins[name]; found {
		return fmt.Errorf("Auth Provider Plugin %q was registered twice", name)
	}
	glog.V(4).Infof("Registered Auth Provider Plugin %q", name)
	plugins[name] = plugin
	return nil
}

func GetAuthProvider(clusterAddress string, apc *clientcmdapi.AuthProviderCon***REMOVED***g, persister AuthProviderCon***REMOVED***gPersister) (AuthProvider, error) {
	pluginsLock.Lock()
	defer pluginsLock.Unlock()
	p, ok := plugins[apc.Name]
	if !ok {
		return nil, fmt.Errorf("No Auth Provider found for name %q", apc.Name)
	}
	return p(clusterAddress, apc.Con***REMOVED***g, persister)
}
