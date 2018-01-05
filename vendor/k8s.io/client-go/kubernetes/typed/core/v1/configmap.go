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

package v1

import (
	v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	scheme "k8s.io/client-go/kubernetes/scheme"
	rest "k8s.io/client-go/rest"
)

// Con***REMOVED***gMapsGetter has a method to return a Con***REMOVED***gMapInterface.
// A group's client should implement this interface.
type Con***REMOVED***gMapsGetter interface {
	Con***REMOVED***gMaps(namespace string) Con***REMOVED***gMapInterface
}

// Con***REMOVED***gMapInterface has methods to work with Con***REMOVED***gMap resources.
type Con***REMOVED***gMapInterface interface {
	Create(*v1.Con***REMOVED***gMap) (*v1.Con***REMOVED***gMap, error)
	Update(*v1.Con***REMOVED***gMap) (*v1.Con***REMOVED***gMap, error)
	Delete(name string, options *meta_v1.DeleteOptions) error
	DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error
	Get(name string, options meta_v1.GetOptions) (*v1.Con***REMOVED***gMap, error)
	List(opts meta_v1.ListOptions) (*v1.Con***REMOVED***gMapList, error)
	Watch(opts meta_v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.Con***REMOVED***gMap, err error)
	Con***REMOVED***gMapExpansion
}

// con***REMOVED***gMaps implements Con***REMOVED***gMapInterface
type con***REMOVED***gMaps struct {
	client rest.Interface
	ns     string
}

// newCon***REMOVED***gMaps returns a Con***REMOVED***gMaps
func newCon***REMOVED***gMaps(c *CoreV1Client, namespace string) *con***REMOVED***gMaps {
	return &con***REMOVED***gMaps{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the con***REMOVED***gMap, and returns the corresponding con***REMOVED***gMap object, and an error if there is any.
func (c *con***REMOVED***gMaps) Get(name string, options meta_v1.GetOptions) (result *v1.Con***REMOVED***gMap, err error) {
	result = &v1.Con***REMOVED***gMap{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("con***REMOVED***gmaps").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and ***REMOVED***eld selectors, and returns the list of Con***REMOVED***gMaps that match those selectors.
func (c *con***REMOVED***gMaps) List(opts meta_v1.ListOptions) (result *v1.Con***REMOVED***gMapList, err error) {
	result = &v1.Con***REMOVED***gMapList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("con***REMOVED***gmaps").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested con***REMOVED***gMaps.
func (c *con***REMOVED***gMaps) Watch(opts meta_v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("con***REMOVED***gmaps").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a con***REMOVED***gMap and creates it.  Returns the server's representation of the con***REMOVED***gMap, and an error, if there is any.
func (c *con***REMOVED***gMaps) Create(con***REMOVED***gMap *v1.Con***REMOVED***gMap) (result *v1.Con***REMOVED***gMap, err error) {
	result = &v1.Con***REMOVED***gMap{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("con***REMOVED***gmaps").
		Body(con***REMOVED***gMap).
		Do().
		Into(result)
	return
}

// Update takes the representation of a con***REMOVED***gMap and updates it. Returns the server's representation of the con***REMOVED***gMap, and an error, if there is any.
func (c *con***REMOVED***gMaps) Update(con***REMOVED***gMap *v1.Con***REMOVED***gMap) (result *v1.Con***REMOVED***gMap, err error) {
	result = &v1.Con***REMOVED***gMap{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("con***REMOVED***gmaps").
		Name(con***REMOVED***gMap.Name).
		Body(con***REMOVED***gMap).
		Do().
		Into(result)
	return
}

// Delete takes name of the con***REMOVED***gMap and deletes it. Returns an error if one occurs.
func (c *con***REMOVED***gMaps) Delete(name string, options *meta_v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("con***REMOVED***gmaps").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *con***REMOVED***gMaps) DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("con***REMOVED***gmaps").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched con***REMOVED***gMap.
func (c *con***REMOVED***gMaps) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.Con***REMOVED***gMap, err error) {
	result = &v1.Con***REMOVED***gMap{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("con***REMOVED***gmaps").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
