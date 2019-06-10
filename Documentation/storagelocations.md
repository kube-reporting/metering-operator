# Storage Locations

A `StorageLocation` is a custom resource that configures where data will be stored by the reporting-operator.
This includes the data collected from Prometheus, and the results produced by generating a `Report`.

Normally, users shouldn't need to configure StorageLocation's unless they want to store data in multiple locations, like multiple S3 buckets or both S3 and HDFS, or if they wish to access a database in Hive/Presto that was not created by metering.
Instead, users should use the [configuring storage](configuring-storage.md) documentation to manage configuration of all components in the metering stack.

The Operator Metering default installation provides a few ways of configuring the [Default StorageLocation](#default-storagelocation), and normally it shouldn't be necessary to create these directly.
Refer to the [Metering Configuration doc](metering-config.md#storing-data-in-s3) for details on using the `Metering` resource to set your default StorageLocation.

## Fields

- `hive`: If this section is present, then the `StorageLocation` will be configured to store data in Presto by creating the table using Hive server.
  - `databaseName`: The name of the database within hive.
  - `unmanagedDatabase`: If true, then this StorageLocation will not be actively managed, and the databaseName is expected to already exist in Hive. If false, this will cause the reporting-operator to create the Database in Hive.
  - `location`: The filesystem URL for Presto and Hive to use for the database. This can be an `hdfs://` or `s3a://` filesystem URL.
  - 'defaultTableProperties': Optional: Contains configuration options for creating tables using Hive.
    - `fileFormat`: Optional: The file format used for storing files in the filesystem. See the [Hive Documentation on File Storage Format for a list of options and more details][hiveFileFormat].
    - `rowFormat`: Optional: Controls the [Hive row format][hiveRowFormat]. This controls how Hive serializes and deserializes rows. See the [Hive Documentation on Row Formats & SerDe for more details][hiveRowFormat].

## Example StorageLocation

This first example is what the built-in local storage option looks like.
As you can see, it's configured to use Hive, and by default data is stored wherever Hive is configured to use storage by default (HDFS, S3, or a ReadWriteMany PVC) since the location isn't set.

```yaml
apiVersion: metering.openshift.io/v1alpha1
kind: StorageLocation
metadata:
  name: hive
  labels:
    operator-metering: "true"
  spec:
    hive:
      databaseName: metering
      unmanagedDatabase: false
      location: ""
```

The example below uses an AWS S3 bucket for storage.
The prefix is appended to the bucket name when constructing the path to use.

```yaml
apiVersion: metering.openshift.io/v1alpha1
kind: StorageLocation
metadata:
  name: example-s3-storage
  labels:
    operator-metering: "true"
  spec:
    hive:
      databaseName: example_s3_storage
      unmanagedDatabase: false
      location: "s3a://bucket-name/path/within/bucket"
```

## Default StorageLocation

If an annotation `storagelocation.metering.openshift.io/is-default` exists and is set to the string "true" on a `StorageLocation` resource, then that resource will be used if a `StorageLocation` is not specified on resources which have a `storage` configuration option.
If more than one resource with the annotation exists, an error will be logged and the operator will consider there to be no default.

```yaml
apiVersion: metering.openshift.io/v1alpha1
kind: StorageLocation
metadata:
  name: example-s3-storage
  labels:
    operator-metering: "true"
  annotations:
    storagelocation.metering.openshift.io/is-default: "true"
  spec:
    hive:
      databaseName: example_s3_storage
      unmanagedDatabase: false
      location: "s3a://bucket-name/path/within/bucket"
```

[hiveFileFormat]: https://cwiki.apache.org/confluence/display/Hive/LanguageManual+DDL#LanguageManualDDL-StorageFormatsStorageFormatsRowFormat,StorageFormat,andSerDe
[hiveRowFormat]: https://cwiki.apache.org/confluence/display/Hive/LanguageManual+DDL#LanguageManualDDL-RowFormats&SerDe
