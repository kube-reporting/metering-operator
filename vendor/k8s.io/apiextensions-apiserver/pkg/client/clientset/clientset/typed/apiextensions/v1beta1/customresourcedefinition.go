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

package v1beta1

import (
	v1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	scheme "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// CustomResourceDe***REMOVED***nitionsGetter has a method to return a CustomResourceDe***REMOVED***nitionInterface.
// A group's client should implement this interface.
type CustomResourceDe***REMOVED***nitionsGetter interface {
	CustomResourceDe***REMOVED***nitions() CustomResourceDe***REMOVED***nitionInterface
}

// CustomResourceDe***REMOVED***nitionInterface has methods to work with CustomResourceDe***REMOVED***nition resources.
type CustomResourceDe***REMOVED***nitionInterface interface {
	Create(*v1beta1.CustomResourceDe***REMOVED***nition) (*v1beta1.CustomResourceDe***REMOVED***nition, error)
	Update(*v1beta1.CustomResourceDe***REMOVED***nition) (*v1beta1.CustomResourceDe***REMOVED***nition, error)
	UpdateStatus(*v1beta1.CustomResourceDe***REMOVED***nition) (*v1beta1.CustomResourceDe***REMOVED***nition, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1beta1.CustomResourceDe***REMOVED***nition, error)
	List(opts v1.ListOptions) (*v1beta1.CustomResourceDe***REMOVED***nitionList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1beta1.CustomResourceDe***REMOVED***nition, err error)
	CustomResourceDe***REMOVED***nitionExpansion
}

// customResourceDe***REMOVED***nitions implements CustomResourceDe***REMOVED***nitionInterface
type customResourceDe***REMOVED***nitions struct {
	client rest.Interface
}

// newCustomResourceDe***REMOVED***nitions returns a CustomResourceDe***REMOVED***nitions
func newCustomResourceDe***REMOVED***nitions(c *ApiextensionsV1beta1Client) *customResourceDe***REMOVED***nitions {
	return &customResourceDe***REMOVED***nitions{
		client: c.RESTClient(),
	}
}

// Create takes the representation of a customResourceDe***REMOVED***nition and creates it.  Returns the server's representation of the customResourceDe***REMOVED***nition, and an error, if there is any.
func (c *customResourceDe***REMOVED***nitions) Create(customResourceDe***REMOVED***nition *v1beta1.CustomResourceDe***REMOVED***nition) (result *v1beta1.CustomResourceDe***REMOVED***nition, err error) {
	result = &v1beta1.CustomResourceDe***REMOVED***nition{}
	err = c.client.Post().
		Resource("customresourcede***REMOVED***nitions").
		Body(customResourceDe***REMOVED***nition).
		Do().
		Into(result)
	return
}

// Update takes the representation of a customResourceDe***REMOVED***nition and updates it. Returns the server's representation of the customResourceDe***REMOVED***nition, and an error, if there is any.
func (c *customResourceDe***REMOVED***nitions) Update(customResourceDe***REMOVED***nition *v1beta1.CustomResourceDe***REMOVED***nition) (result *v1beta1.CustomResourceDe***REMOVED***nition, err error) {
	result = &v1beta1.CustomResourceDe***REMOVED***nition{}
	err = c.client.Put().
		Resource("customresourcede***REMOVED***nitions").
		Name(customResourceDe***REMOVED***nition.Name).
		Body(customResourceDe***REMOVED***nition).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclientstatus=false comment above the type to avoid generating UpdateStatus().

func (c *customResourceDe***REMOVED***nitions) UpdateStatus(customResourceDe***REMOVED***nition *v1beta1.CustomResourceDe***REMOVED***nition) (result *v1beta1.CustomResourceDe***REMOVED***nition, err error) {
	result = &v1beta1.CustomResourceDe***REMOVED***nition{}
	err = c.client.Put().
		Resource("customresourcede***REMOVED***nitions").
		Name(customResourceDe***REMOVED***nition.Name).
		SubResource("status").
		Body(customResourceDe***REMOVED***nition).
		Do().
		Into(result)
	return
}

// Delete takes name of the customResourceDe***REMOVED***nition and deletes it. Returns an error if one occurs.
func (c *customResourceDe***REMOVED***nitions) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Resource("customresourcede***REMOVED***nitions").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *customResourceDe***REMOVED***nitions) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Resource("customresourcede***REMOVED***nitions").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Get takes name of the customResourceDe***REMOVED***nition, and returns the corresponding customResourceDe***REMOVED***nition object, and an error if there is any.
func (c *customResourceDe***REMOVED***nitions) Get(name string, options v1.GetOptions) (result *v1beta1.CustomResourceDe***REMOVED***nition, err error) {
	result = &v1beta1.CustomResourceDe***REMOVED***nition{}
	err = c.client.Get().
		Resource("customresourcede***REMOVED***nitions").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and ***REMOVED***eld selectors, and returns the list of CustomResourceDe***REMOVED***nitions that match those selectors.
func (c *customResourceDe***REMOVED***nitions) List(opts v1.ListOptions) (result *v1beta1.CustomResourceDe***REMOVED***nitionList, err error) {
	result = &v1beta1.CustomResourceDe***REMOVED***nitionList{}
	err = c.client.Get().
		Resource("customresourcede***REMOVED***nitions").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested customResourceDe***REMOVED***nitions.
func (c *customResourceDe***REMOVED***nitions) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Resource("customresourcede***REMOVED***nitions").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Patch applies the patch and returns the patched customResourceDe***REMOVED***nition.
func (c *customResourceDe***REMOVED***nitions) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1beta1.CustomResourceDe***REMOVED***nition, err error) {
	result = &v1beta1.CustomResourceDe***REMOVED***nition{}
	err = c.client.Patch(pt).
		Resource("customresourcede***REMOVED***nitions").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
