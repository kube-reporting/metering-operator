## Con***REMOVED***guration

Chargeback supports a few con***REMOVED***guration options which can currently be set by
creating a secret named `chargeback-settings` with a key of `values.yaml`.

An example con***REMOVED***guration ***REMOVED***le can be found in
[manifests/chargeback-con***REMOVED***g/custom-values.yaml][example-con***REMOVED***g]
For details on customizing this ***REMOVED***le, read the
[common-con***REMOVED***guration-options](common-con***REMOVED***guration-options) section below.

### Documentation Conventions

A common way of describing nested ***REMOVED***elds in `custom-values.yaml` in this
documentation will refer to setting con***REMOVED***guration sections or values, by using
the following notation where dots are separators:

```
chargeback-operator.con***REMOVED***g.awsAccessKeyID
```

This is just a convention for referring to nested ***REMOVED***elds in the yaml, where the
above example is referring to the following yaml structure and value:

```
chargeback-operator:
  con***REMOVED***g:
    awsAccessKeyID: "REPLACEME"
```

If you ever see a reference to a setting using this dotted notation, just know
that it's referring to a nested ***REMOVED***eld in a yaml structure.

### Using a custom con***REMOVED***guration

To install the custom con***REMOVED***guration ***REMOVED***le, you can run the following command:

```
kubectl -n $CHARGEBACK_NAMESPACE create secret generic chargeback-settings --from-***REMOVED***le 'values.yaml=manifests/chargeback-con***REMOVED***g/custom-values.yaml'
```

However, this doesn't make it very easy to make changes without deleting and
re-creating the secret, so you can also use this one-liner which will do the
same thing, but will use `kubectl apply`

```
kubectl -n $CHARGEBACK_NAMESPACE create secret generic chargeback-settings --from-***REMOVED***le 'values.yaml=manifests/chargeback-con***REMOVED***g/custom-values.yaml' -o yaml --dry-run > /tmp/chargeback-settings-secret.yaml && kubectl apply -n $CHARGEBACK_NAMESPACE -f /tmp/chargeback-settings-secret.yaml
```

## Common con***REMOVED***guration options

The example manifest
[manifests/chargeback-con***REMOVED***g/custom-values.yaml][example-con***REMOVED***g]
contains has the most common con***REMOVED***guration values a user might want to modify
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

As of Tectonic 1.8.4, Tectonic doesn't install cloud provider speci***REMOVED***c
StorageClasses by default, however it is possible an admin has created
StorageClasses in the cluster for your use. To check what StorageClasses are
available, run:

```
$ kubectl get storageclasses
```

If the output includes `(default)` next to the name of any StorageClasses, then
that StorageClass is the default for your cluster. The default is used when
StorageClass is unspeci***REMOVED***ed or set to `null` in a PersistentVolumeClaim spec.

If you've got a default StorageClass installed, then you're ready to go and do
not need to do anything extra. However, if you wish to use the non-default
StorageClass, please continue reading. If you need to install a StorageClass,
please read the [of***REMOVED***cial documentation on StorageClasses][storage-classes].

##### Con***REMOVED***guring the StorageClass for Chargeback

You can con***REMOVED***gure and specify a StorageClass to use in chargeback by specifying
the StorageClass in your `custom-values.yaml`.
A trimmed down version of this ***REMOVED***le with just the relevant storage section is located in
[manifests/chargeback-con***REMOVED***g/custom-storageclass-values.yaml][example-storage-con***REMOVED***g].

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

Using [manifests/chargeback-con***REMOVED***g/custom-storageclass-values.yaml][example-storage-con***REMOVED***g]
as a starting point, adjust the `size: "5Gi"` value to the capacity you want
for the following sections:

- `presto.hive.metastore.storage.size`
- `hdfs.datanode.storage.size`
- `hdfs.namenode.storage.size`

### Storing data in S3

By default the data that chargeback collects and generates is stored in a
single node HDFS cluster which is backed by a persistent volume. If you would
instead prefer to store the data in a location outside of the cluster, you can
con***REMOVED***gure Chargeback to store data in s3.

To use S3, for storage, uncomment the `defaultStorage:` section in the example
[manifests/chargeback-con***REMOVED***g/custom-values.yaml][example-con***REMOVED***g] con***REMOVED***guration.
Once it's uncommented, you also need to set `awsAccessKeyID` and
`awsSecretAccessKey` in the `chargeback-operator.con***REMOVED***g` and `presto.con***REMOVED***g`
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
[manifests/chargeback-con***REMOVED***g/custom-values.yaml][example-con***REMOVED***g] con***REMOVED***guration.

To enable AWS Billing correlation, please ensure the AWS Cost and Usage Reports
are enabled on your account, instructions to enable them can be found
[here][enable-aws-billing].

Next, in the example con***REMOVED***guration manifest, you need to update the
`awsBillingDataSource:` section. Change the `enabled` value to `true`, and then
update the `bucket` and `pre***REMOVED***x` to the location of your AWS Detailed billing
report.  Once it's uncommented, you also need to set `awsAccessKeyID` and
`awsSecretAccessKey` in the `chargeback-operator.con***REMOVED***g` and `presto.con***REMOVED***g`
sections.

This can be done either pre-install or post-install. Note that disabling it
post-install can cause errors in the chargeback-operator.

[AWS-billing]: https://docs.aws.amazon.com/awsaccountbilling/latest/aboutv2/billing-reports-costusage.html
[enable-aws-billing]: https://docs.aws.amazon.com/awsaccountbilling/latest/aboutv2/billing-reports-gettingstarted-turnonreports.html
[example-con***REMOVED***g]: manifests/chargeback-con***REMOVED***g/custom-values.yaml
[example-storage-con***REMOVED***g]: manifests/chargeback-con***REMOVED***g/custom-storageclass-values.yaml
[storage-classes]: https://kubernetes.io/docs/concepts/storage/storage-classes/

