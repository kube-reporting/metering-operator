# Reporting V2 API

There are two endpoints for the V2 versions of the endpoint:

- `/api/v2/reports/{namespace}/{name}/full`
- `/api/v2/reports/{namespace}/{name}/table`

`{name}` is the name if the report that you are looking to run. Output format is specified as a query string at the end.

# Sample URLs

Replace `$REPORT_NAME` with the name of your report.
Replace `$REPORT_NAMESPACE` with the namespace of your report.
Replace `$REPORT_FORMAT` with json, csv or tabular.

## V2 Reports Full Endpoint URL

```
/api/v2/reports/$REPORT_NAMESPACE/$REPORT_NAME/full?format=$REPORT_FORMAT
```

## V2 Reports Table Endpoint URL

```
/api/v2/reports/$REPORT_NAMESPACE/$REPORT_NAME/table?format=REPORT_FORMAT
```

### V2 Reports Full

The `/api/v2/reports/{namespace}/{name}/full` endpoint returns reports in either CSV, JSON, or tabular format, similar to /api/v1/reports/get. The difference is in the structure of the JSON results. The JSON results from this endpoint contain more metadata about each field including the unit, and whether or not the field should be shown the in a table (used in the table endpoint below).

This URL `/api/v2/reports/openshift-metering/namespace-cpu-request/full?format=json` returns

```
 {"results":[{"values":[{"name":"data_start","value":"2018-08-13T20:35:00Z","tableHidden":false,"unit":"date"},{"name":"data_end","value":"2018-08-13T23:58:00Z","tableHidden":false,"unit":"date"},{"name":"pod_request_cpu_core_seconds","value":2412,"tableHidden":false,"unit":"cpu_core_seconds"},{"name":"period_start","value":"2018-01-01T00:00:00Z","tableHidden":false,"unit":"date"},{"name":"period_end","value":"2018-12-30T23:59:59Z","tableHidden":false,"unit":"date"},{"name":"namespace","value":"default","tableHidden":false,"unit":"kubernetes_namespace"}]},
 ```

### V2 Reports Table

 The `/api/v2/reports/{namespace}/{name}/table` endpoint returns reports in either CSV, JSON, or tabular format.  tableHidden is a boolean and controls if a column should be shown when displayed in a table. If it's true, then the /api/v2/reports/{namespace}/{name}/table endpoint will omit this column and its values from the response (in all formats).

 This URL  `/api/v2/reports/openshift-metering/namespace-cpu-request/table?format=json` returns

```
 {"results":[{"values":[{"name":"period_start","value":"2018-01-01T00:00:00Z","tableHidden":false,"unit":"date"},{"name":"period_end","value":"2018-12-30T23:59:59Z","tableHidden":false,"unit":"date"},{"name":"namespace","value":"default","tableHidden":false,"unit":"kubernetes_namespace"},{"name":"data_start","value":"2018-08-13T20:35:00Z","tableHidden":false,"unit":"date"},{"name":"data_end","value":"2018-08-13T23:58:00Z","tableHidden":false,"unit":"date"},{"name":"pod_request_cpu_core_seconds","value":2412,"tableHidden":false,"unit":"cpu_core_seconds"}]},
 ```
