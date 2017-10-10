package v1alpha1

import (
	v1alpha1 "github.com/coreos-inc/kube-chargeback/pkg/apis/chargeback/v1alpha1"
	scheme "github.com/coreos-inc/kube-chargeback/pkg/generated/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// ReportDataStoresGetter has a method to return a ReportDataStoreInterface.
// A group's client should implement this interface.
type ReportDataStoresGetter interface {
	ReportDataStores(namespace string) ReportDataStoreInterface
}

// ReportDataStoreInterface has methods to work with ReportDataStore resources.
type ReportDataStoreInterface interface {
	Create(*v1alpha1.ReportDataStore) (*v1alpha1.ReportDataStore, error)
	Update(*v1alpha1.ReportDataStore) (*v1alpha1.ReportDataStore, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.ReportDataStore, error)
	List(opts v1.ListOptions) (*v1alpha1.ReportDataStoreList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.ReportDataStore, err error)
	ReportDataStoreExpansion
}

// reportDataStores implements ReportDataStoreInterface
type reportDataStores struct {
	client rest.Interface
	ns     string
}

// newReportDataStores returns a ReportDataStores
func newReportDataStores(c *ChargebackV1alpha1Client, namespace string) *reportDataStores {
	return &reportDataStores{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the reportDataStore, and returns the corresponding reportDataStore object, and an error if there is any.
func (c *reportDataStores) Get(name string, options v1.GetOptions) (result *v1alpha1.ReportDataStore, err error) {
	result = &v1alpha1.ReportDataStore{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("reportdatastores").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of ReportDataStores that match those selectors.
func (c *reportDataStores) List(opts v1.ListOptions) (result *v1alpha1.ReportDataStoreList, err error) {
	result = &v1alpha1.ReportDataStoreList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("reportdatastores").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested reportDataStores.
func (c *reportDataStores) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("reportdatastores").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a reportDataStore and creates it.  Returns the server's representation of the reportDataStore, and an error, if there is any.
func (c *reportDataStores) Create(reportDataStore *v1alpha1.ReportDataStore) (result *v1alpha1.ReportDataStore, err error) {
	result = &v1alpha1.ReportDataStore{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("reportdatastores").
		Body(reportDataStore).
		Do().
		Into(result)
	return
}

// Update takes the representation of a reportDataStore and updates it. Returns the server's representation of the reportDataStore, and an error, if there is any.
func (c *reportDataStores) Update(reportDataStore *v1alpha1.ReportDataStore) (result *v1alpha1.ReportDataStore, err error) {
	result = &v1alpha1.ReportDataStore{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("reportdatastores").
		Name(reportDataStore.Name).
		Body(reportDataStore).
		Do().
		Into(result)
	return
}

// Delete takes name of the reportDataStore and deletes it. Returns an error if one occurs.
func (c *reportDataStores) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("reportdatastores").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *reportDataStores) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("reportdatastores").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched reportDataStore.
func (c *reportDataStores) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.ReportDataStore, err error) {
	result = &v1alpha1.ReportDataStore{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("reportdatastores").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
