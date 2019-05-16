package reporting

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/Masterminds/sprig"
	"k8s.io/apimachinery/pkg/util/sets"

	cbTypes "github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1"
	"github.com/operator-framework/operator-metering/pkg/operator/prestostore"
	"github.com/operator-framework/operator-metering/pkg/operator/reportingutil"
	"github.com/operator-framework/operator-metering/pkg/presto"
)

type ReportQueryTemplateContext struct {
	Namespace      string
	Query          string
	RequiredInputs []string

	Reports           []*cbTypes.Report
	ReportQueries     []*cbTypes.ReportQuery
	ReportDataSources []*cbTypes.ReportDataSource
	PrestoTables      []*cbTypes.PrestoTable
}

type TemplateContext struct {
	Report ReportTemplateInfo
}

type ReportTemplateInfo struct {
	ReportingStart *time.Time
	ReportingEnd   *time.Time
	Inputs         map[string]interface{}
}

func (ctx *ReportQueryTemplateContext) dataSourceTableName(name string) (string, error) {
	for _, ds := range ctx.ReportDataSources {
		if ds.Name == name {
			if ds.Status.TableRef.Name == "" {
				return "", fmt.Errorf("%s tableRef is empty", ds.Name)
			}
			for _, prestoTable := range ctx.PrestoTables {
				if prestoTable.Name == ds.Status.TableRef.Name {
					return reportingutil.FullyQuali***REMOVED***edTableName(prestoTable), nil
				}
			}
			return "", fmt.Errorf("tableRef PrestoTable %s not found", ds.Status.TableRef.Name)
		}
	}
	return "", fmt.Errorf("ReportDataSource %s dependency not found", name)
}

func (ctx *ReportQueryTemplateContext) reportTableName(name string) (string, error) {
	for _, r := range ctx.Reports {
		if r.Name == name {
			if r.Status.TableRef.Name == "" {
				return "", fmt.Errorf("%s tableRef is empty", r.Name)
			}
			for _, prestoTable := range ctx.PrestoTables {
				if prestoTable.Name == r.Status.TableRef.Name {
					return reportingutil.FullyQuali***REMOVED***edTableName(prestoTable), nil
				}
			}
			return "", fmt.Errorf("tableRef PrestoTable %s not found", r.Status.TableRef.Name)
		}
	}
	return "", fmt.Errorf("Report %s dependency not found", name)
}

func (ctx *ReportQueryTemplateContext) renderReportQuery(name string, tmplCtx TemplateContext) (string, error) {
	var reportQuery *cbTypes.ReportQuery
	for _, q := range ctx.ReportQueries {
		if q.Name == name {
			reportQuery = q
			break
		}
	}
	if reportQuery == nil {
		return "", fmt.Errorf("unknown ReportQuery %s", name)
	}

	// copy context and replace the query we're rendering
	newCtx := *ctx
	newCtx.RequiredInputs = reportingutil.ConvertInputDe***REMOVED***nitionsIntoInputList(reportQuery.Spec.Inputs)
	newCtx.Query = reportQuery.Spec.Query

	renderedQuery, err := RenderQuery(&newCtx, tmplCtx)
	if err != nil {
		return "", fmt.Errorf("unable to render query %s, err: %v", name, err)
	}
	return renderedQuery, nil
}

func (ctx *ReportQueryTemplateContext) newQueryTemplate() (*template.Template, error) {
	var templateFuncMap = template.FuncMap{
		"prestoTimestamp":                 PrestoTimestamp,
		"billingPeriodTimestamp":          reportingutil.AWSBillingPeriodTimestamp,
		"prometheusMetricPartitionFormat": PrometheusMetricPartitionFormat,
		"reportTableName":                 ctx.reportTableName,
		"dataSourceTableName":             ctx.dataSourceTableName,
		"renderReportQuery":               ctx.renderReportQuery,
	}

	tmpl, err := template.New("reportQueryTemplate").Delims("{|", "|}").Funcs(templateFuncMap).Funcs(sprig.TxtFuncMap()).Parse(ctx.Query)
	if err != nil {
		return nil, fmt.Errorf("error parsing query: %v", err)
	}
	return tmpl, nil
}

func RenderQuery(ctx *ReportQueryTemplateContext, tmplCtx TemplateContext) (string, error) {
	requiredInputs := sets.NewString(ctx.RequiredInputs...)
	givenInputs := sets.NewString()
	for inputName := range tmplCtx.Report.Inputs {
		givenInputs.Insert(inputName)
	}

	missingInputs := requiredInputs.Difference(givenInputs)

	if missingInputs.Len() != 0 {
		return "", fmt.Errorf("missing inputs: %s", strings.Join(missingInputs.List(), ", "))
	}

	tmpl, err := ctx.newQueryTemplate()
	if err != nil {
		return "", err
	}
	return renderTemplate(tmpl, tmplCtx)
}

func renderTemplate(tmpl *template.Template, tmplCtx TemplateContext) (string, error) {
	var buf bytes.Buffer
	err := tmpl.Execute(&buf, tmplCtx)
	if err != nil {
		return "", fmt.Errorf("error executing template: %v", err)
	}
	return buf.String(), nil
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
