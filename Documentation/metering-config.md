# Metering Con***REMOVED***guration

Metering supports con***REMOVED***guration options which may be set in the `spec` section of the `Metering` resource.

An example con***REMOVED***guration ***REMOVED***le can be found in [custom-values.yaml][example-con***REMOVED***g].
A minimal con***REMOVED***guration example that doesn't override anything can be found in [default.yaml][default-con***REMOVED***g].
For details on customizing these ***REMOVED***les, read the [common-con***REMOVED***guration-options](#common-con***REMOVED***guration-options) section below.

## Documentation conventions

This document follows the convention of describing nested ***REMOVED***elds in con***REMOVED***guration settings using dots as separators. For example,

```
spec.reporting-operator.spec.con***REMOVED***g.awsAccessKeyID
```

refers to the following YAML structure and value:

```
spec:
  reporting-operator:
    spec:
      con***REMOVED***g:
        awsAccessKeyID: "REPLACEME"
```

## Using a custom con***REMOVED***guration

To install the custom con***REMOVED***guration ***REMOVED***le, run the following command:

```
kubectl -n $METERING_NAMESPACE apply -f manifests/metering-con***REMOVED***g/custom-values.yaml
```

## Common con***REMOVED***guration options

The example manifest [custom-values.yaml][example-con***REMOVED***g] contains the most common user con***REMOVED***guration values, including resource limits. The values listed are the defaults, which may be uncommented and adjusted as needed.

### Prometheus URL

By default, the Metering assumes that your Prometheus service is available at `http://prometheus-k8s.monitoring.svc:9090` within the cluster.
If your not using [kube-prometheus][kube-prometheus], then you will need to override the `reporting-operator.con***REMOVED***g.prometheusURL` con***REMOVED***guration option.

Below is an example of con***REMOVED***guring Metering to use the service `prometheus` on port 9090 in the `cluster-monitoring` namespace:

```
spec:
  reporting-operator:
    spec:
      con***REMOVED***g:
        prometheusURL: "http://prometheus.cluster-monitoring.svc:9090"
```

> Note: currently we do not support https connections or authentication to Prometheus, but support for it is being developed.

### Use MySQL or Postgresql for Hive Metastore database

By default to make installation easier Metering con***REMOVED***gures Hive to use an embedded Java database called [Derby](https://db.apache.org/derby/#What+is+Apache+Derby%3F), however this is unsuited for larger environments or metering installations with a lot of reports and metrics being collected.
Currently two alternative options are available, MySQL and Postgresql, both of which have been tested with operator metering.

There are 4 con***REMOVED***guration options you can use to control the database used by Hive metastore: `dbConnectionURL` , `dbConnectionDriver` , `dbConnectionUsername` , and `dbConnectionPassword`.

Using MySQL:

```
spec:
  presto:
    spec:
      hive:
        con***REMOVED***g:
          dbConnectionURL: "jdbc:mysql://mysql.example.com:3306/hive_metastore"
          dbConnectionDriver: "com.mysql.jdbc.Driver"
          dbConnectionUsername: "REPLACEME"
          dbConnectionPassword: "REPLACEME"
```

You can pass additional JDBC parameters using the `dbConnectionURL`, for more details see [the MySQL Connector/J documentation](https://dev.mysql.com/doc/connector-j/5.1/en/connector-j-reference-con***REMOVED***guration-properties.html).

Using Postgresql:

```
spec:
  presto:
    spec:
      hive:
        con***REMOVED***g:
          dbConnectionURL: "jdbc:postgresql://postgresql.example.com:5432/hive_metastore"
          dbConnectionDriver: "org.postgresql.Driver"
          dbConnectionUsername: "REPLACEME"
          dbConnectionPassword: "REPLACEME"
```

You can pass additional JDBC parameters using the `dbConnectionURL`, for more details see [the Postgresql JDBC driver documentation](https://jdbc.postgresql.org/documentation/head/connect.html#connection-parameters).

### Persistent Volumes

Metering requires at least 1 Persistent Volume to operate. (The example manifest includes 3 by default.) The Persistent Volume Claims (PVCs) are listed below:

- `hive-metastore-db-data` is the only _required_ volume. It is used by
  hive metastore to retain information about the location of Presto data.
- `hdfs-namenode-data-hdfs-namenode-0` and `hdfs-datanode-data-hdfs-datanode-$i`
   are used by the single node HDFS cluster which is deployed by default for
   storing data within the cluster. These two PVCs are not required to [store data in AWS S3](#storing-data-in-s3).

Each of these Persistent Volume Claims is created dynamically by a Stateful Set. Enabling this requires that dynamic volume provisioning be enabled via a Storage Class, or persistent volumes of the correct size must be manually pre-created.

### Dynamically provisioning Persistent Volumes using Storage Classes

Storage Classes may be used when dynamically provisioning Persistent Volume Claims using a Stateful Set. Use `kubectl get` to determine if Storage Classes have been created in your cluster. (By default, Tectonic does not install cloud provider speci***REMOVED***c
Storage Classes.)

```
$ kubectl get storageclasses
```

If the output includes `(default)` next to the `name` of any `StorageClass`, then that `StorageClass` is the default for the cluster. The default is used when `StorageClass` is unspeci***REMOVED***ed or set to `null` in a `PersistentVolumeClaim` spec.

If no `StorageClass` is listed, or if you wish to use a non-default `StorageClass`, see [Con***REMOVED***guring the StorageClass for Metering](#con***REMOVED***guring-the-storage-class-for-metering) below.

For more information, see [Storage Classes][storage-classes] in the Kubernetes documentation.

#### Con***REMOVED***guring the Storage Class for Metering

To con***REMOVED***gure and specify a `StorageClass` for use in Metering, specify the `StorageClass` in `custom-values.yaml`. A example `StorageClass` section is included in [custom-storageclass-values.yaml][example-storage-con***REMOVED***g].

Uncomment the following sections and replace the `null` in `class: null` value with the name of the `StorageClass` to use. Leaving the value `null` will cause Metering to use the default StorageClass for the cluster.

- `spec.presto.spec.hive.metastore.storage.class`
- `spec.hdfs.spec.datanode.storage.class`
- `spec.hdfs.spec.namenode.storage.class`

#### Con***REMOVED***guring the volume sizes for Metering

Use [custom-storageclass-values.yaml][example-storage-con***REMOVED***g] as a template and adjust the `size: "5Gi"` value to the desired capacity for the following sections:

- `presto.spec.hive.metastore.storage.size`
- `hdfs.spec.datanode.storage.size`
- `hdfs.spec.namenode.storage.size`

#### Manually creating Persistent Volumes

If a Storage Class that supports dynamic volume provisioning does not exist in the cluster, it is possible to manually create a Persistent Volume with the correct capacity. By default, the PVCs listed above each request 5Gi of storage. This can be adjusted in the same section as adjusting the Storage Class as documented in [Con***REMOVED***guring the volume sizes for Metering](#con***REMOVED***guring-the-volume-sizes-for-metering).

### Storing data in S3

By default, the data that Metering collects and generates is stored in a single node HDFS cluster which is backed by a Persistent Volume. To store the data in a location outside of the cluster, con***REMOVED***gure Metering to store data in S3.

To use S3 for storage, uncomment the `defaultStorage:` section in the example
[custom-values.yaml][example-con***REMOVED***g] con***REMOVED***guration.
Once uncommented, set `awsAccessKeyID` and `awsSecretAccessKey` in the `reporting-operator.con***REMOVED***g` and `presto.con***REMOVED***g` sections.

To store data in S3, the `awsAccessKeyID` and `awsSecretAccessKey` credentials must have read and write access to the bucket.
For an example of an IAM policy granting the required permissions see the [aws/read-write.json](aws/read-write.json) ***REMOVED***le.
Replace `operator-metering-data` with the name of your bucket.

Please note that this must be done before installation. Changing these settings after installation may result in unexpected behavior.

Because the deployed HDFS cluster will not be used to store data, it may also be disabled. Uncomment the `hdfs.enabled: true` setting in `custom-values.yaml`, and set the
value to `false`.

```
spec:
  hdfs:
    enabled: false
```

### AWS billing correlation

Metering is able to correlate cluster usage information with [AWS detailed billing information][AWS-billing], attaching a dollar amount to resource usage. For clusters running in EC2, this can be enabled by modifying the example [custom-values.yaml][example-con***REMOVED***g] con***REMOVED***guration.

To enable AWS billing correlation, ***REMOVED***rst ensure the AWS Cost and Usage Reports
are enabled. For more information, see [Turning on the AWS Cost and Usage report][enable-aws-billing] in the AWS documentation.

Next, update the `defaultReportDataSources.aws-billing` section in the [custom-values.yaml][example-con***REMOVED***g] example con***REMOVED***guration manifest.

Uncomment the entire `defaultReportDataSources` block , and update the `bucket`, `pre***REMOVED***x` and `region` to the location of your AWS Detailed billing report.

Then, set the `awsAccessKeyID` and `awsSecretAccessKey` in the `spec.reporting-operator.spec.con***REMOVED***g` and `spec.presto.spec.con***REMOVED***g` sections.

To retrieve data in S3, the `awsAccessKeyID` and `awsSecretAccessKey` credentials must have read access to the bucket.
For an example of an IAM policy granting the required permissions see the [aws/read-only.json](aws/read-only.json) ***REMOVED***le.
Replace `operator-metering-data` with the name of your bucket.

This can be done either pre-install or post-install. Note that disabling it post-install can cause errors in the reporting-operator.

[AWS-billing]: https://docs.aws.amazon.com/awsaccountbilling/latest/aboutv2/billing-reports-costusage.html
[enable-aws-billing]: https://docs.aws.amazon.com/awsaccountbilling/latest/aboutv2/billing-reports-gettingstarted-turnonreports.html
[example-con***REMOVED***g]: ../manifests/metering-con***REMOVED***g/custom-values.yaml
[default-con***REMOVED***g]: ../manifests/metering-con***REMOVED***g/default.yaml
[example-storage-con***REMOVED***g]: ../manifests/metering-con***REMOVED***g/custom-storageclass-values.yaml
[storage-classes]: https://kubernetes.io/docs/concepts/storage/storage-classes/
[kube-prometheus]: https://github.com/coreos/prometheus-operator/tree/master/contrib/kube-prometheus
