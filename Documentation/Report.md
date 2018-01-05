# Reports

The `Report` custom Kubernetes resource is used to manage the execution and status of reports. Chargeback produces reports derived from usage data sources which can be used in further analysis and filtering.

## Report object

A single `Report` resource corresponds to a specific run of a report. Once the object is created, Chargeback starts analyzing the data required to perform the report. A report cannot be updated after its creation and must run to completion.

## Example

Example reports ready to be created exist in `manifests/custom-resources/reports`.

The following example report will contain information on every pod's memory requests over the month of September:

```
apiVersion: chargeback.coreos.com/v1alpha1
kind: Report
metadata:
  name: pod-cpu-usage-by-node
spec:
  reportingStart: '2017-09-01T00:00:00Z'
  reportingEnd: '2017-09-30T23:59:59Z'
  generationQuery: "pod-cpu-usage-by-node"
  gracePeriod: "5m"
  runImmediately: true
```

### reportingStart

The timestamp of the beginning of the time period the report will cover. The format of this field is: `[Year]-[Month]-[Day]T[Hour]-[Minute]-[Second]Z`, where all components are numbers with leading zeroes where appropriate.

Timestamps should be [RFC3339][rfc3339] encoded. Times with local offsets will be converted to UTC.

### reportingEnd

The timestamp of the end of the time period the report will cover, with
the same format as `reportingStart`.

Timestamps should be [RFC3339][rfc3339] encoded. Times with local offsets will be converted to UTC.

### generationQuery

Names the generation query used to generate the report. The generation query controls the format of the report as well as the information contained within it.

### gracePeriod

Sets the period of time after `reportingEnd` that the report will be run. This value is `5m` by default.

By default, a report is not run until `reportingEnd` plus the `gracePeriod`
has been reached. The grace period is not used when aggregating over the
reporting period, or if `runImmediately` is true.

This field particularly useful with AWS Billing Reports,
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

The execution of a `Report` can be tracked using its status field. Any errors occurring during the preparation of a report will be recorded here.

A report can have the following states:
* `Started`: Chargeback has started executing the report. No modifications can be made at this point.
* `Finished`: The report successfully completed execution.
* `Error`: A failure occurred running the report. Details are provided in the `output` field.


[rfc3339]: https://tools.ietf.org/html/rfc3339#section-5.8
[storing-data]: Storing-Data-In-S3.md
