# Writing Custom Report Prometheus Queries and Report Queries

One of the main goals of Operator Metering is to be flexible, and extensible.
The way this has been done is to use Kubernetes Custom Resources as a way of letting users add to, or expand upon the built-in reports and metrics that the operator already has.

The primary Custom Resources that allow this are the [ReportDataSource][reportdatasources], and the [ReportQuery][reportqueries].
It's highly recommended you read the documentation on each of these resources before continuing.

A ReportDataSource can be configured to cause Operator Metering to collect additional Prometheus Metrics by allowing users to write custom Prometheus queries and store the metrics collected.
Once this is done a ReportQuery can be used to analyze the metrics.

This guide is going to be structured such that you begin by collecting new Prometheus Metrics, and by the end you will be writing custom SQL queries that analyze these metrics.

## Setup

This guide assumes you've already [installed Metering][install-metering] and have Prometheus in your cluster.

> This guide assumes you've set the `METERING_NAMESPACE` environment variable to the namespace your Operator Metering installation is running. All resources must be created in that namespace for the operator to view and access them.
> Additionally, in some cases, the value of `$METERING_NAMESPACE` will be used directly and the command you need to run may need to be adjusted based on your value. To make it clear where you will need to adjust your values, we will assume the value of "$METERING_NAMESPACE" is `your-namespace`.

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
You should see a list of metrics show up. If you have an empty list, it's possible that you don't have `kube-state-metrics` running, or there's a configuration issue with Prometheus.

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

## Writing a PrometheusMetricsImporter ReportDataSource

To configure a Metric to be collected, you need to create a [ReportDataSource][reportdatasources] with a `spec.prometheusMetricsImporter` section configured to use the Prometheus query of your choice.
Save the snippet below into a file named `unready-deployment-replicas-reportdatasource.yaml`:

```
apiVersion: metering.openshift.io/v1
kind: ReportDataSource
metadata:
  name: unready-deployment-replicas
spec:
  prometheusMetricsImporter:
    query: |
      sum(kube_deployment_status_replicas_unavailable) by (namespace, deployment)
```

Creating the ReportDataSource will cause the metering operator to create a table in Presto and begin collecting metrics using the Prometheus query in the `spec.prometheusMetricsImporter.query` field, so let's create it:

```
kubectl create -n "$METERING_NAMESPACE" -f unready-deployment-replicas-reportdatasource.yaml
```

## Viewing the Metrics in Presto

Before we go any further, we should verify that creating the ReportDataSource did what we wanted, and the data is being collected.
One way to do this is to check the [metering operator logs][reporting-operator-logs], and look for logs mentioning our `ReportDataSource` being collected and stored.

The other way however is to exec into the Presto pod and open up a Presto-cli session which allows us to interactively query Presto.
Open up a Presto-cli session by following the [Query Presto using presto-cli developer documentation][presto-cli-exec].
After you have a session, set the schema to `metering`, and run the following query:

```
use metering;
show tables;
```

This should give you a list of Database Tables created in Presto, within the hive catalog, in the metering schema, and you should see quite a few entries.
Among these entries, `datasource_your_namespace_unready_deployment_replicas` should be in the list (replacing `your_namespace` with the value of `$METERING_NAMESPACE` with `-` replaced with `_`), and if it's not, it's possible the table has not be created yet, or there was an error.
In this case, you should [check the metering operator logs][reporting-operator-logs] for errors.

If the table does exist, it may take up to 5 minutes (the default collection interval) before any data exists in the table.
To check if our data has started getting collected, we can check by issuing a `SELECT` query on our table to see if any rows exist:

```
SELECT * FROM datasource_your_namespace_unready_deployment_replicas LIMIT 10;
```

If at least one row shows up, then everything is working correctly, an example of the output expected is shown below:

