package reporting

import (
	"fmt"
	"strings"

	metering "github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1"
	meteringListers "github.com/operator-framework/operator-metering/pkg/generated/listers/metering/v1alpha1"
)

const maxDepth = 100

type ReportGenerationQueryDependencies struct {
	ReportGenerationQueries        []*metering.ReportGenerationQuery
	DynamicReportGenerationQueries []*metering.ReportGenerationQuery
	ReportDataSources              []*metering.ReportDataSource
}

func ValidateGenerationQueryDependenciesStatus(depsStatus *GenerationQueryDependenciesStatus) (*ReportGenerationQueryDependencies, error) {
	// if the speci***REMOVED***ed ReportGenerationQuery depends on other non-dynamic
	// ReportGenerationQueries, but they have their view disabled, then it's an
	// invalid con***REMOVED***guration.
	var queriesViewDisabled, uninitializedQueries, uninitializedDataSources []string
	for _, query := range depsStatus.UninitializedReportGenerationQueries {
		if query.Spec.View.Disabled {
			queriesViewDisabled = append(queriesViewDisabled, query.Name)
		} ***REMOVED*** if query.ViewName == "" {
			uninitializedQueries = append(uninitializedQueries, query.Name)
		}
	}
	for _, ds := range depsStatus.UninitializedReportDataSources {
		uninitializedDataSources = append(uninitializedDataSources, ds.Name)
	}
	if len(queriesViewDisabled) != 0 {
		return nil, fmt.Errorf("invalid ReportGenerationQuery, references ReportGenerationQueries with spec.view.disabled=true: %s", strings.Join(queriesViewDisabled, ", "))
	}
	if len(uninitializedDataSources) != 0 {
		return nil, fmt.Errorf("ReportGenerationQuery has uninitialized ReportDataSource dependencies: %s", strings.Join(uninitializedDataSources, ", "))
	}
	if len(uninitializedQueries) != 0 {
		return nil, fmt.Errorf("ReportGenerationQuery has uninitialized ReportGenerationQuery dependencies: %s", strings.Join(uninitializedQueries, ", "))
	}

	return &ReportGenerationQueryDependencies{
		ReportGenerationQueries:        depsStatus.InitializedReportGenerationQueries,
		DynamicReportGenerationQueries: depsStatus.InitializedDynamicReportGenerationQueries,
		ReportDataSources:              depsStatus.InitializedReportDataSources,
	}, nil
}

type GenerationQueryDependenciesStatus struct {
	UninitializedReportGenerationQueries      []*metering.ReportGenerationQuery
	InitializedReportGenerationQueries        []*metering.ReportGenerationQuery
	InitializedDynamicReportGenerationQueries []*metering.ReportGenerationQuery

	UninitializedReportDataSources []*metering.ReportDataSource
	InitializedReportDataSources   []*metering.ReportDataSource
}

func GetGenerationQueryDependenciesStatus(reportGenerationQueryLister meteringListers.ReportGenerationQueryLister, reportDataSourceLister meteringListers.ReportDataSourceLister, generationQuery *metering.ReportGenerationQuery) (*GenerationQueryDependenciesStatus, error) {
	// Validate ReportGenerationQuery's that should be views
	dependentQueriesStatus, err := GetDependentGenerationQueries(reportGenerationQueryLister, generationQuery)
	if err != nil {
		return nil, err
	}

	dataSources, err := GetDependentDataSources(reportDataSourceLister, generationQuery)
	if err != nil {
		return nil, err
	}

	var uninitializedDataSources, initializedDataSources []*metering.ReportDataSource
	for _, dataSource := range dataSources {
		if dataSource.TableName == "" {
			uninitializedDataSources = append(uninitializedDataSources, dataSource)
		} ***REMOVED*** {
			initializedDataSources = append(initializedDataSources, dataSource)
		}
	}

	var uninitializedQueries, initializedQueries []*metering.ReportGenerationQuery
	for _, query := range dependentQueriesStatus.ViewReportGenerationQueries {
		if query.ViewName == "" {
			uninitializedQueries = append(uninitializedQueries, query)
		} ***REMOVED*** {
			initializedQueries = append(initializedQueries, query)
		}
	}

	return &GenerationQueryDependenciesStatus{
		UninitializedReportGenerationQueries:      uninitializedQueries,
		InitializedReportGenerationQueries:        initializedQueries,
		InitializedDynamicReportGenerationQueries: dependentQueriesStatus.DynamicReportGenerationQueries,
		UninitializedReportDataSources:            uninitializedDataSources,
		InitializedReportDataSources:              initializedDataSources,
	}, nil
}

