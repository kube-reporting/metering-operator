package fake

import (
	v1alpha1 "github.com/coreos-inc/kube-chargeback/pkg/apis/chargeback/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeReportPrometheusQueries implements ReportPrometheusQueryInterface
type FakeReportPrometheusQueries struct {
	Fake *FakeChargebackV1alpha1
	ns   string
}

var reportprometheusqueriesResource = schema.GroupVersionResource{Group: "chargeback.coreos.com", Version: "v1alpha1", Resource: "reportprometheusqueries"}

var reportprometheusqueriesKind = schema.GroupVersionKind{Group: "chargeback.coreos.com", Version: "v1alpha1", Kind: "ReportPrometheusQuery"}

// Get takes name of the reportPrometheusQuery, and returns the corresponding reportPrometheusQuery object, and an error if there is any.
func (c *FakeReportPrometheusQueries) Get(name string, options v1.GetOptions) (result *v1alpha1.ReportPrometheusQuery, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(reportprometheusqueriesResource, c.ns, name), &v1alpha1.ReportPrometheusQuery{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ReportPrometheusQuery), err
}

// List takes label and field selectors, and returns the list of ReportPrometheusQueries that match those selectors.
func (c *FakeReportPrometheusQueries) List(opts v1.ListOptions) (result *v1alpha1.ReportPrometheusQueryList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(reportprometheusqueriesResource, reportprometheusqueriesKind, c.ns, opts), &v1alpha1.ReportPrometheusQueryList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.ReportPrometheusQueryList{}
	for _, item := range obj.(*v1alpha1.ReportPrometheusQueryList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested reportPrometheusQueries.
func (c *FakeReportPrometheusQueries) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(reportprometheusqueriesResource, c.ns, opts))

}

// Create takes the representation of a reportPrometheusQuery and creates it.  Returns the server's representation of the reportPrometheusQuery, and an error, if there is any.
func (c *FakeReportPrometheusQueries) Create(reportPrometheusQuery *v1alpha1.ReportPrometheusQuery) (result *v1alpha1.ReportPrometheusQuery, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(reportprometheusqueriesResource, c.ns, reportPrometheusQuery), &v1alpha1.ReportPrometheusQuery{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ReportPrometheusQuery), err
}

// Update takes the representation of a reportPrometheusQuery and updates it. Returns the server's representation of the reportPrometheusQuery, and an error, if there is any.
func (c *FakeReportPrometheusQueries) Update(reportPrometheusQuery *v1alpha1.ReportPrometheusQuery) (result *v1alpha1.ReportPrometheusQuery, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(reportprometheusqueriesResource, c.ns, reportPrometheusQuery), &v1alpha1.ReportPrometheusQuery{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ReportPrometheusQuery), err
}

// Delete takes name of the reportPrometheusQuery and deletes it. Returns an error if one occurs.
func (c *FakeReportPrometheusQueries) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(reportprometheusqueriesResource, c.ns, name), &v1alpha1.ReportPrometheusQuery{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeReportPrometheusQueries) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(reportprometheusqueriesResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.ReportPrometheusQueryList{})
	return err
}

// Patch applies the patch and returns the patched reportPrometheusQuery.
func (c *FakeReportPrometheusQueries) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.ReportPrometheusQuery, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(reportprometheusqueriesResource, c.ns, name, data, subresources...), &v1alpha1.ReportPrometheusQuery{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ReportPrometheusQuery), err
}
