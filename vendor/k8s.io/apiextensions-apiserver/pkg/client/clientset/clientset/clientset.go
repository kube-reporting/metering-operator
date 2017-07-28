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

package clientset

import (
	glog "github.com/golang/glog"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	discovery "k8s.io/client-go/discovery"
	rest "k8s.io/client-go/rest"
	flowcontrol "k8s.io/client-go/util/flowcontrol"
)

type Interface interface {
	Discovery() discovery.DiscoveryInterface
	ApiextensionsV1beta1() apiextensionsv1beta1.ApiextensionsV1beta1Interface
	// Deprecated: please explicitly pick a version if possible.
	Apiextensions() apiextensionsv1beta1.ApiextensionsV1beta1Interface
}

// Clientset contains the clients for groups. Each group has exactly one
// version included in a Clientset.
type Clientset struct {
	*discovery.DiscoveryClient
	*apiextensionsv1beta1.ApiextensionsV1beta1Client
}

// ApiextensionsV1beta1 retrieves the ApiextensionsV1beta1Client
func (c *Clientset) ApiextensionsV1beta1() apiextensionsv1beta1.ApiextensionsV1beta1Interface {
	if c == nil {
		return nil
	}
	return c.ApiextensionsV1beta1Client
}

// Deprecated: Apiextensions retrieves the default version of ApiextensionsClient.
// Please explicitly pick a version.
func (c *Clientset) Apiextensions() apiextensionsv1beta1.ApiextensionsV1beta1Interface {
	if c == nil {
		return nil
	}
	return c.ApiextensionsV1beta1Client
}

// Discovery retrieves the DiscoveryClient
func (c *Clientset) Discovery() discovery.DiscoveryInterface {
	if c == nil {
		return nil
	}
	return c.DiscoveryClient
}

// NewForCon***REMOVED***g creates a new Clientset for the given con***REMOVED***g.
func NewForCon***REMOVED***g(c *rest.Con***REMOVED***g) (*Clientset, error) {
	con***REMOVED***gShallowCopy := *c
	if con***REMOVED***gShallowCopy.RateLimiter == nil && con***REMOVED***gShallowCopy.QPS > 0 {
		con***REMOVED***gShallowCopy.RateLimiter = flowcontrol.NewTokenBucketRateLimiter(con***REMOVED***gShallowCopy.QPS, con***REMOVED***gShallowCopy.Burst)
	}
	var cs Clientset
	var err error
	cs.ApiextensionsV1beta1Client, err = apiextensionsv1beta1.NewForCon***REMOVED***g(&con***REMOVED***gShallowCopy)
	if err != nil {
		return nil, err
	}

	cs.DiscoveryClient, err = discovery.NewDiscoveryClientForCon***REMOVED***g(&con***REMOVED***gShallowCopy)
	if err != nil {
		glog.Errorf("failed to create the DiscoveryClient: %v", err)
		return nil, err
	}
	return &cs, nil
}

// NewForCon***REMOVED***gOrDie creates a new Clientset for the given con***REMOVED***g and
// panics if there is an error in the con***REMOVED***g.
func NewForCon***REMOVED***gOrDie(c *rest.Con***REMOVED***g) *Clientset {
	var cs Clientset
	cs.ApiextensionsV1beta1Client = apiextensionsv1beta1.NewForCon***REMOVED***gOrDie(c)

	cs.DiscoveryClient = discovery.NewDiscoveryClientForCon***REMOVED***gOrDie(c)
	return &cs
}

// New creates a new Clientset for the given RESTClient.
func New(c rest.Interface) *Clientset {
	var cs Clientset
	cs.ApiextensionsV1beta1Client = apiextensionsv1beta1.New(c)

	cs.DiscoveryClient = discovery.NewDiscoveryClient(c)
	return &cs
}
