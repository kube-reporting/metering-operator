# Storage Locations

A `StorageLocation` is a custom resource that configures where data will be stored.
This includes the data collected from Prometheus, and the results produced by generating a `Report` or `ScheduledReport`.

The Operator Metering default installation provides a few ways of configuring the [Default StorageLocation](#default-storagelocation), and normally it shouldn't be necessary to create these directly.
Refer to the [Metering Configuration doc](metering-config.md#storing-data-in-s3) for details on using the `Metering` resource to set your default StorageLocation.

## Fields

- `s3`: If this section is present, then the `StorageLocation` will be configured to store data into an AWS S3 bucket.
  - `bucket`: The name of the S3 bucket to store the data in.
  - `prefix`: The path within the bucket to store the data in.
- `local`: If this section is present, then the `StorageLocation` will be configured to store data based on what the operator defines as "local". Currently local storage is defined to be HDFS, and redefining what "local" maps to is not exposed as a configuration option, but is planned in the future. See the examples below for how to specify it.

## Example StorageLocation

This first example is what the built-in local storage option looks like.
As you can see, it takes no options, so you must specify it using an empty object `{}`.

```yaml
apiVersion: chargeback.coreos.com/v1alpha1
kind: StorageLocation
metadata:
  name: local
  labels:
    operator-metering: "true"
  spec:
    local: {}
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
    s3:
      bucket: bucket-name
      prefix: path/within/bucket/
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
    local: {}
```
