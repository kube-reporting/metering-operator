package reporting

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	metering "github.com/operator-framework/operator-metering/pkg/apis/metering/v1"
)

const maxDepth = 50

type ReportQueryDependencies struct {
	ReportQueries     []*metering.ReportQuery
	ReportDataSources []*metering.ReportDataSource
	Reports           []*metering.Report
}

type DependencyResolutionResult struct {
	Dependencies *ReportQueryDependencies
	InputValues  map[string]interface{}
}

type DependencyResolver struct {
	queryGetter      ReportQueryGetter
	dataSourceGetter ReportDataSourceGetter
	reportGetter     ReportGetter
}

func NewDependencyResolver(
	queryGetter ReportQueryGetter,
	dataSourceGetter ReportDataSourceGetter,
	reportGetter ReportGetter) *DependencyResolver {

	return &DependencyResolver{
		queryGetter:      queryGetter,
		dataSourceGetter: dataSourceGetter,
		reportGetter:     reportGetter,
	}
}

func (resolver *DependencyResolver) ResolveDependencies(namespace string, inputDefs []metering.ReportQueryInputDefinition, inputVals []metering.ReportQueryInputValue) (*DependencyResolutionResult, error) {
	resolverCtx := &resolverContext{
		reportAccumulator:     make(map[string]*metering.Report),
		queryAccumulator:      make(map[string]*metering.ReportQuery),
		datasourceAccumulator: make(map[string]*metering.ReportDataSource),
		inputValues:           make(map[string]interface{}),
	}
	err := resolver.resolveDependencies(namespace, resolverCtx, inputDefs, inputVals, 0, maxDepth)
	if err != nil {
		return nil, err
	}

	deps := &ReportQueryDependencies{
		ReportQueries:     make([]*metering.ReportQuery, 0, len(resolverCtx.queryAccumulator)),
		ReportDataSources: make([]*metering.ReportDataSource, 0, len(resolverCtx.datasourceAccumulator)),
		Reports:           make([]*metering.Report, 0, len(resolverCtx.reportAccumulator)),
	}

	for _, datasource := range resolverCtx.datasourceAccumulator {
		deps.ReportDataSources = append(deps.ReportDataSources, datasource)
	}
	for _, query := range resolverCtx.queryAccumulator {
		deps.ReportQueries = append(deps.ReportQueries, query)
	}
	for _, report := range resolverCtx.reportAccumulator {
		deps.Reports = append(deps.Reports, report)
	}

	sort.Slice(deps.ReportDataSources, func(i, j int) bool {
		return deps.ReportDataSources[i].Name < deps.ReportDataSources[j].Name
	})
	sort.Slice(deps.ReportQueries, func(i, j int) bool {
		return deps.ReportQueries[i].Name < deps.ReportQueries[j].Name
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
	queryAccumulator      map[string]*metering.ReportQuery
	datasourceAccumulator map[string]*metering.ReportDataSource
	inputValues           map[string]interface{}
}

func (resolver *DependencyResolver) resolveDependencies(namespace string, resolverCtx *resolverContext, inputDefs []metering.ReportQueryInputDefinition, inputVals []metering.ReportQueryInputValue, depth, maxDepth int) error {
	if depth >= maxDepth {
		return fmt.Errorf("detected a cycle at depth %d", depth)
	}
	depth += 1

	var supportedInputs []string
	for _, def := range inputDefs {
		supportedInputs = append(supportedInputs, def.Name)
	}

	givenInputs := make(map[string]metering.ReportQueryInputValue)
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

		// unmarshal the data based on the input definition type
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
		case "reportquery":
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

func (resolver *DependencyResolver) resolveQuery(namespace string, resolverCtx *resolverContext, inputVals []metering.ReportQueryInputValue, queryName string, depth, maxDepth int) error {
	if _, exists := resolverCtx.queryAccumulator[queryName]; exists {
		return nil
	}
	// fetch the query
	query, err := resolver.queryGetter.GetReportQuery(namespace, queryName)
	if err != nil {
		return err
	}
	// Resolve the dependencies of the reportQuery.
	// We pass nil for the inputValues to resolverDependencies to avoid cycles.
	err = resolver.resolveDependencies(namespace, resolverCtx, query.Spec.Inputs, nil, depth, maxDepth)
	if err != nil {
		return err
	}
	resolverCtx.queryAccumulator[query.Name] = query
	return nil
}

func (resolver *DependencyResolver) resolveDataSource(namespace string, resolverCtx *resolverContext, inputVals []metering.ReportQueryInputValue, dsName string, depth, maxDepth int) error {
	if _, exists := resolverCtx.datasourceAccumulator[dsName]; exists {
		return nil
	}
	// fetch the datasource
	datasource, err := resolver.dataSourceGetter.GetReportDataSource(namespace, dsName)
	if err != nil {
		return err
	}
	// if the datasource is a Query datasource, lookup the query it
	// depends on and resolve it's dependencies
	if datasource.Spec.ReportQueryView != nil {
		err = resolver.resolveQuery(namespace, resolverCtx, inputVals, datasource.Spec.ReportQueryView.QueryName, depth, maxDepth)
		if err != nil {
			return err
		}
	}
	resolverCtx.datasourceAccumulator[datasource.Name] = datasource
	return nil
}

func (resolver *DependencyResolver) resolveReport(namespace string, resolverCtx *resolverContext, inputVals []metering.ReportQueryInputValue, reportName string, depth, maxDepth int) error {
	if _, exists := resolverCtx.reportAccumulator[reportName]; exists {
		return nil
	}
	// this input refers to a report, so fetch the report
	report, err := resolver.reportGetter.GetReport(namespace, reportName)
	if err != nil {
		return err
	}
	err = resolver.resolveQuery(namespace, resolverCtx, inputVals, report.Spec.QueryName, depth, maxDepth)
	if err != nil {
		return err
	}
	resolverCtx.reportAccumulator[report.Name] = report
	return nil
}
