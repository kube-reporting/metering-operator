# Storage Locations

A `StorageLocation` is a custom resource that con***REMOVED***gures where data will be stored.
This includes the data collected from Prometheus, and the results produced by generating a `Report`.

The Operator Metering default installation provides a few ways of con***REMOVED***guring the [Default StorageLocation](#default-storagelocation), and normally it shouldn't be necessary to create these directly.
Refer to the [Metering Con***REMOVED***guration doc](metering-con***REMOVED***g.md#storing-data-in-s3) for details on using the `Metering` resource to set your default StorageLocation.

## Fields

- `hive`: If this section is present, then the `StorageLocation` will be con***REMOVED***gured to store data in Presto by creating the table using Hive server.
  - 'tableProperties': Contains con***REMOVED***guration options for creating tables using Hive.
    - `location`: The ***REMOVED***lesystem URL for Presto and Hive to use. This can be an `hdfs://` or `s3a://` ***REMOVED***lesystem URL.
    - `***REMOVED***leFormat`: The format used for storing ***REMOVED***les in the ***REMOVED***lesystem. See the [Hive Documentation on File Storage Format for a list of options and more details][hiveFileFormat].
    - `serdeFormat`: The [SerDe][hiveSerde] class for Hive to use to serialize and deserialize rows when ***REMOVED***leFormat is `TEXTFILE`. See the [Hive Documentation on Row Formats & SerDe for more details][hiveSerdeFormat].
    - `serdeRowProperties`: Additional properties used to con***REMOVED***gure `serdeFormat`. See the [Hive Documentation on Row Formats & SerDe for more details][hiveSerdeFormat].
    - `external`: If speci***REMOVED***ed, con***REMOVED***gures the table as an external table with existing data. If speci***REMOVED***ed `location` is required. When tables using this storage are dropped, the contents are not deleted. See the [Hive documentation on External tables for more information][hiveExternalTables].

## Example StorageLocation

This ***REMOVED***rst example is what the built-in local storage option looks like.
As you can see, it's con***REMOVED***gured to use HDFS and supplies no additional options.

```yaml
apiVersion: metering.openshift.io/v1alpha1
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
The pre***REMOVED***x is appended to the bucket name when constructing the path to use.

```yaml
apiVersion: metering.openshift.io/v1alpha1
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

If an annotation `storagelocation.metering.openshift.io/is-default` exists and is set to the string "true" on a `StorageLocation` resource, then that resource will be used if a `StorageLocation` is not speci***REMOVED***ed on resources which have a `storage` con***REMOVED***guration option.
If more than one resource with the annotation exists, an error will be logged and the operator will consider there to be no default.

```yaml
apiVersion: metering.openshift.io/v1alpha1
kind: StorageLocation
metadata:
  name: local
  labels:
    operator-metering: "true"
  annotations:
    storagelocation.metering.openshift.io/is-default: "true"
  spec:
    hive:
      location: "s3a://bucket-name/path/within/bucket"
```

[hiveFileFormat]: https://cwiki.apache.org/confluence/display/Hive/LanguageManual+DDL#LanguageManualDDL-StorageFormatsStorageFormatsRowFormat,StorageFormat,andSerDe
[hiveSerdeFormat]: https://cwiki.apache.org/confluence/display/Hive/LanguageManual+DDL#LanguageManualDDL-RowFormats&SerDe
[hiveSerde]: https://cwiki.apache.org/confluence/display/Hive/SerDe
[hiveExternalTables]: https://cwiki.apache.org/confluence/display/Hive/LanguageManual+DDL#LanguageManualDDL-ExternalTables
