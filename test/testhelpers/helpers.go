package testhelpers

import (
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1"
	"github.com/operator-framework/operator-metering/pkg/hive"
	"github.com/operator-framework/operator-metering/pkg/operator/reportingutil"
)

func NewReport(name, namespace, testQueryName string, reportStart, reportEnd *time.Time, status v1alpha1.ReportStatus) *v1alpha1.Report {
	var start, end *meta.Time
	if reportStart != nil {
		start = &meta.Time{*reportStart}
	}
	if reportEnd != nil {
		end = &meta.Time{*reportEnd}
	}
	return &v1alpha1.Report{
		ObjectMeta: meta.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha1.ReportSpec{
			GenerationQueryName: testQueryName,
			ReportingStart:      start,
			ReportingEnd:        end,
		},
		Status: status,
	}
}

func NewReportGenerationQuery(name, namespace string, columns []v1alpha1.ReportGenerationQueryColumn) *v1alpha1.ReportGenerationQuery {
	return &v1alpha1.ReportGenerationQuery{
		ObjectMeta: meta.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha1.ReportGenerationQuerySpec{
			Columns: columns,
		},
	}
}

func NewReportDataSource(name, namespace string) *v1alpha1.ReportDataSource {
	return &v1alpha1.ReportDataSource{
		ObjectMeta: meta.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
}

func NewPrestoTable(name, namespace string, columns []hive.Column) *v1alpha1.PrestoTable {
	return &v1alpha1.PrestoTable{
		ObjectMeta: meta.ObjectMeta{
			Name:      reportingutil.PrestoTableResourceNameFromKind("Report", namespace, name),
			Namespace: namespace,
		},
		Status: v1alpha1.PrestoTableStatus{
			Parameters: v1alpha1.TableParameters{
				Columns: columns,
			},
		},
	}
}

type ReportDataSourceStore struct {
	datasources map[string]*v1alpha1.ReportDataSource
}

func NewReportDataSourceStore(datasources []*v1alpha1.ReportDataSource) (store *ReportDataSourceStore) {
	m := make(map[string]*v1alpha1.ReportDataSource)
	for _, dataSource := range datasources {
		m[dataSource.Namespace+"/"+dataSource.Name] = dataSource
	}
	return &ReportDataSourceStore{m}
}

func (store *ReportDataSourceStore) GetReportDataSource(namespace, name string) (*v1alpha1.ReportDataSource, error) {
	dataSource, ok := store.datasources[namespace+"/"+name]
	if ok {
		return dataSource, nil
	}
	return nil, errors.NewNotFound(v1alpha1.Resource("ReportDataSource"), name)
}

type ReportGenerationQueryStore struct {
	queries map[string]*v1alpha1.ReportGenerationQuery
}

func NewReportGenerationQueryStore(queries []*v1alpha1.ReportGenerationQuery) (store *ReportGenerationQueryStore) {
	m := make(map[string]*v1alpha1.ReportGenerationQuery)
	for _, query := range queries {
		m[query.Namespace+"/"+query.Name] = query
	}
	return &ReportGenerationQueryStore{m}
}

func (store *ReportGenerationQueryStore) GetReportGenerationQuery(namespace, name string) (*v1alpha1.ReportGenerationQuery, error) {
	query, ok := store.queries[namespace+"/"+name]
	if ok {
		return query, nil
	}
	return nil, errors.NewNotFound(v1alpha1.Resource("ReportGenerationQuery"), name)
}

type ReportStore struct {
	reports map[string]*v1alpha1.Report
}

func NewReportStore(reports []*v1alpha1.Report) (store *ReportStore) {
	m := make(map[string]*v1alpha1.Report)
	for _, report := range reports {
		m[report.Namespace+"/"+report.Name] = report
	}
	return &ReportStore{m}
}

func (store *ReportStore) GetReport(namespace, name string) (*v1alpha1.Report, error) {
	report, ok := store.reports[namespace+"/"+name]
	if ok {
		return report, nil
	}
	return nil, errors.NewNotFound(v1alpha1.Resource("Report"), name)
}
