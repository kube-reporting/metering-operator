package v1alpha1

import (
	v1alpha1 "github.com/coreos-inc/kube-chargeback/pkg/apis/chargeback/v1alpha1"
	scheme "github.com/coreos-inc/kube-chargeback/pkg/generated/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// ReportGenerationQueriesGetter has a method to return a ReportGenerationQueryInterface.
// A group's client should implement this interface.
type ReportGenerationQueriesGetter interface {
	ReportGenerationQueries(namespace string) ReportGenerationQueryInterface
}

// ReportGenerationQueryInterface has methods to work with ReportGenerationQuery resources.
type ReportGenerationQueryInterface interface {
	Create(*v1alpha1.ReportGenerationQuery) (*v1alpha1.ReportGenerationQuery, error)
	Update(*v1alpha1.ReportGenerationQuery) (*v1alpha1.ReportGenerationQuery, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.ReportGenerationQuery, error)
	List(opts v1.ListOptions) (*v1alpha1.ReportGenerationQueryList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.ReportGenerationQuery, err error)
	ReportGenerationQueryExpansion
}

// reportGenerationQueries implements ReportGenerationQueryInterface
type reportGenerationQueries struct {
	client rest.Interface
	ns     string
}

// newReportGenerationQueries returns a ReportGenerationQueries
func newReportGenerationQueries(c *ChargebackV1alpha1Client, namespace string) *reportGenerationQueries {
	return &reportGenerationQueries{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the reportGenerationQuery, and returns the corresponding reportGenerationQuery object, and an error if there is any.
func (c *reportGenerationQueries) Get(name string, options v1.GetOptions) (result *v1alpha1.ReportGenerationQuery, err error) {
	result = &v1alpha1.ReportGenerationQuery{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("reportgenerationqueries").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and ***REMOVED***eld selectors, and returns the list of ReportGenerationQueries that match those selectors.
func (c *reportGenerationQueries) List(opts v1.ListOptions) (result *v1alpha1.ReportGenerationQueryList, err error) {
	result = &v1alpha1.ReportGenerationQueryList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("reportgenerationqueries").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested reportGenerationQueries.
func (c *reportGenerationQueries) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("reportgenerationqueries").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a reportGenerationQuery and creates it.  Returns the server's representation of the reportGenerationQuery, and an error, if there is any.
func (c *reportGenerationQueries) Create(reportGenerationQuery *v1alpha1.ReportGenerationQuery) (result *v1alpha1.ReportGenerationQuery, err error) {
	result = &v1alpha1.ReportGenerationQuery{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("reportgenerationqueries").
		Body(reportGenerationQuery).
		Do().
		Into(result)
	return
}

// Update takes the representation of a reportGenerationQuery and updates it. Returns the server's representation of the reportGenerationQuery, and an error, if there is any.
func (c *reportGenerationQueries) Update(reportGenerationQuery *v1alpha1.ReportGenerationQuery) (result *v1alpha1.ReportGenerationQuery, err error) {
	result = &v1alpha1.ReportGenerationQuery{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("reportgenerationqueries").
		Name(reportGenerationQuery.Name).
		Body(reportGenerationQuery).
		Do().
		Into(result)
	return
}

// Delete takes name of the reportGenerationQuery and deletes it. Returns an error if one occurs.
func (c *reportGenerationQueries) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("reportgenerationqueries").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *reportGenerationQueries) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("reportgenerationqueries").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched reportGenerationQuery.
func (c *reportGenerationQueries) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.ReportGenerationQuery, err error) {
	result = &v1alpha1.ReportGenerationQuery{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("reportgenerationqueries").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
