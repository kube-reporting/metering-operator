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

package dynamic

import (
	"sync"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	restclient "k8s.io/client-go/rest"
)

// ClientPool manages a pool of dynamic clients.
type ClientPool interface {
	// ClientForGroupVersionKind returns a client con***REMOVED***gured for the speci***REMOVED***ed groupVersionResource.
	// Resource may be empty.
	ClientForGroupVersionResource(resource schema.GroupVersionResource) (*Client, error)
	// ClientForGroupVersionKind returns a client con***REMOVED***gured for the speci***REMOVED***ed groupVersionKind.
	// Kind may be empty.
	ClientForGroupVersionKind(kind schema.GroupVersionKind) (*Client, error)
}

// APIPathResolverFunc knows how to convert a groupVersion to its API path. The Kind ***REMOVED***eld is
// optional.
type APIPathResolverFunc func(kind schema.GroupVersionKind) string

// LegacyAPIPathResolverFunc can resolve paths properly with the legacy API.
func LegacyAPIPathResolverFunc(kind schema.GroupVersionKind) string {
	if len(kind.Group) == 0 {
		return "/api"
	}
	return "/apis"
}

// clientPoolImpl implements ClientPool and caches clients for the resource group versions
// is asked to retrieve. This type is thread safe.
type clientPoolImpl struct {
	lock                sync.RWMutex
	con***REMOVED***g              *restclient.Con***REMOVED***g
	clients             map[schema.GroupVersion]*Client
	apiPathResolverFunc APIPathResolverFunc
	mapper              meta.RESTMapper
}

// NewClientPool returns a ClientPool from the speci***REMOVED***ed con***REMOVED***g. It reuses clients for the the same
// group version. It is expected this type may be wrapped by speci***REMOVED***c logic that special cases certain
// resources or groups.
func NewClientPool(con***REMOVED***g *restclient.Con***REMOVED***g, mapper meta.RESTMapper, apiPathResolverFunc APIPathResolverFunc) ClientPool {
	confCopy := *con***REMOVED***g

	return &clientPoolImpl{
		con***REMOVED***g:              &confCopy,
		clients:             map[schema.GroupVersion]*Client{},
		apiPathResolverFunc: apiPathResolverFunc,
		mapper:              mapper,
	}
}

// Instantiates a new dynamic client pool with the given con***REMOVED***g.
func NewDynamicClientPool(cfg *restclient.Con***REMOVED***g) ClientPool {
	// restMapper is not needed when using LegacyAPIPathResolverFunc
	emptyMapper := meta.MultiRESTMapper{}
	return NewClientPool(cfg, emptyMapper, LegacyAPIPathResolverFunc)
}

// ClientForGroupVersionResource uses the provided RESTMapper to identify the appropriate resource. Resource may
// be empty. If no matching kind is found the underlying client for that group is still returned.
func (c *clientPoolImpl) ClientForGroupVersionResource(resource schema.GroupVersionResource) (*Client, error) {
	kinds, err := c.mapper.KindsFor(resource)
	if err != nil {
		if meta.IsNoMatchError(err) {
			return c.ClientForGroupVersionKind(schema.GroupVersionKind{Group: resource.Group, Version: resource.Version})
		}
		return nil, err
	}
	return c.ClientForGroupVersionKind(kinds[0])
}

// ClientForGroupVersion returns a client for the speci***REMOVED***ed groupVersion, creates one if none exists. Kind
// in the GroupVersionKind may be empty.
func (c *clientPoolImpl) ClientForGroupVersionKind(kind schema.GroupVersionKind) (*Client, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	gv := kind.GroupVersion()

	// do we have a client already con***REMOVED***gured?
	if existingClient, found := c.clients[gv]; found {
		return existingClient, nil
	}

	// avoid changing the original con***REMOVED***g
	confCopy := *c.con***REMOVED***g
	conf := &confCopy

	// we need to set the api path based on group version, if no group, default to legacy path
	conf.APIPath = c.apiPathResolverFunc(kind)

	// we need to make a client
	conf.GroupVersion = &gv

	dynamicClient, err := NewClient(conf)
	if err != nil {
		return nil, err
	}
	c.clients[gv] = dynamicClient
	return dynamicClient, nil
}
