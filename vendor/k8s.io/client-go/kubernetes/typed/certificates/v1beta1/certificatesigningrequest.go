/*
Copyright 2018 The Kubernetes Authors.

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

package v1beta1

import (
	v1beta1 "k8s.io/api/certi***REMOVED***cates/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	scheme "k8s.io/client-go/kubernetes/scheme"
	rest "k8s.io/client-go/rest"
)

// Certi***REMOVED***cateSigningRequestsGetter has a method to return a Certi***REMOVED***cateSigningRequestInterface.
// A group's client should implement this interface.
type Certi***REMOVED***cateSigningRequestsGetter interface {
	Certi***REMOVED***cateSigningRequests() Certi***REMOVED***cateSigningRequestInterface
}

// Certi***REMOVED***cateSigningRequestInterface has methods to work with Certi***REMOVED***cateSigningRequest resources.
type Certi***REMOVED***cateSigningRequestInterface interface {
	Create(*v1beta1.Certi***REMOVED***cateSigningRequest) (*v1beta1.Certi***REMOVED***cateSigningRequest, error)
	Update(*v1beta1.Certi***REMOVED***cateSigningRequest) (*v1beta1.Certi***REMOVED***cateSigningRequest, error)
	UpdateStatus(*v1beta1.Certi***REMOVED***cateSigningRequest) (*v1beta1.Certi***REMOVED***cateSigningRequest, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1beta1.Certi***REMOVED***cateSigningRequest, error)
	List(opts v1.ListOptions) (*v1beta1.Certi***REMOVED***cateSigningRequestList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1beta1.Certi***REMOVED***cateSigningRequest, err error)
	Certi***REMOVED***cateSigningRequestExpansion
}

// certi***REMOVED***cateSigningRequests implements Certi***REMOVED***cateSigningRequestInterface
type certi***REMOVED***cateSigningRequests struct {
	client rest.Interface
}

// newCerti***REMOVED***cateSigningRequests returns a Certi***REMOVED***cateSigningRequests
func newCerti***REMOVED***cateSigningRequests(c *Certi***REMOVED***catesV1beta1Client) *certi***REMOVED***cateSigningRequests {
	return &certi***REMOVED***cateSigningRequests{
		client: c.RESTClient(),
	}
}

// Get takes name of the certi***REMOVED***cateSigningRequest, and returns the corresponding certi***REMOVED***cateSigningRequest object, and an error if there is any.
func (c *certi***REMOVED***cateSigningRequests) Get(name string, options v1.GetOptions) (result *v1beta1.Certi***REMOVED***cateSigningRequest, err error) {
	result = &v1beta1.Certi***REMOVED***cateSigningRequest{}
	err = c.client.Get().
		Resource("certi***REMOVED***catesigningrequests").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and ***REMOVED***eld selectors, and returns the list of Certi***REMOVED***cateSigningRequests that match those selectors.
func (c *certi***REMOVED***cateSigningRequests) List(opts v1.ListOptions) (result *v1beta1.Certi***REMOVED***cateSigningRequestList, err error) {
	result = &v1beta1.Certi***REMOVED***cateSigningRequestList{}
	err = c.client.Get().
		Resource("certi***REMOVED***catesigningrequests").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested certi***REMOVED***cateSigningRequests.
func (c *certi***REMOVED***cateSigningRequests) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Resource("certi***REMOVED***catesigningrequests").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a certi***REMOVED***cateSigningRequest and creates it.  Returns the server's representation of the certi***REMOVED***cateSigningRequest, and an error, if there is any.
func (c *certi***REMOVED***cateSigningRequests) Create(certi***REMOVED***cateSigningRequest *v1beta1.Certi***REMOVED***cateSigningRequest) (result *v1beta1.Certi***REMOVED***cateSigningRequest, err error) {
	result = &v1beta1.Certi***REMOVED***cateSigningRequest{}
	err = c.client.Post().
		Resource("certi***REMOVED***catesigningrequests").
		Body(certi***REMOVED***cateSigningRequest).
		Do().
		Into(result)
	return
}

// Update takes the representation of a certi***REMOVED***cateSigningRequest and updates it. Returns the server's representation of the certi***REMOVED***cateSigningRequest, and an error, if there is any.
func (c *certi***REMOVED***cateSigningRequests) Update(certi***REMOVED***cateSigningRequest *v1beta1.Certi***REMOVED***cateSigningRequest) (result *v1beta1.Certi***REMOVED***cateSigningRequest, err error) {
	result = &v1beta1.Certi***REMOVED***cateSigningRequest{}
	err = c.client.Put().
		Resource("certi***REMOVED***catesigningrequests").
		Name(certi***REMOVED***cateSigningRequest.Name).
		Body(certi***REMOVED***cateSigningRequest).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *certi***REMOVED***cateSigningRequests) UpdateStatus(certi***REMOVED***cateSigningRequest *v1beta1.Certi***REMOVED***cateSigningRequest) (result *v1beta1.Certi***REMOVED***cateSigningRequest, err error) {
	result = &v1beta1.Certi***REMOVED***cateSigningRequest{}
	err = c.client.Put().
		Resource("certi***REMOVED***catesigningrequests").
		Name(certi***REMOVED***cateSigningRequest.Name).
		SubResource("status").
		Body(certi***REMOVED***cateSigningRequest).
		Do().
		Into(result)
	return
}

// Delete takes name of the certi***REMOVED***cateSigningRequest and deletes it. Returns an error if one occurs.
func (c *certi***REMOVED***cateSigningRequests) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Resource("certi***REMOVED***catesigningrequests").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *certi***REMOVED***cateSigningRequests) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Resource("certi***REMOVED***catesigningrequests").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched certi***REMOVED***cateSigningRequest.
func (c *certi***REMOVED***cateSigningRequests) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1beta1.Certi***REMOVED***cateSigningRequest, err error) {
	result = &v1beta1.Certi***REMOVED***cateSigningRequest{}
	err = c.client.Patch(pt).
		Resource("certi***REMOVED***catesigningrequests").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
