package operator

import (
	"fmt"
	"sort"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"

	cbTypes "github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1"
	"github.com/operator-framework/operator-metering/pkg/operator/reporting"
)

const (
	ReportingStartInputName = "ReportingStart"
	ReportingEndInputName   = "ReportingEnd"
)

func (op *Reporting) generateReport(logger log.FieldLogger, report runtime.Object, reportKind, reportName, tableName string, reportStart, reportEnd *time.Time, inputs []cbTypes.ReportGenerationQueryInputValue, generationQuery *cbTypes.ReportGenerationQuery, deleteExistingData bool) error {
	logger = logger.WithFields(log.Fields{
		"reportKind":         reportKind,
		"deleteExistingData": deleteExistingData,
	})
	logger.Infof("generating usage report")

	depsStatus, err := reporting.GetGenerationQueryDependenciesStatus(
		reporting.NewReportGenerationQueryListerGetter(op.reportGenerationQueryLister),
		reporting.NewReportDataSourceListerGetter(op.reportDataSourceLister),
		reporting.NewReportListerGetter(op.reportLister),
		reporting.NewScheduledReportListerGetter(op.scheduledReportLister),
		generationQuery,
	)
	if err != nil {
		return fmt.Errorf("unable to generateReport for %s %s, ReportGenerationQuery %s, failed to get dependencies: %v", reportKind, reportName, generationQuery.Name, err)
	}
	validateResults, err := op.validateDependencyStatus(depsStatus)
	if err != nil {
		return fmt.Errorf("unable to generateReport for %s %s, ReportGenerationQuery %s, failed to validate dependencies: %v", reportKind, reportName, generationQuery.Name, err)
	}

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
				return fmt.Errorf("%s spec.inputs Name: %s is not a valid timestamp: %s, must be RFC3339 formatted, err: %s", reportKind, v.Name, v.Value, err)
			}
			reportQueryInputs[v.Name] = tVal
		} else {
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
		return fmt.Errorf("unable to generateReport for %s %s, ReportGenerationQuery %s requires %s as inputs, but we have %s", reportKind, reportName, generationQuery.Name, strings.Join(expectedInputs, ","), strings.Join(givenInputs, ","))
	}

	templateInfo := &templateInfo{
		DynamicDependentQueries: validateResults.DynamicReportGenerationQueries,
		Report: &reportTemplateInfo{
			ReportingStart: reportStart,
			ReportingEnd:   reportEnd,
			Inputs:         reportQueryInputs,
		},
	}
	qr := queryRenderer{templateInfo: templateInfo}
	query, err := qr.Render(generationQuery.Spec.Query)
	if err != nil {
		return err
	}

	switch strings.ToLower(reportKind) {
	case "report", "scheduledreport":
		// valid
	default:
		return fmt.Errorf("invalid report kind: %s", reportKind)
	}

	if deleteExistingData {
		logger.Debugf("deleting any preexisting rows in %s", tableName)
		err = op.reportResultsRepo.DeleteReportResults(tableName)
		if err != nil {
			return fmt.Errorf("couldn't empty table %s of preexisting rows: %v", tableName, err)
		}
	}

	// Run the report
	logger.Debugf("running report generation query")
	err = op.reportResultsRepo.StoreReportResults(tableName, query)
	if err != nil {
		logger.WithError(err).Errorf("creating usage report FAILED!")
		return fmt.Errorf("Failed to execute %s usage report: %v", reportName, err)
	}

	return nil
}
