package reporting

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	metering "github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1"
)

const maxDepth = 100

type ReportGenerationQueryDependencies struct {
	DynamicReportGenerationQueries []*metering.ReportGenerationQuery
	ReportDataSources              []*metering.ReportDataSource
	Reports                        []*metering.Report
}

func GetAndValidateGenerationQueryDependencies(
	queryGetter ReportGenerationQueryGetter,
	dataSourceGetter ReportDataSourceGetter,
	reportGetter ReportGetter,
	generationQuery *metering.ReportGenerationQuery,
	handler *UninitialiedDependendenciesHandler,
) (*ReportGenerationQueryDependencies, error) {
	deps, err := GetGenerationQueryDependencies(
		queryGetter,
		dataSourceGetter,
		reportGetter,
		generationQuery,
	)
	if err != nil {
		return nil, err
	}
	err = ValidateGenerationQueryDependencies(deps, handler)
	if err != nil {
		return nil, err
	}
	return deps, nil
}

type UninitialiedDependendenciesHandler struct {
	HandleUninitializedReportDataSource func(*metering.ReportDataSource)
}

func ValidateGenerationQueryDependencies(deps *ReportGenerationQueryDependencies, handler *UninitialiedDependendenciesHandler) error {
	// if the speci***REMOVED***ed ReportGenerationQuery depends on datasources without a
	// table, it's invalid
	var uninitializedDataSources []*metering.ReportDataSource
	validationErr := new(reportGenerationQueryDependenciesValidationError)
	// anything below missing tableName in it's status is uninitialized
	for _, ds := range deps.ReportDataSources {
		if ds.Status.TableRef.Name == "" {
			uninitializedDataSources = append(uninitializedDataSources, ds)
			validationErr.uninitializedDataSourceNames = append(validationErr.uninitializedDataSourceNames, ds.Name)
		}
	}
	for _, report := range deps.Reports {
		if report.Status.TableRef.Name == "" {
			validationErr.uninitializedReportNames = append(validationErr.uninitializedReportNames, report.Name)
		}
	}

	if handler != nil {
		for _, dataSource := range uninitializedDataSources {
			handler.HandleUninitializedReportDataSource(dataSource)
		}
	}

	if len(validationErr.uninitializedDataSourceNames) != 0 || len(validationErr.uninitializedReportNames) != 0 {
		return validationErr
	}
	return nil
}

func IsUninitializedDependencyError(err error) bool {
	validationErr, ok := err.(*reportGenerationQueryDependenciesValidationError)
	return ok && (len(validationErr.uninitializedDataSourceNames) != 0 || len(validationErr.uninitializedReportNames) != 0)

}

func IsInvalidDependencyError(err error) bool {
	_, ok := err.(*reportGenerationQueryDependenciesValidationError)
	return ok
}

type reportGenerationQueryDependenciesValidationError struct {
	uninitializedDataSourceNames,
	uninitializedReportNames []string
}

func (e *reportGenerationQueryDependenciesValidationError) Error() string {
	var errs []string
	if len(e.uninitializedDataSourceNames) != 0 {
		errs = append(errs, fmt.Sprintf("uninitialized ReportDataSource dependencies: %s", strings.Join(e.uninitializedDataSourceNames, ", ")))
	}
	if len(e.uninitializedReportNames) != 0 {
		errs = append(errs, fmt.Sprintf("uninitialized Report dependencies: %s", strings.Join(e.uninitializedReportNames, ", ")))
	}
	if len(errs) != 0 {
		return fmt.Sprintf("ReportGenerationQueryDependencyValidationError: %s", strings.Join(errs, ", "))
	}
	panic("zero uninitialized or invalid dependencies")
}

func GetGenerationQueryDependencies(
	queryGetter ReportGenerationQueryGetter,
	dataSourceGetter ReportDataSourceGetter,
	reportGetter ReportGetter,
	generationQuery *metering.ReportGenerationQuery,
) (*ReportGenerationQueryDependencies, error) {
	queries, dataSources, err := GetDependentGenerationQueriesAndDataSources(queryGetter, dataSourceGetter, generationQuery)
	if err != nil {
		return nil, err
	}

	reports, err := GetDependentReports(reportGetter, generationQuery)
	if err != nil {
		return nil, err
	}

	sort.Slice(queries, func(i, j int) bool {
		return queries[i].Name < queries[j].Name
	})
	sort.Slice(dataSources, func(i, j int) bool {
		return dataSources[i].Name < dataSources[j].Name
	})
	sort.Slice(reports, func(i, j int) bool {
		return reports[i].Name < reports[j].Name
	})

	return &ReportGenerationQueryDependencies{
		DynamicReportGenerationQueries: queries,
		ReportDataSources:              dataSources,
		Reports:                        reports,
	}, nil
}

func GetDependentGenerationQueriesAndDataSources(queryGetter ReportGenerationQueryGetter, dataSourceGetter ReportDataSourceGetter, generationQuery *metering.ReportGenerationQuery) ([]*metering.ReportGenerationQuery, []*metering.ReportDataSource, error) {
	dataSourcesAccumulator := make(map[string]*metering.ReportDataSource)
	queriesAccumulator := make(map[string]*metering.ReportGenerationQuery)

	err := GetDependentGenerationQueriesAndDataSourcesMemoized(queryGetter, dataSourceGetter, generationQuery, 0, maxDepth, queriesAccumulator, dataSourcesAccumulator)
	if err != nil {
		return nil, nil, err
	}

	queries := make([]*metering.ReportGenerationQuery, 0, len(queriesAccumulator))
	for _, query := range queriesAccumulator {
		queries = append(queries, query)
	}
	dataSources := make([]*metering.ReportDataSource, 0, len(dataSourcesAccumulator))
	for _, ds := range dataSourcesAccumulator {
		dataSources = append(dataSources, ds)
	}

	return queries, dataSources, nil
}

