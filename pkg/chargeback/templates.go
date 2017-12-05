package chargeback

import (
	"bytes"
	"fmt"
	"text/template"
	"time"

	cbTypes "github.com/coreos-inc/kube-chargeback/pkg/apis/chargeback/v1alpha1"
	"github.com/coreos-inc/kube-chargeback/pkg/hive"
)

const (
	// PrestoTimestampFormat is the time format string used to produce Presto timestamps.
	PrestoTimestampFormat = "2006-01-02 15:04:05.000"
)

var templateFuncMap = template.FuncMap{
	"hiveAWSPartitionTimestamp": hiveAWSPartitionTimestamp,
	"prestoTimestamp":           prestoTimestamp,
	"dataStoreTableName":        dataStoreTableName,
	"generationQueryViewName":   generationQueryViewName,
}

type templateInfo struct {
	Report *reportTemplateInfo
}

type reportTemplateInfo struct {
	StartPeriod time.Time
	EndPeriod   time.Time
}

func newQueryTemplate(queryTemplate string) (*template.Template, error) {
	tmpl, err := template.New("report-generation-query").Delims("{|", "|}").Funcs(templateFuncMap).Parse(queryTemplate)
	if err != nil {
		return nil, fmt.Errorf("error parsing query: %v", err)
	}
	return tmpl, nil
}

func renderReportGenerationQuery(report *cbTypes.Report, generationQuery *cbTypes.ReportGenerationQuery) (string, error) {
	tmpl, err := newQueryTemplate(generationQuery.Spec.Query)
	if err != nil {
		return "", err
	}
	info := &templateInfo{
		Report: &reportTemplateInfo{
			StartPeriod: report.Spec.ReportingStart.Time,
			EndPeriod:   report.Spec.ReportingEnd.Time,
		},
	}
	return renderTemplateInfo(tmpl, info)
}

func renderGenerationQuery(generationQuery *cbTypes.ReportGenerationQuery) (string, error) {
	tmpl, err := newQueryTemplate(generationQuery.Spec.Query)
	if err != nil {
		return "", err
	}
	return renderTemplateInfo(tmpl, nil)
}

func renderTemplateInfo(tmpl *template.Template, info *templateInfo) (string, error) {
	var buf bytes.Buffer
	err := tmpl.Execute(&buf, info)
	if err != nil {
		return "", fmt.Errorf("error executing template: %v", err)
	}
	return buf.String(), nil
}

func hiveAWSPartitionTimestamp(date time.Time) string {
	return date.Format(hive.HiveDateStringLayout)
}

func prestoTimestamp(date time.Time) string {
	return date.Format(PrestoTimestampFormat)
}
