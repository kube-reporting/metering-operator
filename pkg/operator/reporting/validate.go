package reporting

import (
	"fmt"
	"sort"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	metering "github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1"
	meteringClient "github.com/operator-framework/operator-metering/pkg/generated/clientset/versioned/typed/metering/v1alpha1"
	meteringListers "github.com/operator-framework/operator-metering/pkg/generated/listers/metering/v1alpha1"
)

const maxDepth = 100

type ReportGenerationQueryDependencies struct {
	ReportGenerationQueries        []*metering.ReportGenerationQuery
	DynamicReportGenerationQueries []*metering.ReportGenerationQuery
	ReportDataSources              []*metering.ReportDataSource
	Reports                        []*metering.Report
}

func GetAndValidateGenerationQueryDependencies(
	queryGetter reportGenerationQueryGetter,
	dataSourceGetter reportDataSourceGetter,
	reportGetter reportGetter,
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
	HandleUninitializedReportGenerationQuery func(*metering.ReportGenerationQuery)
	HandleUninitializedReportDataSource      func(*metering.ReportDataSource)
}

func ValidateGenerationQueryDependencies(deps *ReportGenerationQueryDependencies, handler *UninitialiedDependendenciesHandler) error {
	// if the speci***REMOVED***ed ReportGenerationQuery depends on other non-dynamic
	// ReportGenerationQueries, but they have their view disabled, then it's an
	// invalid con***REMOVED***guration.
	var (
		uninitializedQueries     []*metering.ReportGenerationQuery
		uninitializedDataSources []*metering.ReportDataSource
	)
	validationErr := new(reportGenerationQueryDependenciesValidationError)
	for _, query := range deps.ReportGenerationQueries {
		// it's invalid for a ReportGenerationQuery with view.disabled set to
		// true to be a non-dynamic ReportGenerationQuery dependency
		if query.Spec.View.Disabled {
			validationErr.disabledViewQueryNames = append(validationErr.disabledViewQueryNames, query.Name)
			continue
		}
		// if a query doesn't disable view creation, than it is
		// uninitialized if it's view is not created/set yet
		if !query.Spec.View.Disabled && query.Status.ViewName == "" {
			uninitializedQueries = append(uninitializedQueries, query)
			validationErr.uninitializedQueryNames = append(validationErr.uninitializedQueryNames, query.Name)
		}
	}
	// anything below missing tableName in it's status is uninitialized
	for _, ds := range deps.ReportDataSources {
		if ds.Status.TableName == "" {
			uninitializedDataSources = append(uninitializedDataSources, ds)
			validationErr.uninitializedDataSourceNames = append(validationErr.uninitializedDataSourceNames, ds.Name)
		}
	}
	for _, report := range deps.Reports {
		if report.Status.TableName == "" {
			validationErr.uninitializedReportNames = append(validationErr.uninitializedReportNames, report.Name)
		}
	}

	if handler != nil {
		for _, query := range uninitializedQueries {
			handler.HandleUninitializedReportGenerationQuery(query)
		}

		for _, dataSource := range uninitializedDataSources {
			handler.HandleUninitializedReportDataSource(dataSource)
		}
	}

	if len(validationErr.disabledViewQueryNames) != 0 ||
		len(validationErr.uninitializedQueryNames) != 0 ||
		len(validationErr.uninitializedDataSourceNames) != 0 ||
		len(validationErr.uninitializedReportNames) != 0 ||
		len(validationErr.uninitializedReportNames) != 0 {
		return validationErr
	}
	return nil
}

func IsUninitializedDependencyError(err error) bool {
	validationErr, ok := err.(*reportGenerationQueryDependenciesValidationError)
	return ok && (len(validationErr.uninitializedQueryNames) != 0 ||
		len(validationErr.uninitializedDataSourceNames) != 0 ||
		len(validationErr.uninitializedReportNames) != 0 ||
		len(validationErr.uninitializedReportNames) != 0)
}

func IsInvalidDependencyError(err error) bool {
	validationErr, ok := err.(*reportGenerationQueryDependenciesValidationError)
	return ok && len(validationErr.disabledViewQueryNames) != 0
}

