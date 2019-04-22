package reporting

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	metering "github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1"
	meteringClient "github.com/operator-framework/operator-metering/pkg/generated/clientset/versioned/typed/metering/v1alpha1"
	meteringListers "github.com/operator-framework/operator-metering/pkg/generated/listers/metering/v1alpha1"
)

type reportDataSourceGetter interface {
	getReportDataSource(namespace, name string) (*metering.ReportDataSource, error)
}

type reportDataSourceGetterFunc func(string, string) (*metering.ReportDataSource, error)

func (f reportDataSourceGetterFunc) getReportDataSource(namespace, name string) (*metering.ReportDataSource, error) {
	return f(namespace, name)
}

func NewReportDataSourceListerGetter(lister meteringListers.ReportDataSourceLister) reportDataSourceGetter {
	return reportDataSourceGetterFunc(func(namespace, name string) (*metering.ReportDataSource, error) {
		return lister.ReportDataSources(namespace).Get(name)
	})
}

func NewReportDataSourceClientGetter(getter meteringClient.ReportDataSourcesGetter) reportDataSourceGetter {
	return reportDataSourceGetterFunc(func(namespace, name string) (*metering.ReportDataSource, error) {
		return getter.ReportDataSources(namespace).Get(name, metav1.GetOptions{})
	})
}

type reportGetter interface {
	getReport(namespace, name string) (*metering.Report, error)
}

type reportGetterFunc func(string, string) (*metering.Report, error)

func (f reportGetterFunc) getReport(namespace, name string) (*metering.Report, error) {
	return f(namespace, name)
}

func NewReportListerGetter(lister meteringListers.ReportLister) reportGetter {
	return reportGetterFunc(func(namespace, name string) (*metering.Report, error) {
		return lister.Reports(namespace).Get(name)
	})
}

func NewReportClientGetter(getter meteringClient.ReportsGetter) reportGetter {
	return reportGetterFunc(func(namespace, name string) (*metering.Report, error) {
		return getter.Reports(namespace).Get(name, metav1.GetOptions{})
	})
}

type reportGenerationQueryGetter interface {
	getReportGenerationQuery(namespace, name string) (*metering.ReportGenerationQuery, error)
}

type reportGenerationQueryGetterFunc func(string, string) (*metering.ReportGenerationQuery, error)

func (f reportGenerationQueryGetterFunc) getReportGenerationQuery(namespace, name string) (*metering.ReportGenerationQuery, error) {
	return f(namespace, name)
}

func NewReportGenerationQueryListerGetter(lister meteringListers.ReportGenerationQueryLister) reportGenerationQueryGetter {
	return reportGenerationQueryGetterFunc(func(namespace, name string) (*metering.ReportGenerationQuery, error) {
		return lister.ReportGenerationQueries(namespace).Get(name)
	})
}

func NewReportGenerationQueryClientGetter(getter meteringClient.ReportGenerationQueriesGetter) reportGenerationQueryGetter {
	return reportGenerationQueryGetterFunc(func(namespace, name string) (*metering.ReportGenerationQuery, error) {
		return getter.ReportGenerationQueries(namespace).Get(name, metav1.GetOptions{})
	})
}
