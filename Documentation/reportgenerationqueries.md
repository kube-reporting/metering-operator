# Report Generation Queries

Customizing how Operator Metering generates reports is done using a custom resource called a `ReportGenerationQuery`.

These `ReportGenerationQuery` resources control the SQL queries that can be used to produce a report.
When writing a [report](report.md) you can specify the query it will use by setting the `spec.generationQuery` ***REMOVED***eld to `metadata.name` of any `ReportGenerationQuery` in the reporting-operator's namespace.

## Fields

- `query`: A [SQL SELECT statement][presto-select]. This SQL statement supports [go templates][go-templates] and provides additional custom functions speci***REMOVED***c to Operator Metering (de***REMOVED***ned in the [templating](#templating) section below).
- `columns`: A list of columns that match the schema of the results of the query. The order of these columns must match the order of the columns returned by the SELECT statement. Columns have 3 ***REMOVED***elds, `name`, `type`, and `unit`. Each ***REMOVED***eld is covered in more detail below.
  - `name`: This is the name of the column returned in the `SELECT` statement.
  - `type`: This is the [Hive][hive-types] column type. Currently due to implementation details, column types are expressed using hive types. In the future, this will likely be switched to using the Presto native types. This also has an effect that queries with columns containing complex types such as `maps` or `arrays` cannot be used by `Reports` or `ScheduledReports`.
  - `unit`: Unit refers to the unit of measurement of the column.
  - `tableHidden`: Takes a boolean, when true, hides the column from report results depending on the format and endpoint. See [api docs for details][apiTable].
- `reportDataSources`: This is a list of `ReportDataSource` resources that this `ReportGenerationQuery` depends on. These data sources can be referenced as database tables in the `query` using the `dataSourceTableName` template function.
- `reportQueries`: This is a list of other `ReportGenerationQuery` resources that this `ReportGenerationQuery` depends on that have `view.disabled` set to false. Queries in this list can be re-used by querying the database view created, and using `generationQueryViewName` templating function to reference the view by name.
- `dynamicReportQueries`: This is a list of other `ReportGenerationQuery` resources that this `ReportGenerationQuery` depends on, that have `view.disabled` set to true, these are queries that depend on the `.Report` variable. Queries in the list can be re-used by injecting them into the current query using the `renderReportGenerationQuery` template function.
- `view`: This section controls options related to creating a view from the `query` when the `ReportGenerationQuery` resource is created.
    - `view.disabled`: This is false by default, and if set to true, it will prevent the default behavior of creating a database view using the contents of the `query`. This cannot be true if `dynamicReportQueries` is non-empty or if the `query` depends on the `.Report` templating variables.

## Templating

Because much of the type of analysis being done depends on user-input, and because we want to enable users to re-use queries with copying & pasting things around, Operator Metering supports the [go templating language][go-templates] to dynamically generate the SQL statements contained within the `spec.query` ***REMOVED***eld of `ReportGenerationQuery`.
For example, when generating a report, the query needs to know what time range to consider when producing a report, so this is information is exposed within the template context as variables you can use and reference with various [template functions](#template-functions).
Most of these functions are for referring to other resources such as `ReportDataSources` or `ReportGenerationQueries` as either tables, views, or sub-queries, and for formatting various types for use within the SQL query.

### Template variables

- `Report`: This object has two ***REMOVED***elds, `ReportingStart` and `ReportingEnd` which are the value of the `spec.reportingStart` and `spec.reportingEnd` for a `Report`. For a `ScheduledReport` the values map to the speci***REMOVED***c period being collected when the `ScheduledReport` runs.
  - `ReportingStart`: A [time.Time][go-time] object that is generally used to ***REMOVED***lter the results of a `SELECT` query using a `WHERE` clause.
  - `ReportingEnd`: A [time.Time][go-time] object that is generally used to ***REMOVED***lter the results of a `SELECT` query using a `WHERE` clause. Built-in queries select datapoints matching `ReportingStart <= timestamp > ReportingEnd`.
- `DynamicDependentQueries`: This is a list of `ReportGenerationQuery` objects that were listed in the `spec.dynamicReportQueries` ***REMOVED***eld. Generally this list isn't directly referenced in query, but is used indirectly with the `renderReportGenerationQuery` [template function](#template-functions).
- `Inputs`: This is a `map[string]interface{}` of inputs passed in via the Report or ScheduledReport's `spec.inputs`. The value currently is always a string unless the input's name is `ReportingStart` or `ReportingEnd`, in which case it's converted to a [time.Time][go-time].

### Template functions

Below is a list of the available template functions and descriptions on what they do.

- `dataSourceTableName`: Takes a one argument, a string representing a `ReportDataSource` name and outputs a string which is the corresponding table name of the `ReportDataSource` speci***REMOVED***ed.
- `generationQueryViewName`: Takes one argument, a string representing a `ReportGenerationQuery` name and outputs a string which is the corresponding view name of the `ReportGenerationQuery` speci***REMOVED***ed.
- `renderReportGenerationQuery`: Takes two arguments, a string representing a `ReportGenerationQuery` name, the template context (usually this is just `.` in the template), and returns a string containing the speci***REMOVED***ed `ReportGenerationQuery` in it's rendered form, using the 2nd argument as the context for the template rendering.
- `prestoTimestamp`: Takes a [time.Time][go-time] object as the argument, and outputs a string timestamp. Usually this is used on `.Report.ReportingStart` and `.Report.ReportingEnd`.
- `billingPeriodFormat`: Takes a [time.Time][go-time] object as the argument, and outputs a string timestamp that can be used for comparing to `awsBilling` an ReportDataSource's `partition_start` and `partition_stop` columns.

In addition to the above functions, the reporting-operator includes all of the functions from [Sprig - useful template functions for Go templates.
][sprig].

## Example ReportGenerationQueries

Before going into examples, there's an important convention that all the built-in `ReportGenerationQueries` follow that is worth calling out, as these examples will demonstrate them heavily.

The convention I am referring to is the fact that there are quite a few ReportGenerationQueries suf***REMOVED***xed with `-raw` in their `metadata.name`.
These queries are not intended to be used by Reports, or ScheduledReports, but are intended to be purely for re-use.
Currently, these queries suf***REMOVED***xed with `-raw` in their name are generally only ever used as database views that are referenced in other queries.
Additionally, due to implementation reasons, the `types` within the `columns` are using [Hive column types][hive-types] instead of the Presto column types.
This has a negative effect the results in being unable to use `ReportGenerationQueries` containing with complex types (array, maps) with `Reports` and `ScheduledReports`, which is why the `ReportGenerationQueries` that are _not_ suf***REMOVED***xed in `-raw` never expose those types in their output.


The example below is a built-in `ReportGenerationQuery` that is installed with Operator Metering by default.
The query is not intended to be used by Reports, but instead is intended to be re-used by other `ReportGenerationQueries`, which is why it only does simple extraction of ***REMOVED***elds, and calculations.

The important things to note with this query is that it's querying a database table containing Prometheus metric data for the `pod-request-memory-bytes` `ReportDataSource`, and it's getting the table name using the `dataSourceTableName` template function.

```yaml
apiVersion: metering.openshift.io/v1alpha1
kind: ReportGenerationQuery
metadata:
  name: "pod-memory-request-raw"
  labels:
    operator-metering: "true"
spec:
  reportDataSources:
  - "pod-request-memory-bytes"
  columns:
  - name: pod
    type: string
    unit: kubernetes_pod
  - name: namespace
    type: string
    unit: kubernetes_namespace
  - name: node
    type: string
    unit: kubernetes_node
  - name: labels
    type: map<string, string>
    tableHidden: true
  - name: pod_request_memory_bytes
    type: double
    unit: bytes
  - name: timeprecision
    type: double
    unit: seconds
  - name: pod_request_memory_byte_seconds
    type: double
    unit: byte_seconds
  - name: timestamp
    type: timestamp
    unit: date
  query: |
      SELECT labels['pod'] as pod,
          labels['namespace'] as namespace,
          element_at(labels, 'node') as node,
          labels,
          amount as pod_request_memory_bytes,
          timeprecision,
          amount * timeprecision as pod_request_memory_byte_seconds,
          "timestamp"
      FROM {| dataSourceTableName "pod-request-memory-bytes" |}
      WHERE element_at(labels, 'node') IS NOT NULL
```

This next example is also one of the built-in `ReportGenerationQueries`.
This example, unlike the previous is designed to be used with Reports and ScheduledReports.
It summarizes the information exposed in the example above by Kubernetes namespace, and reduces the output ***REMOVED***elds down to the ones that matter for this speci***REMOVED***c use-case.

The important things to note with this example is that it's depending on the previous `ReportGenerationQuery` and referencing the database view by using the `generationQueryViewName` template function.

```yaml
apiVersion: metering.openshift.io/v1alpha1
kind: ReportGenerationQuery
metadata:
  name: "namespace-memory-request"
  labels:
    operator-metering: "true"
spec:
  reportQueries:
  - "pod-memory-request-raw"
  view:
    disabled: true
  columns:
  - name: period_start
    type: timestamp
    unit: date
  - name: period_end
    type: timestamp
    unit: date
  - name: namespace
    type: string
    unit: kubernetes_namespace
  - name: data_start
    type: timestamp
    unit: date
  - name: data_end
    type: timestamp
    unit: date
  - name: pod_request_memory_byte_seconds
    type: double
    unit: byte_seconds
  inputs:
  - name: ReportingStart
  - name: ReportingEnd
  query: |
    SELECT
      timestamp '{| default .Report.ReportingStart .Report.Inputs.ReportingStart| prestoTimestamp |}' AS period_start,
      timestamp '{| default .Report.ReportingEnd .Report.Inputs.ReportingEnd | prestoTimestamp |}' AS period_end,
      namespace,
      min("timestamp") as data_start,
      max("timestamp") as data_end,
      sum(pod_request_memory_byte_seconds) as pod_request_memory_byte_seconds
    FROM {| generationQueryViewName "pod-memory-request-raw" |}
    WHERE "timestamp" >= timestamp '{| default .Report.ReportingStart .Report.Inputs.ReportingStart | prestoTimestamp |}'
    AND "timestamp" < timestamp '{| default .Report.ReportingEnd .Report.Inputs.ReportingEnd | prestoTimestamp |}'
    GROUP BY namespace
    ORDER BY pod_request_memory_byte_seconds DESC
```

## Modifying Columns For Report Display

You can modify the ReportGenerationQuery to display all columns or to hide or show columns as needed. The full endpoint displays all columns.
To query report results using the reporting-operator API for full endpoint the api call is:
`http://127.0.0.1:8001/api/v1/namespaces/metering/services/http:reporting-operator:http/proxy/api/v2/reports/namespace-cpu-request/full?format=json`
This example is showing the `namespace-cpu-request` query in JSON format.

To use the TableHidden display feature:
- enter `TableHidden` ***REMOVED***eld in query as true or false. In `node-cpu-capacity` the `labels` tableHidden value is set to `true`.
- Next run a report.
To query report results using the reporting-operator API for tableHidden endpoint the api call is:
`http://127.0.0.1:8001/api/v1/namespaces/metering/services/http:reporting-operator:http/proxy/api/v2/reports/namespace-cpu-request/table?format=json`

[apiTable]: api.md#v2-reports-table
[presto-select]: https://prestodb.io/docs/current/sql/select.html
[hive-types]: https://cwiki.apache.org/confluence/display/Hive/LanguageManual+Types#LanguageManualTypes-Overview
[presto-functions]: https://prestodb.io/docs/current/functions.html
[go-templates]: https://golang.org/pkg/text/template/
[go-time]: https://golang.org/pkg/time/#Time
[sprig]: https://masterminds.github.io/sprig/
