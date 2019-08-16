# Storage Locations

A `StorageLocation` is a custom resource that con***REMOVED***gures where data will be stored by the reporting-operator.
This includes the data collected from Prometheus, and the results produced by generating a `Report` custom resource.

Normally, users shouldn't need to con***REMOVED***gure a `StorageLocation` resource unless they want to store data in multiple locations, like multiple S3 buckets or both S3 and HDFS, or if they wish to access a database in Hive/Presto that was not created by metering.
Instead, users should use the [con***REMOVED***guring storage](con***REMOVED***guring-storage.md) documentation to manage con***REMOVED***guration of all components in the metering stack.

The Operator Metering default installation provides a few ways of con***REMOVED***guring the [Default StorageLocation](#default-storagelocation), and normally it shouldn't be necessary to create these directly.
Refer to the [Metering Con***REMOVED***guration doc](metering-con***REMOVED***g.md#storing-data-in-s3) for details on using the `Metering` resource to set your default StorageLocation.

## Fields
- `hive`: If this section is present, then the `StorageLocation` will be con***REMOVED***gured to store data in Presto by creating the table using Hive server. Note: only `databaseName` and `unmanagedDatabase` are required ***REMOVED***elds.
  - `databaseName`: The name of the database within hive.
  - `unmanagedDatabase`: If true, then this StorageLocation will not be actively managed, and the databaseName is expected to already exist in Hive. If false, this will cause the reporting-operator to create the Database in Hive.
  - `location`: Optional: The ***REMOVED***lesystem URL for Presto and Hive to use for the database. This can be an `hdfs://` or `s3a://` ***REMOVED***lesystem URL.
  - `defaultTableProperties`: Optional: Contains con***REMOVED***guration options for creating tables using Hive.
    - `***REMOVED***leFormat`: Optional: The ***REMOVED***le format used for storing ***REMOVED***les in the ***REMOVED***lesystem. See the [Hive Documentation on File Storage Format for a list of options and more details][hiveFileFormat].
    - `rowFormat`: Optional: Controls the [Hive row format][hiveRowFormat]. This controls how Hive serializes and deserializes rows. See the [Hive Documentation on Row Formats & SerDe for more details][hiveRowFormat].

## Example StorageLocation

This ***REMOVED***rst example is what the built-in local storage option looks like.
As you can see, it's con***REMOVED***gured to use Hive, and by default data is stored wherever Hive is con***REMOVED***gured to use storage by default (HDFS, S3, or a ReadWriteMany PVC) since the location isn't set.

```yaml
apiVersion: metering.openshift.io/v1
kind: StorageLocation
metadata:
  name: hive
  labels:
    operator-metering: "true"
spec:
  hive:
    databaseName: metering
    unmanagedDatabase: false
```

The example below uses an AWS S3 bucket for storage.
The pre***REMOVED***x is appended to the bucket name when constructing the path to use.

```yaml
apiVersion: metering.openshift.io/v1
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

If an annotation `storagelocation.metering.openshift.io/is-default` exists and is set to the string "true" on a `StorageLocation` resource, then that resource will be used if a `StorageLocation` is not speci***REMOVED***ed on resources which have a `storage` con***REMOVED***guration option.
If more than one resource with the annotation exists, an error will be logged and the operator will consider there to be no default.

```yaml
apiVersion: metering.openshift.io/v1
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
