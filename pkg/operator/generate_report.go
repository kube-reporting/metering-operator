package operator

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	metering "github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1"
	"github.com/operator-framework/operator-metering/pkg/operator/reporting"
)

func (op *Reporting) generateReport(logger log.FieldLogger, reportName, tableName string, reportStart, reportEnd *time.Time, generationQuery *metering.ReportGenerationQuery, dynamicReportGenerationQueries []*metering.ReportGenerationQuery, inputs []metering.ReportGenerationQueryInputValue) error {
	generator := reporting.NewReportGenerator(op.logger, op.reportResultsRepo, true)

	err := generator.GenerateReport(tableName, reportStart, reportEnd, generationQuery, dynamicReportGenerationQueries, inputs)
	if err != nil {
		return fmt.Errorf("failed to generateReport for Report %s, err: %v", reportName, err)
	}
	return nil
}

func (op *Reporting) generateScheduledReport(logger log.FieldLogger, reportName, tableName string, reportStart, reportEnd *time.Time, generationQuery *metering.ReportGenerationQuery, dynamicReportGenerationQueries []*metering.ReportGenerationQuery, inputs []metering.ReportGenerationQueryInputValue, deleteExistingData bool) error {
	generator := reporting.NewReportGenerator(op.logger, op.reportResultsRepo, deleteExistingData)

	err := generator.GenerateReport(tableName, reportStart, reportEnd, generationQuery, dynamicReportGenerationQueries, inputs)
	if err != nil {
		return fmt.Errorf("failed to generateReport for ScheduledReport %s, err: %v", reportName, err)
	}
	return nil
}
