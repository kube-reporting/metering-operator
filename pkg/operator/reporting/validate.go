package reporting

import (
	"fmt"
	"strings"

	cbTypes "github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1"
	metering "github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1"
)

func GetAndValidateGenerationQueryDependencies(
	queryGetter ReportGenerationQueryGetter,
	dataSourceGetter ReportDataSourceGetter,
	reportGetter ReportGetter,
	generationQuery *metering.ReportGenerationQuery,
	inputVals []cbTypes.ReportGenerationQueryInputValue,
	handler *UninitialiedDependendenciesHandler,
) (*ReportGenerationQueryDependencies, error) {
	deps, err := GetGenerationQueryDependencies(queryGetter, dataSourceGetter, reportGetter, generationQuery, inputVals)
	if err != nil {
		return nil, err
	}
	err = ValidateGenerationQueryDependencies(deps, handler)
	if err != nil {
		return nil, err
	}
	return deps, nil
}

func GetGenerationQueryDependencies(
	queryGetter ReportGenerationQueryGetter,
	dataSourceGetter ReportDataSourceGetter,
	reportGetter ReportGetter,
	generationQuery *metering.ReportGenerationQuery,
	inputVals []cbTypes.ReportGenerationQueryInputValue,
) (*ReportGenerationQueryDependencies, error) {
	result, err := NewDependencyResolver(queryGetter, dataSourceGetter, reportGetter).ResolveDependencies(generationQuery.Namespace, generationQuery.Spec.Inputs, inputVals)
	if err != nil {
		return nil, err
	}
	return result.Dependencies, nil
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