type reportGenerationQueryDependenciesValidationError struct {
	uninitializedQueryNames,
	disabledViewQueryNames,
	uninitializedDataSourceNames,
	uninitializedReportNames []string
}

func (e *reportGenerationQueryDependenciesValidationError) Error() string {
	var errs []string
	if len(e.uninitializedDataSourceNames) != 0 {
		errs = append(errs, fmt.Sprintf("uninitialized ReportDataSource dependencies: %s", strings.Join(e.uninitializedDataSourceNames, ", ")))
	}
	if len(e.disabledViewQueryNames) != 0 {
		errs = append(errs, fmt.Sprintf("invalid ReportGenerationQuery dependencies (disabled view): %s", strings.Join(e.disabledViewQueryNames, ", ")))
	}
	if len(e.uninitializedQueryNames) != 0 {
		errs = append(errs, fmt.Sprintf("uninitialized ReportGenerationQuery dependencies: %s", strings.Join(e.uninitializedQueryNames, ", ")))
	}
	if len(e.uninitializedReportNames) != 0 {
		errs = append(errs, fmt.Sprintf("uninitialized Report dependencies: %s", strings.Join(e.uninitializedReportNames, ", ")))
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
	queryGetter reportGenerationQueryGetter,
	dataSourceGetter reportDataSourceGetter,
	reportGetter reportGetter,
	generationQuery *metering.ReportGenerationQuery,
) (*ReportGenerationQueryDependencies, error) {
	dataSourceDeps, err := GetDependentDataSources(dataSourceGetter, generationQuery)
	if err != nil {
		return nil, err
	}
	viewQueries, viewQueriesDataSources, err := GetDependentViewGenerationQueries(queryGetter, dataSourceGetter, generationQuery)
	if err != nil {
		return nil, err
	}
	dynamicQueries, dynamicQueriesDataSources, err := GetDependentDynamicGenerationQueries(queryGetter, dataSourceGetter, generationQuery)
	if err != nil {
		return nil, err
	}

	allDataSources := [][]*metering.ReportDataSource{
		dataSourceDeps,
		viewQueriesDataSources,
		dynamicQueriesDataSources,
	}

	// deduplicate the list of ReportDataSources
	seen := make(map[string]struct{})
	var dataSources []*metering.ReportDataSource
	for _, dsList := range allDataSources {
		for _, ds := range dsList {
			if _, exists := seen[ds.Name]; exists {
				continue
			}
			dataSources = append(dataSources, ds)
			seen[ds.Name] = struct{}{}
		}
	}

	reports, err := GetDependentReports(reportGetter, generationQuery)
	if err != nil {
		return nil, err
	}

	return &ReportGenerationQueryDependencies{
		ReportGenerationQueries:        viewQueries,
		DynamicReportGenerationQueries: dynamicQueries,
		ReportDataSources:              dataSources,
		Reports:                        reports,
	}, nil
}

func GetDependentViewGenerationQueries(queryGetter reportGenerationQueryGetter, dataSourceGetter reportDataSourceGetter, generationQuery *metering.ReportGenerationQuery) ([]*metering.ReportGenerationQuery, []*metering.ReportDataSource, error) {
	viewReportQueriesAccumulator := make(map[string]*metering.ReportGenerationQuery)
	dataSourcesAccumulator := make(map[string]*metering.ReportDataSource)
	err := GetDependentGenerationQueriesWithDataSourcesMemoized(queryGetter, dataSourceGetter, generationQuery, 0, maxDepth, viewReportQueriesAccumulator, dataSourcesAccumulator, false)
	if err != nil {
		return nil, nil, err
	}

	viewQueries := make([]*metering.ReportGenerationQuery, 0, len(viewReportQueriesAccumulator))
	for _, query := range viewReportQueriesAccumulator {
		viewQueries = append(viewQueries, query)
	}
	dataSources := make([]*metering.ReportDataSource, 0, len(dataSourcesAccumulator))
	for _, ds := range dataSourcesAccumulator {
		dataSources = append(dataSources, ds)
	}

	return viewQueries, dataSources, nil
}

func GetDependentDynamicGenerationQueries(queryGetter reportGenerationQueryGetter, dataSourceGetter reportDataSourceGetter, generationQuery *metering.ReportGenerationQuery) ([]*metering.ReportGenerationQuery, []*metering.ReportDataSource, error) {
	dynamicReportQueriesAccumulator := make(map[string]*metering.ReportGenerationQuery)
	dataSourcesAccumulator := make(map[string]*metering.ReportDataSource)
	err := GetDependentGenerationQueriesWithDataSourcesMemoized(queryGetter, dataSourceGetter, generationQuery, 0, maxDepth, dynamicReportQueriesAccumulator, dataSourcesAccumulator, true)
	if err != nil {
		return nil, nil, err
	}

	dynamicQueries := make([]*metering.ReportGenerationQuery, 0, len(dynamicReportQueriesAccumulator))
	for _, query := range dynamicReportQueriesAccumulator {
		dynamicQueries = append(dynamicQueries, query)
	}

	dataSources := make([]*metering.ReportDataSource, 0, len(dataSourcesAccumulator))
	for _, ds := range dataSourcesAccumulator {
		dataSources = append(dataSources, ds)
	}

	return dynamicQueries, dataSources, nil
}

type reportGenerationQueryGetter interface {
	getReportGenerationQuery(namespace, name string) (*metering.ReportGenerationQuery, error)
}

type reportGenerationQueryGetterFunc func(string, string) (*metering.ReportGenerationQuery, error)

func (f reportGenerationQueryGetterFunc) getReportGenerationQuery(namespace, name string) (*metering.ReportGenerationQuery, error) {
	return f(namespace, name)
}

func NewReportGenerationQueryListerGetter(lister meteringListers.ReportGenerationQueryLister) reportGenerationQueryGetter {
	return reportGenerationQueryGetterFunc(func(namespace, name string) (*metering.ReportGenerationQuery, error) {
		return lister.ReportGenerationQueries(namespace).Get(name)
	})
}

func NewReportGenerationQueryClientGetter(getter meteringClient.ReportGenerationQueriesGetter) reportGenerationQueryGetter {
	return reportGenerationQueryGetterFunc(func(namespace, name string) (*metering.ReportGenerationQuery, error) {
		return getter.ReportGenerationQueries(namespace).Get(name, metav1.GetOptions{})
	})
}

func GetDependentGenerationQueriesWithDataSourcesMemoized(queryGetter reportGenerationQueryGetter, dataSourceGetter reportDataSourceGetter, generationQuery *metering.ReportGenerationQuery, depth, maxDepth int, queriesAccumulator map[string]*metering.ReportGenerationQuery, dataSourceAccumulator map[string]*metering.ReportDataSource, dynamicQueries bool) error {
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
		genQuery, err := queryGetter.getReportGenerationQuery(generationQuery.Namespace, queryName)
		if err != nil {
			return err
		}
		// get dependent ReportDataSources
		err = GetDependentDataSourcesMemoized(dataSourceGetter, genQuery, dataSourceAccumulator)
		if err != nil {
			return err
		}
		err = GetDependentGenerationQueriesWithDataSourcesMemoized(queryGetter, dataSourceGetter, genQuery, depth+1, maxDepth, queriesAccumulator, dataSourceAccumulator, dynamicQueries)
		if err != nil {
			return err
		}
		queriesAccumulator[genQuery.Name] = genQuery
	}
	return nil
}

