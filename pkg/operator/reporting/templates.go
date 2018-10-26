package reporting

import (
	"bytes"
	"fmt"
	"text/template"
	"time"

	"github.com/Masterminds/sprig"

	cbTypes "github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1"
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
		"prestoTimestamp":             presto.Timestamp,
		"reportTableName":             ReportTableName,
		"scheduledReportTableName":    ScheduledReportTableName,
		"dataSourceTableName":         DataSourceTableName,
		"generationQueryViewName":     GenerationQueryViewName,
		"billingPeriodTimestamp":      BillingPeriodTimestamp,
		"renderReportGenerationQuery": renderReportGenerationQuery,
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
