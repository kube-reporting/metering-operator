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

// FakePrestoTables implements PrestoTableInterface
type FakePrestoTables struct {
	Fake *FakeChargebackV1alpha1
	ns   string
}

var prestotablesResource = schema.GroupVersionResource{Group: "chargeback.coreos.com", Version: "v1alpha1", Resource: "prestotables"}

var prestotablesKind = schema.GroupVersionKind{Group: "chargeback.coreos.com", Version: "v1alpha1", Kind: "PrestoTable"}

// Get takes name of the prestoTable, and returns the corresponding prestoTable object, and an error if there is any.
func (c *FakePrestoTables) Get(name string, options v1.GetOptions) (result *v1alpha1.PrestoTable, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(prestotablesResource, c.ns, name), &v1alpha1.PrestoTable{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PrestoTable), err
}

// List takes label and field selectors, and returns the list of PrestoTables that match those selectors.
func (c *FakePrestoTables) List(opts v1.ListOptions) (result *v1alpha1.PrestoTableList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(prestotablesResource, prestotablesKind, c.ns, opts), &v1alpha1.PrestoTableList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.PrestoTableList{}
	for _, item := range obj.(*v1alpha1.PrestoTableList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested prestoTables.
func (c *FakePrestoTables) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(prestotablesResource, c.ns, opts))

}

// Create takes the representation of a prestoTable and creates it.  Returns the server's representation of the prestoTable, and an error, if there is any.
func (c *FakePrestoTables) Create(prestoTable *v1alpha1.PrestoTable) (result *v1alpha1.PrestoTable, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(prestotablesResource, c.ns, prestoTable), &v1alpha1.PrestoTable{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PrestoTable), err
}

// Update takes the representation of a prestoTable and updates it. Returns the server's representation of the prestoTable, and an error, if there is any.
func (c *FakePrestoTables) Update(prestoTable *v1alpha1.PrestoTable) (result *v1alpha1.PrestoTable, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(prestotablesResource, c.ns, prestoTable), &v1alpha1.PrestoTable{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PrestoTable), err
}

// Delete takes name of the prestoTable and deletes it. Returns an error if one occurs.
func (c *FakePrestoTables) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(prestotablesResource, c.ns, name), &v1alpha1.PrestoTable{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakePrestoTables) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(prestotablesResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.PrestoTableList{})
	return err
}

// Patch applies the patch and returns the patched prestoTable.
func (c *FakePrestoTables) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.PrestoTable, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(prestotablesResource, c.ns, name, data, subresources...), &v1alpha1.PrestoTable{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PrestoTable), err
}
