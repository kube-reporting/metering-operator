// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1"
	scheme "github.com/operator-framework/operator-metering/pkg/generated/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// PrestoTablesGetter has a method to return a PrestoTableInterface.
// A group's client should implement this interface.
type PrestoTablesGetter interface {
	PrestoTables(namespace string) PrestoTableInterface
}

// PrestoTableInterface has methods to work with PrestoTable resources.
type PrestoTableInterface interface {
	Create(*v1alpha1.PrestoTable) (*v1alpha1.PrestoTable, error)
	Update(*v1alpha1.PrestoTable) (*v1alpha1.PrestoTable, error)
	UpdateStatus(*v1alpha1.PrestoTable) (*v1alpha1.PrestoTable, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.PrestoTable, error)
	List(opts v1.ListOptions) (*v1alpha1.PrestoTableList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.PrestoTable, err error)
	PrestoTableExpansion
}

// prestoTables implements PrestoTableInterface
type prestoTables struct {
	client rest.Interface
	ns     string
}

// newPrestoTables returns a PrestoTables
func newPrestoTables(c *MeteringV1alpha1Client, namespace string) *prestoTables {
	return &prestoTables{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the prestoTable, and returns the corresponding prestoTable object, and an error if there is any.
func (c *prestoTables) Get(name string, options v1.GetOptions) (result *v1alpha1.PrestoTable, err error) {
	result = &v1alpha1.PrestoTable{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("prestotables").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and ***REMOVED***eld selectors, and returns the list of PrestoTables that match those selectors.
func (c *prestoTables) List(opts v1.ListOptions) (result *v1alpha1.PrestoTableList, err error) {
	result = &v1alpha1.PrestoTableList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("prestotables").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested prestoTables.
func (c *prestoTables) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("prestotables").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a prestoTable and creates it.  Returns the server's representation of the prestoTable, and an error, if there is any.
func (c *prestoTables) Create(prestoTable *v1alpha1.PrestoTable) (result *v1alpha1.PrestoTable, err error) {
	result = &v1alpha1.PrestoTable{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("prestotables").
		Body(prestoTable).
		Do().
		Into(result)
	return
}

// Update takes the representation of a prestoTable and updates it. Returns the server's representation of the prestoTable, and an error, if there is any.
func (c *prestoTables) Update(prestoTable *v1alpha1.PrestoTable) (result *v1alpha1.PrestoTable, err error) {
	result = &v1alpha1.PrestoTable{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("prestotables").
		Name(prestoTable.Name).
		Body(prestoTable).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *prestoTables) UpdateStatus(prestoTable *v1alpha1.PrestoTable) (result *v1alpha1.PrestoTable, err error) {
	result = &v1alpha1.PrestoTable{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("prestotables").
		Name(prestoTable.Name).
		SubResource("status").
		Body(prestoTable).
		Do().
		Into(result)
	return
}

// Delete takes name of the prestoTable and deletes it. Returns an error if one occurs.
func (c *prestoTables) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("prestotables").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *prestoTables) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("prestotables").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched prestoTable.
func (c *prestoTables) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.PrestoTable, err error) {
	result = &v1alpha1.PrestoTable{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("prestotables").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
