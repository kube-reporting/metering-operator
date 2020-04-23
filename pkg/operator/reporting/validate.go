package reporting

import (
	"fmt"
	"strings"

	metering "github.com/kube-reporting/metering-operator/pkg/apis/metering/v1"
)

func GetAndValidateQueryDependencies(
	queryGetter ReportQueryGetter,
	dataSourceGetter ReportDataSourceGetter,
	reportGetter ReportGetter,
	query *metering.ReportQuery,
	inputVals []metering.ReportQueryInputValue,
	handler *UninitialiedDependendenciesHandler,
) (*ReportQueryDependencies, error) {
	deps, err := GetQueryDependencies(queryGetter, dataSourceGetter, reportGetter, query, inputVals)
	if err != nil {
		return nil, err
	}
	err = ValidateQueryDependencies(deps, handler)
	if err != nil {
		return nil, err
	}
	return deps, nil
}

func GetQueryDependencies(
	queryGetter ReportQueryGetter,
	dataSourceGetter ReportDataSourceGetter,
	reportGetter ReportGetter,
	query *metering.ReportQuery,
	inputVals []metering.ReportQueryInputValue,
) (*ReportQueryDependencies, error) {
	result, err := NewDependencyResolver(queryGetter, dataSourceGetter, reportGetter).ResolveDependencies(query.Namespace, query.Spec.Inputs, inputVals)
	if err != nil {
		return nil, err
	}
	return result.Dependencies, nil
}

type UninitialiedDependendenciesHandler struct {
	HandleUninitializedReportDataSource func(*metering.ReportDataSource)
}

func ValidateQueryDependencies(deps *ReportQueryDependencies, handler *UninitialiedDependendenciesHandler) error {
	// if the specified ReportQuery depends on datasources without a
	// table, it's invalid
	var uninitializedDataSources []*metering.ReportDataSource
	validationErr := new(reportQueryDependenciesValidationError)
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
	validationErr, ok := err.(*reportQueryDependenciesValidationError)
	return ok && (len(validationErr.uninitializedDataSourceNames) != 0 || len(validationErr.uninitializedReportNames) != 0)

}

func IsInvalidDependencyError(err error) bool {
	_, ok := err.(*reportQueryDependenciesValidationError)
	return ok
}

type reportQueryDependenciesValidationError struct {
	uninitializedDataSourceNames,
	uninitializedReportNames []string
}

func (e *reportQueryDependenciesValidationError) Error() string {
	var errs []string
	if len(e.uninitializedDataSourceNames) != 0 {
		errs = append(errs, fmt.Sprintf("uninitialized ReportDataSource dependencies: %s", strings.Join(e.uninitializedDataSourceNames, ", ")))
	}
	if len(e.uninitializedReportNames) != 0 {
		errs = append(errs, fmt.Sprintf("uninitialized Report dependencies: %s", strings.Join(e.uninitializedReportNames, ", ")))
	}
	if len(errs) != 0 {
		return fmt.Sprintf("ReportQueryDependencyValidationError: %s", strings.Join(errs, ", "))
	}
	panic("zero uninitialized or invalid dependencies")
}
