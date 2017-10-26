# Chargeback CRD Model

Chargeback uses a few different CRDs for configuration. This document describes
the different CRDs, provides examples for each, and explains how they interact
with each other.

## `ReportPrometheusQuery`

The `ReportPrometheusQuery` object simply holds a Prometheus query and a name.

Example:

```
apiVersion: chargeback.coreos.com/v1alpha1
kind: ReportPrometheusQuery
metadata:
    name: "get-memory-by-pod"
spec:
    query: |
        (kube_pod_container_resource_requests_memory_bytes / on(node) group_left kube_node_status_capacity_memory_bytes) * on(node) group_left(provider_id) kube_node_info
```

## `ReportDataStore`

The `ReportDataStore` object lists `ReportPrometheusQuery`s, and lists a
location and for the results of these queries to be stored. When `chargeback`
runs, it lists all `ReportDataStore`s, runs all Prometheus queries listed by
each store, and saves the results into each location. This means that Prometheus
queries which are not pointed to by a `ReportDataStore` will not be run.

Example:
```
apiVersion: chargeback.coreos.com/v1alpha1
kind: ReportDataStore
metadata:
  name: "pod-memory-usage"
  labels:
    tectonic-chargeback: "true"
spec:
  promsum:
    queries:
    - "get-memory-by-pod"
  storage:
    local: {}
```

S3 storage is also supported. An example of using it would replace the `storage`
section with the following:

```
storage:
  s3:
    bucket: BUCKET_NAME
    prefix: PREFIX
```

## `ReportGenerationQuery`

Each `ReportGenerationQuery` object is a different type of report that
Chargeback can generate. The object holds a Presto query that is used to convert
usage data (and potentially AWS billing data) into a report, and the data store
whose data shall be used for the query. Additionally the `ReportGenerationQuery`
object defines the columns that will be present in the produced report.

Example:

```
apiVersion: chargeback.coreos.com/v1alpha1
kind: ReportGenerationQuery
metadata:
  name: "pod-memory-usage-by-node"
spec:
  reportDataStore: "pod-memory-usage"
  columns:
  - name: pod
    type: string
  - name: namespace
    type: string
  - name: node
    type: string
  - name: provider_id
    type: string
  - name: memory_usage
    type: double
  query: |
    WITH usage_period AS (
        SELECT kubeUsage.labels['pod'] as pod,
               kubeUsage.labels['namespace'] as namespace,
               kubeUsage.labels['node'] as node,
               kubeUsage.labels as labels,
               kubeUsage.amount as amount,
               split_part(split_part(element_at(kubeUsage.labels, 'provider_id'), ':///', 2), '/', 2) as provider_id,
               kubeUsage."timestamp" as timestamp
        FROM {{.TableName}} as kubeUsage
    ),
    computed_usage AS (
        SELECT pod,
               namespace,
               node,
               provider_id,
               "timestamp" as period_end,
               lag("timestamp") OVER (PARTITION BY pod, namespace, node ORDER BY "timestamp" ASC) as period_start,
               (amount + lag(amount) OVER (PARTITION BY pod, namespace, node ORDER BY "timestamp" ASC)) / 2 as usage
        FROM usage_period
    )
    SELECT
        pod,
        namespace,
        node,
        provider_id,
        sum(usage * date_diff('millisecond', period_start, period_end)) as memory_usage
    FROM computed_usage
    WHERE period_start >= timestamp '{{.StartPeriod | prestoTimestamp }}'
    AND period_end <= timestamp '{{ .EndPeriod | prestoTimestamp }}'
    GROUP BY pod, namespace, node, provider_id
```

## `Report`

The `Report` object is created by users to trigger reports being generated. The
status of a report, viewable through kubectl, will mark when a report is
finished or errors encountered while generating it.

The `Report` object holds a start and end time over which the report should be
generated, names a generation query to use, and provides a location for the
report to be written to.

Example:

```
apiVersion: chargeback.coreos.com/v1alpha1
kind: Report
metadata:
  name: pod-memory-usage-by-node
spec:
  reportingStart: '2017-01-01T00:00:00Z'
  reportingEnd: '2017-12-30T23:59:59Z'
  generationQuery: "pod-memory-usage-by-node"
  output:
    local: {}
```
