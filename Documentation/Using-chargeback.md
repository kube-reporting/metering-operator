# Using Chargeback

Once chargeback is running, reports can be created and their results can be
fetched. This document will outline how to do this.

*Note: if chargeback has recently been installed, it's recommended to wait at
least 15 minutes before generating reports so that chargeback has a chance to
reach a consistent state and fetch some data from Prometheus*

## What are reports

A report can be created for chargeback to run via kubectl. A report is a custom
resource in Kubernetes, and thus is typically written as a YAML ***REMOVED***le. The report
should be created in the same namespace as chargeback is installed in.

Example reports ready to be created exist in `manifests/custom-resources/reports`.

As an example, here's a report that will contain information on every pod's
memory requests over the month of September:

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
  output:
    local: {}
```

Going over all possible ***REMOVED***elds in a report:

### `generationQuery`

This ***REMOVED***eld names the generation query that should be used to generate this
report. The generation query controls the format of the report and what
information actually ends up in the report.

### `reportingStart`

This is a timestamp of the beginning of the time period the report should cover.

The format of this ***REMOVED***eld is: `[Year]-[Month]-[Day]T[Hour]-[Minute]-[Second]Z`,
where all ***REMOVED***elds are numbers with leading zeroes where appropriate.

### `reportingEnd`

This is a timestamp of the end of the time period the report should cover, with
the same format as `reportingStart`.

### `gracePeriod`

By default, a report is not run until the `reportingEnd` plus the `gracePeriod`
has been reached. The grace period is not used when aggregating over the
reporting period.  This is primarily useful when you have an AWS Billing Report
which may get it's latest information up to 24 hours after the billing period
has ended.  The grace period is a duration of time, which by default is `5m`.
Has no effect if `runImmediately` is true.

### `runImmediately`

If a report should be run immediately with all available data, regardless of if
the end of the reporting period has been reached (plus the grace period),
`runImmediately` can be set to `true`.

### `output`

The output section controls where the results of the report will be stored. The
value of this does not impact how report results are fetched. For more
information on this, please read the documentation on [storing data in S3](Storing-Data-In-S3.md).

## Creating a report

Once the report YAML is written, it can be created via kubectl:

```
$ kubectl -n $CHARGEBACK_NAMESPACE create -f manifests/custom-resources/reports/pod-cpu-usage-by-node.yaml
```

Existing reports can be viewed in Kubernetes with the following command:

```
$ kubectl -n $CHARGEBACK_NAMESPACE get reports
```

A report's status can be inspected by viewing the object with the `-o json`
flag:

```
$ kubectl -n $CHARGEBACK_NAMESPACE get report pod-cpu-usage -o json
```

## Viewing reports

Once a report's status has changed to `Finished`, the report is ready to be
downloaded. The chargeback pod exposes an HTTP API for this.

First, use kubectl to setup a proxy for accessing Kubernetes services:

```
$ kubectl proxy
```

The URL used to fetch a report changes based on the report's name and the format
that the report should be in. The URL scheme is

```
/api/v1/reports/get?name=[Report Name]&format=[Format]
```

Since we're using `kubectl proxy`, we need to access our URL via a pre***REMOVED***x that
points to our Kubernetes service (see the upstream documentation on
[Accessing Services Running on Clusters](https://kubernetes.io/docs/tasks/administer-cluster/access-cluster-services/#manually-constructing-apiserver-proxy-urls)
for more details). This assumes you've deployed chargeback in the `chargeback`
namespace.

```
http://127.0.0.1:8001/api/v1/namespaces/chargeback/services/chargeback/proxy/api/v1/reports/get?name=[Report Name]&format=[Format]
```

For example, the results of the `pod-cpu-usage-by-node` report can be fetched in
CSV, with the following command:

```
$ curl "http://127.0.0.1:8001/api/v1/namespaces/chargeback/services/chargeback/proxy/api/v1/reports/get?name=pod-cpu-usage-by-node&format=csv"
```

The `format` parameter can be either `csv` or `json`.
