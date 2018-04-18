// Code generated by informer-gen. DO NOT EDIT.

// This ***REMOVED***le was automatically generated by informer-gen

package v1alpha1

import (
	time "time"

	chargeback_v1alpha1 "github.com/coreos-inc/kube-chargeback/pkg/apis/chargeback/v1alpha1"
	versioned "github.com/coreos-inc/kube-chargeback/pkg/generated/clientset/versioned"
	internalinterfaces "github.com/coreos-inc/kube-chargeback/pkg/generated/informers/externalversions/internalinterfaces"
	v1alpha1 "github.com/coreos-inc/kube-chargeback/pkg/generated/listers/chargeback/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// ReportPrometheusQueryInformer provides access to a shared informer and lister for
// ReportPrometheusQueries.
type ReportPrometheusQueryInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.ReportPrometheusQueryLister
}

type reportPrometheusQueryInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewReportPrometheusQueryInformer constructs a new informer for ReportPrometheusQuery type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewReportPrometheusQueryInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredReportPrometheusQueryInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredReportPrometheusQueryInformer constructs a new informer for ReportPrometheusQuery type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredReportPrometheusQueryInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.ChargebackV1alpha1().ReportPrometheusQueries(namespace).List(options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.ChargebackV1alpha1().ReportPrometheusQueries(namespace).Watch(options)
			},
		},
		&chargeback_v1alpha1.ReportPrometheusQuery{},
		resyncPeriod,
		indexers,
	)
}

func (f *reportPrometheusQueryInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredReportPrometheusQueryInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *reportPrometheusQueryInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&chargeback_v1alpha1.ReportPrometheusQuery{}, f.defaultInformer)
}

func (f *reportPrometheusQueryInformer) Lister() v1alpha1.ReportPrometheusQueryLister {
	return v1alpha1.NewReportPrometheusQueryLister(f.Informer().GetIndexer())
}
