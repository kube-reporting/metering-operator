package reporting

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

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

	reportTableUnset := testhelpers.NewReport("uninitialized-report", "default", "some-query", nil, nil, metering.ReportStatus{}, nil, false)
	reportTableSet := testhelpers.NewReport("initialized-report", "default", "some-query", nil, nil, metering.ReportStatus{
		TableName: reportingutil.ReportTableName("test-ns", "initialized-report"),
	}, nil, false)

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
				// disabledViewQueries is ***REMOVED***ne for dynamic queries
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
			} ***REMOVED*** {
				assert.NoError(t, err, "expected validation to pass")
			}
		})
	}
}

func TestGetGenerationQueryDependencies(t *testing.T) {
	testNs := "test-ns"

	ds1 := testhelpers.NewReportDataSource("datasource1", testNs)
	ds2 := testhelpers.NewReportDataSource("datasource2", testNs)
	ds3 := testhelpers.NewReportDataSource("datasource3", testNs)
	ds4 := testhelpers.NewReportDataSource("datasource4", testNs)
	ds5 := testhelpers.NewReportDataSource("datasource5", testNs)
	ds6 := testhelpers.NewReportDataSource("datasource6", testNs)

	testQuery := &metering.ReportGenerationQuery{
		ObjectMeta: meta.ObjectMeta{
			Name:      "test-query",
			Namespace: testNs,
		},
		Spec: metering.ReportGenerationQuerySpec{
			DataSources: []string{
				"datasource1",
				"datasource2",
			},
			ReportQueries: []string{
				"query1",
			},
			DynamicReportQueries: []string{
				"dynamicquery1",
				"dynamicquery2",
			},
		},
	}

	query1 := &metering.ReportGenerationQuery{
		ObjectMeta: meta.ObjectMeta{
			Name:      "query1",
			Namespace: testNs,
		},
		Spec: metering.ReportGenerationQuerySpec{
			DataSources: []string{
				"datasource3",
			},
			ReportQueries: []string{
				"query2",
			},
		},
	}

	dynamicquery1 := &metering.ReportGenerationQuery{
		ObjectMeta: meta.ObjectMeta{
			Name:      "dynamicquery1",
			Namespace: testNs,
		},
		Spec: metering.ReportGenerationQuerySpec{
			DataSources: []string{
				"datasource4",
			},
		},
	}

	query2 := &metering.ReportGenerationQuery{
		ObjectMeta: meta.ObjectMeta{
			Name:      "query2",
			Namespace: testNs,
		},
		Spec: metering.ReportGenerationQuerySpec{
			DataSources: []string{
				"datasource5",
			},
		},
	}

	dynamicquery2 := &metering.ReportGenerationQuery{
		ObjectMeta: meta.ObjectMeta{
			Name:      "dynamicquery2",
			Namespace: testNs,
		},
		Spec: metering.ReportGenerationQuerySpec{
			ReportQueries: []string{
				"query3",
			},
		},
	}

	query3 := &metering.ReportGenerationQuery{
		ObjectMeta: meta.ObjectMeta{
			Name:      "query3",
			Namespace: testNs,
		},
		Spec: metering.ReportGenerationQuerySpec{
			DataSources: []string{
				"datasource6",
			},
		},
	}

	expectedDeps := &ReportGenerationQueryDependencies{
		ReportDataSources: []*metering.ReportDataSource{
			ds1, ds2, ds3, ds4, ds5, ds6,
		},
		ReportGenerationQueries: []*metering.ReportGenerationQuery{
			query1, query2, query3,
		},
		DynamicReportGenerationQueries: []*metering.ReportGenerationQuery{
			dynamicquery1, dynamicquery2,
		},
		Reports: []*metering.Report{},
	}

	dataSourceGetter := testhelpers.NewReportDataSourceStore([]*metering.ReportDataSource{
		ds1, ds2, ds3, ds4, ds5, ds6,
	})
	queryGetter := testhelpers.NewReportGenerationQueryStore([]*metering.ReportGenerationQuery{
		testQuery, query1, query2, query3, dynamicquery1, dynamicquery2,
	})
	reportGetter := testhelpers.NewReportStore(nil)

	deps, err := GetGenerationQueryDependencies(
		queryGetter,
		dataSourceGetter,
		reportGetter,
		testQuery,
	)
	require.NoError(t, err)
	require.Equal(t, expectedDeps, deps)
}
