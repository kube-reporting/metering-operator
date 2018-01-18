<br>
<div class="alert alert-info" role="alert">
    <i class="fa fa-exclamation-triangle"></i><b> Note:</b> This documentation is for a pre-alpha feature. To register for the Chargeback Alpha program, email <a href="mailto:tectonic-alpha-feedback@coreos.com">tectonic-alpha-feedback@coreos.com</a>.
</div>

# Chargeback Con***REMOVED***guration

Chargeback supports a few con***REMOVED***guration options which can be set by creating a secret named `chargeback-settings` with a key of `values.yaml`.

An example con***REMOVED***guration ***REMOVED***le can be found in [manifests/chargeback-con***REMOVED***g/custom-values.yaml][example-con***REMOVED***g] For details on customizing this ***REMOVED***le, read the [common-con***REMOVED***guration-options](#common-con***REMOVED***guration-options) section below.

## Documentation conventions

This document follows the convention of describing nested ***REMOVED***elds in con***REMOVED***guration settings using dots as separators. For example,

```
chargeback-operator.con***REMOVED***g.awsAccessKeyID
```

refers to the following YAML structure and value:

```
chargeback-operator:
  con***REMOVED***g:
    awsAccessKeyID: "REPLACEME"
```

## Using a custom con***REMOVED***guration

To install the custom con***REMOVED***guration ***REMOVED***le, run the following command:

```
kubectl -n $CHARGEBACK_NAMESPACE create secret generic chargeback-settings --from-***REMOVED***le 'values.yaml=manifests/chargeback-con***REMOVED***g/custom-values.yaml'
```

This command requires that the secret be deleted and recreated for each installation.

Use `kubectl apply` to avoid this requirement:

```
kubectl -n $CHARGEBACK_NAMESPACE create secret generic chargeback-settings --from-***REMOVED***le 'values.yaml=manifests/chargeback-con***REMOVED***g/custom-values.yaml' -o yaml --dry-run > /tmp/chargeback-settings-secret.yaml && kubectl apply -n $CHARGEBACK_NAMESPACE -f /tmp/chargeback-settings-secret.yaml
```

## Common con***REMOVED***guration options

The example manifest [manifests/chargeback-con***REMOVED***g/custom-values.yaml][example-con***REMOVED***g] contains the most common user con***REMOVED***guration values, including resource limits. The values listed are the defaults, which may be uncommented and adjusted as needed.

### Persistent Volumes

Chargeback requires at least 1 Persistent Volume to operate. (The example manifest includes 3 by default.) The Persistent Volume Claims (PVCs) are listed below:

- `hive-metastore-db-data` is the only _required_ volume. It is used by
  hive metastore to retain information about the location of Presto data.
- `hdfs-namenode-data-hdfs-namenode-0` and `hdfs-datanode-data-hdfs-datanode-$i`
   are used by the single node HDFS cluster which is deployed by default for
   storing data within the cluster. These two PVCs are not required to [store data in s3](#storing-data-in-s3).

Each of these Persistent Volume Claims is created dynamically by a Stateful Set. Enabling this requires that dynamic volume provisioning be enabled via a Storage Class, or persistent volumes of the correct size must be manually pre-created.

### Dynamically provisioning Persistent Volumes using Storage Classes

Storage Classes may be used when dynamically provisioning Persistent Volume Claims using a Stateful Set. Use `kubectl get` to determine if Storage Classes have been created in your cluster. (By default, Tectonic does not install cloud provider speci***REMOVED***c
Storage Classes.)

```
$ kubectl get storageclasses
```

If the output includes `(default)` next to the `name` of any `StorageClass`, then that `StorageClass` is the default for the cluster. The default is used when `StorageClass` is unspeci***REMOVED***ed or set to `null` in a `PersistentVolumeClaim` spec.

If no `StorageClass` is listed, or if you wish to use a non-default `StorageClass`, see [Con***REMOVED***guring the StorageClass for Chargeback](#con***REMOVED***guring-the-storage-class-for-chargeback) below.

For more information, see [Storage Classes][storage-classes] in the Kubernetes documentation.

#### Con***REMOVED***guring the Storage Class for Chargeback

To con***REMOVED***gure and specify a `StorageClass` for use in Chargeback, specify the `StorageClass` in `custom-values.yaml`. A example `StorageClass` section is included in [manifests/chargeback-con***REMOVED***g/custom-storageclass-values.yaml][example-storage-con***REMOVED***g].

Uncomment the following sections and replace the `null` in `class: null` value with the name of the `StorageClass` to use. Leaving the value `null` will cause Chargeback to use the default StorageClass for the cluster.

- `presto.hive.metastore.storage.class`
- `hdfs.datanode.storage.class`
- `hdfs.namenode.storage.class`

### Manually creating Persistent Volumes

If a Storage Class that supports dynamic volume provisioning does not exist in the cluster, it is possible to manually create a Persistent Volume with the correct capacity. By default the PVCs listed above each request 5Gi of storage. This can be adjusted in the same section as adjusting the Storage Class.

Use [manifests/chargeback-con***REMOVED***g/custom-storageclass-values.yaml][example-storage-con***REMOVED***g] as a template and adjust the `size: "5Gi"` value to the desired capacity for the following sections:

- `presto.hive.metastore.storage.size`
- `hdfs.datanode.storage.size`
- `hdfs.namenode.storage.size`

### Storing data in S3

By default the data that Chargeback collects and generates is stored in a single node HDFS cluster which is backed by a Persistent Volume. To store the data in a location outside of the cluster, con***REMOVED***gure Chargeback to store data in S3.

To use S3 for storage, uncomment the `defaultStorage:` section in the example
[manifests/chargeback-con***REMOVED***g/custom-values.yaml][example-con***REMOVED***g] con***REMOVED***guration.
Once uncommented, set `awsAccessKeyID` and `awsSecretAccessKey` in the `chargeback-operator.con***REMOVED***g` and `presto.con***REMOVED***g` sections.

Please note that this must be done before installation. Changing these settings after installation may result in unexpected behavior.

Because the deployed HDFS cluster will not be used to store data, it may also be disabled. Uncomment the `hdfs.enabled: true` setting in `custom-values.yaml`, and set the
value to `false`.

```
hdfs:
  enabled: false
```

### AWS billing correlation

Chargeback is able to correlate cluster usage information with [AWS detailed billing information][AWS-billing], attaching a dollar amount to resource usage. For clusters running in EC2, this can be enabled by modifying the example [manifests/chargeback-con***REMOVED***g/custom-values.yaml][example-con***REMOVED***g] con***REMOVED***guration.

To enable AWS billing correlation, ***REMOVED***rst ensure the AWS Cost and Usage Reports
are enabled. For more information, see [Turning on the AWS Cost and Usage report][enable-aws-billing] in the AWS documentation.

Next, update the `awsBillingDataSource` section in the [custom-values.yaml][example-con***REMOVED***g] example con***REMOVED***guration manifest.

Change the `enabled` value to `true`, and update the `bucket` and `pre***REMOVED***x` to the location of your AWS Detailed billing report.  

Then, set the `awsAccessKeyID` and `awsSecretAccessKey` in the `chargeback-operator.con***REMOVED***g` and `presto.con***REMOVED***g` sections.

This can be done either pre-install or post-install. Note that disabling it post-install can cause errors in the chargeback-operator.

[AWS-billing]: https://docs.aws.amazon.com/awsaccountbilling/latest/aboutv2/billing-reports-costusage.html
[enable-aws-billing]: https://docs.aws.amazon.com/awsaccountbilling/latest/aboutv2/billing-reports-gettingstarted-turnonreports.html
[example-con***REMOVED***g]: ../manifests/chargeback-con***REMOVED***g/custom-values.yaml
[example-storage-con***REMOVED***g]: ../manifests/chargeback-con***REMOVED***g/custom-storageclass-values.yaml
[storage-classes]: https://kubernetes.io/docs/concepts/storage/storage-classes/
