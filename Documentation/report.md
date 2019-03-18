# Reports

The `Report` custom resource is used to manage the execution and status of reports.
Metering produces reports derived from usage data sources which can be used in further analysis and filtering.

## Report object

A single `Report` resource represents a report which is updated with new information according to a schedule.
Reports with a `spec.schedule` field set are always running, and will track what time periods it has collected data for, ensuring that if Metering is shutdown or unavailable for an extended period of time, it will backfill the data starting where it left off.
If the schedule is unset, then the Report will run once for the time specified by the reportingStart and reportingEnd.
By default, reports will wait for ReportDataSources to have the progressed in their import process to cover the report period being processed.
If the report has a schedule, it will wait until the period currently being processed has been covered by the import process.

## Example Report with a Schedule

The following example Report will contain information on every Pod's CPU requests, and will run every hour, adding the last hours worth of data each time it runs.

```
apiVersion: metering.openshift.io/v1alpha1
kind: Report
metadata:
  name: pod-cpu-request-hourly
spec:
  generationQuery: "pod-cpu-request"
  reportingStart: "2018-07-01T00:00:00Z"
  schedule:
    period: "hourly"
    hourly:
      minute: 0
      second: 0
```

## Example Report without a Schedule (Run-Once)

The following example report will contain information on every Pod's CPU requests for all of July.
After completion it does not run again.

```
apiVersion: metering.openshift.io/v1alpha1
kind: Report
metadata:
  name: pod-cpu-request-hourly
spec:
  generationQuery: "pod-cpu-request"
  reportingStart: "2018-07-01T00:00:00Z"
  reportingEnd: "2018-07-31T00:00:00Z"
```

### generationQuery

Names the `ReportGenerationQuery` used to generate the report.
The generation query controls the schema of the report as well how the results are processed.

*`generationQuery` is a required field.*

Use `kubectl` to obtain a list of available `ReportGenerationQuery` objects:

 ```
 kubectl -n $METERING_NAMESPACE get reportgenerationqueries
 NAME                                            AGE
 aws-ec2-billing-data                            11m
 aws-ec2-cluster-cost                            11m
 namespace-cpu-request                           11m
 namespace-memory-request                        11m
 node-cpu-allocatable                            11m
 node-cpu-capacity                               11m
 node-cpu-utilization                            11m
 node-memory-allocatable                         11m
 node-memory-capacity                            11m
 node-memory-utilization                         11m
 pod-cpu-request                                 11m
 pod-cpu-request-aws                             11m
 pod-cpu-request-raw                             11m
 pod-cpu-request-vs-node-cpu-allocatable         11m
 pod-memory-request                              11m
 pod-memory-request-aws                          11m
 pod-memory-request-raw                          11m
 pod-memory-request-vs-node-memory-allocatable   11m
```

ReportGenerationQueries with the `-raw` suffix are used by other ReportGenerationQueries to build more complex queries, and should not be used directly for reports.

`namespace-` prefixed queries aggregate Pod CPU/memory requests by namespace, providing a list of namespaces and their overall usage based on resource requests.

`pod-` prefixed queries are similar to 'namespace-' prefixed, but aggregate information by Pod, rather than namespace. These queries include the Pod's namespace and node.

`node-` prefixed queries return information about each node's total available resources.

`aws-` prefixed queries are specific to AWS. Queries suffixed with `-aws` return the same data as queries of the same name without the suffix, and correlate usage with the EC2 billing data.

The `aws-ec2-billing-data` report is used by other queries, and should not be used as a standalone report. The `aws-ec2-cluster-cost` report provides a total cost based on the nodes included in the cluster, and the sum of their costs for the time period being reported on.

For a complete list of fields each report query produces, use `kubectl` to get the object as JSON, and check the `columns` field:

```
kubectl -n $METERING_NAMESPACE get reportgenerationqueries namespace-memory-request -o json

{
    "apiVersion": "metering.openshift.io/v1alpha1",
    "kind": "ReportGenerationQuery",
    "metadata": {
        "name": "namespace-memory-request",
        "namespace": "metering"
    },
    "spec": {
        "columns": [
            {
                "name": "namespace",
                "type": "string"
            },
            {
                "name": "data_start",
                "type": "timestamp"
            },
            {
                "name": "data_end",
                "type": "timestamp"
            },
            {
                "name": "pod_request_memory_byte_seconds",
                "type": "double"
            }
        ]
    }
}
```

## schedule

The schedule block defines when the report runs. The main fields in the `schedule` section are `period`, and then depending on the value of `period`, the fields `hourly`, `daily`, `weekly` and `monthly` allow you to fine-tune when the report runs.

For example, if `period` is set to `weekly`, you can add a `weekly` key to the `schedule` block. The following example will run once a week on Wednesday, at 1 PM.

```
...
  schedule:
    period: "weekly"
    weekly:
      dayOfWeek: "wednesday"
      hour: 13
```

### period

Valid values of `period` are listed below, and the options available to set for a given period are also listed.

- `hourly`
  - `minute`
  - `second`
- `daily`
  - `hour`
  - `minute`
  - `second`
- `weekly`
  - `dayOfWeek`
  - `hour`
  - `minute`
  - `second`
- `monthly`
  - `dayOfMonth`
  - `hour`
  - `minute`
  - `second`
