package testhelpers

import (
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1"
	"github.com/operator-framework/operator-metering/pkg/operator/reportingutil"
	"github.com/operator-framework/operator-metering/pkg/presto"
)

// NewReport creates a mock report used for testing purposes.
func NewReport(name, namespace, testQueryName string, reportStart, reportEnd *time.Time, status v1alpha1.ReportStatus, schedule *v1alpha1.ReportSchedule, runImmediately bool) *v1alpha1.Report {
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
			QueryName:      testQueryName,
			ReportingStart: start,
			ReportingEnd:   end,
			Schedule:       schedule,
			RunImmediately: runImmediately,
		},
		Status: status,
	}
}

func NewReportQuery(name, namespace string, columns []v1alpha1.ReportQueryColumn) *v1alpha1.ReportQuery {
	return &v1alpha1.ReportQuery{
		ObjectMeta: meta.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha1.ReportQuerySpec{
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

func NewPrestoTable(name, namespace string, columns []presto.Column) *v1alpha1.PrestoTable {
	return &v1alpha1.PrestoTable{
		ObjectMeta: meta.ObjectMeta{
			Name:      reportingutil.TableResourceNameFromKind("Report", namespace, name),
			Namespace: namespace,
		},
		Status: v1alpha1.PrestoTableStatus{
			Columns: columns,
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

type ReportQueryStore struct {
	queries map[string]*v1alpha1.ReportQuery
}

func NewReportQueryStore(queries []*v1alpha1.ReportQuery) (store *ReportQueryStore) {
	m := make(map[string]*v1alpha1.ReportQuery)
	for _, query := range queries {
		m[query.Namespace+"/"+query.Name] = query
	}
	return &ReportQueryStore{m}
}

func (store *ReportQueryStore) GetReportQuery(namespace, name string) (*v1alpha1.ReportQuery, error) {
	query, ok := store.queries[namespace+"/"+name]
	if ok {
		return query, nil
	}
	return nil, errors.NewNotFound(v1alpha1.Resource("ReportQuery"), name)
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
