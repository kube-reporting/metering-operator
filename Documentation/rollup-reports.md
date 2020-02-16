# Guide: Roll-Up Reports

Often, it makes sense to report on data collected in other reports, called roll-up reports.
Roll-up reports can combine datatypes (memory *and* CPU usage per namespace) and can combine smaller periods of time into larger ones — for instance, in an environment where a large amount of data is collected, splitting up processing over a month instead of waiting for the end of the month can be an effective way to spread out compute time and get a quicker final report.

In the following guide, we will create a daily report that aggregates hourly reports.

## 1. Create the sub-report

First, create the hourly report that will be aggregated:

```
apiVersion: metering.openshift.io/v1
kind: Report
metadata:
  name: namespace-cpu-usage-hourly
spec:
  query: "namespace-cpu-usage"
  reportingStart: '2019-10-09T00:00:00Z'
  schedule:
    period: "hourly"
```

## 2. Create the aggregation query

To aggregate the reports together, we need a query that will retrieve the data from the report we wish to aggregate.
The query below is a copy of the built-in `namespace-cpu-usage` query provided to demonstrate how aggregation can be done.
It contains a few a custom inputs: most importantly, `NamespaceCPUUsageReportName` is the input we can use to pass name of our sub-report in:

```
apiVersion: metering.openshift.io/v1
kind: ReportQuery
metadata:
  name: namespace-cpu-usage
  labels:
    operator-metering: "true"
spec:
  columns:
  - name: period_start
    type: timestamp
    unit: date
  - name: period_end
    type: timestamp
    unit: date
  - name: namespace
    type: varchar
    unit: kubernetes_namespace
  - name: pod_usage_cpu_core_seconds
    type: double
    unit: cpu_core_seconds
  inputs:
  - name: ReportingStart
    type: time
  - name: ReportingEnd
    type: time
  - name: NamespaceCPUUsageReportName
    type: Report
  - name: PodCpuUsageRawDataSourceName
    type: ReportDataSource
    default: pod-cpu-usage-raw
  query: |
    SELECT
      timestamp '{| default .Report.ReportingStart .Report.Inputs.ReportingStart| prestoTimestamp |}' AS period_start,
      timestamp '{| default .Report.ReportingEnd .Report.Inputs.ReportingEnd | prestoTimestamp |}' AS period_end,
    {|- if .Report.Inputs.NamespaceCPUUsageReportName |}
      namespace,
      sum(pod_usage_cpu_core_seconds) as pod_usage_cpu_core_seconds
    FROM {| .Report.Inputs.NamespaceCPUUsageReportName | reportTableName |}
    WHERE period_start  >= timestamp '{| default .Report.ReportingStart .Report.Inputs.ReportingStart | prestoTimestamp |}'
    AND period_end <= timestamp '{| default .Report.ReportingEnd .Report.Inputs.ReportingEnd | prestoTimestamp |}'
    GROUP BY namespace
    {|- else |}
      namespace,
      sum(pod_usage_cpu_core_seconds) as pod_usage_cpu_core_seconds
    FROM {| dataSourceTableName .Report.Inputs.PodCpuUsageRawDataSourceName |}
    WHERE "timestamp" >= timestamp '{| default .Report.ReportingStart .Report.Inputs.ReportingStart | prestoTimestamp |}'
    AND "timestamp" < timestamp '{| default .Report.ReportingEnd .Report.Inputs.ReportingEnd | prestoTimestamp |}'
    AND dt >= '{| default .Report.ReportingStart .Report.Inputs.ReportingStart | prometheusMetricPartitionFormat |}'
    AND dt <= '{| default .Report.ReportingEnd .Report.Inputs.ReportingEnd | prometheusMetricPartitionFormat |}'
    GROUP BY namespace
    {|- end |}
```

Note the use of the macro `reportTableName`, which will automatically get the proper table name from the given report name.

## 3. Create the aggregator report

We now have a sub-report and a query that can read data from other reports.
We can create a `Report` that uses that custom report query with the sub-report:

```
apiVersion: metering.openshift.io/v1
kind: Report
metadata:
  name: namespace-cpu-usage-daily
spec:
  query: "namespace-cpu-usage"
  inputs:
  - name: "NamespaceCPUUsageReportName"
    value: "namespace-cpu-usage-hourly"
  reportingStart: '2019-10-09T00:00:00Z'
  schedule:
    period: "daily"
```