type GetDependentGenerationQueriesStatus struct {
	ViewReportGenerationQueries    []*metering.ReportGenerationQuery
	DynamicReportGenerationQueries []*metering.ReportGenerationQuery
}

func GetDependentGenerationQueries(reportGenerationQueryLister meteringListers.ReportGenerationQueryLister, generationQuery *metering.ReportGenerationQuery) (*GetDependentGenerationQueriesStatus, error) {
	viewQueries, err := GetDependentViewGenerationQueries(reportGenerationQueryLister, generationQuery)
	if err != nil {
		return nil, err
	}
	dynamicQueries, err := GetDependentDynamicGenerationQueries(reportGenerationQueryLister, generationQuery)
	if err != nil {
		return nil, err
	}
	return &GetDependentGenerationQueriesStatus{
		ViewReportGenerationQueries:    viewQueries,
		DynamicReportGenerationQueries: dynamicQueries,
	}, nil
}

func GetDependentViewGenerationQueries(reportGenerationQueryLister meteringListers.ReportGenerationQueryLister, generationQuery *metering.ReportGenerationQuery) ([]*metering.ReportGenerationQuery, error) {
	viewReportQueriesAccumulator := make(map[string]*metering.ReportGenerationQuery)
	err := GetDependentGenerationQueriesMemoized(reportGenerationQueryLister, generationQuery, 0, maxDepth, viewReportQueriesAccumulator, false)
	if err != nil {
		return nil, err
	}

	viewQueries := make([]*metering.ReportGenerationQuery, 0, len(viewReportQueriesAccumulator))
	for _, query := range viewReportQueriesAccumulator {
		viewQueries = append(viewQueries, query)
	}
	return viewQueries, nil
}

func GetDependentDynamicGenerationQueries(reportGenerationQueryLister meteringListers.ReportGenerationQueryLister, generationQuery *metering.ReportGenerationQuery) ([]*metering.ReportGenerationQuery, error) {
	dynamicReportQueriesAccumulator := make(map[string]*metering.ReportGenerationQuery)
	err := GetDependentGenerationQueriesMemoized(reportGenerationQueryLister, generationQuery, 0, maxDepth, dynamicReportQueriesAccumulator, true)
	if err != nil {
		return nil, err
	}

	dynamicQueries := make([]*metering.ReportGenerationQuery, 0, len(dynamicReportQueriesAccumulator))
	for _, query := range dynamicReportQueriesAccumulator {
		dynamicQueries = append(dynamicQueries, query)
	}
	return dynamicQueries, nil
}

func GetDependentGenerationQueriesMemoized(reportGenerationQueryLister meteringListers.ReportGenerationQueryLister, generationQuery *metering.ReportGenerationQuery, depth, maxDepth int, queriesAccumulator map[string]*metering.ReportGenerationQuery, dynamicQueries bool) error {
	if depth >= maxDepth {
		return fmt.Errorf("detected a cycle at depth %d for generationQuery %s", depth, generationQuery.Name)
	}
	var queries []string
	if dynamicQueries {
		queries = generationQuery.Spec.DynamicReportQueries
	} ***REMOVED*** {
		queries = generationQuery.Spec.ReportQueries
	}
	for _, queryName := range queries {
		if _, exists := queriesAccumulator[queryName]; exists {
			continue
		}
		genQuery, err := reportGenerationQueryLister.ReportGenerationQueries(generationQuery.Namespace).Get(queryName)
		if err != nil {
			return err
		}
		err = GetDependentGenerationQueriesMemoized(reportGenerationQueryLister, genQuery, depth+1, maxDepth, queriesAccumulator, dynamicQueries)
		if err != nil {
			return err
		}
		queriesAccumulator[genQuery.Name] = genQuery
	}
	return nil
}

func GetDependentDataSources(reportDataSourceLister meteringListers.ReportDataSourceLister, generationQuery *metering.ReportGenerationQuery) ([]*metering.ReportDataSource, error) {
	dataSources := make([]*metering.ReportDataSource, len(generationQuery.Spec.DataSources))
	for i, dataSourceName := range generationQuery.Spec.DataSources {
		dataSource, err := reportDataSourceLister.ReportDataSources(generationQuery.Namespace).Get(dataSourceName)
		if err != nil {
			return nil, err
		}
		dataSources[i] = dataSource
	}
	return dataSources, nil
}
