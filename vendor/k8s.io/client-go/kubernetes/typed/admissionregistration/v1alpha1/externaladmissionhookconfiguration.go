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

package v1alpha1

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	scheme "k8s.io/client-go/kubernetes/scheme"
	v1alpha1 "k8s.io/client-go/pkg/apis/admissionregistration/v1alpha1"
	rest "k8s.io/client-go/rest"
)

// ExternalAdmissionHookCon***REMOVED***gurationsGetter has a method to return a ExternalAdmissionHookCon***REMOVED***gurationInterface.
// A group's client should implement this interface.
type ExternalAdmissionHookCon***REMOVED***gurationsGetter interface {
	ExternalAdmissionHookCon***REMOVED***gurations() ExternalAdmissionHookCon***REMOVED***gurationInterface
}

// ExternalAdmissionHookCon***REMOVED***gurationInterface has methods to work with ExternalAdmissionHookCon***REMOVED***guration resources.
type ExternalAdmissionHookCon***REMOVED***gurationInterface interface {
	Create(*v1alpha1.ExternalAdmissionHookCon***REMOVED***guration) (*v1alpha1.ExternalAdmissionHookCon***REMOVED***guration, error)
	Update(*v1alpha1.ExternalAdmissionHookCon***REMOVED***guration) (*v1alpha1.ExternalAdmissionHookCon***REMOVED***guration, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.ExternalAdmissionHookCon***REMOVED***guration, error)
	List(opts v1.ListOptions) (*v1alpha1.ExternalAdmissionHookCon***REMOVED***gurationList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.ExternalAdmissionHookCon***REMOVED***guration, err error)
	ExternalAdmissionHookCon***REMOVED***gurationExpansion
}

// externalAdmissionHookCon***REMOVED***gurations implements ExternalAdmissionHookCon***REMOVED***gurationInterface
type externalAdmissionHookCon***REMOVED***gurations struct {
	client rest.Interface
}

// newExternalAdmissionHookCon***REMOVED***gurations returns a ExternalAdmissionHookCon***REMOVED***gurations
func newExternalAdmissionHookCon***REMOVED***gurations(c *AdmissionregistrationV1alpha1Client) *externalAdmissionHookCon***REMOVED***gurations {
	return &externalAdmissionHookCon***REMOVED***gurations{
		client: c.RESTClient(),
	}
}

// Create takes the representation of a externalAdmissionHookCon***REMOVED***guration and creates it.  Returns the server's representation of the externalAdmissionHookCon***REMOVED***guration, and an error, if there is any.
func (c *externalAdmissionHookCon***REMOVED***gurations) Create(externalAdmissionHookCon***REMOVED***guration *v1alpha1.ExternalAdmissionHookCon***REMOVED***guration) (result *v1alpha1.ExternalAdmissionHookCon***REMOVED***guration, err error) {
	result = &v1alpha1.ExternalAdmissionHookCon***REMOVED***guration{}
	err = c.client.Post().
		Resource("externaladmissionhookcon***REMOVED***gurations").
		Body(externalAdmissionHookCon***REMOVED***guration).
		Do().
		Into(result)
	return
}

// Update takes the representation of a externalAdmissionHookCon***REMOVED***guration and updates it. Returns the server's representation of the externalAdmissionHookCon***REMOVED***guration, and an error, if there is any.
func (c *externalAdmissionHookCon***REMOVED***gurations) Update(externalAdmissionHookCon***REMOVED***guration *v1alpha1.ExternalAdmissionHookCon***REMOVED***guration) (result *v1alpha1.ExternalAdmissionHookCon***REMOVED***guration, err error) {
	result = &v1alpha1.ExternalAdmissionHookCon***REMOVED***guration{}
	err = c.client.Put().
		Resource("externaladmissionhookcon***REMOVED***gurations").
		Name(externalAdmissionHookCon***REMOVED***guration.Name).
		Body(externalAdmissionHookCon***REMOVED***guration).
		Do().
		Into(result)
	return
}

// Delete takes name of the externalAdmissionHookCon***REMOVED***guration and deletes it. Returns an error if one occurs.
func (c *externalAdmissionHookCon***REMOVED***gurations) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Resource("externaladmissionhookcon***REMOVED***gurations").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *externalAdmissionHookCon***REMOVED***gurations) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Resource("externaladmissionhookcon***REMOVED***gurations").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Get takes name of the externalAdmissionHookCon***REMOVED***guration, and returns the corresponding externalAdmissionHookCon***REMOVED***guration object, and an error if there is any.
func (c *externalAdmissionHookCon***REMOVED***gurations) Get(name string, options v1.GetOptions) (result *v1alpha1.ExternalAdmissionHookCon***REMOVED***guration, err error) {
	result = &v1alpha1.ExternalAdmissionHookCon***REMOVED***guration{}
	err = c.client.Get().
		Resource("externaladmissionhookcon***REMOVED***gurations").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and ***REMOVED***eld selectors, and returns the list of ExternalAdmissionHookCon***REMOVED***gurations that match those selectors.
func (c *externalAdmissionHookCon***REMOVED***gurations) List(opts v1.ListOptions) (result *v1alpha1.ExternalAdmissionHookCon***REMOVED***gurationList, err error) {
	result = &v1alpha1.ExternalAdmissionHookCon***REMOVED***gurationList{}
	err = c.client.Get().
		Resource("externaladmissionhookcon***REMOVED***gurations").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested externalAdmissionHookCon***REMOVED***gurations.
func (c *externalAdmissionHookCon***REMOVED***gurations) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Resource("externaladmissionhookcon***REMOVED***gurations").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Patch applies the patch and returns the patched externalAdmissionHookCon***REMOVED***guration.
func (c *externalAdmissionHookCon***REMOVED***gurations) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.ExternalAdmissionHookCon***REMOVED***guration, err error) {
	result = &v1alpha1.ExternalAdmissionHookCon***REMOVED***guration{}
	err = c.client.Patch(pt).
		Resource("externaladmissionhookcon***REMOVED***gurations").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
