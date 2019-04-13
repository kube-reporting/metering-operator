package reporting

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	metering "github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1"
)

const maxDepth = 50

type ReportGenerationQueryDependencies struct {
	ReportGenerationQueries []*metering.ReportGenerationQuery
	ReportDataSources       []*metering.ReportDataSource
	Reports                 []*metering.Report
}

type DependencyResolutionResult struct {
	Dependencies *ReportGenerationQueryDependencies
	InputValues  map[string]interface{}
}

type DependencyResolver struct {
	queryGetter      ReportGenerationQueryGetter
	dataSourceGetter ReportDataSourceGetter
	reportGetter     ReportGetter
}

func NewDependencyResolver(
	queryGetter ReportGenerationQueryGetter,
	dataSourceGetter ReportDataSourceGetter,
	reportGetter ReportGetter) *DependencyResolver {

	return &DependencyResolver{
		queryGetter:      queryGetter,
		dataSourceGetter: dataSourceGetter,
		reportGetter:     reportGetter,
	}
}

func (resolver *DependencyResolver) ResolveDependencies(namespace string, inputDefs []metering.ReportGenerationQueryInputDe***REMOVED***nition, inputVals []metering.ReportGenerationQueryInputValue) (*DependencyResolutionResult, error) {
	resolverCtx := &resolverContext{
		reportAccumulator:     make(map[string]*metering.Report),
		queryAccumulator:      make(map[string]*metering.ReportGenerationQuery),
		datasourceAccumulator: make(map[string]*metering.ReportDataSource),
		inputValues:           make(map[string]interface{}),
	}
	err := resolver.resolveDependencies(namespace, resolverCtx, inputDefs, inputVals, 0, maxDepth)
	if err != nil {
		return nil, err
	}

	deps := &ReportGenerationQueryDependencies{
		ReportGenerationQueries: make([]*metering.ReportGenerationQuery, 0, len(resolverCtx.queryAccumulator)),
		ReportDataSources:       make([]*metering.ReportDataSource, 0, len(resolverCtx.datasourceAccumulator)),
		Reports:                 make([]*metering.Report, 0, len(resolverCtx.reportAccumulator)),
	}

	for _, datasource := range resolverCtx.datasourceAccumulator {
		deps.ReportDataSources = append(deps.ReportDataSources, datasource)
	}
	for _, query := range resolverCtx.queryAccumulator {
		deps.ReportGenerationQueries = append(deps.ReportGenerationQueries, query)
	}
	for _, report := range resolverCtx.reportAccumulator {
		deps.Reports = append(deps.Reports, report)
	}

	sort.Slice(deps.ReportDataSources, func(i, j int) bool {
		return deps.ReportDataSources[i].Name < deps.ReportDataSources[j].Name
	})
	sort.Slice(deps.ReportGenerationQueries, func(i, j int) bool {
		return deps.ReportGenerationQueries[i].Name < deps.ReportGenerationQueries[j].Name
	})
	sort.Slice(deps.Reports, func(i, j int) bool {
		return deps.Reports[i].Name < deps.Reports[j].Name
	})

	return &DependencyResolutionResult{
		Dependencies: deps,
		InputValues:  resolverCtx.inputValues,
	}, nil
}

type resolverContext struct {
	reportAccumulator     map[string]*metering.Report
	queryAccumulator      map[string]*metering.ReportGenerationQuery
	datasourceAccumulator map[string]*metering.ReportDataSource
	inputValues           map[string]interface{}
}

