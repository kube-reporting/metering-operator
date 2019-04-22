package reporting

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	metering "github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1"
	meteringClient "github.com/operator-framework/operator-metering/pkg/generated/clientset/versioned/typed/metering/v1alpha1"
	meteringListers "github.com/operator-framework/operator-metering/pkg/generated/listers/metering/v1alpha1"
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
		return getter.ReportDataSources(namespace).Get(name, metav1.GetOptions{})
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
		return getter.Reports(namespace).Get(name, metav1.GetOptions{})
	})
}

type ReportGenerationQueryGetter interface {
	GetReportGenerationQuery(namespace, name string) (*metering.ReportGenerationQuery, error)
}

type ReportGenerationQueryGetterFunc func(string, string) (*metering.ReportGenerationQuery, error)

func (f ReportGenerationQueryGetterFunc) GetReportGenerationQuery(namespace, name string) (*metering.ReportGenerationQuery, error) {
	return f(namespace, name)
}

func NewReportGenerationQueryListerGetter(lister meteringListers.ReportGenerationQueryLister) ReportGenerationQueryGetter {
	return ReportGenerationQueryGetterFunc(func(namespace, name string) (*metering.ReportGenerationQuery, error) {
		return lister.ReportGenerationQueries(namespace).Get(name)
	})
}

func NewReportGenerationQueryClientGetter(getter meteringClient.ReportGenerationQueriesGetter) ReportGenerationQueryGetter {
	return ReportGenerationQueryGetterFunc(func(namespace, name string) (*metering.ReportGenerationQuery, error) {
		return getter.ReportGenerationQueries(namespace).Get(name, metav1.GetOptions{})
	})
}
