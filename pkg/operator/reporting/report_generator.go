package reporting

import (
	"errors"
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/operator-framework/operator-metering/pkg/operator/prestostore"
)

const (
	ReportingStartInputName = "ReportingStart"
	ReportingEndInputName   = "ReportingEnd"
)

var (
	errInvalidTableName       = errors.New("tableName cannot be empty")
	errInvalidReportQueryName = errors.New("reportQuery cannot be empty")
	errEmptyQueryField        = errors.New("ReportQuery spec.query cannot be empty")
)

type ReportGenerator interface {
	GenerateReport(tableName, query string, deleteExistingData bool) error
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

func (g *reportGenerator) GenerateReport(tableName, query string, deleteExistingData bool) error {
	if tableName == "" {
		return errInvalidTableName
	}
	logger := g.logger.WithFields(log.Fields{
		"tableName": tableName,
	})
	logger.Infof("generating Report")

	if deleteExistingData {
		logger.Debugf("deleting any preexisting rows in %s", tableName)
		err := g.reportResultsRepo.DeleteReportResults(tableName)
		if err != nil {
			return fmt.Errorf("couldn't empty table %s of preexisting rows: %v", tableName, err)
		}
	}

	logger.Debugf("StoreReportResults: executing ReportQuery")
	err := g.reportResultsRepo.StoreReportResults(tableName, query)
	if err != nil {
		logger.WithError(err).Errorf("creating usage report FAILED!")
		return fmt.Errorf("Failed to execute query for Report table %s: %v", tableName, err)
	}

	return nil
}
