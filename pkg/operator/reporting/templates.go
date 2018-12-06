package reporting

import (
	"bytes"
	"errors"
	"fmt"
	"text/template"
	"time"

	"github.com/Masterminds/sprig"

	cbTypes "github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1"
	"github.com/operator-framework/operator-metering/pkg/operator/prestostore"
	"github.com/operator-framework/operator-metering/pkg/operator/reportingutil"
	"github.com/operator-framework/operator-metering/pkg/presto"
)

type ReportQueryTemplateContext struct {
	Report                  *ReportTemplateInfo
	DynamicDependentQueries []*cbTypes.ReportGenerationQuery
}

type ReportTemplateInfo struct {
	ReportingStart *time.Time
	ReportingEnd   *time.Time
	Inputs         map[string]interface{}
}

func newQueryTemplate(queryTemplate string) (*template.Template, error) {
	var templateFuncMap = template.FuncMap{
		"prestoTimestamp":                 PrestoTimestamp,
		"prometheusMetricPartitionFormat": PrometheusMetricPartitionFormat,
		"reportTableName":        reportingutil.ReportTableName,
		"dataSourceTableName":             reportingutil.DataSourceTableName,
		"generationQueryViewName":         reportingutil.GenerationQueryViewName,
		"billingPeriodTimestamp":          reportingutil.BillingPeriodTimestamp,
		"renderReportGenerationQuery":     renderReportGenerationQuery,
	}

	tmpl, err := template.New("report-generation-query").Delims("{|", "|}").Funcs(templateFuncMap).Funcs(sprig.TxtFuncMap()).Parse(queryTemplate)
	if err != nil {
		return nil, fmt.Errorf("error parsing query: %v", err)
	}
	return tmpl, nil
}

func RenderQuery(query string, tmplCtx *ReportQueryTemplateContext) (string, error) {
	tmpl, err := newQueryTemplate(query)
	if err != nil {
		return "", err
	}
	return renderTemplate(tmpl, tmplCtx)
}

func renderTemplate(tmpl *template.Template, tmplCtx *ReportQueryTemplateContext) (string, error) {
	var buf bytes.Buffer
	err := tmpl.Execute(&buf, tmplCtx)
	if err != nil {
		return "", fmt.Errorf("error executing template: %v", err)
	}
	return buf.String(), nil
}

func renderReportGenerationQuery(queryName string, tmplCtx *ReportQueryTemplateContext) (string, error) {
	var query string
	for _, q := range tmplCtx.DynamicDependentQueries {
		if q.Name == queryName {
			query = q.Spec.Query
			break
		}
	}
	if query == "" {
		return "", fmt.Errorf("unknown ReportGenerationQuery %s", queryName)
	}

	renderedQuery, err := RenderQuery(query, tmplCtx)
	if err != nil {
		return "", fmt.Errorf("unable to render query %s, err: %v", queryName, err)
	}
	return renderedQuery, nil
}

func TimestampFormat(input interface{}, format string) (string, error) {
	var err error
	var d time.Time
	switch v := input.(type) {
	case time.Time:
		d = v
	case *time.Time:
		if v == nil {
			return "", errors.New("got nil timestamp")
		}
		d = *v
	case string:
		d, err = time.Parse(time.RFC3339, v)
	default:
		return "", fmt.Errorf("couldn't convert %#v to a Presto timestamp", input)
	}
	return d.Format(format), err
}

func PrometheusMetricPartitionFormat(input interface{}) (string, error) {
	return TimestampFormat(input, prestostore.PrometheusMetricTimestampPartitionFormat)
}

func PrestoTimestamp(input interface{}) (string, error) {
	return TimestampFormat(input, presto.TimestampFormat)
}
