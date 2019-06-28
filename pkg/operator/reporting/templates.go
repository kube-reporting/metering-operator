package reporting

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/Masterminds/sprig"
	metering "github.com/operator-framework/operator-metering/pkg/apis/metering/v1"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/operator-framework/operator-metering/pkg/operator/prestostore"
	"github.com/operator-framework/operator-metering/pkg/operator/reportingutil"
	"github.com/operator-framework/operator-metering/pkg/presto"
)

// ReportQueryTemplateContext is used to hold all information about a ReportQuery that will be
// needed when rendering the templating inside of a ReportQuery's query ***REMOVED***eld.
type ReportQueryTemplateContext struct {
	Namespace      string
	Query          string
	RequiredInputs []string

	Reports           []*metering.Report
	ReportQueries     []*metering.ReportQuery
	ReportDataSources []*metering.ReportDataSource
	PrestoTables      []*metering.PrestoTable
}

// TemplateContext is the context passed to each template and contains variables related to the Report
// and ReportQuery such as inputs and dependencies on other resources.
type TemplateContext struct {
	Report ReportTemplateInfo
}

// ReportTemplateInfo contains the variables for the Reprot being executed with this template.
type ReportTemplateInfo struct {
	ReportingStart *time.Time
	ReportingEnd   *time.Time
	Inputs         map[string]interface{}
}

// dataSourceTableName is a receiver method for ReportQueryTemplateContext, which validates that
// certain ***REMOVED***elds in the ctx.DataSources are properly set. This returns the name of the Presto Table
// that the DataSource references (DataSource.Status.TableRef) and nil, or an empty string and an error
// if the TableRef.Name is unset, or unable to be found in the ctx.PrestoTables.
func (ctx *ReportQueryTemplateContext) dataSourceTableName(name string) (string, error) {
	for _, ds := range ctx.ReportDataSources {
		if ds.Name == name {
			if ds.Status.TableRef.Name == "" {
				return "", fmt.Errorf("%s tableRef is empty", ds.Name)
			}
			for _, prestoTable := range ctx.PrestoTables {
				if prestoTable.Name == ds.Status.TableRef.Name {
					return reportingutil.FullyQuali***REMOVED***edTableName(prestoTable)
				}
			}
			return "", fmt.Errorf("tableRef PrestoTable %s not found", ds.Status.TableRef.Name)
		}
	}
	return "", fmt.Errorf("ReportDataSource %s dependency not found", name)
}

// reportTableName is a receiver method for ReportQueryTemplateContext, which validates that
// that certain ***REMOVED***elds in ctx.Reports are properly set. This returns the name of the Presto Table
// that the Report references (Report.Status.TableRef) and nil, or an empty string and an error
// if the TableRef.Name is unset, or unable to be found in the ctx.PrestoTables
func (ctx *ReportQueryTemplateContext) reportTableName(name string) (string, error) {
	for _, r := range ctx.Reports {
		if r.Name == name {
			if r.Status.TableRef.Name == "" {
				return "", fmt.Errorf("%s tableRef is empty", r.Name)
			}
			for _, prestoTable := range ctx.PrestoTables {
				if prestoTable.Name == r.Status.TableRef.Name {
					return reportingutil.FullyQuali***REMOVED***edTableName(prestoTable)
				}
			}
			return "", fmt.Errorf("tableRef PrestoTable %s not found", r.Status.TableRef.Name)
		}
	}
	return "", fmt.Errorf("Report %s dependency not found", name)
}

// renderReportQuery takes two parameters: a string parameter referencing a ReportQuery's name, and a TemplateContext
// parameter, which is typically just `.` in the template. If the tmplCtx.ReportQuery is valid, this returns
// a string containing the speci***REMOVED***ed ReportQuery in its rendered form, using the second argument as the context
// for templating rendering, or an error due to an unkown ReportQuery name or the inability to render that query.
func (ctx *ReportQueryTemplateContext) renderReportQuery(name string, tmplCtx TemplateContext) (string, error) {
	var reportQuery *metering.ReportQuery
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

// RenderQuery creates a new query template by calling the ctx parameter's method, newQueryTemplate,
// and checks if the returned error is nil. If nil, return an empty string and the error, ***REMOVED*** return
// the function call to renderTemplate, passing in the new query template and tmplCtx parameter.
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

// TimestampFormat checks the type of the input interface parameter and returns that parameter in
// the form speci***REMOVED***ed by the format string parameter, or an error if it's not able to be converted.
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

// PrometheusMetricPartitionFormat is a helper function that returns the Prometheus timestamp partition format
func PrometheusMetricPartitionFormat(input interface{}) (string, error) {
	return TimestampFormat(input, prestostore.PrometheusMetricTimestampPartitionFormat)
}

// PrestoTimestamp is a helper function that returns the Presto timestamp format
func PrestoTimestamp(input interface{}) (string, error) {
	return TimestampFormat(input, presto.TimestampFormat)
}
