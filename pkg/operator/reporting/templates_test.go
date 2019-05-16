package reporting

import (
	"fmt"
	"testing"
	"time"

	"github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1"
	metering "github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1"
	"github.com/operator-framework/operator-metering/pkg/operator/prestostore"
	"github.com/operator-framework/operator-metering/pkg/presto"
	"github.com/operator-framework/operator-metering/test/testhelpers"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestRenderQuery(t *testing.T) {
	const (
		testNamespace   = "default"
		testReportName  = "namespace-cpu-usage"
		testReportQuery = "unready-deployment-replicas-query"
		testCatalogName = "hive"
		testSchemaName  = "default"
		testTimeString  = "2019-05-01T15:00:05Z"
	)

	reportStart := &time.Time{}
	reportEndTmp := reportStart.AddDate(0, 1, 0)
	reportEnd := &reportEndTmp
	reportTestTimeTmp := reportStart.AddDate(2019, 05, 01)
	reportTestTime := &reportTestTimeTmp

	ds1 := testhelpers.NewReportDataSource("datasource1", testNamespace)
	ds1.Status.TableRef = v1.LocalObjectReference{Name: "initialized_datasource"}
	ds2 := testhelpers.NewReportDataSource("datasource2", testNamespace)

	testValidQuery := &metering.ReportQuery{
		ObjectMeta: meta.ObjectMeta{
			Name:      testReportQuery,
			Namespace: testNamespace,
		},
		Spec: metering.ReportQuerySpec{
			Query: "SELECT * FROM test_table",
		},
	}

	testValidQueryInputs := &metering.ReportQuery{
		ObjectMeta: meta.ObjectMeta{
			Name:      testReportQuery,
			Namespace: testNamespace,
		},
		Spec: metering.ReportQuerySpec{
			Query: "SELECT * FROM test_table",
			Inputs: []metering.ReportQueryInputDe***REMOVED***nition{
				metering.ReportQueryInputDe***REMOVED***nition{
					Name:     "ValidRenderReportQuery",
					Type:     "ReportQuery",
					Required: true,
				},
			},
		},
	}

	testTable := []struct {
		name            string
		reportTemplate  *ReportQueryTemplateContext
		templateContext TemplateContext
		expectErr       bool
		expectErrMsg    string
		expectOutput    string
	}{
		{
			name: "valid report query with no templating returns nil and the expected query output",
			reportTemplate: &ReportQueryTemplateContext{
				Query: "SELECT * FROM test_table;",
			},
			templateContext: TemplateContext{},
			expectOutput:    "SELECT * FROM test_table;",
		},
		{
			name: "valid report query with valid templating returns nil and the expected query output",
			reportTemplate: &ReportQueryTemplateContext{
				Query:             "SELECT * FROM {| dataSourceTableName .Report.Inputs.ValidDataSourceName |}",
				RequiredInputs:    []string{"ValidDataSourceName"},
				ReportDataSources: []*metering.ReportDataSource{ds1},
				PrestoTables: []*metering.PrestoTable{
					newTestPrestoTable(ds1.Status.TableRef.Name, testNamespace, testSchemaName, testCatalogName, nil),
				},
			},
			templateContext: TemplateContext{
				Report: ReportTemplateInfo{
					Inputs: map[string]interface{}{"ValidDataSourceName": ds1.Name},
				},
			},
			expectOutput: fmt.Sprintf("SELECT * FROM %s.%s.%s", testCatalogName, testSchemaName, ds1.Status.TableRef.Name),
		},
		{
			name: "invalid report query with invalid templating (missing Inputs ***REMOVED***eld) returns error",
			reportTemplate: &ReportQueryTemplateContext{
				Query: "SELECT * FROM {| dataSourceTableName .Report.Inputs. |}",
			},
			templateContext: TemplateContext{},
			expectErr:       true,
			expectErrMsg:    "error parsing query: template: reportQueryTemplate:1: unexpected <.> in operand",
		},
		{
			name: "valid report query with valid templating but unde***REMOVED***ned data source table ref returns error",
			reportTemplate: &ReportQueryTemplateContext{
				Query:             "SELECT * FROM {| dataSourceTableName .Report.Inputs.MissingDataSourceTableRef |}",
				RequiredInputs:    []string{"MissingDataSourceTableRef"},
				ReportDataSources: []*metering.ReportDataSource{ds2},
				PrestoTables: []*metering.PrestoTable{
					newTestPrestoTable("", testNamespace, testSchemaName, testCatalogName, nil),
				},
			},
			templateContext: TemplateContext{
				Report: ReportTemplateInfo{
					Inputs: map[string]interface{}{"MissingDataSourceTableRef": ds2.Name},
				},
			},
			expectErr:    true,
			expectErrMsg: fmt.Sprintf("error executing template: template: reportQueryTemplate:1:17: executing \"reportQueryTemplate\" at <dataSourceTableName ...>: error calling dataSourceTableName: %s tableRef is empty", ds2.Name),
		},
		{
			name: "valid report query with valid templating but the presto table's table ref returns error",
			reportTemplate: &ReportQueryTemplateContext{
				Query:             "SELECT * FROM {| dataSourceTableName .Report.Inputs.MissingPrestoTableRef |}",
				RequiredInputs:    []string{"MissingPrestoTableRef"},
				ReportDataSources: []*metering.ReportDataSource{ds1},
				PrestoTables: []*metering.PrestoTable{
					newTestPrestoTable("", testNamespace, testSchemaName, testCatalogName, nil),
				},
			},
			templateContext: TemplateContext{
				Report: ReportTemplateInfo{
					Inputs: map[string]interface{}{"MissingPrestoTableRef": ds1.Name},
				},
			},
			expectErr:    true,
			expectErrMsg: fmt.Sprintf("error executing template: template: reportQueryTemplate:1:17: executing \"reportQueryTemplate\" at <dataSourceTableName ...>: error calling dataSourceTableName: tableRef PrestoTable %s not found", ds1.Status.TableRef.Name),
		},
		{
			name: "ReportDataSource dependencies were not found (due to empty prestoTable.Name) results in error",
			reportTemplate: &ReportQueryTemplateContext{
				Query:             "SELECT * FROM {| dataSourceTableName .Report.Inputs.EmptyPrestoTableName |}",
				RequiredInputs:    []string{"EmptyPrestoTableName"},
				ReportDataSources: []*metering.ReportDataSource{ds1},
				PrestoTables: []*metering.PrestoTable{
					newTestPrestoTable("", testNamespace, testSchemaName, testCatalogName, nil),
				},
			},
			templateContext: TemplateContext{
				Report: ReportTemplateInfo{
					Inputs: map[string]interface{}{"EmptyPrestoTableName": ""},
				},
			},
			expectErr:    true,
			expectErrMsg: "error executing template: template: reportQueryTemplate:1:17: executing \"reportQueryTemplate\" at <dataSourceTableName ...>: error calling dataSourceTableName: ReportDataSource  dependency not found",
		},
		{
			name: "Report dependency was not found (due to no reports) returns error",
			reportTemplate: &ReportQueryTemplateContext{
				Query:          "SELECT * FROM {| reportTableName .Report.Inputs.MissingReportsField |}",
				RequiredInputs: []string{"MissingReportsField"},
				PrestoTables: []*metering.PrestoTable{
					newTestPrestoTable(testReportQuery, testNamespace, testSchemaName, testCatalogName, nil),
				},
			},
			templateContext: TemplateContext{
				Report: ReportTemplateInfo{
					Inputs: map[string]interface{}{"MissingReportsField": testReportName},
				},
			},
			expectErr:    true,
			expectErrMsg: fmt.Sprintf("error executing template: template: reportQueryTemplate:1:17: executing \"reportQueryTemplate\" at <reportTableName .Rep...>: error calling reportTableName: Report %s dependency not found", testReportName),
		},
		{
			name: "Report dependency was not found (due to missing Report.Name) returns error",
			reportTemplate: &ReportQueryTemplateContext{
				Query:          "SELECT * FROM {| reportTableName .Report.Inputs.MissingReportNameInReportsField |}",
				RequiredInputs: []string{"MissingReportNameInReportsField"},
				Reports: []*metering.Report{
					testhelpers.NewReport("", testNamespace, testReportQuery, reportStart, reportEnd, metering.ReportStatus{}, nil, false),
				},
				PrestoTables: []*metering.PrestoTable{
					newTestPrestoTable(ds1.Status.TableRef.Name, testNamespace, testCatalogName, testCatalogName, nil),
				},
			},
			templateContext: TemplateContext{
				Report: ReportTemplateInfo{
					Inputs: map[string]interface{}{"MissingReportNameInReportsField": testReportName},
				},
			},
			expectErr:    true,
			expectErrMsg: fmt.Sprintf("error executing template: template: reportQueryTemplate:1:17: executing \"reportQueryTemplate\" at <reportTableName .Rep...>: error calling reportTableName: Report %s dependency not found", testReportName),
		},
		{
			name: "Report dependency was not found (due to Report.Name not matching name) returns error",
			reportTemplate: &ReportQueryTemplateContext{
				Reports: []*metering.Report{
					testhelpers.NewReport(testReportName, testNamespace, testReportQuery, reportStart, reportEnd, metering.ReportStatus{}, nil, false),
				},
				Query:             "SELECT * FROM {| reportTableName .Report.Inputs.ReportNameDoesNotMatchInputName |}",
				RequiredInputs:    []string{"ReportNameDoesNotMatchInputName"},
				ReportDataSources: []*metering.ReportDataSource{ds1},
				PrestoTables: []*metering.PrestoTable{
					newTestPrestoTable(ds1.Status.TableRef.Name, testNamespace, testSchemaName, testCatalogName, nil),
				},
			},
			templateContext: TemplateContext{
				Report: ReportTemplateInfo{
					Inputs: map[string]interface{}{"ReportNameDoesNotMatchInputName": "this-does-not-match"},
				},
			},
			expectErr:    true,
			expectErrMsg: "error executing template: template: reportQueryTemplate:1:17: executing \"reportQueryTemplate\" at <reportTableName .Rep...>: error calling reportTableName: Report this-does-not-match dependency not found",
		},
		{
			name: "invalid reportTableName templating (Status.TableRef.Name is missing) results in error and the expected errror output",
			reportTemplate: &ReportQueryTemplateContext{
				Query:          "SELECT * FROM {| reportTableName .Report.Inputs.MissingReportStatusTableRef|}",
				RequiredInputs: []string{"MissingReportStatusTableRef"},
				Reports: []*metering.Report{
					testhelpers.NewReport(testReportName, testNamespace, testReportQuery, reportStart, reportEnd, metering.ReportStatus{
						TableRef: v1.LocalObjectReference{Name: ""},
					}, nil, false),
				},
			},
			templateContext: TemplateContext{
				Report: ReportTemplateInfo{
					Inputs: map[string]interface{}{"MissingReportStatusTableRef": testReportName},
				},
			},
			expectErr:    true,
			expectErrMsg: fmt.Sprintf("error executing template: template: reportQueryTemplate:1:17: executing \"reportQueryTemplate\" at <reportTableName .Rep...>: error calling reportTableName: %s tableRef is empty", testReportName),
		},
		{
			name: "query's table ref and presto table don't match up returns error",
			reportTemplate: &ReportQueryTemplateContext{
				Query:          "SELECT * FROM {| reportTableName .Report.Inputs.EmptyReportStatusTableRef |}",
				RequiredInputs: []string{"EmptyReportStatusTableRef"},
				Reports: []*metering.Report{
					testhelpers.NewReport(testReportName, testNamespace, testReportQuery, reportStart, reportEnd, metering.ReportStatus{
						TableRef: v1.LocalObjectReference{Name: testReportName},
					}, nil, false),
				},
				PrestoTables: []*metering.PrestoTable{
					newTestPrestoTable(ds1.Status.TableRef.Name, testNamespace, testSchemaName, testCatalogName, nil),
				},
			},
			templateContext: TemplateContext{
				Report: ReportTemplateInfo{
					Inputs: map[string]interface{}{"EmptyReportStatusTableRef": testReportName},
				},
			},
			expectErr:    true,
			expectErrMsg: fmt.Sprintf("error executing template: template: reportQueryTemplate:1:17: executing \"reportQueryTemplate\" at <reportTableName .Rep...>: error calling reportTableName: tableRef PrestoTable %s not found", testReportName),
		},
		{
			name: "valid reportTableName templating (Status.TableRef.Name matches PrestoTable name) results in nil and expected query output",
			reportTemplate: &ReportQueryTemplateContext{
				Query:          "SELECT * FROM {| reportTableName .Report.Inputs.ValidReportTableNameQuery |}",
				RequiredInputs: []string{"ValidReportTableNameQuery"},
				Reports: []*metering.Report{
					testhelpers.NewReport(testReportName, testNamespace, testReportQuery, reportStart, reportEnd, metering.ReportStatus{
						TableRef: v1.LocalObjectReference{Name: testReportName},
					}, nil, false),
				},
				PrestoTables: []*metering.PrestoTable{
					newTestPrestoTable(testReportName, testNamespace, testSchemaName, testCatalogName, nil),
				},
			},
			templateContext: TemplateContext{
				Report: ReportTemplateInfo{
					Inputs: map[string]interface{}{"ValidReportTableNameQuery": testReportName},
				},
			},
			expectOutput: fmt.Sprintf("SELECT * FROM %s.%s.%s", testCatalogName, testNamespace, testReportName),
		},
		{
			name: "Query containing renderReportQuery is unable to be rendered returns error",
			reportTemplate: &ReportQueryTemplateContext{
				Query:          "SELECT * FROM {| renderReportQuery .Report.Inputs.MissingTemplateContext . |}",
				RequiredInputs: []string{"MissingTemplateContext"},
				ReportQueries: []*metering.ReportQuery{
					newTestReportQuery(testReportQuery, testNamespace, "SELECT * FROM {| renderReportQuery .Report.Inputs.QueryReportNameDoesNotMatchInputName . |}", nil),
				},
			},
			templateContext: TemplateContext{
				Report: ReportTemplateInfo{
					Inputs: map[string]interface{}{"MissingTemplateContext": testReportQuery},
				},
			},
			expectErr:    true,
			expectErrMsg: fmt.Sprintf("error executing template: template: reportQueryTemplate:1:17: executing \"reportQueryTemplate\" at <renderReportQuery .R...>: error calling renderReportQuery: unable to render query %s, err: error executing template: template: reportQueryTemplate:1:42: executing \"reportQueryTemplate\" at <.Report.Inputs.Query...>: invalid value; expected string", testReportQuery),
		},
		{
			name: "Query containing an unknown ReportQuery returns error",
			reportTemplate: &ReportQueryTemplateContext{
				Query:          "SELECT * FROM {| renderReportQuery .Report.Inputs.ReportQueryNameDoesNotExist . |}",
				RequiredInputs: []string{"ReportQueryNameDoesNotExist"},
				ReportQueries: []*metering.ReportQuery{
					testValidQuery,
				},
			},
			templateContext: TemplateContext{
				Report: ReportTemplateInfo{
					Inputs: map[string]interface{}{"ReportQueryNameDoesNotExist": "FakeReportQueryName"},
				},
			},
			expectErr:    true,
			expectErrMsg: "error executing template: template: reportQueryTemplate:1:17: executing \"reportQueryTemplate\" at <renderReportQuery .R...>: error calling renderReportQuery: unknown ReportQuery FakeReportQueryName",
		},
		{
			name: "renderReportQuery ReportQueries unable to render reportQuery (due to wrong number of args in renderReportQuery templating)",
			reportTemplate: &ReportQueryTemplateContext{
				Query:          "SELECT * FROM {| renderReportQuery .Report.Inputs.MissingTemplateContext |}",
				RequiredInputs: []string{"MissingTemplateContext"},
				ReportQueries: []*metering.ReportQuery{
					newTestReportQuery(testReportQuery, testNamespace, "SELECT * FROM {| renderReportQuery .Report.Inputs.MissingTemplateContext |}", nil),
				},
			},
			templateContext: TemplateContext{
				Report: ReportTemplateInfo{
					Inputs: map[string]interface{}{"MissingTemplateContext": testReportQuery},
				},
			},
			expectErr:    true,
			expectErrMsg: "error executing template: template: reportQueryTemplate:1:17: executing \"reportQueryTemplate\" at <renderReportQuery>: wrong number of args for renderReportQuery: want 2 got 1",
		},
		{
			name: "valid renderReportQuery templating returns nil and the expected query output",
			reportTemplate: &ReportQueryTemplateContext{
				Query:          "SELECT * FROM ({| renderReportQuery .Report.Inputs.ValidRenderReportQuery . |}) AS test_sub_query",
				RequiredInputs: []string{"ValidRenderReportQuery"},
				ReportQueries: []*metering.ReportQuery{
					testValidQueryInputs,
				},
			},
			templateContext: TemplateContext{
				Report: ReportTemplateInfo{
					Inputs: map[string]interface{}{"ValidRenderReportQuery": testValidQuery.Name},
				},
			},
			expectOutput: "SELECT * FROM (SELECT * FROM test_table) AS test_sub_query",
		},
		{
			name: "invalid prestoTimestamp templating (no timestamp in Report.Inputs) returns error",
			reportTemplate: &ReportQueryTemplateContext{
				Query:          "SELECT timestamp {| default .Report.ReportingStart .Report.Inputs.ReportingStart | prestoTimestamp |} AS period_start",
				RequiredInputs: []string{},
			},
			templateContext: TemplateContext{
				Report: ReportTemplateInfo{
					Inputs: map[string]interface{}{"ReportingStart": nil},
				},
			},
			expectErr:    true,
			expectErrMsg: "error executing template: template: reportQueryTemplate:1:83: executing \"reportQueryTemplate\" at <prestoTimestamp>: error calling prestoTimestamp: got nil timestamp",
		},
		{
			name: "valid prestoTimestamp templating returns nil and the expected query output",
			reportTemplate: &ReportQueryTemplateContext{
				Query:          "SELECT timestamp {| .Report.Inputs.ReportingStart | prestoTimestamp |} AS period_start",
				RequiredInputs: []string{"ReportingStart"},
			},
			templateContext: TemplateContext{
				Report: ReportTemplateInfo{
					Inputs: map[string]interface{}{"ReportingStart": reportStart},
				},
			},
			expectOutput: fmt.Sprintf("SELECT timestamp %s AS period_start", reportStart.Format(presto.TimestampFormat)),
		},
		{
			name: "invalid prestoTimestamp input (references an invalid format - integer) returns error",
			reportTemplate: &ReportQueryTemplateContext{
				Query:          "SELECT timestamp {| prestoTimestamp .Report.Inputs.ReportingStart |} AS period_start",
				RequiredInputs: []string{"ReportingStart"},
			},
			templateContext: TemplateContext{
				Report: ReportTemplateInfo{
					Inputs: map[string]interface{}{"ReportingStart": 5},
				},
			},
			expectErr:    true,
			expectErrMsg: "error executing template: template: reportQueryTemplate:1:20: executing \"reportQueryTemplate\" at <prestoTimestamp .Rep...>: error calling prestoTimestamp: couldn't convert 5 to a Presto timestamp",
		},
		{
			name: "invalid prestoTimestamp input (references an invalid format - integer and default is overrided) returns error",
			reportTemplate: &ReportQueryTemplateContext{
				Query:          "SELECT timestamp {| default .Report.ReportingStart .Report.Inputs.ReportingStart | prestoTimestamp |} AS period_start",
				RequiredInputs: []string{},
			},
			templateContext: TemplateContext{
				Report: ReportTemplateInfo{
					ReportingStart: reportStart,
					Inputs:         map[string]interface{}{"ReportingStart": 5},
				},
			},
			expectErr:    true,
			expectErrMsg: "error executing template: template: reportQueryTemplate:1:83: executing \"reportQueryTemplate\" at <prestoTimestamp>: error calling prestoTimestamp: couldn't convert 5 to a Presto timestamp",
		},
		{
			name: "valid prestoTimestamp input (references time.Time object) returns nil and the expected query output",
			reportTemplate: &ReportQueryTemplateContext{
				Query:          "SELECT timestamp {| prestoTimestamp .Report.Inputs.ReportingStart |} AS period_start",
				RequiredInputs: []string{"ReportingStart"},
			},
			templateContext: TemplateContext{
				Report: ReportTemplateInfo{
					Inputs: map[string]interface{}{"ReportingStart": *reportStart},
				},
			},
			expectOutput: fmt.Sprintf("SELECT timestamp %s AS period_start", reportStart.Format(presto.TimestampFormat)),
		},
		{
			name: "valid prestoTimestamp input (references string) returns nil and the expected query output",
			reportTemplate: &ReportQueryTemplateContext{
				Query:          "SELECT {| .Report.Inputs.Namespace |} as namespace FROM test_table",
				RequiredInputs: []string{"Namespace"},
			},
			templateContext: TemplateContext{
				Report: ReportTemplateInfo{
					Inputs: map[string]interface{}{"Namespace": testNamespace},
				},
			},
			expectOutput: fmt.Sprintf("SELECT %s as namespace FROM test_table", testNamespace),
		},
		{
			name: "valid prestoTimestamp input (references string and dataSourceTableName) returns nil and the expected query output",
			reportTemplate: &ReportQueryTemplateContext{
				Query:             "SELECT {| .Report.Inputs.Namespace |} as namespace FROM {| dataSourceTableName .Report.Inputs.PodRequestCpuCoresDataSourceName |}",
				RequiredInputs:    []string{"Namespace", "PodRequestCpuCoresDataSourceName"},
				ReportDataSources: []*metering.ReportDataSource{ds1},
				PrestoTables: []*metering.PrestoTable{
					newTestPrestoTable(ds1.Status.TableRef.Name, testNamespace, testNamespace, testCatalogName, nil),
				},
			},
			templateContext: TemplateContext{
				Report: ReportTemplateInfo{
					Inputs: map[string]interface{}{"Namespace": testNamespace, "PodRequestCpuCoresDataSourceName": ds1.Name},
				},
			},
			expectOutput: fmt.Sprintf("SELECT %s as namespace FROM %s.%s.%s", testNamespace, testCatalogName, testNamespace, ds1.Status.TableRef.Name),
		},
		{
			name: "valid prestoTimestamp input (references string to be converted to time) returns nil and the expected query output",
			reportTemplate: &ReportQueryTemplateContext{
				Query:          "SELECT timestamp {| .Report.Inputs.ReportingStart | prestoTimestamp |} AS period_start",
				RequiredInputs: []string{"ReportingStart"},
			},
			templateContext: TemplateContext{
				Report: ReportTemplateInfo{
					Inputs: map[string]interface{}{"ReportingStart": testTimeString},
				},
			},
			expectOutput: "SELECT timestamp 2019-05-01 15:00:05.000 AS period_start",
		},
		{
			name: "prestoTimestamp input is missing but default is de***REMOVED***ned returns nil and the expected query output",
			reportTemplate: &ReportQueryTemplateContext{
				Query:          "SELECT timestamp {| default .Report.ReportingStart .Report.Inputs.ReportingStart | prestoTimestamp |} AS period_start",
				RequiredInputs: []string{},
			},
			templateContext: TemplateContext{
				Report: ReportTemplateInfo{
					ReportingStart: reportStart,
					Inputs:         map[string]interface{}{"ReportingStart": nil},
				},
			},
			expectOutput: fmt.Sprintf("SELECT timestamp %s AS period_start", reportStart.Format(presto.TimestampFormat)),
		},
		{
			name: "prestoTimestamp input and default are nil and returns error",
			reportTemplate: &ReportQueryTemplateContext{
				Query:          "SELECT timestamp {| default .Report.ReportingStart .Report.Inputs.ReportingStart | prestoTimestamp |} AS period_start",
				RequiredInputs: []string{},
			},
			templateContext: TemplateContext{
				Report: ReportTemplateInfo{
					Inputs: map[string]interface{}{"ReportingStart": nil},
				},
			},
			expectErr:    true,
			expectErrMsg: "error executing template: template: reportQueryTemplate:1:83: executing \"reportQueryTemplate\" at <prestoTimestamp>: error calling prestoTimestamp: got nil timestamp",
		},
		{
			name: "valid prometheusMetricPartitionFormat templating results in nil and the expected query output",
			reportTemplate: &ReportQueryTemplateContext{
				Query:          "SELECT datetime {| .Report.Inputs.ReportingStart | prometheusMetricPartitionFormat |} AS period_start",
				RequiredInputs: []string{"ReportingStart"},
			},
			templateContext: TemplateContext{
				Report: ReportTemplateInfo{
					Inputs: map[string]interface{}{"ReportingStart": reportStart},
				},
			},
			expectOutput: fmt.Sprintf("SELECT datetime %s AS period_start", reportStart.Format(prestostore.PrometheusMetricTimestampPartitionFormat)),
		},
		{
			name: "query containing valid prestoTimestamp and prometheusMetricPartitionFormat templating results in nil and the expected query output",
			reportTemplate: &ReportQueryTemplateContext{
				Query:          "SELECT timestamp {| default .Report.ReportingStart .Report.Inputs.ReportingStart | prestoTimestamp |} AS period_start WHERE dt >= {| default .Report.ReportingStart .Report.Inputs.ReportingStart | prometheusMetricPartitionFormat |}",
				RequiredInputs: []string{},
			},
			templateContext: TemplateContext{
				Report: ReportTemplateInfo{
					Inputs: map[string]interface{}{"ReportingStart": reportStart},
				},
			},
			expectOutput: fmt.Sprintf("SELECT timestamp %s AS period_start WHERE dt >= %s", reportStart.Format(presto.TimestampFormat), reportStart.Format(prestostore.PrometheusMetricTimestampPartitionFormat)),
		},
		{
			name: "query containing missing inputs but defaults are de***REMOVED***ned templating results in nil and the expected query output",
			reportTemplate: &ReportQueryTemplateContext{
				Query:          "SELECT timestamp {| default .Report.ReportingStart .Report.Inputs.ReportingStart | prestoTimestamp |} AS period_start WHERE dt >= {| default .Report.Inputs.TestTime .Report.Inputs.ReportingStart | prometheusMetricPartitionFormat |}",
				RequiredInputs: []string{},
			},
			templateContext: TemplateContext{
				Report: ReportTemplateInfo{
					ReportingStart: reportStart,
					Inputs:         map[string]interface{}{"ReportingStart": nil, "TestTime": reportTestTime},
				},
			},
			expectOutput: fmt.Sprintf("SELECT timestamp %s AS period_start WHERE dt >= %s", reportStart.Format(presto.TimestampFormat), reportTestTime.Format(prestostore.PrometheusMetricTimestampPartitionFormat)),
		},
		{
			name: "incorrect number of inputs provided in RequiredInputs",
			reportTemplate: &ReportQueryTemplateContext{
				Query:          "SELECT * FROM {| reportTableName .Report.Inputs.NamespaceCpuUsage |} WHERE namespace = {| .Report.Inputs.Namespace |}",
				RequiredInputs: []string{"NamespaceCpuUsage", "Namespace"},
				Reports: []*metering.Report{
					testhelpers.NewReport(testReportName, testNamespace, testReportQuery, reportStart, reportEnd, metering.ReportStatus{
						TableRef: v1.LocalObjectReference{Name: ds1.Status.TableRef.Name},
					}, nil, false),
				},
				PrestoTables: []*metering.PrestoTable{
					newTestPrestoTable(ds1.Status.TableRef.Name, testNamespace, testCatalogName, testCatalogName, nil),
				},
			},
			templateContext: TemplateContext{
				Report: ReportTemplateInfo{
					Inputs: map[string]interface{}{"NamespaceCpuUsage": testReportName},
				},
			},
			expectErr:    true,
			expectErrMsg: "missing inputs: Namespace",
		},
		{
			name: "valid report query with valid templating (references an integer input type) returns nil and expected query output",
			reportTemplate: &ReportQueryTemplateContext{
				Query:          "SELECT * FROM test_table WHERE int_col > {| .Report.Inputs.IntegerInput |}",
				RequiredInputs: []string{"IntegerInput"},
			},
			templateContext: TemplateContext{
				Report: ReportTemplateInfo{
					Inputs: map[string]interface{}{"IntegerInput": 5},
				},
			},
			expectOutput: "SELECT * FROM test_table WHERE int_col > 5",
		},
	}

	for _, testCase := range testTable {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			output, err := RenderQuery(testCase.reportTemplate, testCase.templateContext)

			if testCase.expectErr {
				assert.EqualErrorf(t, err, testCase.expectErrMsg, "expected that RenderQuery would return the correct error message")
			} ***REMOVED*** {
				assert.NoErrorf(t, err, "expected the report would return no error, but got an error.")
				assert.EqualValuesf(t, testCase.expectOutput, output, "expected that RenderQuery would return the expected output")
			}
		})
	}
}

// newPrestoTableCustom allows us to insert a custom name, as opposed to testhelpers.NewPrestoTable
func newTestPrestoTable(name, namespace, schema, catalog string, columns []presto.Column) *v1alpha1.PrestoTable {
	return &v1alpha1.PrestoTable{
		ObjectMeta: meta.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Status: v1alpha1.PrestoTableStatus{
			Catalog:   catalog,
			Schema:    schema,
			TableName: name,
			Columns:   columns,
		},
		Spec: v1alpha1.PrestoTableSpec{
			Catalog:   catalog,
			Schema:    schema,
			TableName: name,
			Columns:   columns,
		},
	}
}

// newReportQueryCustomQuery helps create a ReportQuery with a custom ReportDataSource and adds a Query ***REMOVED***eld
func newTestReportQuery(testReportName, testNamespace, query string, inputs []metering.ReportQueryInputDe***REMOVED***nition) *metering.ReportQuery {
	return &metering.ReportQuery{
		ObjectMeta: meta.ObjectMeta{
			Name:      testReportName,
			Namespace: testNamespace,
		},
		Spec: metering.ReportQuerySpec{
			Inputs: inputs,
			Query:  query,
		},
	}
}
