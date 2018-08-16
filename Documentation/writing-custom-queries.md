# Writing Custom Report Prometheus Queries and Report Generation Queries

One of the main goals of Operator Metering is to be flexible, and extendible.
The way this has been done is to use Kubernetes Custom Resources as a way of letting users add to, or expand upon the built-in reports and metrics that the operator already has.

The primary Custom Resources that allow this are the [ReportPrometheusQuery][reportprometheusqueries], [ReportDataSource][reportdatasources], and the [ReportGenerationQuery][reportgenerationqueries].
It's highly recommended you read the documentation on each of these resources before continuing.

A ReportDataSource combined with a ReportPrometheusQueries cause Operator Metering to collect additional Prometheus Metrics by allowing users to write custom Prometheus queries and store the metrics collected for analysis with a ReportGenerationQuery.

This guide is going to be structured so that you begin by collecting new Prometheus Metrics, and then end by writing custom SQL queries that analyze these metrics.

## Setup

This guide assumes you've already [installed Metering][install-metering] and have Prometheus in your cluster.

> This guide assumes you've set the `METERING_NAMESPACE` environment variable to the namespace your Operator Metering installation is running. All resources must be created in that namespace for the operator to view and access them.

Many of the examples below, will have you run a Prometheus query outside of the Metering operator.
The easiest way to do this is open a port-forward to your Prometheus pod in your cluster.
Below is a command that will open up a port-forward, exposing Prometheus on `127.0.0.1:9090`.

First, set the `PROMETHEUS_NAMESPACE` environment variable to the Kubernetes namespace where your Prometheus pod is running, and the run the following:

```
export PROMETHEUS_NAMESPACE=monitoring
kubectl -n $PROMETHEUS_NAMESPACE get pods -l app=prometheus -o name | cut -d/ -f2 | xargs -I{} kubectl -n $PROMETHEUS_NAMESPACE port-forward {} 9090:9090
```

