# Report Prometheus Queries

A lot of the flexibility in Operator Metering is in the ability to customize what data is collected by writing Custom Prometheus Queries in a custom resource called a `ReportPrometheusQuery`.

These `ReportPrometheusQuery` resources allow you to control what Prometheus queries the operator executes periodically, and makes the data collected by the query available as another database table that can be reported on by [ReportGenerationQueries](reportgenerationqueries.md).

## Fields

All fields that can be controlled on an individual `ReportPrometheusQuery` level are contained in the `spec` section of the resource.

- `query`: A string containing the Prometheus Query to be executed by the operator. For details on writing Prometheus queries read the official [Querying Prometheus documentation][querying-prometheus].

## Example ReportPrometheusQuery

The example resource below is installed by default when installing Operator Metering.

```yaml
apiVersion: metering.openshift.io/v1alpha1
kind: ReportPrometheusQuery
metadata:
  name: "pod-request-memory-bytes"
  labels:
    operator-metering: "true"
spec:
  query: |
    sum(kube_pod_container_resource_requests_memory_bytes) by (pod, namespace, node)
```

The important part is the `spec.query` field which is the actual Prometheus query executed by the operator:

```
sum(kube_pod_container_resource_requests_memory_bytes) by (pod, namespace, node)
```

[querying-prometheus]: https://prometheus.io/docs/prometheus/latest/querying/basics/
