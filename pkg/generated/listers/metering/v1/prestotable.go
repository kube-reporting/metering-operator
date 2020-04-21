// Code generated by lister-gen. DO NOT EDIT.

package v1

import (
	v1 "github.com/kubernetes-reporting/metering-operator/pkg/apis/metering/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// PrestoTableLister helps list PrestoTables.
type PrestoTableLister interface {
	// List lists all PrestoTables in the indexer.
	List(selector labels.Selector) (ret []*v1.PrestoTable, err error)
	// PrestoTables returns an object that can list and get PrestoTables.
	PrestoTables(namespace string) PrestoTableNamespaceLister
	PrestoTableListerExpansion
}

// prestoTableLister implements the PrestoTableLister interface.
type prestoTableLister struct {
	indexer cache.Indexer
}

// NewPrestoTableLister returns a new PrestoTableLister.
func NewPrestoTableLister(indexer cache.Indexer) PrestoTableLister {
	return &prestoTableLister{indexer: indexer}
}

// List lists all PrestoTables in the indexer.
func (s *prestoTableLister) List(selector labels.Selector) (ret []*v1.PrestoTable, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.PrestoTable))
	})
	return ret, err
}

// PrestoTables returns an object that can list and get PrestoTables.
func (s *prestoTableLister) PrestoTables(namespace string) PrestoTableNamespaceLister {
	return prestoTableNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// PrestoTableNamespaceLister helps list and get PrestoTables.
type PrestoTableNamespaceLister interface {
	// List lists all PrestoTables in the indexer for a given namespace.
	List(selector labels.Selector) (ret []*v1.PrestoTable, err error)
	// Get retrieves the PrestoTable from the indexer for a given namespace and name.
	Get(name string) (*v1.PrestoTable, error)
	PrestoTableNamespaceListerExpansion
}

// prestoTableNamespaceLister implements the PrestoTableNamespaceLister
// interface.
type prestoTableNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all PrestoTables in the indexer for a given namespace.
func (s prestoTableNamespaceLister) List(selector labels.Selector) (ret []*v1.PrestoTable, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.PrestoTable))
	})
	return ret, err
}

// Get retrieves the PrestoTable from the indexer for a given namespace and name.
func (s prestoTableNamespaceLister) Get(name string) (*v1.PrestoTable, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1.Resource("prestotable"), name)
	}
	return obj.(*v1.PrestoTable), nil
}