- `cron`
  - `expression`

Generally, the `hour`, `minute`, `second` fields control when in the day the report runs, and `dayOfWeek`/`dayOfMonth` control what day of the week, or day of month the report runs on, if it's a weekly or monthly report period.

For each of these fields, there is a range of valid values:

- `hour` is an integer value between 0-23.
- `minute` is an integer value between 0-59.
- `second` is an integer value between 0-59.
- `dayOfWeek` is a string value that expects the day of the week (spelled out).
- `dayOfMonth` is an integer value between 1-31.

For cron periods, normal cron expressions are valid:

- `expression: "*/5 * * * *"`

### reportingStart

To support running a Report against existing data, you can set the `spec.reportingStart` field to a RFC3339 timestamp to tell the Report to run according to its `schedule` starting from `reportingStart` rather than the current time.
One important thing to understand is that this will result in the reporting-operator running many queries in succession for each interval in the schedule that's between the `reportingStart` time and the current time.
This could be thousands of queries if the period is less than daily and the `reportingStart` is more than a few months back.
If `reportingStart` is left unset, the Report will run at the next full reportingPeriod after the time the report is created.

As an example of how to use this field, if you had data already collected dating back to January 1st, 2018 which you wanted to be included in your Report, you could create a report with the following values:

```
apiVersion: metering.openshift.io/v1alpha1
kind: Report
metadata:
  name: pod-cpu-request-hourly
spec:
  generationQuery: "pod-cpu-request"
  schedule:
    period: "hourly"
  reportingStart: "2018-01-01T00:00:00Z"
```


### reportingEnd

To configure a Report to only run until a specified time, you can set the `spec.reportingEnd` field to an RFC3339 timestamp.
The value of this field will cause the Report to stop running on its schedule after it has finished generating reporting data for the period covered from its start time until `reportingEnd`.
Because a schedule will most likely not align with reportingEnd, the last period in the schedule will be shortened to end at the specified reportingEnd time.
If left unset, then the Report will run forever, or until a `reportingEnd` is set on the Report.

For example, if you wanted to create a report that runs once a week for the month of July:

```
apiVersion: metering.openshift.io/v1alpha1
kind: Report
metadata:
  name: pod-cpu-request-hourly
spec:
  generationQuery: "pod-cpu-request"
  schedule:
    period: "weekly"
  reportingStart: "2018-07-01T00:00:00Z"
  reportingEnd: "2018-07-31T00:00:00Z"
```

### runImmediately

Set `runImmediately` to `true` to run the report immediately with all available data, regardless of the `reportingStart` or `reportingEnd` values, and without checking if there is any data for the report period.
For reports with a schedule set, it will not wait for each period's reportingEnd to elapse before processing and all reportPeriods between `reportingStart` and `reportingEnd`.

### Inputs

The `inputs` field of a Report `spec` can be used to pass custom values into a `ReportGenerationQuery`.

It is a list of name-value pairs:

```
spec:
  inputs:
  - name: AggregatedReportName
    value: namespace-cpu-usage-hourly
```

For an example of how this can be used, see it in action [in a roll-up report](rollup-reports.md#3-create-the-aggregator-report).

## Roll-up Reports

Report data is stored in the database much like metrics themselves, and can thus be used in aggregated or roll-up reports. A simple use case for a roll-up report is to spread the time required to produce a report over a longer period of time: instead of requiring a monthly report to query and add all data over an entire month, the task can be split into daily reports that each run over a thirtieth of the data.

A custom roll-up report requires a custom generation query. The ReportGenerationQuery processor provides a macro that can get the necessary table name [from a report name](rollup-reports.md#2-create-the-aggregation-query):

```
# namespace-cpu-usage-aggregated-query.yaml
inputs:
- name: AggregatedReportName
  required: true
...
 WHERE {| .Report.Inputs.AggregatedReportName | reportTableName |}.period_start >= timestamp '{| default .Report.ReportingStart .Report.Inputs.ReportingStart | prestoTimestamp |}'
```

```
# aggregated-report.yaml
spec:
  generationQuery: "namespace-cpu-usage-aggregated"
  inputs:
  - name: "AggregatedReportName"
    value: "namespace-cpu-usage-hourly"
```

For more information on setting up a roll-up report, see the [roll-up report guide](rollup-reports.md).

### Scheduled Report Status

The execution of a scheduled report can be tracked using its status field. Any errors occurring during the preparation of a report will be recorded here.

The `status` field of a `Report` currently has two fields:

- `conditions`: Conditions is a list of conditions, each of which have a `type`, `status`, `reason`, and `message` field. Possible values of a condition's `type` field are `Running` and `Failure`, indicating the current state of the scheduled report. The `reason` indicates why its `condition` is in its current state with the `status` being either `true`, `false` or `unknown`. The `message` provides a human readable indicating why the condition is in the current state. For detailed information on the `reason` values see [`pkg/apis/metering/v1alpha1/util/report_util.go`](https://github.com/operator-framework/operator-metering/blob/master/pkg/apis/metering/v1alpha1/util/report_util.go#L10).
- `lastReportTime`: Indicates the time Metering has collected data up to.

[rfc3339]: https://tools.ietf.org/html/rfc3339#section-5.8