Next, open up your web browser to [http://127.0.0.1:9090](http://127.0.0.1:9090).
If you get a 404, it's possible your Prometheus is behind an ingress controller, in which case, the port-forward will not work, and you will need to find the documentation for your Prometheus installation on how to access the Prometheus Web UI.

For the purposes of this guide, we are going to report on the `kube_deployment_status_replicas_unavailable` metric that is produced by `kube-state-metrics`, which should be available in your Prometheus instance if you're using the prometheus-operator in some form.

To validate, open your web browser to your Prometheus UI that we set up earlier and in the box above the `Execute` button, enter in `kube_deployment_status_replicas_unavailable`, and hit the `Execute` button.
You should see a list of metrics show up, if you have an empty list, it's possible you don't have `kube-state-metrics` running, or there's a configuration issue with Prometheus.

## Collecting additional Prometheus metrics

The first thing to figure out is what question do we want to ask, and what information is needed to do it.
For the example, we're going to try to answer the following question.

> What is the average and total amount of time that a particular deployment's replicas are unready?

To answer this question, we're going to use the `kube_deployment_status_replicas_unavailable` metric which tells us how many unready replicas a particular deployment has at a given moment in time.

Next, we need to figure out what information we care about from this metric, as a lot of it's not particularly useful to us.
The most relevant information available in this metric is the `namespace`, `deployment` labels and the actual value of the metric.

To strip out everything besides this information, the following Prometheus query [sums](https://prometheus.io/docs/prometheus/latest/querying/operators/#aggregation-operators) the value of the metric grouped by the `namespace` and `deployment`.

```
sum(kube_deployment_status_replicas_unavailable) by (namespace, deployment)
```

Try the above query in the Prometheus UI to get an idea of what this changes from the original metric and what it's doing.

## Writing a ReportPrometheusQuery and ReportDataSource

Now that we have figured out the Prometheus query, we need to create a [ReportPrometheusQuery][reportprometheusqueries] to make this query available to the metering operator.
Save the snippet below into a file named `unready-deployment-replicas-reportprometheusquery.yaml`:

```
apiVersion: metering.openshift.io/v1alpha1
kind: ReportPrometheusQuery
metadata:
  name: unready-deployment-replicas
spec:
  query: |
    sum(kube_deployment_status_replicas_unavailable) by (namespace, deployment)
```

Creating the ReportPrometheusQuery only makes the query available for use, but doesn't actually cause the metrics to be collected, to cause a Metric to be collected you need to create a [ReportDataSource][reportdatasources] with a `spec.promsum` section configured to use the ReportPrometheusQuery of your choice.
Save the snippet below into a file named `unready-deployment-replicas-reportdatasource.yaml`:

```
apiVersion: metering.openshift.io/v1alpha1
kind: ReportDataSource
metadata:
  name: unready-deployment-replicas
spec:
  promsum:
    query: "unready-deployment-replicas"
```

Creating the ReportDataSource will cause the metering operator to create a table in Presto and begin collecting metrics using the Prometheus query in the ReportPrometheusQuery, so let's create them:

```
kubectl create -n "$METERING_NAMESPACE" -f unready-deployment-replicas-reportprometheusquery.yaml
kubectl create -n "$METERING_NAMESPACE" -f unready-deployment-replicas-reportdatasource.yaml
```

## Viewing the Metrics in Presto

Before we go any further, we should verify that creating the ReportDataSource did what we wanted, and the data is being collected.
One way to do this is to check the [metering operator logs][reporting-operator-logs], and look for logs mentioning our `ReportDataSource` being collected and stored.

The other way however is to exec into the Presto pod and open up a Presto-cli session which allows us to interactively query Presto.
Open up a Presto-cli session by following the [Query Presto using presto-cli developer documentation][presto-cli-exec].
After you have a session run the following query:

```
show tables;
```

This should give you a list of Database Tables created in Presto, you should see quite a few entries, and among them `datasource_unready_deployment_replicas` should be in the list, if it is not, it's possible the table has not be created yet, or there was an error.
In this case, you should [check the metering operator logs][reporting-operator-logs] for errors.

If the table does exist, it may take up to 5 minutes (the default collection interval) before any data exists in the table.
To check if our data has started getting collected, we can check by issuing a `SELECT` query on our table to see if any rows exist:

```
SELECT * FROM datasource_unready_deployment_replicas LIMIT 10;
```

If at least one row shows up, then everything is working correctly, an example of the output expected is shown below:

```
presto:default> SELECT * FROM datasource_unready_deployment_replicas LIMIT 10;
 amount |        timestamp        | timeprecision |                                 labels
--------+-------------------------+---------------+-------------------------------------------------------------------------
    0.0 | 2018-05-18 17:00:00.000 |          60.0 | {namespace=tectonic-system, deployment=alm-operator}
    0.0 | 2018-05-18 17:00:00.000 |          60.0 | {namespace=tectonic-system, deployment=catalog-operator}
    0.0 | 2018-05-18 17:00:00.000 |          60.0 | {namespace=metering-testing, deployment=reporting-operator}
    0.0 | 2018-05-18 17:00:00.000 |          60.0 | {namespace=metering-testing, deployment=metering-operator}
    0.0 | 2018-05-18 17:00:00.000 |          60.0 | {namespace=tectonic-system, deployment=container-linux-update-operator}
    0.0 | 2018-05-18 17:00:00.000 |          60.0 | {namespace=tectonic-system, deployment=default-http-backend}
    0.0 | 2018-05-18 17:00:00.000 |          60.0 | {namespace=default, deployment=etcd-operator}
    0.0 | 2018-05-18 17:00:00.000 |          60.0 | {namespace=tectonic-system, deployment=etcd-operator}
    1.0 | 2018-05-18 17:00:00.000 |          60.0 | {namespace=default, deployment=example}
    0.0 | 2018-05-18 17:00:00.000 |          60.0 | {namespace=tectonic-system, deployment=grafana}
```

Now, it's time to start talking about our data in terms of SQL.
As you can see, our new table has 4 columns which are documented [in the ReportDataSource Table Schema documentation][datasource-table-schema].

## Writing Presto queries against Promsum ReportDataSource Data

Now that we have a database table that we can experiment with, we can begin to answer the question we started with:

> What is the average and total amount of time that a particular deployment's replicas are unready?

To answer this, we need to do a few things:
- for each timestamp, find the time each individual pod was unready
- divide up, or group the results by the deployment.
- find the average and total duration pods are unready for each deployment

We'll start with getting the unready time at an individual metric level.
Since the `amount` corresponds to the number of unready pods at that moment in time, and the `timeprecision` gives how long the metric was that value, we just need to multiply the `amount` (number of pods) by the `timeprecision` (length of time it was at that value):

```
SELECT
    labels['namespace'] as namespace,
    labels['deployment'] as deployment,
    amount * "timeprecision" as pod_unready_seconds
FROM datasource_unready_deployment_replicas
ORDER BY pod_unready_seconds DESC, namespace ASC, deployment ASC;
```

Next, we need to group our results by the deployment, this can be done using a SQL `GROUP BY` clause.
One thing we need to consider is a deployment of a specific name can appear in multiple namespaces, so we need to actually group by both namespace, and deployment.
Additionally a limitation of the `GROUP BY` is that columns not mentioned in the `GROUP BY` clause cannot be in the `SELECT` statement without an aggregation function, so we'll temporarily remove the `pod_unready_seconds` for this, but we'll come back to that shortly.

The query below uses the `GROUP BY` clause to create a list of deployments by namespace from the metric:

```
SELECT
    labels['namespace'] as namespace,
    labels['deployment'] as deployment
FROM datasource_unready_deployment_replicas
GROUP BY labels['namespace'], labels['deployment']
ORDER BY namespace ASC, deployment ASC;
```

Next, we want to get the average and total time that each deployment has replicas that are unready.
This can be done using the SQL `avg()` and `sum()` aggregation functions.
Before we can take the average and sum however, we need to figure out what we're averaging or summing, thankfully we already figure this out, it's the `pod_unready_seconds` column from the first SQL query.

Our final query is the following:

```
SELECT
    labels['namespace'] as namespace,
    labels['deployment'] as deployment,
    sum(amount * "timeprecision") AS total_replica_unready_seconds,
    avg(amount * "timeprecision") AS avg_replica_unready_seconds
FROM datasource_unready_deployment_replicas
GROUP BY labels['namespace'], labels['deployment']
ORDER BY total_replica_unready_seconds DESC, avg_replica_unready_seconds DESC, namespace ASC, deployment ASC;
```

## Writing a ReportGenerationQuery

Now that we have our final query, the time has come to put it into a [ReportGenerationQuery][reportgenerationqueries] resource.

The basic things you need to know when creating a `ReportGenerationQuery` is the query your going to use, the schema for that query, and the `ReportDataSources` or `ReportGenerationQueries` your query depends on.

For our example, we will add the `unready-deployment-replicas` `ReportDataSources` to the `spec.ReportDataSource`, and we'll add the query to `spec.query`.
The schema, is defined in the `spec.columns` field and is basically a list of the columns from the `SELECT` query and their SQL data types.
The column information is what the operator uses to create the table when a report is being generated.
If this doesn't match the query, there will be issues when running the query and trying to store the data into the database.

Below is an example of our final query from the steps above put into a `ReportGenerationQuery`.
One thing to note is we replaced `FROM datasource_unready_deployment_replicas` with `FROM {| dataSourceTableName "unready-deployment-replicas" |}` to avoid hard coding the table name.
The format of the table names could change in the future, so always use the `dataSourceTableName` template function to ensure it's always using the correct table name.

```
apiVersion: metering.openshift.io/v1alpha1
kind: ReportGenerationQuery
metadata:
  name: "unready-deployment-replicas"
spec:
  reportDataSources:
  - "unready-deployment-replicas"
  columns:
  - name: namespace
    type: string
  - name: deployment
    type: string
  - name: total_replica_unready_seconds,
    type: double
  - name: avg_replica_unready_seconds,
    type: double
  query: |
    SELECT
        labels['namespace'] as namespace,
        labels['deployment'] as deployment,
        sum(amount * "timeprecision") AS total_replica_unready_seconds,
        avg(amount * "timeprecision") AS avg_replica_unready_seconds
    FROM {| dataSourceTableName "unready-deployment-replicas" |}
    GROUP BY labels['namespace'], labels['deployment']
    ORDER BY total_replica_unready_seconds DESC, avg_replica_unready_seconds DESC, namespace ASC, deployment ASC
```

However, the above example is missing one crucial bit, and that's the ability to constrain the period of time that this query is reporting over.
To handle this, the `.Report` variable is accessible within templates and contains a `.Report.StartPeriod` and `.Report.EndPeriod` field which will be filled in with values corresponding to the Report or ScheduledReport's reporting period.
We can use these variables in a `WHERE` clause within our query to filter the results to those time ranges.

The `WHERE` clause generally for this generally looks the same for all `ReportGenerationQueries` that expect to be used by a Report:

```
WHERE "timestamp" >= timestamp '{|.Report.StartPeriod | prestoTimestamp |}'
AND "timestamp" <= timestamp '{| .Report.EndPeriod | prestoTimestamp |}'
```

However, once we begin accessing variables in the `ReportGenerationQuery` this means we have to set `spec.view.disabled` to `true` because our query is now dynamic and depends on user-input, meaning we cannot create a view.
Adding this filter to our query and updating our `spec.view.disabled` to true and we get the final version of our ReportGenerationQuery.
Save the snippet below into a file named `unready-deployment-replicas-reportgenerationquery.yaml`:

```
apiVersion: metering.openshift.io/v1alpha1
kind: ReportGenerationQuery
metadata:
  name: "unready-deployment-replicas"
spec:
  reportDataSources:
  - "unready-deployment-replicas"
  view:
    disabled: true
  columns:
  - name: namespace
    type: string
  - name: deployment
    type: string
  - name: total_replica_unready_seconds
    type: double
  - name: avg_replica_unready_seconds
    type: double
  query: |
    SELECT
        labels['namespace'] as namespace,
        labels['deployment'] as deployment,
        sum(amount * "timeprecision") AS total_replica_unready_seconds,
        avg(amount * "timeprecision") AS avg_replica_unready_seconds
    FROM {| dataSourceTableName "unready-deployment-replicas" |}
    WHERE "timestamp" >= timestamp '{|.Report.StartPeriod | prestoTimestamp |}'
    AND "timestamp" <= timestamp '{| .Report.EndPeriod | prestoTimestamp |}'
    GROUP BY labels['namespace'], labels['deployment']
    ORDER BY total_replica_unready_seconds DESC, avg_replica_unready_seconds DESC, namespace ASC, deployment ASC
```

Next, let's create the `ReportGenerationQuery` so it can be used by Reports:

```
kubectl create -n "$METERING_NAMESPACE" -f unready-deployment-replicas-reportgenerationquery.yaml
```

## Creating a Report

Save the snippet below into a file named `unready-deployment-replicas-report.yaml`:

```
apiVersion: metering.openshift.io/v1alpha1
kind: Report
metadata:
  name: unready-deployment-replicas
spec:
  reportingStart: '2018-01-01T00:00:00Z'
  reportingEnd: '2018-12-31T23:59:59Z'
  generationQuery: "unready-deployment-replicas"
  runImmediately: true
```

Next, let's create the report and let the operator generate the results:

```
kubectl create -n "$METERING_NAMESPACE" -f unready-deployment-replicas-reportgenerationquery.yaml
```

Taking a report may take a while, but you can check on the report's status by reading the `status` field from output of the command below:

```
kubectl -n $METERING_NAMESPACE get report unready-deployment-replicas -o json
```

Once the Report's status has changed to `Finished` (this can take a few minutes depending on cluster size and amount of data collected), we can query the operator's HTTP API for the results:

```
kubectl proxy &
sleep 2
curl "http://127.0.0.1:8001/api/v1/namespaces/$METERING_NAMESPACE/services/metering:http/proxy/api/v1/reports/get?name=unready-deployment-replicas&format=csv"
```

This should output a CSV report that looks similar to this:

```
period_start,period_end,namespace,deployment,total_replica_unready_seconds,avg_replica_unready_seconds
2018-01-01 00:00:00 +0000 UTC,2018-12-31 23:59:59 +0000 UTC,metering-testing,reporting-operator,3420,13.464566929133857
2018-01-01 00:00:00 +0000 UTC,2018-12-31 23:59:59 +0000 UTC,metering-testing,metering-operator,0,0
2018-01-01 00:00:00 +0000 UTC,2018-12-31 23:59:59 +0000 UTC,metering-testing,presto,0,0
2018-01-01 00:00:00 +0000 UTC,2018-12-31 23:59:59 +0000 UTC,default,etcd-operator,0,0
2018-01-01 00:00:00 +0000 UTC,2018-12-31 23:59:59 +0000 UTC,default,example,15240,60
2018-01-01 00:00:00 +0000 UTC,2018-12-31 23:59:59 +0000 UTC,default,test-deploy,15240,60
2018-01-01 00:00:00 +0000 UTC,2018-12-31 23:59:59 +0000 UTC,default,test-hostnet-deploy,0,0
2018-01-01 00:00:00 +0000 UTC,2018-12-31 23:59:59 +0000 UTC,kube-system,heapster,0,0
2018-01-01 00:00:00 +0000 UTC,2018-12-31 23:59:59 +0000 UTC,kube-system,kube-controller-manager,0,0
2018-01-01 00:00:00 +0000 UTC,2018-12-31 23:59:59 +0000 UTC,kube-system,kube-dns,0,0
2018-01-01 00:00:00 +0000 UTC,2018-12-31 23:59:59 +0000 UTC,kube-system,kube-scheduler,0,0
2018-01-01 00:00:00 +0000 UTC,2018-12-31 23:59:59 +0000 UTC,metering,metering,15240,60
2018-01-01 00:00:00 +0000 UTC,2018-12-31 23:59:59 +0000 UTC,metering,metering-operator,0,0
...
```

Creating and using reports is covered in more detail in the [Using Metering documentation][using-metering].

## Summary

Here's a summary of what we did in this guide:

- We wrote a Prometheus query that collects metrics on unready deployment replicas
- We created a `ReportPrometheusQuery` containing this query
- We created a `ReportDataSource` that gave us a Presto table containing the metrics from our Prometheus query
- We wrote a Presto SQL query that calculates the average and total number of seconds pods are unready for each deployment
- We wrote a `ReportGenerationQuery` that does our calculation, and handles filtering the results to a Reports configured time range
- We created a `Report` that uses our `ReportGenerationQuery`
- We checked that the Report finished, and then fetched the results from the metering operator HTTP API.

[reportdatasources]: reportdatasources.md
[reportprometheusqueries]: reportprometheusqueries.md
[reportgenerationqueries]: reportgenerationqueries.md
[reports]: report.md
[install-metering]: install-metering.md
[presto-cli-exec]:  dev/debugging.md#query-presto-using-presto-cli
[reporting-operator-logs]:  dev/debugging.md#get-reporting-operator-logs
[presto-types]: https://prestodb.io/docs/current/language/types.html
[using-metering]: using-metering.md
[datasource-table-schema]: reportdatasources.md#table-schemas
