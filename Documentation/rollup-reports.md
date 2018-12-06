# Guide: Roll-Up Reports

Often, it makes sense to report on data collected in other reports, called roll-up reports. Roll-up reports can combine datatypes (memory *and* CPU usage per namespace) and can combine smaller periods of time into larger ones — for instance, in an environment where a large amount of data is collected, splitting up processing over a month instead of waiting for the end of the month can be an effective way to spread out compute time and get a quicker ***REMOVED***nal report.

Currently, there are no built-in roll-up reports. Therefore, a custom roll-up report requires a custom generation query. In the following guide, we will create a daily report that aggregates hourly reports:

## 1. Create the sub-report

```
apiVersion: metering.openshift.io/v1alpha1
kind: ScheduledReport
metadata:
  name: namespace-cpu-usage-hourly
spec:
  generationQuery: "namespace-cpu-usage"
  reportingStart: '2018-10-09T00:00:00Z'
  schedule:
    period: "hourly"
```

## 2. Create the aggregation query

To aggregate the reports together, we need a query that will retrieve the data from the report we wish to aggregate. This report is essentially a duplicate of the original `namespace-cpu-usage` query with the `data_end` and `data_start` timestamps stripped out, and a custom input `AggregatedReportName` that we use to pass the name of our sub-report in:

```
apiVersion: metering.openshift.io/v1alpha1
kind: ReportGenerationQuery
metadata:
  name: namespace-cpu-usage-aggregated
spec:
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
  - name: pod_usage_cpu_core_seconds
    type: double
    unit: cpu_core_seconds
  inputs:
  - name: ReportingStart
  - name: ReportingEnd
  - name: AggregatedReportName
    required: true
  query: |
    SELECT
      timestamp '{| default .Report.ReportingStart .Report.Inputs.ReportingStart| prestoTimestamp |}' AS period_start,
      timestamp '{| default .Report.ReportingEnd .Report.Inputs.ReportingEnd | prestoTimestamp |}' AS period_end,
      namespace,
      sum(pod_usage_cpu_core_seconds) as pod_usage_cpu_core_seconds
    FROM {| .Report.Inputs.AggregatedReportName | reportTableName |}
    WHERE {| .Report.Inputs.AggregatedReportName | reportTableName |}.period_start >= timestamp '{| default .Report.ReportingStart .Report.Inputs.ReportingStart | prestoTimestamp |}'
    AND {| .Report.Inputs.AggregatedReportName | reportTableName |}.period_end < timestamp '{| default .Report.ReportingEnd .Report.Inputs.ReportingEnd | prestoTimestamp |}'
    GROUP BY namespace
    ORDER BY pod_usage_cpu_core_seconds DESC
```

Note the use of the macro `reportTableName`, which will automatically get the proper table name from the given scheduled report name.

## 3. Create the aggregator report

We now have a sub-report and a query that can read data from other reports. We can create a `ScheduledReport` that uses that custom generation query with the sub-report:

```
apiVersion: metering.openshift.io/v1alpha1
kind: ScheduledReport
metadata:
  name: namespace-cpu-usage-daily-aggregate
spec:
  generationQuery: "namespace-cpu-usage-aggregated"
  inputs:
  - name: "AggregatedReportName"
    value: "namespace-cpu-usage-hourly"
  reportingStart: '2018-10-09T00:00:00Z'
  gracePeriod: "10m" # wait for sub-query to ***REMOVED***nish
  schedule:
    period: "daily"
```
