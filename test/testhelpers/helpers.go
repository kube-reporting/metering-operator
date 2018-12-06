package testhelpers

import (
	"time"

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
			Name:      reportingutil.PrestoTableResourceNameFromKind("Report", name),
			Namespace: namespace,
		},
		Status: v1alpha1.PrestoTableStatus{
			Parameters: v1alpha1.TableParameters{
				Columns: columns,
			},
		},
	}
}
