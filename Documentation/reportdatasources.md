# Report Data Sources

A `ReportDataSource` is a custom resource that represents how to store data, such as where it should be stored, and in some cases, how the data is to be collected.

There are currently four types of ReportDataSource's: `promsum`, `awsBilling`, `generationQueryView` and `prestoTable`.
Each has a corresponding con***REMOVED***guration section within the `spec` of a `ReportDataSource`.
The main effect that creating a ReportDataSource has is that it causes the metering operator to create a table in Presto or Hive.
Depending on the type of ReportDataSource it then may do other additional tasks.
For `promsum` datasources the operator periodically collects metrics and stores them in the table.
For `awsBilling`, the operator con***REMOVED***gures the table to point at an S3 bucket containing [AWS Cost and Usage reports][AWS-billing], making these reports exposed as a database table.
To read more details on how the different ReportDataSources work, read the [metering architecture document][architecture].

## Fields

- `promsum`: If this section is present, then the `ReportDataSource` will be con***REMOVED***gured to periodically poll Prometheus for metrics using the speci***REMOVED***ed `ReportPrometheusQuery`.
  - `query`: The name of the `ReportPrometheusQuery` resource.
  - `storage`: This section controls the `StorageLocation` options, allowing you to control on a per ReportDataSource level, where data is stored.
    - `storageLocationName`: The name of the `StorageLocation` resource to use.
  - `prometheusCon***REMOVED***g`:
    - `url`: If present, the URL of the Prometheus instance to scrape for this ReportDataSource.
- `awsBilling`: If speci***REMOVED***ed, the `ReportDataSource` will be con***REMOVED***gured to use an S3 bucket containing AWS billing reports as its source of data.
  - `source`:
    - `bucket`: Bucket name to store data into.
    - `pre***REMOVED***x`: Path within the bucket where to store data.
    - `region`: The region where bucket is located.
- `generationQueryView`: If this section is present, then the `ReportDataSource` will be con***REMOVED***gured to create a View in Presto using the rendered `spec.query` as the query for the view.
  - `queryName`: The name of a [ReportGenerationQuery][reportgenerationquery] to create a view from.
  - `inputs`: A list of inputs this report query accepts to control its behavior. For more in depth details, see the [ReportGenerationQuery query inputs][query-inputs] documentation.
  - `storage`: This section controls the `StorageLocation` options, allowing you to control on a per ReportDataSource level, where data is stored.
    - `storageLocationName`: The name of the `StorageLocation` resource to use.
- `prestoTable`: If present, then the `ReportDataSource` will simply make it possible to reference a database table within Presto as a ReportDataSource.
  - `tableRef`: The name of the [PrestoTable][prestotable] that this ReportDataSource should refer to.

## Promsum Datasource

For ReportDataSources with a `spec.promsum` present, their tables have the following database table schema:

- `timestamp`: The type of this column is `timestamp`. This is the time which the metric was collected.
   - Note: `timestamp` is also a reserved keyword (for the column type) in Presto, meaning any queries using it must use quotes to refer to the column, like so: `SELECT "timestamp" FROM datasource_unready_deployment_replicas LIMIT 1;`
- `timeprecision`: The type of this column is a `double`. This is "query resolution step width" used to query this metric from Prometheus. This de***REMOVED***nes how accurate the data is. The bigger the value, the less accurate. This value is controlled globally by the operator, and has a default value of 60.
- `labels`: The type of this column is a `map(varchar, varchar)`. This is the set of Prometheus labels and their values for the metric.
- `amount`: The type of this column is a `double`. Amount is the value of the metric at that `timestamp`

### Example Promsum Datasource

Below is an example of one of the built-in `ReportDataSource` resources that is installed with Operator Metering by default.

```
apiVersion: metering.openshift.io/v1alpha1
kind: ReportDataSource
metadata:
  name: "pod-request-memory-bytes"
  labels:
    operator-metering: "true"
spec:
  promsum:
    query: "pod-request-memory-bytes"
```

If the data to be scraped is on a non-default Prometheus instance:

```
apiVersion: metering.openshift.io/v1alpha1
kind: ReportDataSource
metadata:
  name: "pod-request-memory-bytes"
  labels:
    operator-metering: "true"
spec:
  promsum:
    query: "pod-request-memory-bytes"
    prometheusCon***REMOVED***g:
      url: http://custom-prometheus-instance:9090
```

## ReportGenerationQuery View Datasource

For ReportDataSources with a `spec.generationQueryView` present, a Presto view will be created using the rendered output of a speci***REMOVED***ed [ReportGenerationQuery][reportgenerationquery]'s `spec.query` ***REMOVED***eld.
This enables abstracting away the details of more complex queries by exposing them as a database table whose content is based on the result of of the query the view is based on.
It also enables re-use by allowing you to create a view containing the complexities of a query allowing other queries to simply query it as a regular table.

### Example ReportGenerationQuery View Datasource

This example exposes the `pod-memory-request-raw` ReportGenerationQuery as a view.
The schema is based on the `spec.columns` of the ReportGenerationQuery.

```
apiVersion: metering.openshift.io/v1alpha1
kind: ReportDataSource
metadata:
  name: "pod-memory-request-raw"
  labels:
    operator-metering: "true"
spec:
  generationQueryView:
    queryName: pod-memory-request-raw
```


## AWS Billing Datasource

For ReportDataSources with a `spec.awsBilling` present, see [here](aws-billing-datasource-schema.md) for an example of what the table schema looks like.

### Example AWS Billing Datasource

```
apiVersion: metering.openshift.io/v1alpha1
kind: ReportDataSource
metadata:
  name: "aws-billing"
  labels:
    operator-metering: "true"
spec:
  awsBilling:
    source:
      bucket: "your-aws-cost-report-bucket"
      pre***REMOVED***x: "path/to/report"
      region: "your-buckets-region"
```

## PrestoTable Datasource

For ReportDataSources with a `spec.prestoTable` present, the reporting-operator will simply verify that a [PrestoTable resource][prestotables] resource exists and it's `status.tableName` is set.
If it does, then the ReportDataSource will simply point at the existing PrestoTable.
A PrestoTable ReportDataSource is merely a way to expose an arbitrary table to the rest of the metering resources which expect to interact with a ReportDataSource.

### Example PrestoTable Datasource

```
apiVersion: metering.openshift.io/v1alpha1
kind: ReportDataSource
metadata:
  name: example-baremetal-cost
spec:
  prestoTable:
    tableRef:
      name: example-baremetal-cost
```

[storage-locations]: storagelocations.md
[AWS-billing]: https://docs.aws.amazon.com/awsaccountbilling/latest/aboutv2/billing-reports-costusage.html
[metering-aws-billing-conf]: metering-con***REMOVED***g.md#aws-billing-correlation
[default-storage-location]: storagelocations.md#default-storagelocation
[architecture]: metering-architecture.md
[presto-types]: https://prestodb.io/docs/current/language/types.html
[reportgenerationqueries]: reportgenerationqueries.md
[query-inputs]: reportgenerationqueries.md#query-inputs
[prestotables]: prestotables.md
