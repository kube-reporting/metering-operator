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
	v1beta1 "k8s.io/api/admissionregistration/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	scheme "k8s.io/client-go/kubernetes/scheme"
	rest "k8s.io/client-go/rest"
)

// ValidatingWebhookCon***REMOVED***gurationsGetter has a method to return a ValidatingWebhookCon***REMOVED***gurationInterface.
// A group's client should implement this interface.
type ValidatingWebhookCon***REMOVED***gurationsGetter interface {
	ValidatingWebhookCon***REMOVED***gurations() ValidatingWebhookCon***REMOVED***gurationInterface
}

// ValidatingWebhookCon***REMOVED***gurationInterface has methods to work with ValidatingWebhookCon***REMOVED***guration resources.
type ValidatingWebhookCon***REMOVED***gurationInterface interface {
	Create(*v1beta1.ValidatingWebhookCon***REMOVED***guration) (*v1beta1.ValidatingWebhookCon***REMOVED***guration, error)
	Update(*v1beta1.ValidatingWebhookCon***REMOVED***guration) (*v1beta1.ValidatingWebhookCon***REMOVED***guration, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1beta1.ValidatingWebhookCon***REMOVED***guration, error)
	List(opts v1.ListOptions) (*v1beta1.ValidatingWebhookCon***REMOVED***gurationList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1beta1.ValidatingWebhookCon***REMOVED***guration, err error)
	ValidatingWebhookCon***REMOVED***gurationExpansion
}

// validatingWebhookCon***REMOVED***gurations implements ValidatingWebhookCon***REMOVED***gurationInterface
type validatingWebhookCon***REMOVED***gurations struct {
	client rest.Interface
}

// newValidatingWebhookCon***REMOVED***gurations returns a ValidatingWebhookCon***REMOVED***gurations
func newValidatingWebhookCon***REMOVED***gurations(c *AdmissionregistrationV1beta1Client) *validatingWebhookCon***REMOVED***gurations {
	return &validatingWebhookCon***REMOVED***gurations{
		client: c.RESTClient(),
	}
}

// Get takes name of the validatingWebhookCon***REMOVED***guration, and returns the corresponding validatingWebhookCon***REMOVED***guration object, and an error if there is any.
func (c *validatingWebhookCon***REMOVED***gurations) Get(name string, options v1.GetOptions) (result *v1beta1.ValidatingWebhookCon***REMOVED***guration, err error) {
	result = &v1beta1.ValidatingWebhookCon***REMOVED***guration{}
	err = c.client.Get().
		Resource("validatingwebhookcon***REMOVED***gurations").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and ***REMOVED***eld selectors, and returns the list of ValidatingWebhookCon***REMOVED***gurations that match those selectors.
func (c *validatingWebhookCon***REMOVED***gurations) List(opts v1.ListOptions) (result *v1beta1.ValidatingWebhookCon***REMOVED***gurationList, err error) {
	result = &v1beta1.ValidatingWebhookCon***REMOVED***gurationList{}
	err = c.client.Get().
		Resource("validatingwebhookcon***REMOVED***gurations").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested validatingWebhookCon***REMOVED***gurations.
func (c *validatingWebhookCon***REMOVED***gurations) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Resource("validatingwebhookcon***REMOVED***gurations").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a validatingWebhookCon***REMOVED***guration and creates it.  Returns the server's representation of the validatingWebhookCon***REMOVED***guration, and an error, if there is any.
func (c *validatingWebhookCon***REMOVED***gurations) Create(validatingWebhookCon***REMOVED***guration *v1beta1.ValidatingWebhookCon***REMOVED***guration) (result *v1beta1.ValidatingWebhookCon***REMOVED***guration, err error) {
	result = &v1beta1.ValidatingWebhookCon***REMOVED***guration{}
	err = c.client.Post().
		Resource("validatingwebhookcon***REMOVED***gurations").
		Body(validatingWebhookCon***REMOVED***guration).
		Do().
		Into(result)
	return
}

// Update takes the representation of a validatingWebhookCon***REMOVED***guration and updates it. Returns the server's representation of the validatingWebhookCon***REMOVED***guration, and an error, if there is any.
func (c *validatingWebhookCon***REMOVED***gurations) Update(validatingWebhookCon***REMOVED***guration *v1beta1.ValidatingWebhookCon***REMOVED***guration) (result *v1beta1.ValidatingWebhookCon***REMOVED***guration, err error) {
	result = &v1beta1.ValidatingWebhookCon***REMOVED***guration{}
	err = c.client.Put().
		Resource("validatingwebhookcon***REMOVED***gurations").
		Name(validatingWebhookCon***REMOVED***guration.Name).
		Body(validatingWebhookCon***REMOVED***guration).
		Do().
		Into(result)
	return
}

// Delete takes name of the validatingWebhookCon***REMOVED***guration and deletes it. Returns an error if one occurs.
func (c *validatingWebhookCon***REMOVED***gurations) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Resource("validatingwebhookcon***REMOVED***gurations").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *validatingWebhookCon***REMOVED***gurations) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Resource("validatingwebhookcon***REMOVED***gurations").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched validatingWebhookCon***REMOVED***guration.
func (c *validatingWebhookCon***REMOVED***gurations) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1beta1.ValidatingWebhookCon***REMOVED***guration, err error) {
	result = &v1beta1.ValidatingWebhookCon***REMOVED***guration{}
	err = c.client.Patch(pt).
		Resource("validatingwebhookcon***REMOVED***gurations").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
