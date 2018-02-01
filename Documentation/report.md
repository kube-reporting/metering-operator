<br>
<div class="alert alert-info" role="alert">
    <i class="fa fa-exclamation-triangle"></i><b> Note:</b> This documentation is for a pre-alpha feature. To register for the Chargeback Alpha program, email <a href="mailto:tectonic-alpha-feedback@coreos.com">tectonic-alpha-feedback@coreos.com</a>.
</div>

# Reports

The `Report` custom Kubernetes resource is used to manage the execution and status of reports. Chargeback produces reports derived from usage data sources which can be used in further analysis and ***REMOVED***ltering.

## Report object

A single `Report` resource corresponds to a speci***REMOVED***c run of a report. Once the object is created, Chargeback starts analyzing the data required to perform the report. A report cannot be updated after its creation and must run to completion.

## Example

The following example report will contain information on every pod's cpu requests over the month of September:

```
apiVersion: chargeback.coreos.com/v1alpha1
kind: Report
metadata:
  name: pod-cpu-usage-by-node
spec:
  reportingStart: '2018-09-01T00:00:00Z'
  reportingEnd: '2018-09-30T23:59:59Z'
  generationQuery: "pod-cpu-usage-by-node"
  runImmediately: true
```

### reportingStart

The timestamp of the beginning of the time period the report will cover. The format of this ***REMOVED***eld is: `[Year]-[Month]-[Day]T[Hour]-[Minute]-[Second]Z`, where all components are numbers with leading zeroes where appropriate.

Timestamps should be [RFC3339][rfc3339] encoded. Times with local offsets will be converted to UTC.

### reportingEnd

The timestamp of the end of the time period the report will cover, with
the same format as `reportingStart`.

Timestamps should be [RFC3339][rfc3339] encoded. Times with local offsets will be converted to UTC.

### generationQuery

Names the `ReportGenerationQuery` used to generate the report. The generation query controls the format of the report as well as the information contained within it.

You can obtain a list of available `ReportGenerationQuery` objects by using `kubectl get reportgenerationqueries -n $CHARGEBACK_NAMESPACE`.

Here is a list of the ReportGenerationQueries available:

- aws-ec2-billing-data
- aws-ec2-cluster-cost
- namespace-cpu-request
- namespace-memory-request
- node-cpu-allocatable
- node-cpu-capacity
- node-cpu-utilization
- node-memory-allocatable
- node-memory-capacity
- node-memory-utilization
- pod-cpu-request
- pod-cpu-request-aws
- pod-cpu-request-raw
- pod-cpu-request-vs-node-cpu-allocatable
- pod-memory-request
- pod-memory-request-aws
- pod-memory-request-raw
- pod-memory-request-vs-node-memory-allocatable

The ReportGenerationQueries with the `-raw` suf***REMOVED***x shouldn't generally be used directly for reports, they're currently used by other ReportGenerationQueries as a building block to more complex queries.

The `namespace-` pre***REMOVED***xed queries are aggregating pod cpu/memory requests by namespace, giving you a list of namespaces and their overall usage based on resource requests.
The queries with a `pod-` pre***REMOVED***x are basically the same, but have it broken down by pod, which includes the pods namespace, and node.
The report queries pre***REMOVED***xed `node-` contain information about each node's total available resources.
Report queries pre***REMOVED***xed with `aws-` are queries speci***REMOVED***cally about AWS information, while queries suf***REMOVED***xed with `-aws` are effectively the same as queries of the same name without the suf***REMOVED***x, but correlate usage with the EC2 billing data.
The `aws-ec2-billing-data` report shouldn't generally be used directly, it is used by other queries. The `aws-ec2-cluster-cost` report provides a total cost based entirely on what nodes are in the cluster, and what the sum of their costs are for the time period being reported on.

For a complete list of ***REMOVED***elds each report query produces, use the kubectl to get the object as JSON, and check the `columns` ***REMOVED***eld:

```
kubectl -n $CHARGEBACK_NAMESPACE get reportgenerationqueries namespace-memory-request -o json

{
    "apiVersion": "chargeback.coreos.com/v1alpha1",
    "kind": "ReportGenerationQuery",
    "metadata": {
        "name": "namespace-memory-request",
        "namespace": "chargeback"
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

### gracePeriod

Sets the period of time after `reportingEnd` that the report will be run. This value is `5m` by default.

By default, a report is not run until `reportingEnd` plus the `gracePeriod`
has been reached. The grace period is not used when aggregating over the
reporting period, or if `runImmediately` is true.

This ***REMOVED***eld particularly useful with AWS Billing Reports,
which may get their latest information up to 24 hours after the billing period
has ended.

### runImmediately

Set `runImmediately` to true to run the report immediately with all available data, regardless of the `gracePeriod` or `reportingEnd` flag settings.

## Execution

Reports take a variable amount of time to complete and may run for very long periods.

The amount of time required is determined by:
* report type
* amount of data being analyzed
* system performance (memory, CPU)
* network performance

## Status

The execution of a `Report` can be tracked using its status ***REMOVED***eld. Any errors occurring during the preparation of a report will be recorded here.

A report can have the following states:
* `Started`: Chargeback has started executing the report. No modi***REMOVED***cations can be made at this point.
* `Finished`: The report successfully completed execution.
* `Error`: A failure occurred running the report. Details are provided in the `output` ***REMOVED***eld.


[rfc3339]: https://tools.ietf.org/html/rfc3339#section-5.8
