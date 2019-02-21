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

package discovery

import (
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"path/***REMOVED***lepath"
	"sync"
	"time"

	"github.com/googleapis/gnostic/OpenAPIv2"
	"k8s.io/klog"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/kubernetes/scheme"
	restclient "k8s.io/client-go/rest"
)

// CachedDiscoveryClient implements the functions that discovery server-supported API groups,
// versions and resources.
type CachedDiscoveryClient struct {
	delegate DiscoveryInterface

	// cacheDirectory is the directory where discovery docs are held.  It must be unique per host:port combination to work well.
	cacheDirectory string

	// ttl is how long the cache should be considered valid
	ttl time.Duration

	// mutex protects the variables below
	mutex sync.Mutex

	// ourFiles are all ***REMOVED***lenames of cache ***REMOVED***les created by this process
	ourFiles map[string]struct{}
	// invalidated is true if all cache ***REMOVED***les should be ignored that are not ours (e.g. after Invalidate() was called)
	invalidated bool
	// fresh is true if all used cache ***REMOVED***les were ours
	fresh bool
}

var _ CachedDiscoveryInterface = &CachedDiscoveryClient{}

// ServerResourcesForGroupVersion returns the supported resources for a group and version.
func (d *CachedDiscoveryClient) ServerResourcesForGroupVersion(groupVersion string) (*metav1.APIResourceList, error) {
	***REMOVED***lename := ***REMOVED***lepath.Join(d.cacheDirectory, groupVersion, "serverresources.json")
	cachedBytes, err := d.getCachedFile(***REMOVED***lename)
	// don't fail on errors, we either don't have a ***REMOVED***le or won't be able to run the cached check. Either way we can fallback.
	if err == nil {
		cachedResources := &metav1.APIResourceList{}
		if err := runtime.DecodeInto(scheme.Codecs.UniversalDecoder(), cachedBytes, cachedResources); err == nil {
			klog.V(10).Infof("returning cached discovery info from %v", ***REMOVED***lename)
			return cachedResources, nil
		}
	}

	liveResources, err := d.delegate.ServerResourcesForGroupVersion(groupVersion)
	if err != nil {
		klog.V(3).Infof("skipped caching discovery info due to %v", err)
		return liveResources, err
	}
	if liveResources == nil || len(liveResources.APIResources) == 0 {
		klog.V(3).Infof("skipped caching discovery info, no resources found")
		return liveResources, err
	}

	if err := d.writeCachedFile(***REMOVED***lename, liveResources); err != nil {
		klog.V(1).Infof("failed to write cache to %v due to %v", ***REMOVED***lename, err)
	}

	return liveResources, nil
}

// ServerResources returns the supported resources for all groups and versions.
func (d *CachedDiscoveryClient) ServerResources() ([]*metav1.APIResourceList, error) {
	return ServerResources(d)
}

// ServerGroups returns the supported groups, with information like supported versions and the
// preferred version.
func (d *CachedDiscoveryClient) ServerGroups() (*metav1.APIGroupList, error) {
	***REMOVED***lename := ***REMOVED***lepath.Join(d.cacheDirectory, "servergroups.json")
	cachedBytes, err := d.getCachedFile(***REMOVED***lename)
	// don't fail on errors, we either don't have a ***REMOVED***le or won't be able to run the cached check. Either way we can fallback.
	if err == nil {
		cachedGroups := &metav1.APIGroupList{}
		if err := runtime.DecodeInto(scheme.Codecs.UniversalDecoder(), cachedBytes, cachedGroups); err == nil {
			klog.V(10).Infof("returning cached discovery info from %v", ***REMOVED***lename)
			return cachedGroups, nil
		}
	}

	liveGroups, err := d.delegate.ServerGroups()
	if err != nil {
		klog.V(3).Infof("skipped caching discovery info due to %v", err)
		return liveGroups, err
	}
	if liveGroups == nil || len(liveGroups.Groups) == 0 {
		klog.V(3).Infof("skipped caching discovery info, no groups found")
		return liveGroups, err
	}

	if err := d.writeCachedFile(***REMOVED***lename, liveGroups); err != nil {
		klog.V(1).Infof("failed to write cache to %v due to %v", ***REMOVED***lename, err)
	}

	return liveGroups, nil
}

func (d *CachedDiscoveryClient) getCachedFile(***REMOVED***lename string) ([]byte, error) {
	// after invalidation ignore cache ***REMOVED***les not created by this process
	d.mutex.Lock()
	_, ourFile := d.ourFiles[***REMOVED***lename]
	if d.invalidated && !ourFile {
		d.mutex.Unlock()
		return nil, errors.New("cache invalidated")
	}
	d.mutex.Unlock()

	***REMOVED***le, err := os.Open(***REMOVED***lename)
	if err != nil {
		return nil, err
	}
	defer ***REMOVED***le.Close()

	***REMOVED***leInfo, err := ***REMOVED***le.Stat()
	if err != nil {
		return nil, err
	}

	if time.Now().After(***REMOVED***leInfo.ModTime().Add(d.ttl)) {
		return nil, errors.New("cache expired")
	}

	// the cache is present and its valid.  Try to read and use it.
	cachedBytes, err := ioutil.ReadAll(***REMOVED***le)
	if err != nil {
		return nil, err
	}

	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.fresh = d.fresh && ourFile

	return cachedBytes, nil
}

