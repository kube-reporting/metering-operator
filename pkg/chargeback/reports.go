package chargeback

import (
	"bytes"
	"database/sql"
	"fmt"
	"math/rand"
	"strings"
	"text/template"
	"time"

	log "github.com/sirupsen/logrus"

	cb "github.com/coreos-inc/kube-chargeback/pkg/chargeback/v1"
	cbTypes "github.com/coreos-inc/kube-chargeback/pkg/chargeback/v1/types"
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

func generateHiveColumns(report *cbTypes.Report, genQuery cbTypes.ReportGenerationQuery) []string {
	columns := []string{}
	for _, c := range genQuery.Spec.Columns {
		columns = append(columns, fmt.Sprintf("%s %s", c.Name, c.Type))
	}
	for _, l := range report.Spec.AdditionalLabels {
		columns = append(columns, fmt.Sprintf("%s string", l))
	}
	return columns
}

func generateReport(logger *log.Entry, report *cbTypes.Report, genQuery cbTypes.ReportGenerationQuery, rng cb.Range, promsumTbl string, hiveCon *hive.Connection, prestoCon *sql.DB) error {
	logger.Infof("generating usage report")

	// Perform query templating
	tmpl, err := template.New("request").Funcs(templateFuncMap).Parse(genQuery.Spec.Query)
	if err != nil {
		return fmt.Errorf("error parsing query: %v", err)
	}
	buf := &bytes.Buffer{}
	err = tmpl.Execute(buf, newTemplateInfo(promsumTbl, report.Spec.ReportingStart.Time, report.Spec.ReportingEnd.Time, report.Spec.AdditionalLabels))
	if err != nil {
		return fmt.Errorf("error executing template: %v", err)
	}
	query := string(buf.Bytes())

	logger.Debugf("query generated:\n%s", query)

	// Create a table to write to
	reportTable := fmt.Sprintf("%s_%d", strings.Replace(genQuery.Name, "-", "_", -1), rand.Int31())
	bucket, pre***REMOVED***x := report.Spec.Output.Bucket, report.Spec.Output.Pre***REMOVED***x
	logger.Infof("found S3 bucket to write to: %q", bucket+"/"+pre***REMOVED***x)
	logger.Debugf("creating output table %s", reportTable)
	err = hive.CreateReportTable(hiveCon, reportTable, bucket, pre***REMOVED***x, generateHiveColumns(report, genQuery))
	if err != nil {
		return fmt.Errorf("Couldn't create table for output report: %v", err)
	}

	// Run the report
	logger.Debugf("running report generation query")
	err = presto.ExecuteInsertQuery(prestoCon, reportTable, query)
	if err != nil {
		logger.WithError(err).Errorf("usage report FAILED!")
		return fmt.Errorf("Failed to execute %s usage report: %v", genQuery.Name, err)
	}
	return nil
}
