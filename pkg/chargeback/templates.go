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
	"dataSourceTableName":        dataSourceTableName,
	"generationQueryViewName":   generationQueryViewName,
	"billingPeriodFormat":       billingPeriodFormat,
	"filterAWSData":             filterAWSData,
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

func renderReportGenerationQuery(startPeriod, endPeriod time.Time, queryStr string) (string, error) {
	tmpl, err := newQueryTemplate(queryStr)
	if err != nil {
		return "", err
	}
	info := &templateInfo{
		Report: &reportTemplateInfo{
			StartPeriod: startPeriod,
			EndPeriod:   endPeriod,
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

func filterAWSData(r *reportTemplateInfo, awsBillingDataSourceName string) string {
	start := prestoTimestamp(r.StartPeriod)
	stop := prestoTimestamp(r.EndPeriod)
	partitionStart := hiveAWSPartitionTimestamp(r.StartPeriod)
	partitionStop := hiveAWSPartitionTimestamp(r.EndPeriod)
	return fmt.Sprintf(`
        SELECT aws_billing.*,
               CASE
                   -- AWS data covers entire reporting period
                   WHEN (aws_billing.period_start <= timestamp '%s') AND ( timestamp '%s' <= aws_billing.period_stop)
                       THEN cast(date_diff('millisecond', timestamp '%s', timestamp '%s') as double) / cast(date_diff('millisecond', aws_billing.period_start, aws_billing.period_stop) as double)

                   -- AWS data covers start to middle
                   WHEN (aws_billing.period_start <= timestamp '%s')
                       THEN cast(date_diff('millisecond', timestamp '%s', aws_billing.period_stop) as double) / cast(date_diff('millisecond', aws_billing.period_start, aws_billing.period_stop) as double)

                   -- AWS data covers middle to end
                   WHEN ( timestamp '%s' <= aws_billing.period_stop)
                       THEN cast(date_diff('millisecond', aws_billing.period_start, timestamp '%s') as double) / cast(date_diff('millisecond', aws_billing.period_start, aws_billing.period_stop) as double)

                   ELSE 1
               END as period_percent
        FROM %s as aws_billing

        -- make sure the partition overlaps with our range
        WHERE (partition_stop >= '%s' AND partition_start <= '%s')

        -- make sure lineItem entries overlap with our range
        AND (period_stop >= timestamp '%s' AND period_start <= timestamp '%s')
`, start, stop, start, stop, start, start, stop, stop, generationQueryViewName(awsBillingDataSourceName), partitionStart, partitionStop, start, stop)
}
