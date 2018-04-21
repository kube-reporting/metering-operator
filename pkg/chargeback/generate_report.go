package chargeback

import (
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"

	cbTypes "github.com/coreos-inc/kube-chargeback/pkg/apis/chargeback/v1alpha1"
	"github.com/coreos-inc/kube-chargeback/pkg/hive"
	"github.com/coreos-inc/kube-chargeback/pkg/presto"
)

func (c *Chargeback) generateReport(logger log.FieldLogger, report runtime.Object, reportKind, reportName, tableName string, reportStart, reportEnd time.Time, storage *cbTypes.StorageLocationRef, columns []hive.Column, query string, deleteExistingData bool) ([]map[string]interface{}, error) {
	logger = logger.WithFields(log.Fields{
		"reportKind":         reportKind,
		"deleteExistingData": deleteExistingData,
	})
	logger.Infof("generating usage report")
	query, err := renderReportGenerationQuery(reportStart, reportEnd, query)
	if err != nil {
		return nil, err
	}

	switch strings.ToLower(reportKind) {
	case "report", "scheduledreport":
		// valid
	default:
		return nil, fmt.Errorf("invalid report kind: %s", reportKind)
	}

	// presto uses `map(varchar, varchar)` while hive is `MAP<STRING, STRING>`
	for idx, c := range columns {
		if mapIndex := strings.Index(c.Type, "map("); mapIndex != -1 && strings.Index(c.Type, ")") > mapIndex {
			columns[idx].Type = "MAP<STRING, STRING>"
		}
	}

	storageSpec, err := c.getStorageSpec(logger, storage, reportKind)
	if err != nil {
		return nil, err
	}
	err = c.createReportTable(logger, report, reportKind, reportName, tableName, storageSpec, columns, deleteExistingData)
	if err != nil {
		return nil, err
	}

	if deleteExistingData {
		logger.Debugf("deleting any preexisting rows in %s", tableName)
		_, err = presto.ExecuteSelect(c.prestoConn, fmt.Sprintf("DELETE FROM %s", tableName))
		if err != nil {
			return nil, fmt.Errorf("couldn't empty table %s of preexisting rows: %v", tableName, err)
		}
	}

	// Run the report
	logger.Debugf("running report generation query")
	wrappedQuery := fmt.Sprintf("SELECT timestamp '%s' as period_start, timestamp '%s' as period_end, * FROM (%s)", prestoTimestamp(reportStart), prestoTimestamp(reportEnd), query)
	err = presto.ExecuteInsertQuery(c.prestoConn, tableName, wrappedQuery)
	if err != nil {
		logger.WithError(err).Errorf("creating usage report FAILED!")
		return nil, fmt.Errorf("Failed to execute %s usage report: %v", reportName, err)
	}

	getReportQuery := fmt.Sprintf("SELECT * FROM %s", tableName)
	results, err := presto.ExecuteSelect(c.prestoConn, getReportQuery)
	if err != nil {
		logger.WithError(err).Errorf("getting usage report FAILED!")
		return nil, fmt.Errorf("Failed to get usage report results: %v", err)
	}
	return results, nil
}

func (c *Chargeback) createReportTable(logger log.FieldLogger, report runtime.Object, reportKind, reportName string, tableName string, storageSpec cbTypes.StorageLocationSpec, columns []hive.Column, dropTable bool) error {
	var (
		tableParams hive.CreateTableParameters
		err         error
	)
	logger = logger.WithField("dropTable", dropTable)
	if storageSpec.Local != nil {
		logger.Debugf("Creating table %s backed by local storage", tableName)
		tableParams, err = hive.CreateLocalReportTable(c.hiveQueryer, tableName, columns, dropTable)
	} else if storageSpec.S3 != nil {
		bucket, prefix := storageSpec.S3.Bucket, storageSpec.S3.Prefix
		logger.Debugf("Creating table %s pointing to s3 bucket %s at prefix %s", tableName, bucket, prefix)
		tableParams, err = hive.CreateS3ReportTable(c.hiveQueryer, tableName, bucket, prefix, columns, dropTable)
	} else {
		return fmt.Errorf("storage incorrectly configured on report: %s", reportName)
	}

	if err != nil {
		return fmt.Errorf("couldn't create table for output report: %v", err)
	}

	logger.Infof("creating presto table resource for table %q", tableName)
	err = c.createPrestoTableCR(report, cbTypes.GroupName, reportKind, tableParams)
	if err != nil {
		if errors.IsAlreadyExists(err) {
			logger.Infof("presto table resource already exists")
		} else {
			return fmt.Errorf("couldn't create PrestoTable resource for report: %v", err)
		}
	}
	return nil
}
