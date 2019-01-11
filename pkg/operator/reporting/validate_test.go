package reporting

import (
	"testing"

	"github.com/stretchr/testify/assert"

	metering "github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1"
	"github.com/operator-framework/operator-metering/pkg/operator/reportingutil"
	"github.com/operator-framework/operator-metering/test/testhelpers"
)

func TestValidateGenerationQueryDependencies(t *testing.T) {
	reportQueryViewUnset := testhelpers.NewReportGenerationQuery("uninitialized-query", "default", nil)
	reportQueryViewDisabled := testhelpers.NewReportGenerationQuery("query-view-disabled", "default", nil)
	reportQueryViewDisabled.Spec.View.Disabled = true
	reportQueryViewSet := testhelpers.NewReportGenerationQuery("initialized-query", "default", nil)
	reportQueryViewSet.Status.ViewName = reportingutil.GenerationQueryViewName("test-ns", "initialized-query")

	dataSourceTableUnset := testhelpers.NewReportDataSource("uninitialized-datasource", "default")
	dataSourceTableSet := testhelpers.NewReportDataSource("initialized-datasource", "default")
	dataSourceTableSet.Status.TableName = reportingutil.DataSourceTableName("test-ns", "initialized-datasource")

	reportTableUnset := testhelpers.NewReport("uninitialized-report", "default", "some-query", nil, nil, metering.ReportStatus{})
	reportTableSet := testhelpers.NewReport("initialized-report", "default", "some-query", nil, nil, metering.ReportStatus{
		TableName: reportingutil.ReportTableName("test-ns", "initialized-report"),
	})

	// we keep a set of our test objects here since we re-use them in different
	// combinations in the test cases
	uninitializedQueries := []*metering.ReportGenerationQuery{
		reportQueryViewUnset,
	}
	disabledViewQueries := []*metering.ReportGenerationQuery{
		reportQueryViewDisabled,
	}
	combinedUninitializedQueries := []*metering.ReportGenerationQuery{
		reportQueryViewUnset,
		reportQueryViewDisabled,
	}
	initializedQueries := []*metering.ReportGenerationQuery{
		reportQueryViewSet,
	}
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
		"ReportGenerationQueryDependencies dependencies with their view created is valid": {
			deps: ReportGenerationQueryDependencies{
				ReportGenerationQueries: initializedQueries,
			},
		},
		"ReportGenerationQuery dependencies missing their status.viewName unset is a validation error": {
			deps: ReportGenerationQueryDependencies{
				ReportGenerationQueries: uninitializedQueries,
			},
			expectErr: true,
		},
		// if view is disabled, then it should be a DynamicReportQueries
		// dependency, not a regular one
		"ReportGenerationQuery dependencies with view disabled is invalid": {
			deps: ReportGenerationQueryDependencies{
				ReportGenerationQueries: disabledViewQueries,
			},
			expectErr: true,
		},
		"DynamicReportGenerationQuery dependencies with view disabled is valid": {
			deps: ReportGenerationQueryDependencies{
				DynamicReportGenerationQueries: disabledViewQueries,
			},
		},
		"multiple invalid/uninitialized ReportGenerationQuery dependencies is a validation error": {
			deps: ReportGenerationQueryDependencies{
				ReportGenerationQueries: combinedUninitializedQueries,
			},
			expectErr: true,
		},
		"ReportDataSource dependencies with status.tableName unset is a validation error": {
			deps: ReportGenerationQueryDependencies{
				ReportDataSources: uninitializedDataSources,
			},
			expectErr: true,
		},
		"ReportDataSource dependencies with status.tableName set is valid": {
			deps: ReportGenerationQueryDependencies{
				ReportDataSources: initializedDataSources,
			},
		},
		"Report dependencies with status.tableName unset is a validation error": {
			deps: ReportGenerationQueryDependencies{
				Reports: uninitializedReports,
			},
			expectErr: true,
		},
		"Report dependencies with status.tableName set is valid": {
			deps: ReportGenerationQueryDependencies{
				Reports: initializedReports,
			},
		},
		"mixing valid and invalid dependencies is a validation error": {
			deps: ReportGenerationQueryDependencies{
				Reports:                 uninitializedReports,
				ReportGenerationQueries: combinedUninitializedQueries,
				ReportDataSources:       uninitializedDataSources,
				// disabledViewQueries is fine for dynamic queries
				DynamicReportGenerationQueries: disabledViewQueries,
			},
			expectErr: true,
		},
		"mixing valid dependencies is a valid": {
			deps: ReportGenerationQueryDependencies{
				Reports:                        uninitializedReports,
				ReportGenerationQueries:        initializedQueries,
				ReportDataSources:              initializedDataSources,
				DynamicReportGenerationQueries: disabledViewQueries,
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
			} else {
				assert.NoError(t, err, "expected validation to pass")
			}
		})
	}
}
