# Reports and Scheduled Reports

The `Report` and `ScheduledReport` custom resources are used to manage the execution and status of reports. Metering produces reports derived from usage data sources which can be used in further analysis and ***REMOVED***ltering.

## Scheduled Report object

A single `ScheduledReport` resource represents a report which is updated with new information according to a schedule. Scheduled reports are always running, and will track what time periods it has collected data for, ensuring that if Metering is shutdown or unavailable for an extended period of time, it will back***REMOVED***ll the data starting where it left off.

## Example Scheduled Report

The following example scheduled report will contain information on every Pod's CPU requests, and will run every hour, adding the last hours worth of data each time it runs.

```
apiVersion: metering.openshift.io/v1alpha1
kind: ScheduledReport
metadata:
  name: pod-cpu-request-hourly
spec:
  generationQuery: "pod-cpu-request"
  gracePeriod: "5m"
  schedule:
    period: "hourly"
    hourly:
      minute: 0
      second: 0
```

## generationQuery

Names the `ReportGenerationQuery` used to generate the report. The generation query controls the schema of the report as well as the information contained within it.

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

## schedule

The schedule block de***REMOVED***nes when the report runs. The main ***REMOVED***elds in the `schedule` section are `period`, and then depending on the value of `period`, the ***REMOVED***elds `hourly`, `daily`, `weekly` and `monthly` allow you to ***REMOVED***ne-tune when the report runs.

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

Generally, the `hour`, `minute`, `second` ***REMOVED***elds control when in the day the report runs, and `dayOfWeek`/`dayOfMonth` control what day of the week, or day of month the report runs on, if it's a weekly or monthly report period.

For each of these ***REMOVED***elds, there is a range of valid values:

- `hour` is an integer value between 0-23.
- `minute` is an integer value between 0-59.
- `second` is an integer value between 0-59.
- `dayOfWeek` is a string value that expects the day of the week (spelled out).
- `dayOfMonth` is an integer value between 1-31.


### reportingStart

To support running a ScheduledReport against existing data, you can set the `spec.reportingStart` ***REMOVED***eld to a RFC3339 timestamp to tell the ScheduledReport to run according to it's `schedule` starting from `reportingStart` rather than the current time.
One important thing to understand is that this will result in the reporting-operator running many queries in succession for each interval in the schedule that's between the `reportingStart` time and the current time. This could be thousands of queries if the period is less than daily and the `reportingStart` is more than a few months back.

As an example of how to use this ***REMOVED***eld, if you had data already collected dating back to January 1st, 2018 which you wanted to be included in your scheduledReport, you could create a report with the following values:

```
apiVersion: metering.openshift.io/v1alpha1
kind: ScheduledReport
metadata:
  name: pod-cpu-request-hourly
spec:
  generationQuery: "pod-cpu-request"
  gracePeriod: "5m"
  schedule:
    period: "hourly"
  reportingStart: "2018-01-01T00:00:00Z"
```

### Scheduled Report Status

The execution of a scheduled report can be tracked using its status ***REMOVED***eld. Any errors occurring during the preparation of a report will be recorded here.

The `status` ***REMOVED***eld of a `ScheduledReport` currently has two ***REMOVED***elds:

- `conditions`: Conditions is an list of conditions, each have a `Type`, `Reason`, and `Message` ***REMOVED***eld. Possible values of a condition's `Type` ***REMOVED***eld are `Running` and `Failure`, indicating the current state of the scheduled report. The `Reason` indicates why it's the `Condition` is in it's current state, with and the `Message` provides a detailed information on the `Reason`.
- `lastReportTime`: Indicates the time Metering has collected data up to.

## Report object

A single `Report` resource represents a report which runs the provided query for the speci***REMOVED***ed time range. Once the object is created, Metering starts analyzing the data required to perform the report. A report cannot be updated after its creation and must run to completion.

## Example Report

The following example report will contain information on every Pod's CPU requests over the month of September:

```
apiVersion: metering.openshift.io/v1alpha1
kind: Report
metadata:
  name: pod-cpu-request
spec:
  reportingStart: '2018-09-01T00:00:00Z'
  reportingEnd: '2018-09-30T23:59:59Z'
  generationQuery: "pod-cpu-request"
  runImmediately: true
```

### reportingStart

The timestamp of the beginning of the time period the report will cover. The format of this ***REMOVED***eld is: `[Year]-[Month]-[Day]T[Hour]-[Minute]-[Second]Z`, where all components are numbers with leading zeroes where appropriate.

Timestamps should be [RFC3339][rfc3339] encoded. Times with local offsets will be converted to UTC.

### reportingEnd

The timestamp of the end of the time period the report will cover, with
the same format as `reportingStart`.

Timestamps should be [RFC3339][rfc3339] encoded. Times with local offsets will be converted to UTC.

### gracePeriod

Sets the period of time after `reportingEnd` that the report will be run. This value is `5m` by default.

By default, a report is not run until `reportingEnd` plus the `gracePeriod`
has been reached. The grace period is not used when aggregating over the
reporting period, or if `runImmediately` is true.

This ***REMOVED***eld is particularly useful with AWS Billing Reports,
which may get their latest information up to 24 hours after the billing period
has ended.

### runImmediately

Set `runImmediately` to `true` to run the report immediately with all available data, regardless of the `gracePeriod` or `reportingEnd` flag settings.

### generationQuery

Names the `ReportGenerationQuery` used to generate the report. The generation query controls the format of the report as well as the information contained within it.

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

ReportGenerationQueries with the `-raw` suf***REMOVED***x are used by other ReportGenerationQueries to build more complex queries, and should not be should not be used directly for reports.

`namespace-` pre***REMOVED***xed queries aggregate Pod CPU/memory requests by namespace, providing a list of namespaces and their overall usage based on resource requests.

`pod-` pre***REMOVED***xed queries are similar to 'namespace-' pre***REMOVED***xed, but aggregate information by Pod, rather than namespace. These queries include the Pod's namespace and node.

`node-` pre***REMOVED***xed queries return information about each node's total available resources.

`aws-` pre***REMOVED***xed queries are speci***REMOVED***c to AWS. Queries suf***REMOVED***xed with `-aws` return the same data as queries of the same name without the suf***REMOVED***x, and correlate usage with the EC2 billing data.

The `aws-ec2-billing-data` report is used by other queries, and should not be used as a standalone report. The `aws-ec2-cluster-cost` report provides a total cost based on the nodes included in the cluster, and the sum of their costs for the time period being reported on.

For a complete list of ***REMOVED***elds each report query produces, use `kubectl` to get the object as JSON, and check the `columns` ***REMOVED***eld:

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

### Report Status

The execution of a report can be tracked using its status ***REMOVED***eld. Any errors occurring during the preparation of a report will be recorded here.

A report can have the following states:
* `Started`: Metering has started executing the report. No modi***REMOVED***cations can be made at this point.
* `Finished`: The report successfully completed execution.
* `Error`: A failure occurred running the report. Details are provided in the `output` ***REMOVED***eld.


[rfc3339]: https://tools.ietf.org/html/rfc3339#section-5.8