func (d *CachedDiscoveryClient) writeCachedFile(***REMOVED***lename string, obj runtime.Object) error {
	if err := os.MkdirAll(***REMOVED***lepath.Dir(***REMOVED***lename), 0755); err != nil {
		return err
	}

	bytes, err := runtime.Encode(scheme.Codecs.LegacyCodec(), obj)
	if err != nil {
		return err
	}

	f, err := ioutil.TempFile(***REMOVED***lepath.Dir(***REMOVED***lename), ***REMOVED***lepath.Base(***REMOVED***lename)+".")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	_, err = f.Write(bytes)
	if err != nil {
		return err
	}

	err = os.Chmod(f.Name(), 0755)
	if err != nil {
		return err
	}

	name := f.Name()
	err = f.Close()
	if err != nil {
		return err
	}

	// atomic rename
	d.mutex.Lock()
	defer d.mutex.Unlock()
	err = os.Rename(name, ***REMOVED***lename)
	if err == nil {
		d.ourFiles[***REMOVED***lename] = struct{}{}
	}
	return err
}

// RESTClient returns a RESTClient that is used to communicate with API server
// by this client implementation.
func (d *CachedDiscoveryClient) RESTClient() restclient.Interface {
	return d.delegate.RESTClient()
}

// ServerPreferredResources returns the supported resources with the version preferred by the
// server.
func (d *CachedDiscoveryClient) ServerPreferredResources() ([]*metav1.APIResourceList, error) {
	return ServerPreferredResources(d)
}

// ServerPreferredNamespacedResources returns the supported namespaced resources with the
// version preferred by the server.
func (d *CachedDiscoveryClient) ServerPreferredNamespacedResources() ([]*metav1.APIResourceList, error) {
	return ServerPreferredNamespacedResources(d)
}

// ServerVersion retrieves and parses the server's version (git version).
func (d *CachedDiscoveryClient) ServerVersion() (*version.Info, error) {
	return d.delegate.ServerVersion()
}

// OpenAPISchema retrieves and parses the swagger API schema the server supports.
func (d *CachedDiscoveryClient) OpenAPISchema() (*openapi_v2.Document, error) {
	return d.delegate.OpenAPISchema()
}

// Fresh is supposed to tell the caller whether or not to retry if the cache
// fails to ***REMOVED***nd something (false = retry, true = no need to retry).
func (d *CachedDiscoveryClient) Fresh() bool {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	return d.fresh
}

// Invalidate enforces that no cached data is used in the future that is older than the current time.
func (d *CachedDiscoveryClient) Invalidate() {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.ourFiles = map[string]struct{}{}
	d.fresh = true
	d.invalidated = true
}

// NewCachedDiscoveryClientForCon***REMOVED***g creates a new DiscoveryClient for the given con***REMOVED***g, and wraps
// the created client in a CachedDiscoveryClient. The provided con***REMOVED***guration is updated with a
// custom transport that understands cache responses.
// We receive two distinct cache directories for now, in order to preserve old behavior
// which makes use of the --cache-dir flag value for storing cache data from the CacheRoundTripper,
// and makes use of the hardcoded destination (~/.kube/cache/discovery/...) for storing
// CachedDiscoveryClient cache data. If httpCacheDir is empty, the restcon***REMOVED***g's transport will not
// be updated with a roundtripper that understands cache responses.
// If discoveryCacheDir is empty, cached server resource data will be looked up in the current directory.
// TODO(juanvallejo): the value of "--cache-dir" should be honored. Consolidate discoveryCacheDir with httpCacheDir
// so that server resources and http-cache data are stored in the same location, provided via con***REMOVED***g flags.
func NewCachedDiscoveryClientForCon***REMOVED***g(con***REMOVED***g *restclient.Con***REMOVED***g, discoveryCacheDir, httpCacheDir string, ttl time.Duration) (*CachedDiscoveryClient, error) {
	if len(httpCacheDir) > 0 {
		// update the given restcon***REMOVED***g with a custom roundtripper that
		// understands how to handle cache responses.
		wt := con***REMOVED***g.WrapTransport
		con***REMOVED***g.WrapTransport = func(rt http.RoundTripper) http.RoundTripper {
			if wt != nil {
				rt = wt(rt)
			}
			return newCacheRoundTripper(httpCacheDir, rt)
		}
	}

	discoveryClient, err := NewDiscoveryClientForCon***REMOVED***g(con***REMOVED***g)
	if err != nil {
		return nil, err
	}

	return newCachedDiscoveryClient(discoveryClient, discoveryCacheDir, ttl), nil
}

// NewCachedDiscoveryClient creates a new DiscoveryClient.  cacheDirectory is the directory where discovery docs are held.  It must be unique per host:port combination to work well.
func newCachedDiscoveryClient(delegate DiscoveryInterface, cacheDirectory string, ttl time.Duration) *CachedDiscoveryClient {
	return &CachedDiscoveryClient{
		delegate:       delegate,
		cacheDirectory: cacheDirectory,
		ttl:            ttl,
		ourFiles:       map[string]struct{}{},
		fresh:          true,
	}
}
