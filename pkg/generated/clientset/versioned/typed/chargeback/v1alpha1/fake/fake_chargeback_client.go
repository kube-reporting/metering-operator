package fake

import (
	v1alpha1 "github.com/coreos-inc/kube-chargeback/pkg/generated/clientset/versioned/typed/chargeback/v1alpha1"
	rest "k8s.io/client-go/rest"
	testing "k8s.io/client-go/testing"
)

type FakeChargebackV1alpha1 struct {
	*testing.Fake
}

func (c *FakeChargebackV1alpha1) PrestoTables(namespace string) v1alpha1.PrestoTableInterface {
	return &FakePrestoTables{c, namespace}
}

func (c *FakeChargebackV1alpha1) Reports(namespace string) v1alpha1.ReportInterface {
	return &FakeReports{c, namespace}
}

func (c *FakeChargebackV1alpha1) ReportDataSources(namespace string) v1alpha1.ReportDataSourceInterface {
	return &FakeReportDataSources{c, namespace}
}

func (c *FakeChargebackV1alpha1) ReportGenerationQueries(namespace string) v1alpha1.ReportGenerationQueryInterface {
	return &FakeReportGenerationQueries{c, namespace}
}

func (c *FakeChargebackV1alpha1) ReportPrometheusQueries(namespace string) v1alpha1.ReportPrometheusQueryInterface {
	return &FakeReportPrometheusQueries{c, namespace}
}

func (c *FakeChargebackV1alpha1) ScheduledReports(namespace string) v1alpha1.ScheduledReportInterface {
	return &FakeScheduledReports{c, namespace}
}

func (c *FakeChargebackV1alpha1) StorageLocations(namespace string) v1alpha1.StorageLocationInterface {
	return &FakeStorageLocations{c, namespace}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *FakeChargebackV1alpha1) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}