```
presto:metering> SELECT * FROM datasource_your_namespace_unready_deployment_replicas LIMIT 10;
 amount |        timestamp        | timeprecision |                              labels                              |     dt
--------+-------------------------+---------------+------------------------------------------------------------------+------------
    0.0 | 2019-05-03 16:01:00.000 |          60.0 | {namespace=telemeter-tschuy, deployment=presto-worker}           | 2019-05-03
    0.0 | 2019-05-03 16:01:00.000 |          60.0 | {namespace=metering-emoss, deployment=metering-operator}         | 2019-05-03
    0.0 | 2019-05-03 16:01:00.000 |          60.0 | {namespace=openshift-monitoring, deployment=prometheus-operator} | 2019-05-03
    0.0 | 2019-05-03 16:01:00.000 |          60.0 | {namespace=metering-tschuy, deployment=presto-worker}            | 2019-05-03
    0.0 | 2019-05-03 16:34:00.000 |          60.0 | {namespace=telemeter-tschuy, deployment=presto-worker}           | 2019-05-03
    1.0 | 2019-05-03 16:18:00.000 |          60.0 | {namespace=metering-tschuy, deployment=reporting-operator}       | 2019-05-03
    0.0 | 2019-05-03 16:12:00.000 |          60.0 | {namespace=telemeter-tschuy, deployment=presto-worker}           | 2019-05-03
    0.0 | 2019-05-03 16:01:00.000 |          60.0 | {namespace=metering-tflannag, deployment=reporting-operator}     | 2019-05-03
    0.0 | 2019-05-03 16:01:00.000 |          60.0 | {namespace=metering-tflannag, deployment=presto-worker}          | 2019-05-03
    1.0 | 2019-05-03 16:01:00.000 |          60.0 | {namespace=metering-chancez, deployment=reporting-operator}      | 2019-05-03
(10 rows)
```

Now, it's time to start talking about our data in terms of SQL.
As you can see, our new table has 4 columns which are documented [in the ReportDataSource Table Schema documentation][datasource-table-schema].

## Writing Presto queries against PrometheusMetricsImporter ReportDataSource Data

Now that we have a database table that we can experiment with, we can begin to answer the question we started with:

> What is the average and total amount of time that a particular deployment's replicas are unready?

To answer this, we need to do a few things:
- for each timestamp, find the time each individual pod was unready.
- divide up, or group the results by the deployment.
- find the average and total duration pods are unready for each deployment.

We'll start with getting the unready time at an individual metric level for each timestamp.
Since the `amount` corresponds to the number of unready pods at that moment in time, and the `timeprecision` gives how long the metric was that value, we just need to multiply the `amount` (number of pods) by the `timeprecision` (length of time it was at that value):

```
SELECT
    "timestamp",
    labels['namespace'] as namespace,
    labels['deployment'] as deployment,
    amount * "timeprecision" as pod_unready_seconds
FROM datasource_your_namespace_unready_deployment_replicas
ORDER BY pod_unready_seconds DESC, namespace ASC, deployment ASC
LIMIT 10;
```

Next, we need to group our results by the deployment, this can be done using a SQL `GROUP BY` clause.
One thing we need to consider is a deployment of a specific name can appear in multiple namespaces, so we need to actually group by both namespace, and deployment.
Additionally a limitation of the `GROUP BY` is that columns not mentioned in the `GROUP BY` clause cannot be in the `SELECT` statement without an aggregation function, so we'll temporarily remove the `pod_unready_seconds` for this, but we'll come back to that shortly.
Of course we include timestamp again so we can get the value at each time.

The query below uses the `GROUP BY` clause to create a list of deployments by namespace from the metric:

```
SELECT
    "timestamp",
    labels['namespace'] as namespace,
    labels['deployment'] as deployment
FROM datasource_your_namespace_unready_deployment_replicas
GROUP BY "timestamp", labels['namespace'], labels['deployment']
ORDER BY namespace ASC, deployment ASC
LIMIT 10;
```

