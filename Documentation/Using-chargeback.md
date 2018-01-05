# Using Chargeback

Use Chargeback to create reports and fetch their results.

*Note: Wait at least 15 minutes after installing Chargeback before generating reports to enable Chargeback to reach a consistent state and fetch some data from Prometheus*

## Running Chargeback reports

A report can be created for Chargeback to run using `kubectl`. The report
should be created in the same namespace as Chargeback is installed.

Example reports ready to be created exist in `manifests/custom-resources/reports`.

For more information on Report YAML configuration options see [Reports][report-md].

## Creating a report

Once the report YAML is written, use `kubectl` to create the report:

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
downloaded. The Chargeback pod exposes an HTTP API for this.

First, use `kubectl` to set up a proxy for accessing Kubernetes services:

```
$ kubectl proxy
```

The URL used to fetch a report changes based on the report's name and format.
The `format` parameter may be either `csv` or `json`. The URL scheme is:

```
/api/v1/reports/get?name=[Report Name]&format=[Format]
```

Using `kubectl proxy` requires that the URL be accessed through a prefix that
points to the Kubernetes service. (See the upstream documentation on
[Manually constructing apiserver proxy URLs][accessing-services] for more details.) The following example assumes Chargeback is deployed in the `chargeback` namespace.

```
http://127.0.0.1:8001/api/v1/namespaces/chargeback/services/chargeback/proxy/api/v1/reports/get?name=[Report Name]&format=[Format]
```

For example, the results of the `pod-cpu-usage-by-node` report can be fetched in
CSV, with the following command:

```
$ curl "http://127.0.0.1:8001/api/v1/namespaces/chargeback/services/chargeback/proxy/api/v1/reports/get?name=pod-cpu-usage-by-node&format=csv"
```


[accessing-services]: https://kubernetes.io/docs/tasks/administer-cluster/access-cluster-services/#manually-constructing-apiserver-proxy-urls
[reports-md]: report.md
