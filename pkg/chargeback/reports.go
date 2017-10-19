package chargeback

import (
	"bytes"
	"database/sql"
	"fmt"
	"text/template"
	"time"

	log "github.com/sirupsen/logrus"

	cbTypes "github.com/coreos-inc/kube-chargeback/pkg/apis/chargeback/v1alpha1"
	cb "github.com/coreos-inc/kube-chargeback/pkg/chargeback/v1"
	"github.com/coreos-inc/kube-chargeback/pkg/hive"
	"github.com/coreos-inc/kube-chargeback/pkg/presto"
)

const (
	// TimestampFormat is the time format string used to produce Presto timestamps.
	TimestampFormat = "2006-01-02 15:04:05.000"
)

func prestoTime(t time.Time) string {
	return t.Format(TimestampFormat)
}

var templateFuncMap = template.FuncMap{
	"listAdditionalLabels": listAdditionalLabels,
	"addAdditionalLabels":  addAdditionalLabels,
}

type TemplateInfo struct {
	TableName   string
	StartPeriod string
	EndPeriod   string
	Labels      []string
}

func newTemplateInfo(tableName string, startPeriod, endPeriod time.Time, labels []string) TemplateInfo {
	return TemplateInfo{tableName, prestoTime(startPeriod), prestoTime(endPeriod), labels}
}

func listAdditionalLabels(labels []string) string {
	output := ""
	for _, l := range labels {
		output += fmt.Sprintf(", label_%s", l)
	}
	return output
}

func addAdditionalLabels(labels []string) string {
	output := ""
	for _, l := range labels {
		output += fmt.Sprintf(", kubeUsage.labels['%s'] as label_%s", l, l)
	}
	return output
}

func generateHiveColumns(report *cbTypes.Report, genQuery *cbTypes.ReportGenerationQuery) []hive.Column {
	columns := make([]hive.Column, 0)
	for _, c := range genQuery.Spec.Columns {
		columns = append(columns, hive.Column{Name: c.Name, Type: c.Type})
	}
	for _, label := range report.Spec.AdditionalLabels {
		columns = append(columns, hive.Column{Name: label, Type: "string"})
	}
	return columns
}

func generateReport(logger *log.Entry, report *cbTypes.Report, genQuery *cbTypes.ReportGenerationQuery, rng cb.Range, promsumTbl string, queryer hive.Queryer, prestoCon *sql.DB) ([]map[string]interface{}, error) {
	logger.Infof("generating usage report")

	// Perform query templating
	tmpl, err := template.New("request").Funcs(templateFuncMap).Parse(genQuery.Spec.Query)
	if err != nil {
		return nil, fmt.Errorf("error parsing query: %v", err)
	}
	buf := &bytes.Buffer{}
	err = tmpl.Execute(buf, newTemplateInfo(promsumTbl, report.Spec.ReportingStart.Time, report.Spec.ReportingEnd.Time, report.Spec.AdditionalLabels))
	if err != nil {
		return nil, fmt.Errorf("error executing template: %v", err)
	}
	query := string(buf.Bytes())

	// Create a table to write to
	reportTable := reportTableName(report.Name)
	bucket, pre***REMOVED***x := report.Spec.Output.Bucket, report.Spec.Output.Pre***REMOVED***x
	logger.Debugf("Creating table %s pointing to s3 bucket %s at pre***REMOVED***x %s", reportTable, bucket, pre***REMOVED***x)
	err = hive.CreateReportTable(queryer, reportTable, bucket, pre***REMOVED***x, generateHiveColumns(report, genQuery))
	if err != nil {
		return nil, fmt.Errorf("Couldn't create table for output report: %v", err)
	}

	logger.Debugf("deleting any preexisting rows in %s", reportTable)
	err = hive.ExecuteTruncate(queryer, reportTable)
	if err != nil {
		return nil, fmt.Errorf("couldn't empty table %s of preexisting rows: %v", reportTable, err)
	}

	// Run the report
	logger.Debugf("running report generation query")
	err = presto.ExecuteInsertQuery(prestoCon, reportTable, query)
	if err != nil {
		logger.WithError(err).Errorf("creating usage report FAILED!")
		return nil, fmt.Errorf("Failed to execute %s usage report: %v", genQuery.Name, err)
	}

	getReportQuery := fmt.Sprintf("SELECT * FROM %s", reportTable)
	results, err := presto.ExecuteSelect(prestoCon, getReportQuery)
	if err != nil {
		logger.WithError(err).Errorf("getting usage report FAILED!")
		return nil, fmt.Errorf("Failed to get usage report results: %v", err)
	}
	return results, nil
}