Next, we want to get the average and total time that each deployment has replicas that are unready for each timestamp.
This can be done using the SQL `avg()` and `sum()` aggregation functions.
Before we can take the average and sum however, we need to figure out what we're averaging or summing, thankfully we already figure this out, it's the `pod_unready_seconds` column from the first SQL query.

Our final query is the following:

```
SELECT
    "timestamp",
    labels['namespace'] as namespace,
    labels['deployment'] as deployment,
    sum(amount * "timeprecision") AS total_replica_unready_seconds,
    avg(amount * "timeprecision") AS avg_replica_unready_seconds
FROM datasource_your_namespace_unready_deployment_replicas
GROUP BY "timestamp", labels['namespace'], labels['deployment']
ORDER BY total_replica_unready_seconds DESC, avg_replica_unready_seconds DESC, namespace ASC, deployment ASC
LIMIT 10;
```

## Writing a ReportQuery

Now that we have our final query, the time has come to put it into a [ReportQuery][reportqueries] resource.

The basic things you need to know when creating a `ReportQuery` is the query you're going to use, the schema for that query, and the `ReportDataSources` or `ReportQueries` your query depends on.

For our example, we will add the `unready-deployment-replicas` `ReportDataSources` to the `spec.ReportDataSource`, and we'll add the query to `spec.query`.
The schema, which is defined in the `spec.columns` field, is basically a list of the columns from the `SELECT` query and their SQL data types.
The column information is what the operator uses to create the table when a report is being generated.
If this doesn't match the query, there will be issues when running the query and trying to store the data into the database.

Below is an example of our final query from the steps above put into a `ReportQuery`.
One thing to note is we replaced `FROM datasource_your_namespace_unready_deployment_replicas` with `{| dataSourceTableName .Report.Inputs.UnreadyDeploymentReplicasDataSourceName |}` and added an `inputs` configuration to avoid hard coding the table name.
By using inputs, we can override the default ReportDataSource used and by marking it as `type: ReportDataSource`, it will be considered a dependency and will ensure it exists before running.
The format of the table names could change in the future, so always use the `dataSourceTableName` template function to ensure it's always using the correct table name.

```
apiVersion: metering.openshift.io/v1
kind: ReportQuery
metadata:
  name: "unready-deployment-replicas"
spec:
  columns:
  - name: timestamp
    type: timestamp
  - name: namespace
    type: varchar
  - name: deployment
    type: varchar
  - name: total_replica_unready_seconds
    type: double
  - name: avg_replica_unready_seconds
    type: double
  inputs:
  - name: UnreadyDeploymentReplicasDataSourceName
    type: ReportDataSource
    default: unready-deployment-replicas
  query: |
    SELECT
        "timestamp",
        labels['namespace'] AS namespace,
        labels['deployment'] AS deployment,
        sum(amount * "timeprecision") AS total_replica_unready_seconds,
        avg(amount * "timeprecision") AS avg_replica_unready_seconds
    FROM {| dataSourceTableName .Report.Inputs.UnreadyDeploymentReplicasDataSourceName |}
    GROUP BY "timestamp", labels['namespace'], labels['deployment']
    ORDER BY total_replica_unready_seconds DESC, avg_replica_unready_seconds DESC, namespace ASC, deployment ASC
```

However, the above example is missing one crucial bit, and that's the ability to constrain the period of time that this query is reporting over.
To handle this, the `.Report` variable is accessible within templates and contains a `.Report.StartPeriod` and `.Report.EndPeriod` field which will be filled in with values corresponding to the Report's reporting period.
We can use these variables in a `WHERE` clause within our query to filter the results to those time ranges.

The `WHERE` clause generally looks the same for all `ReportQueries` that expect to be used by a Report:

```
WHERE "timestamp" >= timestamp '{| default .Report.ReportingStart .Report.Inputs.ReportingStart | prestoTimestamp |}'
AND "timestamp" < timestamp '{| default .Report.ReportingEnd .Report.Inputs.ReportingEnd | prestoTimestamp |}'
```