func GetDependentGenerationQueriesAndDataSourcesMemoized(queryGetter ReportGenerationQueryGetter, dataSourceGetter ReportDataSourceGetter, generationQuery *metering.ReportGenerationQuery, depth, maxDepth int, queriesAccumulator map[string]*metering.ReportGenerationQuery, dataSourceAccumulator map[string]*metering.ReportDataSource) error {
	if depth >= maxDepth {
		return fmt.Errorf("detected a cycle at depth %d for generationQuery %s", depth, generationQuery.Name)
	}
	for _, dataSourceName := range generationQuery.Spec.DataSources {
		if _, exists := dataSourceAccumulator[dataSourceName]; exists {
			continue
		}
		dataSource, err := dataSourceGetter.GetReportDataSource(generationQuery.Namespace, dataSourceName)
		if err != nil {
			return err
		}
		if dataSource.Spec.GenerationQueryView != nil {
			genQuery, err := queryGetter.GetReportGenerationQuery(generationQuery.Namespace, dataSource.Spec.GenerationQueryView.QueryName)
			if err != nil {
				return err
			}
			err = GetDependentGenerationQueriesAndDataSourcesMemoized(queryGetter, dataSourceGetter, genQuery, depth+1, maxDepth, queriesAccumulator, dataSourceAccumulator)
			if err != nil {
				return err
			}
			queriesAccumulator[genQuery.Name] = genQuery
		}
		dataSourceAccumulator[dataSource.Name] = dataSource
	}
	for _, queryName := range generationQuery.Spec.DynamicReportQueries {
		if _, exists := queriesAccumulator[queryName]; exists {
			continue
		}
		genQuery, err := queryGetter.GetReportGenerationQuery(generationQuery.Namespace, queryName)
		if err != nil {
			return err
		}
		err = GetDependentGenerationQueriesAndDataSourcesMemoized(queryGetter, dataSourceGetter, genQuery, depth+1, maxDepth, queriesAccumulator, dataSourceAccumulator)
		if err != nil {
			return err
		}
		queriesAccumulator[genQuery.Name] = genQuery
	}
	return nil
}

func GetDependentReports(reportGetter ReportGetter, generationQuery *metering.ReportGenerationQuery) ([]*metering.Report, error) {
	reports := make([]*metering.Report, len(generationQuery.Spec.Reports))
	for i, reportName := range generationQuery.Spec.Reports {
		report, err := reportGetter.GetReport(generationQuery.Namespace, reportName)
		if err != nil {
			return nil, err
		}
		reports[i] = report
	}
	return reports, nil
}

func ValidateReportGenerationQueryInputs(generationQuery *metering.ReportGenerationQuery, inputs []metering.ReportGenerationQueryInputValue) (map[string]interface{}, error) {
	var givenInputs, missingInputs, expectedInputs []string
	reportQueryInputs := make(map[string]interface{})
	inputDe***REMOVED***nitions := make(map[string]metering.ReportGenerationQueryInputDe***REMOVED***nition)

	for _, inputDef := range generationQuery.Spec.Inputs {
		inputDe***REMOVED***nitions[inputDef.Name] = inputDef
	}

	for _, inputVal := range inputs {
		inputDef := inputDe***REMOVED***nitions[inputVal.Name]
		val, err := convertQueryInputValueFromDe***REMOVED***nition(inputVal, inputDef)
		if err != nil {
			return nil, err
		}
		reportQueryInputs[inputVal.Name] = val
		givenInputs = append(givenInputs, inputVal.Name)
	}

	// now validate the inputs match what the query is expecting
	for _, input := range generationQuery.Spec.Inputs {
		expectedInputs = append(expectedInputs, input.Name)
		// If the input isn't required than don't include it in the missing
		if !input.Required {
			continue
		}
		if _, ok := reportQueryInputs[input.Name]; !ok {
			missingInputs = append(missingInputs, input.Name)
		}
	}

	if len(missingInputs) != 0 {
		sort.Strings(expectedInputs)
		sort.Strings(givenInputs)
		return nil, fmt.Errorf("unable to validate ReportGenerationQuery %s inputs: requires %s as inputs, got %s", generationQuery.Name, strings.Join(expectedInputs, ","), strings.Join(givenInputs, ","))
	}

	return reportQueryInputs, nil
}

func convertQueryInputValueFromDe***REMOVED***nition(inputVal metering.ReportGenerationQueryInputValue, inputDef metering.ReportGenerationQueryInputDe***REMOVED***nition) (interface{}, error) {
	if inputVal.Value == nil {
		return nil, nil
	}

	inputType := strings.ToLower(inputDef.Type)
	if inputVal.Name == ReportingStartInputName || inputVal.Name == ReportingEndInputName {
		inputType = "time"
	}
	// unmarshal the data based on the input de***REMOVED***nition type
	var dst interface{}
	switch inputType {
	case "", "string":
		dst = new(string)
	case "time":
		dst = new(time.Time)
	case "int", "integer":
		dst = new(int)
	default:
		return nil, fmt.Errorf("unsupported input type %s", inputType)
	}
	err := json.Unmarshal(*inputVal.Value, dst)
	if err != nil {
		return nil, fmt.Errorf("inputs Name: %s is not valid a %s: value: %s, err: %s", inputVal.Name, inputType, string(*inputVal.Value), err)
	}
	return dst, nil
}
