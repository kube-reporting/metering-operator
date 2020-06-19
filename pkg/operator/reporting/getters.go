package reporting

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	metering "github.com/kube-reporting/metering-operator/pkg/apis/metering/v1"
	meteringClient "github.com/kube-reporting/metering-operator/pkg/generated/clientset/versioned/typed/metering/v1"
	meteringListers "github.com/kube-reporting/metering-operator/pkg/generated/listers/metering/v1"
)

type ReportDataSourceGetter interface {
	GetReportDataSource(namespace, name string) (*metering.ReportDataSource, error)
}

type ReportDataSourceGetterFunc func(string, string) (*metering.ReportDataSource, error)

func (f ReportDataSourceGetterFunc) GetReportDataSource(namespace, name string) (*metering.ReportDataSource, error) {
	return f(namespace, name)
}

func NewReportDataSourceListerGetter(lister meteringListers.ReportDataSourceLister) ReportDataSourceGetter {
	return ReportDataSourceGetterFunc(func(namespace, name string) (*metering.ReportDataSource, error) {
		return lister.ReportDataSources(namespace).Get(name)
	})
}

func NewReportDataSourceClientGetter(getter meteringClient.ReportDataSourcesGetter) ReportDataSourceGetter {
	return ReportDataSourceGetterFunc(func(namespace, name string) (*metering.ReportDataSource, error) {
		return getter.ReportDataSources(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	})
}

type ReportGetter interface {
	GetReport(namespace, name string) (*metering.Report, error)
}

type ReportGetterFunc func(string, string) (*metering.Report, error)

func (f ReportGetterFunc) GetReport(namespace, name string) (*metering.Report, error) {
	return f(namespace, name)
}

func NewReportListerGetter(lister meteringListers.ReportLister) ReportGetter {
	return ReportGetterFunc(func(namespace, name string) (*metering.Report, error) {
		return lister.Reports(namespace).Get(name)
	})
}

func NewReportClientGetter(getter meteringClient.ReportsGetter) ReportGetter {
	return ReportGetterFunc(func(namespace, name string) (*metering.Report, error) {
		return getter.Reports(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	})
}

type ReportQueryGetter interface {
	GetReportQuery(namespace, name string) (*metering.ReportQuery, error)
}

type ReportQueryGetterFunc func(string, string) (*metering.ReportQuery, error)

func (f ReportQueryGetterFunc) GetReportQuery(namespace, name string) (*metering.ReportQuery, error) {
	return f(namespace, name)
}

func NewReportQueryListerGetter(lister meteringListers.ReportQueryLister) ReportQueryGetter {
	return ReportQueryGetterFunc(func(namespace, name string) (*metering.ReportQuery, error) {
		return lister.ReportQueries(namespace).Get(name)
	})
}

func NewReportQueryClientGetter(getter meteringClient.ReportQueriesGetter) ReportQueryGetter {
	return ReportQueryGetterFunc(func(namespace, name string) (*metering.ReportQuery, error) {
		return getter.ReportQueries(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	})
}