Queries should be [left-closed and right-open](https://en.wikipedia.org/wiki/Interval_(mathematics)#Classification_of_intervals); that is, we should collect data with timestamps equal to or greater than the start time and less than the end time, as seen in the example above.

In addition to the query time constraints, we often want to be able to track the time period for each row of data. In order to do this, we can append two columns to the above schema definition: `period_start` and `period_end` and remove "timestamp", since we're looking at a range of time rather than an instant in time. Both of these columns will be of type `timestamp`, which requires us to add an additional field, `spec.input`, to our `ReportQuery` as this is a custom input. To see more about `spec.inputs`, `ReportingStart`, and `ReportingEnd` see [reports.md.](https://github.com/kubernetes-reporting/metering-operator/blob/master/Documentation/reports.md#reportingstart)
Lastly, we need to update the SELECT portion of the `spec.query` field:

```
query: |
  SELECT
    timestamp '{| default .Report.ReportingStart .Report.Inputs.ReportingStart | prestoTimestamp |}' AS period_start,
    timestamp '{| default .Report.ReportingEnd .Report.Inputs.ReportingEnd | prestoTimestamp |}' AS period_end,
    labels['namespace'] AS namespace,
    ...
```

Once we add these columns filters to our query we get the final version of our ReportQuery.
Save the snippet below into a file named `unready-deployment-replicas-reportquery.yaml`:

```
apiVersion: metering.openshift.io/v1
kind: ReportQuery
metadata:
  name: "unready-deployment-replicas"
spec:
  columns:
  - name: period_start
    type: timestamp
  - name: period_end
    type: timestamp
  - name: namespace
    type: varchar
  - name: deployment
    type: varchar
  - name: total_replica_unready_seconds
    type: double
  - name: avg_replica_unready_seconds
    type: double
  inputs:
  - name: ReportingStart
    type: time
  - name: ReportingEnd
    type: time
  - name: UnreadyDeploymentReplicasDataSourceName
    type: ReportDataSource
    default: unready-deployment-replicas
  query: |
    SELECT
        timestamp '{| default .Report.ReportingStart .Report.Inputs.ReportingStart | prestoTimestamp |}' AS period_start,
        timestamp '{| default .Report.ReportingEnd .Report.Inputs.ReportingEnd | prestoTimestamp |}' AS period_end,
        labels['namespace'] AS namespace,
        labels['deployment'] AS deployment,
        sum(amount * "timeprecision") AS total_replica_unready_seconds,
        avg(amount * "timeprecision") AS avg_replica_unready_seconds
    FROM {| dataSourceTableName .Report.Inputs.UnreadyDeploymentReplicasDataSourceName |}
    WHERE "timestamp" >= timestamp '{| default .Report.ReportingStart .Report.Inputs.ReportingStart | prestoTimestamp |}'
    AND "timestamp" < timestamp '{| default .Report.ReportingEnd .Report.Inputs.ReportingEnd | prestoTimestamp |}'
    GROUP BY labels['namespace'], labels['deployment']
    ORDER BY total_replica_unready_seconds DESC, avg_replica_unready_seconds DESC, namespace ASC, deployment ASC
```

Next, let's create the `ReportQuery` so it can be used by Reports:

```
kubectl create -n "$METERING_NAMESPACE" -f unready-deployment-replicas-reportquery.yaml
```

## Creating a Report

Save the snippet below into a file named `unready-deployment-replicas-report.yaml`:

```
apiVersion: metering.openshift.io/v1
kind: Report
metadata:
  name: unready-deployment-replicas
spec:
  reportingStart: '2019-01-01T00:00:00Z'
  reportingEnd: '2019-12-31T23:59:59Z'
  query: "unready-deployment-replicas"
  runImmediately: true
```

Next, let's create the report and let the operator generate the results:

```
kubectl create -n "$METERING_NAMESPACE" -f unready-deployment-replicas-report.yaml
```

Creating a report may take a while, but you can check on the report's status by reading the `status` field from output of the command below:

```
kubectl -n $METERING_NAMESPACE get report unready-deployment-replicas -o json
```

Once the Report's status has changed to `Finished` (this can take a few minutes depending on cluster size and amount of data collected), we can query the operator's HTTP API for the results:

```
METERING_ROUTE_HOSTNAME=$(oc -n $METERING_NAMESPACE get routes metering -o json | jq -r '.status.ingress[].host')
TOKEN=$(oc -n $METERING_NAMESPACE serviceaccounts get-token reporting-operator)
curl -H "Authorization: Bearer $TOKEN" -k "https://$METERING_ROUTE_HOSTNAME/api/v1/reports/get?name=unready-deployment-replicas&namespace=$METERING_NAMESPACE&format=csv"
```

If you aren't using Openshift, or have disabled `spec.tls.enabled` top-level key, follow the [viewing reports][viewing-reports] section, and come back to this document afterwards.

This should output a CSV report that looks similar to this:

```
period_start,period_end,namespace,deployment,total_replica_unready_seconds,avg_replica_unready_seconds
2019-01-01 00:00:00 +0000 UTC,2019-12-31 23:59:59 +0000 UTC,kube-system,tiller-deploy,0.000000,0.000000
2019-01-01 00:00:00 +0000 UTC,2019-12-31 23:59:59 +0000 UTC,metering-chancez,metering-operator,120.000000,1.000000
2019-01-01 00:00:00 +0000 UTC,2019-12-31 23:59:59 +0000 UTC,metering-chancez,presto-coordinator,360.000000,3.050847
2019-01-01 00:00:00 +0000 UTC,2019-12-31 23:59:59 +0000 UTC,metering-chancez,presto-worker,0.000000,0.000000
2019-01-01 00:00:00 +0000 UTC,2019-12-31 23:59:59 +0000 UTC,metering-chancez,reporting-operator,1680.000000,14.237288
2019-01-01 00:00:00 +0000 UTC,2019-12-31 23:59:59 +0000 UTC,openshift-monitoring,cluster-monitoring-operator,0.000000,0.000000
2019-01-01 00:00:00 +0000 UTC,2019-12-31 23:59:59 +0000 UTC,openshift-monitoring,grafana,0.000000,0.000000
2019-01-01 00:00:00 +0000 UTC,2019-12-31 23:59:59 +0000 UTC,openshift-monitoring,kube-state-metrics,0.000000,0.000000
2019-01-01 00:00:00 +0000 UTC,2019-12-31 23:59:59 +0000 UTC,openshift-monitoring,prometheus-operator,0.000000,0.000000
2019-01-01 00:00:00 +0000 UTC,2019-12-31 23:59:59 +0000 UTC,openshift-web-console,webconsole,0.000000,0.000000
...
```

Creating and using reports is covered in more detail in the [Using Metering documentation][using-metering].

## Summary

Here's a summary of what we did in this guide:

- We wrote a Prometheus query that collects metrics on unready deployment replicas.
- We created a `ReportDataSource` that gave us a Presto table containing the metrics from our Prometheus query.
- We wrote a Presto SQL query that calculates the average and total number of seconds that pods are unready for each deployment.
- We wrote a `ReportQuery` that does our calculation, and handles filtering the results to a Report's configured time range.
- We created a `Report` that uses our `ReportQuery`.
- We checked that the Report finished, and then fetched the results from the metering operator HTTP API.

[reportdatasources]: reportdatasources.md
[reportqueries]: reportqueries.md
[reports]: reports.md
[install-metering]: install-metering.md
[presto-cli-exec]:  dev/debugging.md#query-presto-using-presto-cli
[reporting-operator-logs]:  dev/debugging.md#get-reporting-operator-logs
[presto-types]: https://prestodb.io/docs/current/language/types.html
[using-metering]: using-metering.md
[datasource-table-schema]: reportdatasources.md#table-schemas
[viewing-reports]: using-metering.md#viewing-reports
