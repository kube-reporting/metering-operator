// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// ReportPrometheusQueryLister helps list ReportPrometheusQueries.
type ReportPrometheusQueryLister interface {
	// List lists all ReportPrometheusQueries in the indexer.
	List(selector labels.Selector) (ret []*v1alpha1.ReportPrometheusQuery, err error)
	// ReportPrometheusQueries returns an object that can list and get ReportPrometheusQueries.
	ReportPrometheusQueries(namespace string) ReportPrometheusQueryNamespaceLister
	ReportPrometheusQueryListerExpansion
}

// reportPrometheusQueryLister implements the ReportPrometheusQueryLister interface.
type reportPrometheusQueryLister struct {
	indexer cache.Indexer
}

// NewReportPrometheusQueryLister returns a new ReportPrometheusQueryLister.
func NewReportPrometheusQueryLister(indexer cache.Indexer) ReportPrometheusQueryLister {
	return &reportPrometheusQueryLister{indexer: indexer}
}

// List lists all ReportPrometheusQueries in the indexer.
func (s *reportPrometheusQueryLister) List(selector labels.Selector) (ret []*v1alpha1.ReportPrometheusQuery, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.ReportPrometheusQuery))
	})
	return ret, err
}

// ReportPrometheusQueries returns an object that can list and get ReportPrometheusQueries.
func (s *reportPrometheusQueryLister) ReportPrometheusQueries(namespace string) ReportPrometheusQueryNamespaceLister {
	return reportPrometheusQueryNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// ReportPrometheusQueryNamespaceLister helps list and get ReportPrometheusQueries.
type ReportPrometheusQueryNamespaceLister interface {
	// List lists all ReportPrometheusQueries in the indexer for a given namespace.
	List(selector labels.Selector) (ret []*v1alpha1.ReportPrometheusQuery, err error)
	// Get retrieves the ReportPrometheusQuery from the indexer for a given namespace and name.
	Get(name string) (*v1alpha1.ReportPrometheusQuery, error)
	ReportPrometheusQueryNamespaceListerExpansion
}

// reportPrometheusQueryNamespaceLister implements the ReportPrometheusQueryNamespaceLister
// interface.
type reportPrometheusQueryNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all ReportPrometheusQueries in the indexer for a given namespace.
func (s reportPrometheusQueryNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.ReportPrometheusQuery, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.ReportPrometheusQuery))
	})
	return ret, err
}

// Get retrieves the ReportPrometheusQuery from the indexer for a given namespace and name.
func (s reportPrometheusQueryNamespaceLister) Get(name string) (*v1alpha1.ReportPrometheusQuery, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("reportprometheusquery"), name)
	}
	return obj.(*v1alpha1.ReportPrometheusQuery), nil
}
