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
	"io"
	"sync"

	"k8s.io/klog"

	restclient "k8s.io/client-go/rest"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// DeferredLoadingClientCon***REMOVED***g is a ClientCon***REMOVED***g interface that is backed by a client con***REMOVED***g loader.
// It is used in cases where the loading rules may change after you've instantiated them and you want to be sure that
// the most recent rules are used.  This is useful in cases where you bind flags to loading rule parameters before
// the parse happens and you want your calling code to be ignorant of how the values are being mutated to avoid
// passing extraneous information down a call stack
type DeferredLoadingClientCon***REMOVED***g struct {
	loader         ClientCon***REMOVED***gLoader
	overrides      *Con***REMOVED***gOverrides
	fallbackReader io.Reader

	clientCon***REMOVED***g ClientCon***REMOVED***g
	loadingLock  sync.Mutex

	// provided for testing
	icc InClusterCon***REMOVED***g
}

// InClusterCon***REMOVED***g abstracts details of whether the client is running in a cluster for testing.
type InClusterCon***REMOVED***g interface {
	ClientCon***REMOVED***g
	Possible() bool
}

// NewNonInteractiveDeferredLoadingClientCon***REMOVED***g creates a Con***REMOVED***gClientClientCon***REMOVED***g using the passed context name
func NewNonInteractiveDeferredLoadingClientCon***REMOVED***g(loader ClientCon***REMOVED***gLoader, overrides *Con***REMOVED***gOverrides) ClientCon***REMOVED***g {
	return &DeferredLoadingClientCon***REMOVED***g{loader: loader, overrides: overrides, icc: &inClusterClientCon***REMOVED***g{overrides: overrides}}
}

// NewInteractiveDeferredLoadingClientCon***REMOVED***g creates a Con***REMOVED***gClientClientCon***REMOVED***g using the passed context name and the fallback auth reader
func NewInteractiveDeferredLoadingClientCon***REMOVED***g(loader ClientCon***REMOVED***gLoader, overrides *Con***REMOVED***gOverrides, fallbackReader io.Reader) ClientCon***REMOVED***g {
	return &DeferredLoadingClientCon***REMOVED***g{loader: loader, overrides: overrides, icc: &inClusterClientCon***REMOVED***g{overrides: overrides}, fallbackReader: fallbackReader}
}

func (con***REMOVED***g *DeferredLoadingClientCon***REMOVED***g) createClientCon***REMOVED***g() (ClientCon***REMOVED***g, error) {
	if con***REMOVED***g.clientCon***REMOVED***g == nil {
		con***REMOVED***g.loadingLock.Lock()
		defer con***REMOVED***g.loadingLock.Unlock()

		if con***REMOVED***g.clientCon***REMOVED***g == nil {
			mergedCon***REMOVED***g, err := con***REMOVED***g.loader.Load()
			if err != nil {
				return nil, err
			}

			var mergedClientCon***REMOVED***g ClientCon***REMOVED***g
			if con***REMOVED***g.fallbackReader != nil {
				mergedClientCon***REMOVED***g = NewInteractiveClientCon***REMOVED***g(*mergedCon***REMOVED***g, con***REMOVED***g.overrides.CurrentContext, con***REMOVED***g.overrides, con***REMOVED***g.fallbackReader, con***REMOVED***g.loader)
			} ***REMOVED*** {
				mergedClientCon***REMOVED***g = NewNonInteractiveClientCon***REMOVED***g(*mergedCon***REMOVED***g, con***REMOVED***g.overrides.CurrentContext, con***REMOVED***g.overrides, con***REMOVED***g.loader)
			}

			con***REMOVED***g.clientCon***REMOVED***g = mergedClientCon***REMOVED***g
		}
	}

	return con***REMOVED***g.clientCon***REMOVED***g, nil
}

func (con***REMOVED***g *DeferredLoadingClientCon***REMOVED***g) RawCon***REMOVED***g() (clientcmdapi.Con***REMOVED***g, error) {
	mergedCon***REMOVED***g, err := con***REMOVED***g.createClientCon***REMOVED***g()
	if err != nil {
		return clientcmdapi.Con***REMOVED***g{}, err
	}

	return mergedCon***REMOVED***g.RawCon***REMOVED***g()
}

// ClientCon***REMOVED***g implements ClientCon***REMOVED***g
func (con***REMOVED***g *DeferredLoadingClientCon***REMOVED***g) ClientCon***REMOVED***g() (*restclient.Con***REMOVED***g, error) {
	mergedClientCon***REMOVED***g, err := con***REMOVED***g.createClientCon***REMOVED***g()
	if err != nil {
		return nil, err
	}

	// load the con***REMOVED***guration and return on non-empty errors and if the
	// content differs from the default con***REMOVED***g
	mergedCon***REMOVED***g, err := mergedClientCon***REMOVED***g.ClientCon***REMOVED***g()
	switch {
	case err != nil:
		if !IsEmptyCon***REMOVED***g(err) {
			// return on any error except empty con***REMOVED***g
			return nil, err
		}
	case mergedCon***REMOVED***g != nil:
		// the con***REMOVED***guration is valid, but if this is equal to the defaults we should try
		// in-cluster con***REMOVED***guration
		if !con***REMOVED***g.loader.IsDefaultCon***REMOVED***g(mergedCon***REMOVED***g) {
			return mergedCon***REMOVED***g, nil
		}
	}

	// check for in-cluster con***REMOVED***guration and use it
	if con***REMOVED***g.icc.Possible() {
		klog.V(4).Infof("Using in-cluster con***REMOVED***guration")
		return con***REMOVED***g.icc.ClientCon***REMOVED***g()
	}

	// return the result of the merged client con***REMOVED***g
	return mergedCon***REMOVED***g, err
}

// Namespace implements KubeCon***REMOVED***g
func (con***REMOVED***g *DeferredLoadingClientCon***REMOVED***g) Namespace() (string, bool, error) {
	mergedKubeCon***REMOVED***g, err := con***REMOVED***g.createClientCon***REMOVED***g()
	if err != nil {
		return "", false, err
	}

	ns, overridden, err := mergedKubeCon***REMOVED***g.Namespace()
	// if we get an error and it is not empty con***REMOVED***g, or if the merged con***REMOVED***g de***REMOVED***ned an explicit namespace, or
	// if in-cluster con***REMOVED***g is not possible, return immediately
	if (err != nil && !IsEmptyCon***REMOVED***g(err)) || overridden || !con***REMOVED***g.icc.Possible() {
		// return on any error except empty con***REMOVED***g
		return ns, overridden, err
	}

	if len(ns) > 0 {
		// if we got a non-default namespace from the kubecon***REMOVED***g, use it
		if ns != "default" {
			return ns, false, nil
		}

		// if we got a default namespace, determine whether it was explicit or implicit
		if raw, err := mergedKubeCon***REMOVED***g.RawCon***REMOVED***g(); err == nil {
			if context := raw.Contexts[raw.CurrentContext]; context != nil && len(context.Namespace) > 0 {
				return ns, false, nil
			}
		}
	}

	klog.V(4).Infof("Using in-cluster namespace")

	// allow the namespace from the service account token directory to be used.
	return con***REMOVED***g.icc.Namespace()
}

// Con***REMOVED***gAccess implements ClientCon***REMOVED***g
func (con***REMOVED***g *DeferredLoadingClientCon***REMOVED***g) Con***REMOVED***gAccess() Con***REMOVED***gAccess {
	return con***REMOVED***g.loader
}
