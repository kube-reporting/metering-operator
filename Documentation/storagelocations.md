# Storage Locations

A `StorageLocation` is a custom resource that configures where data will be stored.
This includes the data collected from Prometheus, and the results produced by generating a `Report` or `ScheduledReport`.

The Operator Metering default installation provides a few ways of configuring the [Default StorageLocation](#default-storagelocation), and normally it shouldn't be necessary to create these directly.
Refer to the [Metering Configuration doc](metering-config.md#storing-data-in-s3) for details on using the `Metering` resource to set your default StorageLocation.

## Fields

- `hive`: If this section is present, then the `StorageLocation` will be configured to store data in Presto by creating the table using Hive server.
  - 'tableProperties': Contains configuration options for creating tables using Hive.
    - `location`: The filesystem URL for Presto and Hive to use. This can be an `hdfs://` or `s3a://` filesystem URL.
    - `fileFormat`: The format used for storing files in the filesystem. See the [Hive Documentation on File Storage Format for a list of options and more details][hiveFileFormat].
    - `serdeFormat`: The [SerDe][hiveSerde] class for Hive to use to serialize and deserialize rows when fileFormat is `TEXTFILE`. See the [Hive Documentation on Row Formats & SerDe for more details][hiveSerdeFormat].
    - `serdeRowProperties`: Additional properties used to configure `serdeFormat`. See the [Hive Documentation on Row Formats & SerDe for more details][hiveSerdeFormat].
    - `external`: If specified, configures the table as an external table with existing data. If specified `location` is required. When tables using this storage are dropped, the contents are not deleted. See the [Hive documentation on External tables for more information][hiveExternalTables].

## Example StorageLocation

This first example is what the built-in local storage option looks like.
As you can see, it's configured to use HDFS and supplies no additional options.

```yaml
apiVersion: chargeback.coreos.com/v1alpha1
kind: StorageLocation
metadata:
  name: local
  labels:
    operator-metering: "true"
  spec:
    hive:
      location: "hdfs://hdfs-namenode-proxy:8020"
```

The example below uses an AWS S3 bucket for storage.
The prefix is appended to the bucket name when constructing the path to use.

```yaml
apiVersion: chargeback.coreos.com/v1alpha1
kind: StorageLocation
metadata:
  name: example-s3-storage
  labels:
    operator-metering: "true"
  spec:
    hive:
      location: "s3a://bucket-name/path/within/bucket"
```

## Default StorageLocation

If an annotation `storagelocation.chargeback.coreos.com/is-default` exists and is set to the string "true" on a `StorageLocation` resource, then that resource will be used if a `StorageLocation` is not specified on resources which have a `storage` configuration option.
If more than one resource with the annotation exists, an error will be logged and the operator will consider there to be no default.

```yaml
apiVersion: chargeback.coreos.com/v1alpha1
kind: StorageLocation
metadata:
  name: local
  labels:
    operator-metering: "true"
  annotations:
    storagelocation.chargeback.coreos.com/is-default: "true"
  spec:
    hive:
      location: "s3a://bucket-name/path/within/bucket"
```

[hiveFileFormat]: https://cwiki.apache.org/confluence/display/Hive/LanguageManual+DDL#LanguageManualDDL-StorageFormatsStorageFormatsRowFormat,StorageFormat,andSerDe
[hiveSerdeFormat]: https://cwiki.apache.org/confluence/display/Hive/LanguageManual+DDL#LanguageManualDDL-RowFormats&SerDe
[hiveSerde]: https://cwiki.apache.org/confluence/display/Hive/SerDe
[hiveExternalTables]: https://cwiki.apache.org/confluence/display/Hive/LanguageManual+DDL#LanguageManualDDL-ExternalTables
