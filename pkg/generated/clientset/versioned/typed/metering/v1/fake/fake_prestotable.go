// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	meteringv1 "github.com/kube-reporting/metering-operator/pkg/apis/metering/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakePrestoTables implements PrestoTableInterface
type FakePrestoTables struct {
	Fake *FakeMeteringV1
	ns   string
}

var prestotablesResource = schema.GroupVersionResource{Group: "metering.openshift.io", Version: "v1", Resource: "prestotables"}

var prestotablesKind = schema.GroupVersionKind{Group: "metering.openshift.io", Version: "v1", Kind: "PrestoTable"}

// Get takes name of the prestoTable, and returns the corresponding prestoTable object, and an error if there is any.
func (c *FakePrestoTables) Get(name string, options v1.GetOptions) (result *meteringv1.PrestoTable, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(prestotablesResource, c.ns, name), &meteringv1.PrestoTable{})

	if obj == nil {
		return nil, err
	}
	return obj.(*meteringv1.PrestoTable), err
}

// List takes label and field selectors, and returns the list of PrestoTables that match those selectors.
func (c *FakePrestoTables) List(opts v1.ListOptions) (result *meteringv1.PrestoTableList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(prestotablesResource, prestotablesKind, c.ns, opts), &meteringv1.PrestoTableList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &meteringv1.PrestoTableList{ListMeta: obj.(*meteringv1.PrestoTableList).ListMeta}
	for _, item := range obj.(*meteringv1.PrestoTableList).Items {
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
func (c *FakePrestoTables) Create(prestoTable *meteringv1.PrestoTable) (result *meteringv1.PrestoTable, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(prestotablesResource, c.ns, prestoTable), &meteringv1.PrestoTable{})

	if obj == nil {
		return nil, err
	}
	return obj.(*meteringv1.PrestoTable), err
}

// Update takes the representation of a prestoTable and updates it. Returns the server's representation of the prestoTable, and an error, if there is any.
func (c *FakePrestoTables) Update(prestoTable *meteringv1.PrestoTable) (result *meteringv1.PrestoTable, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(prestotablesResource, c.ns, prestoTable), &meteringv1.PrestoTable{})

	if obj == nil {
		return nil, err
	}
	return obj.(*meteringv1.PrestoTable), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakePrestoTables) UpdateStatus(prestoTable *meteringv1.PrestoTable) (*meteringv1.PrestoTable, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(prestotablesResource, "status", c.ns, prestoTable), &meteringv1.PrestoTable{})

	if obj == nil {
		return nil, err
	}
	return obj.(*meteringv1.PrestoTable), err
}

// Delete takes name of the prestoTable and deletes it. Returns an error if one occurs.
func (c *FakePrestoTables) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(prestotablesResource, c.ns, name), &meteringv1.PrestoTable{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakePrestoTables) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(prestotablesResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &meteringv1.PrestoTableList{})
	return err
}

// Patch applies the patch and returns the patched prestoTable.
func (c *FakePrestoTables) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *meteringv1.PrestoTable, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(prestotablesResource, c.ns, name, pt, data, subresources...), &meteringv1.PrestoTable{})

	if obj == nil {
		return nil, err
	}
	return obj.(*meteringv1.PrestoTable), err
}
