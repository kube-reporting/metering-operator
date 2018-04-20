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

type templateInfo struct {
	Report                  *reportTemplateInfo
	DynamicDependentQueries []*cbTypes.ReportGenerationQuery
}

type reportTemplateInfo struct {
	StartPeriod time.Time
	EndPeriod   time.Time
}

func newQueryTemplate(queryTemplate string) (*template.Template, error) {
	var templateFuncMap = template.FuncMap{
		"hiveAWSPartitionTimestamp":   hiveAWSPartitionTimestamp,
		"prestoTimestamp":             prestoTimestamp,
		"dataSourceTableName":         dataSourceTableName,
		"generationQueryViewName":     generationQueryViewName,
		"billingPeriodFormat":         billingPeriodFormat,
		"filterAWSData":               filterAWSData,
		"renderReportGenerationQuery": renderReportGenerationQuery,
	}

	tmpl, err := template.New("report-generation-query").Delims("{|", "|}").Funcs(templateFuncMap).Parse(queryTemplate)
	if err != nil {
		return nil, fmt.Errorf("error parsing query: %v", err)
	}
	return tmpl, nil
}

type queryRenderer struct {
	templateInfo *templateInfo
}

func (qr queryRenderer) Render(query string) (string, error) {
	tmpl, err := newQueryTemplate(query)
	if err != nil {
		return "", err
	}
	return qr.renderTemplate(tmpl)
}

func (qr queryRenderer) renderTemplate(tmpl *template.Template) (string, error) {
	var buf bytes.Buffer
	err := tmpl.Execute(&buf, qr.templateInfo)
	if err != nil {
		return "", fmt.Errorf("error executing template: %v", err)
	}
	return buf.String(), nil
}

func renderReportGenerationQuery(generationQueryName string, templateInfo *templateInfo) (string, error) {
	var query string
	for _, q := range templateInfo.DynamicDependentQueries {
		if q.Name == generationQueryName {
			query = q.Spec.Query
			break
		}
	}
	if query == "" {
		return "", fmt.Errorf("unknown generationQuery %s", generationQueryName)
	}
	qr := queryRenderer{templateInfo: templateInfo}
	renderedQuery, err := qr.Render(query)
	if err != nil {
		return "", fmt.Errorf("unable to render query %s, err: %v", generationQueryName, err)
	}
	return renderedQuery, nil
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
                   WHEN (aws_billing.usage_start_date <= timestamp '%s') AND ( timestamp '%s' <= aws_billing.usage_end_date)
                       THEN cast(date_diff('millisecond', timestamp '%s', timestamp '%s') as double) / cast(date_diff('millisecond', aws_billing.usage_start_date, aws_billing.usage_end_date) as double)

                   -- AWS data covers start to middle
                   WHEN (aws_billing.usage_start_date <= timestamp '%s')
                       THEN cast(date_diff('millisecond', timestamp '%s', aws_billing.usage_end_date) as double) / cast(date_diff('millisecond', aws_billing.usage_start_date, aws_billing.usage_end_date) as double

                   -- AWS data covers middle to end
                   WHEN ( timestamp '%s' <= aws_billing.usage_end_date)
                       THEN cast(date_diff('millisecond', aws_billing.usage_start_date, timestamp '%s') as double) / cast(date_diff('millisecond', aws_billing.usage_start_date, aws_billing.usage_end_date) as doub
                   ELSE 1
               END as period_percent
        FROM %s as aws_billing

        -- make sure the partition overlaps with our range
        WHERE (partition_stop >= '%s' AND partition_start <= '%s')

        -- make sure lineItem entries overlap with our range
        AND (usage_end_date >= timestamp '%s' AND usage_start_date <= timestamp '%s')
`, start, stop, start, stop, start, start, stop, stop, generationQueryViewName(awsBillingDataSourceName), partitionStart, partitionStop, start, stop)
}
