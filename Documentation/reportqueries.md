# Report Queries

Customizing how Operator Metering generates reports is done using a custom resource called a `ReportQuery`.

These `ReportQuery` resources control the SQL queries that can be used to produce a report.
When writing a [report](reports.md) you can specify the query it will use by setting the `spec.query` ***REMOVED***eld to `metadata.name` of any `ReportQuery` in the reporting-operator's namespace.

## Fields

- `query`: A [SQL SELECT statement][presto-select]. This SQL statement supports [go templates][go-templates] and provides additional custom functions speci***REMOVED***c to Operator Metering (de***REMOVED***ned in the [templating](#templating) section below).
- `columns`: A list of columns that match the schema of the results of the query. The order of these columns must match the order of the columns returned by the SELECT statement. Columns have 3 ***REMOVED***elds, `name`, `type`, and `unit`. Each ***REMOVED***eld is covered in more detail below.
  - `name`: This is the name of the column returned in the `SELECT` statement.
  - `type`: This is the [Presto][presto-types] column type.
  - `unit`: Unit refers to the unit of measurement of the column.
  - `tableHidden`: Takes a boolean, when true, hides the column from report results depending on the format and endpoint. See [api docs for details][apiTable].
- `inputs`: A list of inputs this report query accepts to control its behavior. For more in depth details, see the [query inputs](#query-inputs) section.
  - `name`: The name used to refer to the input in the `Report` or `ScheduledReport` `spec.inputs` and within the queries template variables (see below).
  - `required`: A boolean indicating if this input is required for the query to run. Defaults to false.
  - `type`: An optional type indicating what data type this input takes. Available options are `string`, `time`, and `int`, `ReportDataSource`, `ReportQuery`, and `Report`. If left empty, it defaults to `string`. For more details, see the [query input types](#query-input-types) section.
  - `default`: An optional default value to use if unspeci***REMOVED***ed.

## Templating

Because much of the type of analysis being done depends on user-input, and because we want to enable users to re-use queries with copying & pasting things around, Operator Metering supports the [go templating language][go-templates] to dynamically generate the SQL statements contained within the `spec.query` ***REMOVED***eld of `ReportQuery`.
For example, when generating a report, the query needs to know what time range to consider when producing a report, so this is information is exposed within the template context as variables you can use and reference with various [template functions](#template-functions).
Most of these functions are for referring to other resources such as `ReportDataSources` or `ReportQueries` as either tables, views, or sub-queries, and for formatting various types for use within the SQL query.

### Template variables

- `Report`: This object has two ***REMOVED***elds, `ReportingStart` and `ReportingEnd` which are the value of the `spec.reportingStart` and `spec.reportingEnd` for a `Report`. For a `Report` with a `spec.schedule` set, the values map to the speci***REMOVED***c period being collected when the `Report` runs.
  - `ReportingStart`: A [time.Time][go-time] object that is generally used to ***REMOVED***lter the results of a `SELECT` query using a `WHERE` clause.
  - `ReportingEnd`: A [time.Time][go-time] object that is generally used to ***REMOVED***lter the results of a `SELECT` query using a `WHERE` clause. Built-in queries select datapoints matching `ReportingStart <= timestamp > ReportingEnd`.
  - `Inputs`: This is a `map[string]interface{}` of inputs passed in via the Report's `spec.inputs`. The values type is based on the report queries input de***REMOVED***nition [type](#***REMOVED***elds), and defaults to string unless the input's name is `ReportingStart` or `ReportingEnd`, in which case it's converted to a [time.Time][go-time] automatically.

#### Query Inputs

Within the template context, the `.Report.Inputs` template variable contains a map where each key is the name of an input de***REMOVED***ned by the ReportQuery's `spec.inputs`.
The value of an input depends on a if the user speci***REMOVED***ed a value in their `Report` or `ReportDataSource`, or if there is a default value de***REMOVED***ned.

- Depending on what is referencing the ReportQuery: The value comes from a `Report`'s `spec.inputs` or from a `ReportDataSource`'s `spec.reportQueryView.inputs`.
- If not provided, it will use the `default` value as de***REMOVED***ned the ReportQuery's `spec.inputs`, if one exists.
- Otherwise: the zero value for the type, according to Go's zero value rules.

##### Query Input types

Each input can have a different `type`, which determines how the input should be processed.

Available options are `varchar`, `time`, and `int`, `ReportDataSource`, `ReportQuery`, and `Report`.
If left empty, it defaults to `varchar`.

For each of these types, the behavior varies:

- `varchar`: Varchars are passed through directly.
- `time`: A string value is parsed as an RFC3339 timestamp. Within the template context, the variable with be a Go [time.Time][go-time] object.
- `int`: An int value is passed through as a Go [int](https://golang.org/pkg/builtin/#int).
- `ReportDataSource`: A string value referencing the name of a [ReportDataSource][reportdatasources] within the same namespace as the query. When this query is referenced by a Report or ReportDataSource, all `ReportDataSource` inputs are validated by checking that all the ReportDataSources speci***REMOVED***ed exist.
- `ReportQuery`: A string value referencing the name of a [ReportQuery][reportqueries] within the same namespace as the query. When this query is referenced by a Report or ReportDataSource, all `ReportQuery` inputs are validated by checking that all the ReportQueries speci***REMOVED***ed exist.
- `Report`: A string value referencing the name of a [Report][reports] within the same namespace as the query. When this query is referenced by a Report or ReportDataSource, all `Report` inputs are validated by checking that all the Reports speci***REMOVED***ed exist.

### Template functions

Below is a list of the available template functions and descriptions on what they do.

- `dataSourceTableName`: Takes a one argument, a string referencing a `ReportDataSource` by name, and outputs a string which is the corresponding table name of the `ReportDataSource` speci***REMOVED***ed.
- `reportTableName`: Takes a one argument, a string referencing a `Report` by name, and outputs a string which is the corresponding table name of the `Report` speci***REMOVED***ed.
- `renderReportQuery`: Takes two arguments, a string referencing a `ReportQuery` by name, the template context (usually this is just `.` in the template), and returns a string containing the speci***REMOVED***ed `ReportQuery` in its rendered form, using the 2nd argument as the context for the template rendering.
- `prestoTimestamp`: Takes a [time.Time][go-time] object as the argument, and outputs a string timestamp. Usually this is used on `.Report.ReportingStart` and `.Report.ReportingEnd`.
- `prometheusMetricPartitionFormat`: Takes a [time.Time][go-time] object as the argument, and outputs a string in the form of `year-month-day`, eg: `2006-01-02`. Usually this is used on `.Report.ReportingStart` and `.Report.ReportingEnd`.
- `billingPeriodFormat`: Takes a [time.Time][go-time] object as the argument, and outputs a string timestamp that can be used for comparing to `awsBilling` an ReportDataSource's `partition_start` and `partition_stop` columns.

In addition to the above functions, the reporting-operator includes all of the functions from [Sprig - useful template functions for Go templates.][sprig].

## Example ReportQueries

Before going into examples, there's an important convention that all the built-in `ReportQueries` follow that is worth calling out, as these examples will demonstrate them heavily.

The convention I am referring to is the fact that there are quite a few ReportQueries suf***REMOVED***xed with `-raw` in their `metadata.name`.
These queries are not intended to be used by Reports, but are intended to be purely for re-use.
Currently, these queries suf***REMOVED***xed with `-raw` in their name are generally have no ***REMOVED***ltering and are used by [ReportDataSources to create views][view-datasources].
Additionally, `-raw` queries often expose complex types (array, maps) which are incompatible with `Reports`, which is why the `ReportQueries` that are _not_ suf***REMOVED***xed in `-raw` never expose those types in their columns list.

The example below is a built-in `ReportQuery` that is installed with Operator Metering by default.
The query is not intended to be used by Reports, but instead is intended to be re-used by other `ReportQueries`, which is why it only does simple extraction of ***REMOVED***elds, and calculations.

The important things to note with this query is that it's querying a database table containing Prometheus metric data for the `pod-request-memory-bytes` `ReportDataSource`, and it's getting the table name using the `dataSourceTableName` template function.

```yaml
apiVersion: metering.openshift.io/v1alpha1
kind: ReportQuery
metadata:
  name: pod-memory-request-raw
  labels:
    operator-metering: "true"
spec:
  columns:
  - name: pod
    type: varchar
    unit: kubernetes_pod
  - name: namespace
    type: varchar
    unit: kubernetes_namespace
  - name: node
    type: varchar
    unit: kubernetes_node
  - name: labels
    tableHidden: true
    type: map<string, string>
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
  - name: dt
    type: varchar
  inputs:
  - name: PodRequestMemoryBytesDataSourceName
    type: ReportDataSource
    default: pod-request-memory-bytes
  query: |
    SELECT labels['pod'] as pod,
        labels['namespace'] as namespace,
        element_at(labels, 'node') as node,
        labels,
        amount as pod_request_memory_bytes,
        timeprecision,
        amount * timeprecision as pod_request_memory_byte_seconds,
        "timestamp",
        dt
    FROM {| dataSourceTableName .Report.Inputs.PodRequestMemoryBytesDataSourceName |}
    WHERE element_at(labels, 'node') IS NOT NULL
```

This next example is also one of the built-in `ReportQueries`.
This example, unlike the previous is designed to be used with Reports.

## Modifying Columns For Report Display

You can modify the ReportQuery to display all columns or to hide or show columns as needed. The full endpoint displays all columns.
To query report results using the reporting-operator API for full endpoint the api call is:
`http://127.0.0.1:8001/api/v1/namespaces/metering/services/http:reporting-operator:http/proxy/api/v2/reports/metering/namespace-cpu-request/full?format=json`
This example is showing the `namespace-cpu-request` query in JSON format.

To use the TableHidden display feature:
- enter `TableHidden` ***REMOVED***eld in query as true or false. In `node-cpu-capacity` the `labels` tableHidden value is set to `true`.
- Next run a report.
To query report results using the reporting-operator API for tableHidden endpoint the api call is:
`http://127.0.0.1:8001/api/v1/namespaces/metering/services/http:reporting-operator:http/proxy/api/v2/reports/metering/namespace-cpu-request/table?format=json`

[apiTable]: api.md#v2-reports-table
[presto-select]: https://prestodb.io/docs/current/sql/select.html
[presto-types]: https://prestosql.io/docs/current/language/types.html
[presto-functions]: https://prestodb.io/docs/current/functions.html
[go-templates]: https://golang.org/pkg/text/template/
[go-time]: https://golang.org/pkg/time/#Time
[sprig]: https://masterminds.github.io/sprig/
[view-datasources]: reportdatasources.md#ReportQuery-View-Datasource
[storagelocations]: storagelocations.md
[reportdatasources]: reportdatasources.md
[reportqueries]: reportqueries.md
[reports]: reports.md
