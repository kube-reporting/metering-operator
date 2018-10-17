# Report Data Sources

A `ReportDataSource` is a custom resource that represents how to store data, such as where it should be stored, and in some cases, how the data is to be collected.

There are currently two types of ReportDataSource's, `promsum`, and `awsBilling`.
Each has a corresponding con***REMOVED***guration section within the `spec` of a `ReportDataSource`.
The main effect that creating a ReportDataSource has is that it causes the metering operator to create a table in Presto. Depending on the type of ReportDataSource it then may do other additional tasks. For `promsum` data sources the operator periodically collects metrics and stores them in the table.
For `awsBilling`, the operator con***REMOVED***gures the table to point at an S3 bucket containing [AWS Cost and Usage reports][AWS-billing], making these reports exposed as a database table.
To read more details on how the different ReportDataSources work, read the [metering architecture document][architecture].

## Fields

- `promsum`: If this section is present, then the `ReportDataSource` will be con***REMOVED***gured to periodically poll Prometheus for metrics using the speci***REMOVED***ed `ReportPrometheusQuery`.
  - `query`: The name of the `ReportPrometheusQuery` resource.
  - `storage`: This section controls the `StorageLocation` options, allowing you to control on a per ReportDataSource level, where data is stored.
    - `storageLocationName`: The name of the `StorageLocation` resource to use.
    - `spec`: If `storageLocationName` is not set, then this section is used to control the storage location settings. See the [StorageLocation documentation][storage-locations] for details on what can be speci***REMOVED***ed here. Anything valid in a `StorageLocation`'s `spec` is valid here.
  - `prometheusCon***REMOVED***g`:
    - `url`: If present, the URL of the Prometheus instance to scrape for this ReportDataSource.
- `awsBilling`:
  - `source`:
    - `bucket`: Bucket name to store data into.
    - `pre***REMOVED***x`: Path within the bucket where to store data.
    - `region`: The region where bucket is located.

## Table Schemas

For ReportDataSources with a `spec.promsum` present, their tables have the following database table schema:

- `timestamp`: The type of this column is `timestamp`. This is the time which the metric was collected.
   - Note: `timestamp` is also a reserved keyword (for the column type) in Presto, meaning any queries using it must use quotes to refer to the column, like so: `SELECT "timestamp" FROM datasource_unready_deployment_replicas LIMIT 1;`
- `timeprecision`: The type of this column is a `double`. This is "query resolution step width" used to query this metric from Prometheus. This de***REMOVED***nes how accurate the data is. The bigger the value, the less accurate. This value is controlled globally by the operator, and has a default value of 60.
- `labels`: The type of this column is a `map(varchar, varchar)`. This is the set of Prometheus labels and their values for the metric.
- `amount`: The type of this column is a `double`. Amount is the value of the metric at that `timestamp`

For ReportDataSources with a `spec.awsBilling` present, see [here](aws-billing-datasource-schema.md) for an example of what the table schema looks like.

For more details read [the Presto Data Type documentation][presto-types].

## Example ReportDataSource

Below is an example of one of the built-in `ReportDataSource` resources that is installed with Operator Metering by default.
This example doesn't specify a `spec.storage` section, which will result in it using the [default StorageLocation resource][default-storage-location].

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

This example is a slightly modi***REMOVED***ed version of the one above that does set the `storage` to the `StorageLocation` with a `metadata.name` of "local".

```
apiVersion: metering.openshift.io/v1alpha1
kind: ReportDataSource
metadata:
  name: "pod-request-memory-bytes-custom-storage-location"
  labels:
    operator-metering: "true"
spec:
  promsum:
    query: "pod-request-memory-bytes"
    storage:
      storageLocationName: local
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

[storage-locations]: storagelocations.md
[AWS-billing]: https://docs.aws.amazon.com/awsaccountbilling/latest/aboutv2/billing-reports-costusage.html
[metering-aws-billing-conf]: metering-con***REMOVED***g.md#aws-billing-correlation
[default-storage-location]: storagelocations.md#default-storagelocation
[architecture]: metering-architecture.md
[presto-types]: https://prestodb.io/docs/current/language/types.html
