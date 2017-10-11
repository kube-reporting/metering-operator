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

// FakeReportDataStores implements ReportDataStoreInterface
type FakeReportDataStores struct {
	Fake *FakeChargebackV1alpha1
	ns   string
}

var reportdatastoresResource = schema.GroupVersionResource{Group: "chargeback.coreos.com", Version: "v1alpha1", Resource: "reportdatastores"}

var reportdatastoresKind = schema.GroupVersionKind{Group: "chargeback.coreos.com", Version: "v1alpha1", Kind: "ReportDataStore"}

// Get takes name of the reportDataStore, and returns the corresponding reportDataStore object, and an error if there is any.
func (c *FakeReportDataStores) Get(name string, options v1.GetOptions) (result *v1alpha1.ReportDataStore, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(reportdatastoresResource, c.ns, name), &v1alpha1.ReportDataStore{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ReportDataStore), err
}

// List takes label and field selectors, and returns the list of ReportDataStores that match those selectors.
func (c *FakeReportDataStores) List(opts v1.ListOptions) (result *v1alpha1.ReportDataStoreList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(reportdatastoresResource, reportdatastoresKind, c.ns, opts), &v1alpha1.ReportDataStoreList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.ReportDataStoreList{}
	for _, item := range obj.(*v1alpha1.ReportDataStoreList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested reportDataStores.
func (c *FakeReportDataStores) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(reportdatastoresResource, c.ns, opts))

}

// Create takes the representation of a reportDataStore and creates it.  Returns the server's representation of the reportDataStore, and an error, if there is any.
func (c *FakeReportDataStores) Create(reportDataStore *v1alpha1.ReportDataStore) (result *v1alpha1.ReportDataStore, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(reportdatastoresResource, c.ns, reportDataStore), &v1alpha1.ReportDataStore{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ReportDataStore), err
}

// Update takes the representation of a reportDataStore and updates it. Returns the server's representation of the reportDataStore, and an error, if there is any.
func (c *FakeReportDataStores) Update(reportDataStore *v1alpha1.ReportDataStore) (result *v1alpha1.ReportDataStore, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(reportdatastoresResource, c.ns, reportDataStore), &v1alpha1.ReportDataStore{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ReportDataStore), err
}

// Delete takes name of the reportDataStore and deletes it. Returns an error if one occurs.
func (c *FakeReportDataStores) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(reportdatastoresResource, c.ns, name), &v1alpha1.ReportDataStore{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeReportDataStores) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(reportdatastoresResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.ReportDataStoreList{})
	return err
}

// Patch applies the patch and returns the patched reportDataStore.
func (c *FakeReportDataStores) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.ReportDataStore, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(reportdatastoresResource, c.ns, name, data, subresources...), &v1alpha1.ReportDataStore{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ReportDataStore), err
}
