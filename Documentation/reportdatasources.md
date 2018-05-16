# Report Data Sources

A `ReportDataSource` is a custom resource that represents how to store data, such as where it should be stored, and in some cases, how the data is to be collected.

There are currently two types of ReportDataSource's, `promsum`, and `awsBilling`.
Each has a corresponding configuration section within the `spec` of a `ReportDataSource`.
The main effect that creating a ReportDataSource has is that it causes the metering operator to create a table in Presto. Depending on the type of ReportDataSource it then may do other additional tasks. For `promsum` data sources the operator periodically collects metrics and stores them in the table.
For `awsBilling`, the operator configures the table to point at an S3 bucket containing [AWS Cost and Usage reports][AWS-billing], making these reports exposed as a database table.
To read more details on how the different ReportDataSources work, read the [metering architecture document][architecture].

## Fields

- `promsum`: If this section is present, then the `ReportDataSource` will be configured to periodically poll Prometheus for metrics using the specified `ReportPrometheusQuery`.
 - `query`: The name of the `ReportPrometheusQuery` resource.
 - `storage`: This section controls the `StorageLocation` options, allowing you to control on a per ReportDataSource level, where data is stored.
   - `storageLocationName`: The name of the `StorageLocation` resource to use.
   - `spec`: If `storageLocationName` is not set, then this section is used to control the storage location settings. See the [StorageLocation documentation][storage-locations] for details on what can be specified here. Anything valid in a `StorageLocation`'s `spec` is valid here.
- `awsBilling`:
  - `source`:
    - `bucket`: Bucket name to store data into.
    - `prefix`: Path within the bucket where to store data.

## Example ReportDataSource

Below is an example of one of the built-in `ReportDataSource` resources that is installed with Operator Metering by default.
This example doesn't specify a `spec.storage` section, which will result in it using the [default StorageLocation resource][default-storage-location].

```
apiVersion: chargeback.coreos.com/v1alpha1
kind: ReportDataSource
metadata:
  name: "pod-request-memory-bytes"
  labels:
    operator-metering: "true"
spec:
  promsum:
    query: "pod-request-memory-bytes"
```

This example is a slightly modified version of the one above that does set the `storage` to the `StorageLocation` with a `metadata.name` of "local".

```
apiVersion: chargeback.coreos.com/v1alpha1
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

[storage-locations]: storagelocations.md
[AWS-billing]: https://docs.aws.amazon.com/awsaccountbilling/latest/aboutv2/billing-reports-costusage.html
[metering-aws-billing-conf]: metering-config.md#aws-billing-correlation
[default-storage-location]: storagelocations.md#default-storagelocation
[architecture]: metering-architecture.md
