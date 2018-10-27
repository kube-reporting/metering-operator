package reporting

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	metering "github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1"
	"github.com/operator-framework/operator-metering/pkg/operator/prestostore"
)

const (
	ReportingStartInputName = "ReportingStart"
	ReportingEndInputName   = "ReportingEnd"
)

type ReportGenerator interface {
	GenerateReport(tableName string, reportStart, reportEnd *time.Time, generationQuery *metering.ReportGenerationQuery, dynamicReportGenerationQueries []*metering.ReportGenerationQuery, inputs []metering.ReportGenerationQueryInputValue, deleteExistingData bool) error
}

type reportGenerator struct {
	logger            log.FieldLogger
	reportResultsRepo prestostore.ReportResultsRepo
}

func NewReportGenerator(logger log.FieldLogger, reportResultsRepo prestostore.ReportResultsRepo) *reportGenerator {
	return &reportGenerator{
		logger:            logger,
		reportResultsRepo: reportResultsRepo,
	}
}

func (g *reportGenerator) GenerateReport(tableName string, reportStart, reportEnd *time.Time, generationQuery *metering.ReportGenerationQuery, dynamicReportGenerationQueries []*metering.ReportGenerationQuery, inputs []metering.ReportGenerationQueryInputValue, deleteExistingData bool) error {
	logger := g.logger.WithFields(log.Fields{
		"tableName":             tableName,
		"reportGenerationQuery": generationQuery.Name,
	})
	logger.Infof("generating Report")

	reportQueryInputs, err := ValidateReportGenerationQueryInputs(generationQuery, inputs)
	if err != nil {
		return fmt.Errorf("unable to GenerateReport for Report Table %s, ReportGenerationQuery %s, failed to validate ReportGenerationQueryInputs: %s", tableName, generationQuery.Name, err)
	}

	tmplCtx := &ReportQueryTemplateContext{
		DynamicDependentQueries: dynamicReportGenerationQueries,
		Report: &ReportTemplateInfo{
			ReportingStart: reportStart,
			ReportingEnd:   reportEnd,
			Inputs:         reportQueryInputs,
		},
	}
	query, err := RenderQuery(generationQuery.Spec.Query, tmplCtx)
	if err != nil {
		return err
	}

	if deleteExistingData {
		logger.Debugf("deleting any preexisting rows in %s", tableName)
		err = g.reportResultsRepo.DeleteReportResults(tableName)
		if err != nil {
			return fmt.Errorf("couldn't empty table %s of preexisting rows: %v", tableName, err)
		}
	}

	logger.Debugf("StoreReportResults: executing ReportGenerationQuery")
	err = g.reportResultsRepo.StoreReportResults(tableName, query)
	if err != nil {
		logger.WithError(err).Errorf("creating usage report FAILED!")
		return fmt.Errorf("Failed to execute query %s for Report table %s: %v", generationQuery.Name, tableName, err)
	}

	return nil
}