func GetDependentGenerationQueriesMemoized(queryGetter reportGenerationQueryGetter, generationQuery *metering.ReportGenerationQuery, depth, maxDepth int, queriesAccumulator map[string]*metering.ReportGenerationQuery, dynamicQueries bool) error {
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
		genQuery, err := queryGetter.getReportGenerationQuery(generationQuery.Namespace, queryName)
		if err != nil {
			return err
		}
		err = GetDependentGenerationQueriesMemoized(queryGetter, genQuery, depth+1, maxDepth, queriesAccumulator, dynamicQueries)
		if err != nil {
			return err
		}
		queriesAccumulator[genQuery.Name] = genQuery
	}
	return nil
}

type reportDataSourceGetter interface {
	getReportDataSource(namespace, name string) (*metering.ReportDataSource, error)
}

type reportDataSourceGetterFunc func(string, string) (*metering.ReportDataSource, error)

func (f reportDataSourceGetterFunc) getReportDataSource(namespace, name string) (*metering.ReportDataSource, error) {
	return f(namespace, name)
}

func NewReportDataSourceListerGetter(lister meteringListers.ReportDataSourceLister) reportDataSourceGetter {
	return reportDataSourceGetterFunc(func(namespace, name string) (*metering.ReportDataSource, error) {
		return lister.ReportDataSources(namespace).Get(name)
	})
}

