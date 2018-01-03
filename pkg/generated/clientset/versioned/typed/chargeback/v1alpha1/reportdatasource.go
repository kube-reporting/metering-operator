package v1alpha1

import (
	v1alpha1 "github.com/coreos-inc/kube-chargeback/pkg/apis/chargeback/v1alpha1"
	scheme "github.com/coreos-inc/kube-chargeback/pkg/generated/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// ReportDataSourcesGetter has a method to return a ReportDataSourceInterface.
// A group's client should implement this interface.
type ReportDataSourcesGetter interface {
	ReportDataSources(namespace string) ReportDataSourceInterface
}

// ReportDataSourceInterface has methods to work with ReportDataSource resources.
type ReportDataSourceInterface interface {
	Create(*v1alpha1.ReportDataSource) (*v1alpha1.ReportDataSource, error)
	Update(*v1alpha1.ReportDataSource) (*v1alpha1.ReportDataSource, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.ReportDataSource, error)
	List(opts v1.ListOptions) (*v1alpha1.ReportDataSourceList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.ReportDataSource, err error)
	ReportDataSourceExpansion
}

// reportDataSources implements ReportDataSourceInterface
type reportDataSources struct {
	client rest.Interface
	ns     string
}

// newReportDataSources returns a ReportDataSources
func newReportDataSources(c *ChargebackV1alpha1Client, namespace string) *reportDataSources {
	return &reportDataSources{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the reportDataSource, and returns the corresponding reportDataSource object, and an error if there is any.
func (c *reportDataSources) Get(name string, options v1.GetOptions) (result *v1alpha1.ReportDataSource, err error) {
	result = &v1alpha1.ReportDataSource{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("reportdatasources").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and ***REMOVED***eld selectors, and returns the list of ReportDataSources that match those selectors.
func (c *reportDataSources) List(opts v1.ListOptions) (result *v1alpha1.ReportDataSourceList, err error) {
	result = &v1alpha1.ReportDataSourceList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("reportdatasources").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested reportDataSources.
func (c *reportDataSources) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("reportdatasources").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a reportDataSource and creates it.  Returns the server's representation of the reportDataSource, and an error, if there is any.
func (c *reportDataSources) Create(reportDataSource *v1alpha1.ReportDataSource) (result *v1alpha1.ReportDataSource, err error) {
	result = &v1alpha1.ReportDataSource{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("reportdatasources").
		Body(reportDataSource).
		Do().
		Into(result)
	return
}

// Update takes the representation of a reportDataSource and updates it. Returns the server's representation of the reportDataSource, and an error, if there is any.
func (c *reportDataSources) Update(reportDataSource *v1alpha1.ReportDataSource) (result *v1alpha1.ReportDataSource, err error) {
	result = &v1alpha1.ReportDataSource{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("reportdatasources").
		Name(reportDataSource.Name).
		Body(reportDataSource).
		Do().
		Into(result)
	return
}

// Delete takes name of the reportDataSource and deletes it. Returns an error if one occurs.
func (c *reportDataSources) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("reportdatasources").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *reportDataSources) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("reportdatasources").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched reportDataSource.
func (c *reportDataSources) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.ReportDataSource, err error) {
	result = &v1alpha1.ReportDataSource{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("reportdatasources").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
