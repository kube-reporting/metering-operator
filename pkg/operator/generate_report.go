package operator

import (
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"

	cbTypes "github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1"
	"github.com/operator-framework/operator-metering/pkg/operator/reporting"
	"github.com/operator-framework/operator-metering/pkg/presto"
)

func (op *Reporting) generateReport(logger log.FieldLogger, report runtime.Object, reportKind, reportName, tableName string, reportStart, reportEnd time.Time, generationQuery *cbTypes.ReportGenerationQuery, deleteExistingData bool) error {
	logger = logger.WithFields(log.Fields{
		"reportKind":         reportKind,
		"deleteExistingData": deleteExistingData,
	})
	logger.Infof("generating usage report")

	reportGenerationQueryLister := op.informers.Metering().V1alpha1().ReportGenerationQueries().Lister()
	reportDataSourceLister := op.informers.Metering().V1alpha1().ReportDataSources().Lister()

	depsStatus, err := reporting.GetGenerationQueryDependenciesStatus(
		reporting.NewReportGenerationQueryListerGetter(reportGenerationQueryLister),
		reporting.NewReportDataSourceListerGetter(reportDataSourceLister),
		generationQuery,
	)
	if err != nil {
		return fmt.Errorf("unable to generateReport for %s %s, ReportGenerationQuery %s, failed to validate dependencies: %v", reportKind, reportName, generationQuery.Name, err)
	}
	validateResults, err := op.validateDependencyStatus(depsStatus)

	templateInfo := &templateInfo{
		DynamicDependentQueries: validateResults.DynamicReportGenerationQueries,
		Report: &reportTemplateInfo{
			StartPeriod: reportStart,
			EndPeriod:   reportEnd,
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
		err = presto.DeleteFrom(op.prestoQueryer, tableName)
		if err != nil {
			return fmt.Errorf("couldn't empty table %s of preexisting rows: %v", tableName, err)
		}
	}

	// Run the report
	logger.Debugf("running report generation query")
	err = presto.InsertInto(op.prestoQueryer, tableName, query)
	if err != nil {
		logger.WithError(err).Errorf("creating usage report FAILED!")
		return fmt.Errorf("Failed to execute %s usage report: %v", reportName, err)
	}

	return nil
}
