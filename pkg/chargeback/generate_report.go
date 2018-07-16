package chargeback

import (
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"

	cbTypes "github.com/operator-framework/operator-metering/pkg/apis/chargeback/v1alpha1"
	"github.com/operator-framework/operator-metering/pkg/hive"
	"github.com/operator-framework/operator-metering/pkg/presto"
)

func (c *Chargeback) generateReport(logger log.FieldLogger, report runtime.Object, reportKind, reportName, tableName string, reportStart, reportEnd time.Time, storage *cbTypes.StorageLocationRef, generationQuery *cbTypes.ReportGenerationQuery, dropTable, deleteExistingData bool) error {
	logger = logger.WithFields(log.Fields{
		"reportKind":         reportKind,
		"deleteExistingData": deleteExistingData,
		"dropTable":          dropTable,
	})
	logger.Infof("generating usage report")

	dependentQueries, err := c.getDependentGenerationQueries(generationQuery, true)
	if err != nil {
		return fmt.Errorf("unable to get dependent generationQueries for %s, err: %v", generationQuery.Name, err)
	}

	columns := generateHiveColumns(generationQuery)

	templateInfo := &templateInfo{
		DynamicDependentQueries: dependentQueries,
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

	if dropTable {
		logger.Debugf("dropping table %s", tableName)
		err := hive.ExecuteDropTable(c.hiveQueryer, tableName, true)
		if err != nil {
			return err
		}
	}

	err = c.createTableForStorage(logger, report, reportKind, reportName, storage, tableName, columns)
	if err != nil {
		return err
	}

	if deleteExistingData {
		logger.Debugf("deleting any preexisting rows in %s", tableName)
		err = presto.DeleteFrom(c.prestoQueryer, tableName)
		if err != nil {
			return fmt.Errorf("couldn't empty table %s of preexisting rows: %v", tableName, err)
		}
	}

	// Run the report
	logger.Debugf("running report generation query")
	err = presto.InsertInto(c.prestoQueryer, tableName, query)
	if err != nil {
		logger.WithError(err).Errorf("creating usage report FAILED!")
		return fmt.Errorf("Failed to execute %s usage report: %v", reportName, err)
	}

	return nil
}