func NewReportDataSourceClientGetter(getter meteringClient.ReportDataSourcesGetter) reportDataSourceGetter {
	return reportDataSourceGetterFunc(func(namespace, name string) (*metering.ReportDataSource, error) {
		return getter.ReportDataSources(namespace).Get(name, metav1.GetOptions{})
	})
}

func GetDependentDataSourcesMemoized(dataSourceGetter reportDataSourceGetter, generationQuery *metering.ReportGenerationQuery, dataSourceAccumulator map[string]*metering.ReportDataSource) error {
	for _, dataSourceName := range generationQuery.Spec.DataSources {
		if _, exists := dataSourceAccumulator[dataSourceName]; exists {
			continue
		}
		dataSource, err := dataSourceGetter.getReportDataSource(generationQuery.Namespace, dataSourceName)
		if err != nil {
			return err
		}
		dataSourceAccumulator[dataSource.Name] = dataSource
	}
	return nil
}

func GetDependentDataSources(dataSourceGetter reportDataSourceGetter, generationQuery *metering.ReportGenerationQuery) ([]*metering.ReportDataSource, error) {
	dataSourceAccumulator := make(map[string]*metering.ReportDataSource)
	err := GetDependentDataSourcesMemoized(dataSourceGetter, generationQuery, dataSourceAccumulator)
	if err != nil {
		return nil, err
	}
	dataSources := make([]*metering.ReportDataSource, 0, len(dataSourceAccumulator))
	for _, ds := range dataSourceAccumulator {
		dataSources = append(dataSources, ds)
	}
	return dataSources, nil
}

type reportGetter interface {
	getReport(namespace, name string) (*metering.Report, error)
}

type reportGetterFunc func(string, string) (*metering.Report, error)

func (f reportGetterFunc) getReport(namespace, name string) (*metering.Report, error) {
	return f(namespace, name)
}

func NewReportListerGetter(lister meteringListers.ReportLister) reportGetter {
	return reportGetterFunc(func(namespace, name string) (*metering.Report, error) {
		return lister.Reports(namespace).Get(name)
	})
}

func NewReportClientGetter(getter meteringClient.ReportsGetter) reportGetter {
	return reportGetterFunc(func(namespace, name string) (*metering.Report, error) {
		return getter.Reports(namespace).Get(name, metav1.GetOptions{})
	})
}

func GetDependentReports(reportGetter reportGetter, generationQuery *metering.ReportGenerationQuery) ([]*metering.Report, error) {
	reports := make([]*metering.Report, len(generationQuery.Spec.Reports))
	for i, reportName := range generationQuery.Spec.Reports {
		report, err := reportGetter.getReport(generationQuery.Namespace, reportName)
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
	for _, v := range inputs {
		// currently inputs can only have string values, but we want to support
		// other types in the future.
		// To support overriding the default ReportingStart and ReportingEnd
		// using inputs, we have to treat them specially and turn them into
		// time.Time objects before passing to the template context.
		if v.Name == ReportingStartInputName || v.Name == ReportingEndInputName {
			tVal, err := time.Parse(time.RFC3339, v.Value)
			if err != nil {
				return nil, fmt.Errorf("inputs Name: %s is not a valid timestamp: %s, must be RFC3339 formatted, err: %s", v.Name, v.Value, err)
			}
			reportQueryInputs[v.Name] = tVal
		} ***REMOVED*** {
			reportQueryInputs[v.Name] = v.Value
		}
		givenInputs = append(givenInputs, v.Name)
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