func (resolver *DependencyResolver) resolveDependencies(namespace string, resolverCtx *resolverContext, inputDefs []metering.ReportGenerationQueryInputDe***REMOVED***nition, inputVals []metering.ReportGenerationQueryInputValue, depth, maxDepth int) error {
	if depth >= maxDepth {
		return fmt.Errorf("detected a cycle at depth %d", depth)
	}
	depth += 1

	var supportedInputs []string
	for _, def := range inputDefs {
		supportedInputs = append(supportedInputs, def.Name)
	}

	givenInputs := make(map[string]metering.ReportGenerationQueryInputValue)
	for _, val := range inputVals {
		var seen bool
		for _, def := range inputDefs {
			if def.Name == val.Name {
				seen = true
				break
			}
		}
		if !seen {
			return fmt.Errorf("invalid input %q, supported inputs: %s", val.Name, strings.Join(supportedInputs, " ,"))
		}
		givenInputs[val.Name] = val
	}

	for _, def := range inputDefs {
		// already resolved
		if _, exists := resolverCtx.inputValues[def.Name]; exists {
			continue
		}

		inputVal := givenInputs[def.Name].Value
		// use the default value if it exists
		if inputVal == nil && def.Default != nil {
			inputVal = def.Default
		}

		if inputVal == nil {
			continue
		}

		inputType := strings.ToLower(def.Type)
		if def.Name == ReportingStartInputName || def.Name == ReportingEndInputName {
			inputType = "time"
		}

		// unmarshal the data based on the input de***REMOVED***nition type
		var dst interface{}
		var err error
		switch inputType {
		case "", "string":
			dst = new(string)
			err = json.Unmarshal(*inputVal, dst)
		case "time":
			dst = new(time.Time)
			err = json.Unmarshal(*inputVal, dst)
		case "int", "integer":
			dst = new(int)
			err = json.Unmarshal(*inputVal, dst)
		case "reportdatasource":
			dst = new(string)
			err = json.Unmarshal(*inputVal, dst)
			if err == nil {
				name := dst.(*string)
				if name != nil {
					err = resolver.resolveDataSource(namespace, resolverCtx, inputVals, *name, depth, maxDepth)
					if err != nil {
						return err
					}
				}
			}
		case "reportgenerationquery":
			dst = new(string)
			err = json.Unmarshal(*inputVal, dst)
			if err == nil {
				name := dst.(*string)
				if name != nil {
					err = resolver.resolveQuery(namespace, resolverCtx, inputVals, *name, depth, maxDepth)
					if err != nil {
						return err
					}
				}
			}
		case "report":
			dst = new(string)
			err = json.Unmarshal(*inputVal, dst)
			if err == nil {
				name := dst.(*string)
				if name != nil {
					err = resolver.resolveReport(namespace, resolverCtx, inputVals, *name, depth, maxDepth)
					if err != nil {
						return err
					}
				}
			}
		default:
			return fmt.Errorf("unsupported input type %s", inputType)
		}
		if err != nil {
			return fmt.Errorf("inputs Name: %s is not valid a '%s': value: '%s', err: %s", def.Name, inputType, string(*inputVal), err)
		}
		resolverCtx.inputValues[def.Name] = dst
	}
	return nil
}

func (resolver *DependencyResolver) resolveQuery(namespace string, resolverCtx *resolverContext, inputVals []metering.ReportGenerationQueryInputValue, queryName string, depth, maxDepth int) error {
	if _, exists := resolverCtx.queryAccumulator[queryName]; exists {
		return nil
	}
	// fetch the query
	query, err := resolver.queryGetter.GetReportGenerationQuery(namespace, queryName)
	if err != nil {
		return err
	}
	// Resolve the dependencies of the reportGenerationQuery.
	// We pass nil for the inputValues to resolverDependencies to avoid cycles.
	err = resolver.resolveDependencies(namespace, resolverCtx, query.Spec.Inputs, nil, depth, maxDepth)
	if err != nil {
		return err
	}
	resolverCtx.queryAccumulator[query.Name] = query
	return nil
}

func (resolver *DependencyResolver) resolveDataSource(namespace string, resolverCtx *resolverContext, inputVals []metering.ReportGenerationQueryInputValue, dsName string, depth, maxDepth int) error {
	if _, exists := resolverCtx.datasourceAccumulator[dsName]; exists {
		return nil
	}
	// fetch the datasource
	datasource, err := resolver.dataSourceGetter.GetReportDataSource(namespace, dsName)
	if err != nil {
		return err
	}
	// if the datasource is a GenerationQuery datasource, lookup the query it
	// depends on and resolve it's dependencies
	if datasource.Spec.GenerationQueryView != nil {
		err = resolver.resolveQuery(namespace, resolverCtx, inputVals, datasource.Spec.GenerationQueryView.QueryName, depth, maxDepth)
		if err != nil {
			return err
		}
	}
	resolverCtx.datasourceAccumulator[datasource.Name] = datasource
	return nil
}

func (resolver *DependencyResolver) resolveReport(namespace string, resolverCtx *resolverContext, inputVals []metering.ReportGenerationQueryInputValue, reportName string, depth, maxDepth int) error {
	if _, exists := resolverCtx.reportAccumulator[reportName]; exists {
		return nil
	}
	// this input refers to a report, so fetch the report
	report, err := resolver.reportGetter.GetReport(namespace, reportName)
	if err != nil {
		return err
	}
	err = resolver.resolveQuery(namespace, resolverCtx, inputVals, report.Spec.GenerationQueryName, depth, maxDepth)
	if err != nil {
		return err
	}
	resolverCtx.reportAccumulator[report.Name] = report
	return nil
}
