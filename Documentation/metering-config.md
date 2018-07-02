# Metering Configuration

Metering supports configuration options which may be set in the `spec` section of the `Metering` resource.

An example configuration file can be found in [custom-values.yaml][example-config].
A minimal configuration example that doesn't override anything can be found in [default.yaml][default-config].
For details on customizing these files, read the [common-configuration-options](#common-configuration-options) section below.

## Documentation conventions

This document follows the convention of describing nested fields in configuration settings using dots as separators. For example,

```
spec.metering-operator.config.awsAccessKeyID
```

refers to the following YAML structure and value:

```
spec:
  metering-operator:
    config:
      awsAccessKeyID: "REPLACEME"
```

## Using a custom configuration

To install the custom configuration file, run the following command:

```
kubectl -n $METERING_NAMESPACE apply -f manifests/metering-config/custom-values.yaml
```

## Common configuration options

The example manifest [custom-values.yaml][example-config] contains the most common user configuration values, including resource limits. The values listed are the defaults, which may be uncommented and adjusted as needed.

### Prometheus URL

By default, the Metering assumes that your Prometheus service is available at `http://prometheus-k8s.monitoring.svc:9090` within the cluster.
If your not using [kube-prometheus][kube-prometheus], then you will need to override the `metering-operator.config.prometheusURL` configuration option.

Below is an example of configuring Metering to use the service `prometheus` on port 9090 in the `cluster-monitoring` namespace:

```
spec:
  metering-operator:
    config:
      prometheusURL: "http://prometheus.cluster-monitoring.svc:9090"
```

> Note: currently we do not support https connections or authentication to Prometheus, but support for it is being developed.

### Use MySQL or Postgresql for Hive Metastore database

By default to make installation easier Metering configures Hive to use an embedded Java database called [Derby](https://db.apache.org/derby/#What+is+Apache+Derby%3F), however this is unsuited for larger environments or metering installations with a lot of reports and metrics being collected.
Currently two alternative options are available, MySQL and Postgresql, both of which have been tested with operator metering.

There are 4 configuration options you can use to control the database used by Hive metastore: `dbConnectionURL` , `dbConnectionDriver` , `dbConnectionUsername` , and `dbConnectionPassword`.

Using MySQL:

```
spec:
  presto:
    hive:
      config:
        dbConnectionURL: "jdbc:mysql://mysql.example.com:3306/hive_metastore"
        dbConnectionDriver: "com.mysql.jdbc.Driver"
        dbConnectionUsername: "REPLACEME"
        dbConnectionPassword: "REPLACEME"
```

You can pass additional JDBC parameters using the `dbConnectionURL`, for more details see [the MySQL Connector/J documentation](https://dev.mysql.com/doc/connector-j/5.1/en/connector-j-reference-configuration-properties.html).

Using Postgresql:

```
spec:
  presto:
    hive:
      config:
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

Storage Classes may be used when dynamically provisioning Persistent Volume Claims using a Stateful Set. Use `kubectl get` to determine if Storage Classes have been created in your cluster. (By default, Tectonic does not install cloud provider specific
Storage Classes.)

```
$ kubectl get storageclasses
```

If the output includes `(default)` next to the `name` of any `StorageClass`, then that `StorageClass` is the default for the cluster. The default is used when `StorageClass` is unspecified or set to `null` in a `PersistentVolumeClaim` spec.

If no `StorageClass` is listed, or if you wish to use a non-default `StorageClass`, see [Configuring the StorageClass for Metering](#configuring-the-storage-class-for-metering) below.

For more information, see [Storage Classes][storage-classes] in the Kubernetes documentation.

#### Configuring the Storage Class for Metering

To configure and specify a `StorageClass` for use in Metering, specify the `StorageClass` in `custom-values.yaml`. A example `StorageClass` section is included in [custom-storageclass-values.yaml][example-storage-config].

Uncomment the following sections and replace the `null` in `class: null` value with the name of the `StorageClass` to use. Leaving the value `null` will cause Metering to use the default StorageClass for the cluster.

- `spec.presto.hive.metastore.storage.class`
- `spec.hdfs.datanode.storage.class`
- `spec.hdfs.namenode.storage.class`

### Manually creating Persistent Volumes

If a Storage Class that supports dynamic volume provisioning does not exist in the cluster, it is possible to manually create a Persistent Volume with the correct capacity. By default, the PVCs listed above each request 5Gi of storage. This can be adjusted in the same section as adjusting the Storage Class.

Use [custom-storageclass-values.yaml][example-storage-config] as a template and adjust the `size: "5Gi"` value to the desired capacity for the following sections:

- `presto.hive.metastore.storage.size`
- `hdfs.datanode.storage.size`
- `hdfs.namenode.storage.size`

### Storing data in S3

By default, the data that Metering collects and generates is stored in a single node HDFS cluster which is backed by a Persistent Volume. To store the data in a location outside of the cluster, configure Metering to store data in S3.

To use S3 for storage, uncomment the `defaultStorage:` section in the example
[custom-values.yaml][example-config] configuration.
Once uncommented, set `awsAccessKeyID` and `awsSecretAccessKey` in the `metering-operator.config` and `presto.config` sections.

To store data in S3, the `awsAccessKeyID` and `awsSecretAccessKey` credentials must have read and write access to the bucket.
For an example of an IAM policy granting the required permissions see the [aws/read-write.json][aws/read-write.json] file.
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

Metering is able to correlate cluster usage information with [AWS detailed billing information][AWS-billing], attaching a dollar amount to resource usage. For clusters running in EC2, this can be enabled by modifying the example [custom-values.yaml][example-config] configuration.

To enable AWS billing correlation, first ensure the AWS Cost and Usage Reports
are enabled. For more information, see [Turning on the AWS Cost and Usage report][enable-aws-billing] in the AWS documentation.

Next, update the `awsBillingDataSource` section in the [custom-values.yaml][example-config] example configuration manifest.

Change the `enabled` value to `true`, and update the `bucket` and `prefix` to the location of your AWS Detailed billing report.

Then, set the `awsAccessKeyID` and `awsSecretAccessKey` in the `spec.metering-operator.config` and `spec.presto.config` sections.

To retrieve data in S3, the `awsAccessKeyID` and `awsSecretAccessKey` credentials must have read access to the bucket.
For an example of an IAM policy granting the required permissions see the [aws/read-only.json][aws/read-only.json] file.
Replace `operator-metering-data` with the name of your bucket.

This can be done either pre-install or post-install. Note that disabling it post-install can cause errors in the metering-operator.


[AWS-billing]: https://docs.aws.amazon.com/awsaccountbilling/latest/aboutv2/billing-reports-costusage.html
[enable-aws-billing]: https://docs.aws.amazon.com/awsaccountbilling/latest/aboutv2/billing-reports-gettingstarted-turnonreports.html
[example-config]: ../manifests/metering-config/custom-values.yaml
[default-config]: ../manifests/metering-config/default.yaml
[example-storage-config]: ../manifests/metering-config/custom-storageclass-values.yaml
[storage-classes]: https://kubernetes.io/docs/concepts/storage/storage-classes/
[kube-prometheus]: https://github.com/coreos/prometheus-operator/tree/master/contrib/kube-prometheus
