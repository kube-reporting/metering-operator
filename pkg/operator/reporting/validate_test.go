package reporting

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"

	metering "github.com/operator-framework/operator-metering/pkg/apis/metering/v1"
	"github.com/operator-framework/operator-metering/pkg/operator/reportingutil"
	"github.com/operator-framework/operator-metering/test/testhelpers"
)

func TestValidateQueryDependencies(t *testing.T) {
	// reportQuery := testhelpers.NewReportQuery("uninitialized-query", "default", nil)
	dataSourceTableUnset := testhelpers.NewReportDataSource("uninitialized-datasource", "default")
	dataSourceTableSet := testhelpers.NewReportDataSource("initialized-datasource", "default")
	dataSourceTableSet.Status.TableRef.Name = reportingutil.DataSourceTableName("test-ns", "initialized-datasource")

	reportTableUnset := testhelpers.NewReport("uninitialized-report", "default", "some-query", nil, nil, metering.ReportStatus{}, nil, false)
	reportTableSet := testhelpers.NewReport("initialized-report", "default", "some-query", nil, nil, metering.ReportStatus{
		TableRef: v1.LocalObjectReference{Name: reportingutil.ReportTableName("test-ns", "initialized-report")},
	}, nil, false)

	uninitializedDataSources := []*metering.ReportDataSource{
		dataSourceTableUnset,
	}
	initializedDataSources := []*metering.ReportDataSource{
		dataSourceTableSet,
	}
	uninitializedReports := []*metering.Report{
		reportTableUnset,
	}
	initializedReports := []*metering.Report{
		reportTableSet,
	}

	tests := map[string]struct {
		deps      ReportQueryDependencies
		expectErr bool
	}{
		"no dependencies results in no errors": {
			deps: ReportQueryDependencies{},
		},
		"ReportQueryDependencies dependencies on other queries is valid": {
			deps: ReportQueryDependencies{
				// ReportQueries: initializedQueries,
			},
		},
		"ReportDataSource dependencies with status.tableRef.name unset is a validation error": {
			deps: ReportQueryDependencies{
				ReportDataSources: uninitializedDataSources,
			},
			expectErr: true,
		},
		"ReportDataSource dependencies with status.tableRef.name set is valid": {
			deps: ReportQueryDependencies{
				ReportDataSources: initializedDataSources,
			},
		},
		"Report dependencies with status.tableRef.name unset is a validation error": {
			deps: ReportQueryDependencies{
				Reports: uninitializedReports,
			},
			expectErr: true,
		},
		"Report dependencies with status.tableRef.name set is valid": {
			deps: ReportQueryDependencies{
				Reports: initializedReports,
			},
		},
		"mixing valid and invalid dependencies is a validation error": {
			deps: ReportQueryDependencies{
				Reports: uninitializedReports,
				// ReportQueries: initializedQueries,
				ReportDataSources: uninitializedDataSources,
			},
			expectErr: true,
		},
		"mixing valid dependencies is a valid": {
			deps: ReportQueryDependencies{
				Reports: uninitializedReports,
				// ReportQueries: initializedQueries,
				ReportDataSources: initializedDataSources,
			},
			expectErr: true,
		},
	}

	for testName, tt := range tests {
		testName := testName
		tt := tt
		t.Run(testName, func(t *testing.T) {
			err := ValidateQueryDependencies(&tt.deps, nil)
			if tt.expectErr {
				assert.NotNil(t, err, "expected a validation error")
			} ***REMOVED*** {
				assert.NoError(t, err, "expected validation to pass")
			}
		})
	}
}
