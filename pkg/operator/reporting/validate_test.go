package reporting

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"

	metering "github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1"
	"github.com/operator-framework/operator-metering/pkg/operator/reportingutil"
	"github.com/operator-framework/operator-metering/test/testhelpers"
)

func TestValidateGenerationQueryDependencies(t *testing.T) {
	// reportQuery := testhelpers.NewReportGenerationQuery("uninitialized-query", "default", nil)
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
		deps      ReportGenerationQueryDependencies
		expectErr bool
	}{
		"no dependencies results in no errors": {
			deps: ReportGenerationQueryDependencies{},
		},
		"ReportGenerationQueryDependencies dependencies on other queries is valid": {
			deps: ReportGenerationQueryDependencies{
				// ReportGenerationQueries: initializedQueries,
			},
		},
		"ReportDataSource dependencies with status.tableRef.name unset is a validation error": {
			deps: ReportGenerationQueryDependencies{
				ReportDataSources: uninitializedDataSources,
			},
			expectErr: true,
		},
		"ReportDataSource dependencies with status.tableRef.name set is valid": {
			deps: ReportGenerationQueryDependencies{
				ReportDataSources: initializedDataSources,
			},
		},
		"Report dependencies with status.tableRef.name unset is a validation error": {
			deps: ReportGenerationQueryDependencies{
				Reports: uninitializedReports,
			},
			expectErr: true,
		},
		"Report dependencies with status.tableRef.name set is valid": {
			deps: ReportGenerationQueryDependencies{
				Reports: initializedReports,
			},
		},
		"mixing valid and invalid dependencies is a validation error": {
			deps: ReportGenerationQueryDependencies{
				Reports: uninitializedReports,
				// ReportGenerationQueries: initializedQueries,
				ReportDataSources: uninitializedDataSources,
			},
			expectErr: true,
		},
		"mixing valid dependencies is a valid": {
			deps: ReportGenerationQueryDependencies{
				Reports: uninitializedReports,
				// ReportGenerationQueries: initializedQueries,
				ReportDataSources: initializedDataSources,
			},
			expectErr: true,
		},
	}

	for testName, tt := range tests {
		testName := testName
		tt := tt
		t.Run(testName, func(t *testing.T) {
			err := ValidateGenerationQueryDependencies(&tt.deps, nil)
			if tt.expectErr {
				assert.NotNil(t, err, "expected a validation error")
			} ***REMOVED*** {
				assert.NoError(t, err, "expected validation to pass")
			}
		})
	}
}
