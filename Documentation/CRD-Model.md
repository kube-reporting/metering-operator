# Chargeback CRD Model

Chargeback uses a few different CRDs for con***REMOVED***guration. This document outlines
the different CRDs, examples for each, and how they interact with each
other.

## `ReportPrometheusQuery`

The `ReportPrometheusQuery` object simply holds a Prometheus query.

Example:

```
apiVersion: chargeback.coreos.com/prealpha
kind: ReportPrometheusQuery
metadata:
    name: "get-memory-by-pod"
spec:
    query: |
        (kube_pod_container_resource_requests_memory_bytes / on(node) group_left kube_node_status_capacity_memory_bytes) * on(node) group_left(provider_id) kube_node_info
```

## `ReportDataStore`

The `ReportDataStore` object lists `ReportPrometheusQuery`s, and lists a
location and format for the results of these queries to be stored. When
`promsum` runs, it lists all `ReportDataStore`s, runs all Prometheus queries
listed by each store, and saves the results into each location. This means that
Prometheus queries which are not pointed to by a `ReportDataStore` will not be
run.

Example:
```
apiVersion: chargeback.coreos.com/prealpha
kind: ReportDataStore
metadata:
    name: "default-datastore"
    labels:
        tectonic-chargeback: "true"
spec:
    storage:
        type: "s3"
        format: "json"
        bucket: "chargeback-datastore"
        pre***REMOVED***x: "promsum/memory_by_pod"
    queries:
    - "get-memory-by-pod"
```

Note: currently `s3` is the only supported storage type and `json` is the only
supported storage format.

## `ReportGenerationQuery`

Each `ReportGenerationQuery` object is a different type of report that
Chargeback can generate. The object holds a Presto query that is used to convert
usage data (and potentially AWS billing data) into a report. Additionally the
`ReportGenerationQuery` object de***REMOVED***nes the columns that will be present in the
produced report.

Example:

```
apiVersion: chargeback.coreos.com/prealpha
kind: ReportGenerationQuery
metadata:
    name: "pod-usage-by-node"
spec:
    reportDataStore: "default-datastore"
    columns:
        - name: pod
          type: string
        - name: namespace
          type: string
        - name: node
          type: string
        - name: memoryUsage
          type: double
    query: |
        WITH usage_period AS (
            SELECT kubeUsage.labels['pod'] as pod,
                   kubeUsage.labels['namespace'] as namespace,
                   kubeUsage.labels as labels,
                   kubeUsage.amount as amount,
                   split_part(split_part(kubeUsage.labels['provider_id'], ':///', 2), '/', 2) as node,
                   kubeUsage."timestamp" as timestamp
                   {{addAdditionalLabels .Labels}}
            FROM {{.TableName}} as kubeUsage
            WHERE kubeUsage."timestamp" >= timestamp '{{.StartPeriod}}'
            AND kubeUsage."timestamp" <= timestamp '{{.EndPeriod}}'
        ),
        computed_usage AS (
            SELECT pod,
                   namespace,
                   node,
                   "timestamp" as curr,
                   lag("timestamp") OVER (PARTITION BY pod, namespace, node ORDER BY "timestamp" ASC) as prev,
                   (amount + lag(amount) OVER (PARTITION BY pod, namespace, node ORDER BY "timestamp" ASC)) / 2 as usage
                   {{listAdditionalLabels .Labels}}
            FROM usage_period
        )
        SELECT
            pod,
            namespace,
            node,
            sum(usage * date_diff('millisecond', prev, curr)) as usage
           {{listAdditionalLabels .Labels}}
        FROM computed_usage
        GROUP BY pod, namespace, node {{listAdditionalLabels .Labels}}
```

## `Report`

The `Report` object is created by users to trigger reports being generated. The
status of a report, viewable through kubectl, will mark when a report is
***REMOVED***nished or errors encountered while generating it.

The `Report` object holds a start and end time over which the report should be
generated, names a generation query to use, provides a location for the report
to be written to, and lists any additional labels that the user wishes to show
up in the report from the object being reported on.

Example:

```
apiVersion: chargeback.coreos.com/prealpha
kind: Report
metadata:
    name: pods
spec:
    reportingStart: '2017-09-01T00:00:00Z'
    reportingEnd: '2017-09-30T23:59:59Z'
    generationQuery: "pod-usage-by-node"
    output:
        bucket: chargeback-datastore
        pre***REMOVED***x: results
    additionalLabels:
    - "provider_id"
```
