// Code generated by lister-gen. DO NOT EDIT.

// This ***REMOVED***le was automatically generated by lister-gen

package v1alpha1

import (
	v1alpha1 "github.com/coreos-inc/kube-chargeback/pkg/apis/chargeback/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// ReportGenerationQueryLister helps list ReportGenerationQueries.
type ReportGenerationQueryLister interface {
	// List lists all ReportGenerationQueries in the indexer.
	List(selector labels.Selector) (ret []*v1alpha1.ReportGenerationQuery, err error)
	// ReportGenerationQueries returns an object that can list and get ReportGenerationQueries.
	ReportGenerationQueries(namespace string) ReportGenerationQueryNamespaceLister
	ReportGenerationQueryListerExpansion
}

// reportGenerationQueryLister implements the ReportGenerationQueryLister interface.
type reportGenerationQueryLister struct {
	indexer cache.Indexer
}

// NewReportGenerationQueryLister returns a new ReportGenerationQueryLister.
func NewReportGenerationQueryLister(indexer cache.Indexer) ReportGenerationQueryLister {
	return &reportGenerationQueryLister{indexer: indexer}
}

// List lists all ReportGenerationQueries in the indexer.
func (s *reportGenerationQueryLister) List(selector labels.Selector) (ret []*v1alpha1.ReportGenerationQuery, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.ReportGenerationQuery))
	})
	return ret, err
}

// ReportGenerationQueries returns an object that can list and get ReportGenerationQueries.
func (s *reportGenerationQueryLister) ReportGenerationQueries(namespace string) ReportGenerationQueryNamespaceLister {
	return reportGenerationQueryNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// ReportGenerationQueryNamespaceLister helps list and get ReportGenerationQueries.
type ReportGenerationQueryNamespaceLister interface {
	// List lists all ReportGenerationQueries in the indexer for a given namespace.
	List(selector labels.Selector) (ret []*v1alpha1.ReportGenerationQuery, err error)
	// Get retrieves the ReportGenerationQuery from the indexer for a given namespace and name.
	Get(name string) (*v1alpha1.ReportGenerationQuery, error)
	ReportGenerationQueryNamespaceListerExpansion
}

// reportGenerationQueryNamespaceLister implements the ReportGenerationQueryNamespaceLister
// interface.
type reportGenerationQueryNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all ReportGenerationQueries in the indexer for a given namespace.
func (s reportGenerationQueryNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.ReportGenerationQuery, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.ReportGenerationQuery))
	})
	return ret, err
}

// Get retrieves the ReportGenerationQuery from the indexer for a given namespace and name.
func (s reportGenerationQueryNamespaceLister) Get(name string) (*v1alpha1.ReportGenerationQuery, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("reportgenerationquery"), name)
	}
	return obj.(*v1alpha1.ReportGenerationQuery), nil
}
