# Con***REMOVED***guring Storage

Metering by default requires persistent storage in a few areas, but can be con***REMOVED***gured so it doesn't require any real persistent storage within the cluster.
The primary purpose of the storage requirements are to persist data collected by the reporting-operator and store the results of reports.

## Persistent Volumes

Metering by default requires at least 3 Persistent Volume to operate. The Persistent Volume Claims (PVCs) are listed below:

- `hive-metastore-db-data` is generally the only _required_ volume.
  It is used by hive metastore to retain information about the location of where data is stored, which Presto and Hive server use.
  In practice, it is possible to remove this requirement by using [MySQL or Postgresql for the Hive metastore database][con***REMOVED***guring-hive-metastore].
- `hdfs-namenode-data-hdfs-namenode-0`
  Used by the hdfs-namenode pod to store metadata about the ***REMOVED***les and blocks stored in the hdfs-datanodes.
  This PVCs are not required to [store data in AWS S3](#storing-data-in-s3).
- One `hdfs-datanode-data-hdfs-datanode-$i` PV per hdfs-datanode replica.
  Used by each hdfs-datanode pod to store blocks for ***REMOVED***les in the HDFS cluster.
  These PVCs are not required to [store data in AWS S3](#storing-data-in-s3).

Each of these Persistent Volume Claims is created dynamically by a Stateful Set.
Enabling this requires that dynamic volume provisioning be enabled via a Storage Class, or persistent volumes of the correct size must be manually pre-created.

## Dynamically provisioning Persistent Volumes using Storage Classes

Storage Classes may be used when dynamically provisioning Persistent Volume Claims using a Stateful Set.
Use `kubectl get` to determine if Storage Classes have been created in your cluster.

```
$ kubectl get storageclasses
```

If the output includes `(default)` next to the `name` of any `StorageClass`, then that `StorageClass` is the default for the cluster.
The default is used when `StorageClass` is unspeci***REMOVED***ed or set to `null` in a `PersistentVolumeClaim` spec.

If no `StorageClass` is listed, or if you wish to use a non-default `StorageClass`, see [Con***REMOVED***guring the StorageClass for Metering](#con***REMOVED***guring-the-storage-class-for-metering) below.

For more information, see [Storage Classes][storage-classes] in the Kubernetes documentation.

### Con***REMOVED***guring the Storage Class for Metering

To con***REMOVED***gure and specify a `StorageClass` for use in Metering, specify the `StorageClass` in `custom-values.yaml`. A example `StorageClass` section is included in [custom-storage.yaml][custom-storage-con***REMOVED***g].

Uncomment the following sections and replace the `null` in `class: null` value with the name of the `StorageClass` to use. Leaving the value `null` will cause Metering to use the default StorageClass for the cluster.

- `spec.presto.spec.hive.metastore.storage.class`
- `spec.hdfs.spec.datanode.storage.class`
- `spec.hdfs.spec.namenode.storage.class`

### Con***REMOVED***guring the volume sizes for Metering

Use [custom-storage.yaml][custom-storage-con***REMOVED***g] as a template and adjust the `size: "5Gi"` value to the desired capacity for the following sections:

- `presto.spec.hive.metastore.storage.size`
- `hdfs.spec.datanode.storage.size`
- `hdfs.spec.namenode.storage.size`

### Manually creating Persistent Volumes

If a Storage Class that supports dynamic volume provisioning does not exist in the cluster, it is possible to manually create a Persistent Volume with the correct capacity.
By default, the PVCs listed above each request 5Gi of storage.
This can be adjusted in the same section as adjusting the Storage Class as documented in [Con***REMOVED***guring the volume sizes for Metering](#con***REMOVED***guring-the-volume-sizes-for-metering).

## Storing data in S3

By default, the data that Metering collects and generates is stored in a single node HDFS cluster which is backed by a Persistent Volume.
To store the data in a location outside of the cluster, con***REMOVED***gure Metering to store data in S3.

To use S3 for storage, edit the `spec.defaultStorage` section in the example [s3-storage.yaml][s3-storage-con***REMOVED***g] con***REMOVED***guration.
Set `awsAccessKeyID` and `awsSecretAccessKey` in the `reporting-operator.con***REMOVED***g` and `presto.con***REMOVED***g` sections.

To store data in S3, the `awsAccessKeyID` and `awsSecretAccessKey` credentials must have read and write access to the bucket.
For an example of an IAM policy granting the required permissions see the [aws/read-write.json](aws/read-write.json) ***REMOVED***le.
Replace `operator-metering-data` with the name of your bucket.

Please note that this must be done before installation. Changing these settings after installation may result in unexpected behavior.

Because the deployed HDFS cluster will not be used to store data, it may also be disabled.
In `s3-storage.yaml`, this has already been done by setting `hdfs.enabled` to `false` and setting `presto.spec.hive.con***REMOVED***g.useHdfsCon***REMOVED***gMap` to `false`.

## Using shared volumes for storage

Metering uses HDFS for storage by default, but can use any ReadWriteMany PersistentVolume or StorageClass.

To use a ReadWriteMany for storage, modify the [shared-storage.yaml][shared-storage-con***REMOVED***g] con***REMOVED***guration.

Con***REMOVED***gure the `presto.spec.con***REMOVED***g.sharedVolume.storage.persistentVolumeClaimStorageClass` to a StorageClass with ReadWriteMany access mode.

Note that our example [shared-storage.yaml][shared-storage-con***REMOVED***g] disables HDFS by setting `hdfs.enabled` to false since it will not be used.

> Note: NFS is not recommended to use with Metering.

[storage-classes]: https://kubernetes.io/docs/concepts/storage/storage-classes/
[custom-storage-con***REMOVED***g]: ../manifests/metering-con***REMOVED***g/custom-storage.yaml
[s3-storage-con***REMOVED***g]: ../manifests/metering-con***REMOVED***g/s3-storage.yaml
[shared-storage-con***REMOVED***g]: ../manifests/metering-con***REMOVED***g/shared-storage.yaml
[con***REMOVED***guring-hive-metastore]: con***REMOVED***guring-hive-metastore.md
