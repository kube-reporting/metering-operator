package operator

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	cbTypes "github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1"
	"github.com/operator-framework/operator-metering/pkg/operator/reporting"
)

func (op *Reporting) generateReport(logger log.FieldLogger, reportName, tableName string, reportStart, reportEnd *time.Time, generationQuery *cbTypes.ReportGenerationQuery, inputs []cbTypes.ReportGenerationQueryInputValue) error {
	generator := reporting.NewReportGenerator(
		op.logger, op.reportResultsRepo, op.uninitialiedDependendenciesHandler(),
		op.reportLister, op.scheduledReportLister, op.reportDataSourceLister, op.reportGenerationQueryLister,
		true,
	)

	err := generator.GenerateReport(tableName, reportStart, reportEnd, generationQuery, inputs)
	if err != nil {
		return fmt.Errorf("failed to generateReport for Report %s, err: %v", reportName, err)
	}
	return nil
}

func (op *Reporting) generateScheduledReport(logger log.FieldLogger, reportName, tableName string, reportStart, reportEnd *time.Time, generationQuery *cbTypes.ReportGenerationQuery, inputs []cbTypes.ReportGenerationQueryInputValue, deleteExistingData bool) error {
	generator := reporting.NewReportGenerator(
		op.logger, op.reportResultsRepo, op.uninitialiedDependendenciesHandler(),
		op.reportLister, op.scheduledReportLister, op.reportDataSourceLister, op.reportGenerationQueryLister,
		deleteExistingData,
	)

	err := generator.GenerateReport(tableName, reportStart, reportEnd, generationQuery, inputs)
	if err != nil {
		return fmt.Errorf("failed to generateReport for ScheduledReport %s, err: %v", reportName, err)
	}
	return nil
}
