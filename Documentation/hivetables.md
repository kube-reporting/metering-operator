# Hive Tables

A `HiveTable` is a custom resource that represents a database table within Hive.

When created, a HiveTable resource causes the reporting-operator to create a table within Hive according to the configuration provided.

## Fields

- `databaseName`: The name of the Hive database to use. This generally should be `default` or the value of the databaseName in a `Hive` [StorageLocation][storage-locations].
- `tableName`: The name of the table to create in Hive.
- `columns`: A list of columns that match the schema of the of the HiveTable. Columns in `partitionedBy` and `columns` must not overlap.
  - `name`: The name of the column.
  - `type`: The column data type. [See the Hive Language Manual section on types for more details][hiveTypes]. Currently the only complex types supported are map's of primitive types.
- `partitionedBy`: A list of columns that are used as partition columns. Columns in `partitionedBy` and `columns` must not overlap.
  - `name`: The name of the column.
  - `type`: The column data type. [See the Hive Language Manual section on types for more details][hiveTypes]. Currently the only complex types supported are map's of primitive types.
- `clusteredBy`: A list of columns from `columns` to use for [bucketed tables][hiveBucketedTables]. Must set `numBuckets` if specified.
- `sortedBy`: A list of column names from `columns` to use for [bucketed tables][hiveBucketedTables]. Must set `clusteredBy` and `numBuckets` if specified.
  - `name`: The name of the column from `columns`.
  - `descending`: Optional: if true, the column is descending, if false, it's ascending. If unspecified, it defaults to the hive default behavior.
- `numBuckets`: The number of buckets to create for a [bucketed table][hiveBucketedTables]. Must set `clusteredBy` if set.
- `location`: Optional: Specifies the HDFS path to store this table in. Can be any URI supported by Hive. Currently supports `s3a://`, `hdfs://` and `/local/path` based URIs.
- `rowFormat`: Controls the [Hive row format][hiveRowFormat]. This controls how Hive serializes and deserializes rows. See the [Hive Documentation on Row Formats & SerDe for more details][hiveRowFormat].
- `fileFormat`: The file format used for storing files in the filesystem. See the [Hive Documentation on File Storage Format for a list of options and more details][hiveFileFormat].
- `tableProperties`: Allows tagging the table definition with your own key/value metadata. Some predefined properties exist to control behavior of the table as well. See the [Hive table properties documentation][hiveTableProperties] for details.
- `external`: If true, creates an external table instead of a managed table, causing Hive to point at an existing location as specified by `location` where data lives. See the [Hive external tables documentation][hiveExternalTable] for details. Location must be specified if `external` is true.
- `managePartitions`: If true, configures the reporting-operator ensure the Table partitions match those specified in `partitions`.
- `partitions`: A list of partitions that this table should have. Only valid if `managePartitions` is true.
  - `partitionSpec`: A map of string keys and string values where each key is expected to be the name of a partition column, and the value is the value of the partition column.
  - `location`: Specifies where the data for this partition is stored. This should be a sub-directory of `location`.

## Example HiveTables

```
apiVersion: metering.openshift.io/v1alpha1
kind: HiveTable
metadata:
  name: apache-log
  annotations:
    reference: "based on the RegEx example from https://cwiki.apache.org/confluence/display/Hive/LanguageManual+DDL#LanguageManualDDL-RowFormats&SerDe"
spec:
  databaseName: default
  tableName: apache_log
  # bucket containing apache log files
  location: s3a://my-bucket/apache_logs
  columns:
  - name: host
    type: string
  - name: identity
    type: string
  - name: user
    type: string
  - name: time
    type: string
  - name: request
    type: string
  - name: status
    type: string
  - name: size
    type: string
  - name: referer
    type: string
  - name: agent
    type: string
  rowFormat: |
    SERDE 'org.apache.hadoop.hive.serde2.RegexSerDe'
    WITH SERDEPROPERTIES (
      "input.regex" = "([^ ]*) ([^ ]*) ([^ ]*) (-|\\[[^\\]]*\\]) ([^ \"]*|\"[^\"]*\") (-|[0-9]*) (-|[0-9]*)(?: ([^ \"]*|\"[^\"]*\") ([^ \"]*|\"[^\"]*\"))?"
    )
  fileFormat: TEXTFILE
  external: true
```

[storage-locations]: storagelocations.md
[hiveFileFormat]: https://cwiki.apache.org/confluence/display/Hive/LanguageManual+DDL#LanguageManualDDL-StorageFormatsStorageFormatsRowFormat,StorageFormat,andSerDe
[hiveRowFormat]: https://cwiki.apache.org/confluence/display/Hive/LanguageManual+DDL#LanguageManualDDL-RowFormats&SerDe
[hiveBucketedTables]: https://cwiki.apache.org/confluence/display/Hive/LanguageManual+DDL+BucketedTables
[hiveTypes]: https://cwiki.apache.org/confluence/display/Hive/LanguageManual+Types
[hiveTableProperties]: https://cwiki.apache.org/confluence/display/Hive/LanguageManual+DDL#LanguageManualDDL-listTableProperties
[hiveExternalTable]: https://cwiki.apache.org/confluence/display/Hive/LanguageManual+DDL#LanguageManualDDL-ExternalTables
