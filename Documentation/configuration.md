## Configuration

Chargeback supports a few configuration options which can currently be set by
creating a secret named `chargeback-settings` with a key of `values.yaml`.

An example configuration file can be found in
[manifests/chargeback-config/custom-values.yaml][example-config]
For details on customizing this file, read the
[common-configuration-options](common-configuration-options) section below.

### Documentation Conventions

A common way of describing nested fields in `custom-values.yaml` in this
documentation will refer to setting configuration sections or values, by using
the following notation where dots are separators:

```
chargeback-operator.config.awsAccessKeyID
```

This is just a convention for referring to nested fields in the yaml, where the
above example is referring to the following yaml structure and value:

```
chargeback-operator:
  config:
    awsAccessKeyID: "REPLACEME"
```

If you ever see a reference to a setting using this dotted notation, just know
that it's referring to a nested field in a yaml structure.

### Using a custom configuration

To install the custom configuration file, you can run the following command:

```
kubectl -n $CHARGEBACK_NAMESPACE create secret generic chargeback-settings --from-file 'values.yaml=manifests/chargeback-config/custom-values.yaml'
```

However, this doesn't make it very easy to make changes without deleting and
re-creating the secret, so you can also use this one-liner which will do the
same thing, but will use `kubectl apply`

```
kubectl -n $CHARGEBACK_NAMESPACE create secret generic chargeback-settings --from-file 'values.yaml=manifests/chargeback-config/custom-values.yaml' -o yaml --dry-run > /tmp/chargeback-settings-secret.yaml && kubectl apply -n $CHARGEBACK_NAMESPACE -f /tmp/chargeback-settings-secret.yaml
```

## Common configuration options

The example manifest
[manifests/chargeback-config/custom-values.yaml][example-config]
contains has the most common configuration values a user might want to modify
included in it already, the options are commented out and can be uncommented
and adjusted as needed. The values are the defaults, and should work for most
users, but if you run into resource issues, you can increase the limits as
needed.

### StorageClasses and Volumes

Chargeback currently requires at minimum 1 (or 3, if using the defaults)
persistent volume(s) to operate. The PersistentVolumeClaims (PVC's) are listed below:

- `hive-metastore-db-data` is the only _required_ volume. It is used by
  hive metastore to retain information about where data for Presto is located.
- `hdfs-namenode-data-hdfs-namenode-0` and `hdfs-datanode-data-hdfs-datanode-$i`
   are used by the single node HDFS cluster which is deployed by default for
   storing data within the cluster. If your [storing data in s3](#storing-data-in-s3)
   then these aren't required.

Each of these PersistentVolumeClaims is created dynamically by a StatefulSet,
meaning you must have dynamic volume provisioning enabled via a StorageClass,
or you will need to manually pre-create persistent volumes of the correct size.

#### Using StorageClasses

As of Tectonic 1.8.4, Tectonic doesn't install cloud provider specific
StorageClasses by default, however it is possible an admin has created
StorageClasses in the cluster for your use. To check what StorageClasses are
available, run:

```
$ kubectl get storageclasses
```

If the output includes `(default)` next to the name of any StorageClasses, then
that StorageClass is the default for your cluster. The default is used when
StorageClass is unspecified or set to `null` in a PersistentVolumeClaim spec.

If you've got a default StorageClass installed, then you're ready to go and do
not need to do anything extra. However, if you wish to use the non-default
StorageClass, please continue reading. If you need to install a StorageClass,
please read the [official documentation on StorageClasses][storage-classes].

##### Configuring the StorageClass for Chargeback

You can configure and specify a StorageClass to use in chargeback by specifying
the StorageClass in your `custom-values.yaml`.
A trimmed down version of this file with just the relevant storage section is located in
[manifests/chargeback-config/custom-storageclass-values.yaml][example-storage-config].

Uncomment the sections and replace the `null` in `class: null` value
with the name of the StorageClass you want to use. If the value is null, that
will use the default StorageClass for your cluster. You will need to make this
change to the following sections:

- `presto.hive.metastore.storage.class`
- `hdfs.datanode.storage.class`
- `hdfs.namenode.storage.class`

#### Manually creating PersistentVolumes

If you do not have a StorageClass you can use that supports dynamic volume
provisioning then it is possible to manually create a PersistentVolume with
the correct capacity manually. By default the PVCs listed above each request
5Gi of storage. This can be adjusted in the same section as adjusting the
StorageClass.

Using [manifests/chargeback-config/custom-storageclass-values.yaml][example-storage-config]
as a starting point, adjust the `size: "5Gi"` value to the capacity you want
for the following sections:

- `presto.hive.metastore.storage.size`
- `hdfs.datanode.storage.size`
- `hdfs.namenode.storage.size`

### Storing data in S3

By default the data that chargeback collects and generates is stored in a
single node HDFS cluster which is backed by a persistent volume. If you would
instead prefer to store the data in a location outside of the cluster, you can
configure Chargeback to store data in s3.

To use S3, for storage, uncomment the `defaultStorage:` section in the example
[manifests/chargeback-config/custom-values.yaml][example-config] configuration.
Once it's uncommented, you also need to set `awsAccessKeyID` and
`awsSecretAccessKey` in the `chargeback-operator.config` and `presto.config`
sections.

Please note that this must be done before installation, and if it is changed
post-install you may see unexpected behavior.

At this point, you can also disable the deployed HDFS cluster as it's not going
to be used for storing data any longer. To disable the HDFS cluster uncomment
the `hdfs.enabled: true` setting in your `custom-values.yaml`, and set the
value to `false`:


```
hdfs:
  enabled: false
```

### AWS Billing Correlation

Chargeback is able to correlate cluster usage information with [AWS detailed
billing information][AWS-billing], attaching a dollar amount to resource usage.
For clusters running in EC2, this can be enabled by modifying the example
[manifests/chargeback-config/custom-values.yaml][example-config] configuration.

To enable AWS Billing correlation, please ensure the AWS Cost and Usage Reports
are enabled on your account, instructions to enable them can be found
[here][enable-aws-billing].

Next, in the example configuration manifest, you need to update the
`awsBillingDataSource:` section. Change the `enabled` value to `true`, and then
update the `bucket` and `prefix` to the location of your AWS Detailed billing
report.  Once it's uncommented, you also need to set `awsAccessKeyID` and
`awsSecretAccessKey` in the `chargeback-operator.config` and `presto.config`
sections.

This can be done either pre-install or post-install. Note that disabling it
post-install can cause errors in the chargeback-operator.

[AWS-billing]: https://docs.aws.amazon.com/awsaccountbilling/latest/aboutv2/billing-reports-costusage.html
[enable-aws-billing]: https://docs.aws.amazon.com/awsaccountbilling/latest/aboutv2/billing-reports-gettingstarted-turnonreports.html
[example-config]: manifests/chargeback-config/custom-values.yaml
[example-storage-config]: manifests/chargeback-config/custom-storageclass-values.yaml
[storage-classes]: https://kubernetes.io/docs/concepts/storage/storage-classes/

