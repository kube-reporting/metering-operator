# Con***REMOVED***guring Storage

By Default, Metering requires persistent storage in two main ways.
The primary storage requirement is to persist data collected by the reporting-operator and store the results of reports. This is usually some form of object storage or distributed ***REMOVED***le system.

Additionally, Hive metastore requires storage for it's database containing metadata about database tables managed by Presto or Hive. By default, this information is stored in an embedded database called Derby, which keeps it's data on disk in a PersistentVolume, but metastore can also be con***REMOVED***gured to use an existing Mysql or Postgresql database, instead of Derby. Read the [con***REMOVED***guring the Hive metastore documentation][con***REMOVED***guring-hive-metastore] for more details.

## Storing data in Amazon S3
**Note**: Metering only supports Amazon S3, and not any S3 compatible API at this time.

To use Amazon S3 for storage, edit the `spec.storage` section in the example [s3-storage.yaml][s3-storage-con***REMOVED***g] con***REMOVED***guration.
Set the `spec.storage.hive.s3.bucket`, `spec.storage.hive.s3.region` and `spec.storage.hive.s3.secretName` values.

The `bucket` and `region` ***REMOVED***elds should be the name, and optionally the path within the bucket you wish to store Metering data at, and the region in which you wish to create your bucket in, respectively.

If you want to provide an existing S3 bucket, or do not want to provide IAM credentials that have CreateBucket permissions, set `spec.storage.hive.s3.createBucket` to `false` and provide the name of a pre-existing bucket for `spec.storage.hive.s3.bucket`.
The `secretName` should be the name of a secret in the Metering namespace containing the AWS credentials in the `data.aws-access-key-id` and `data.aws-secret-access-key` ***REMOVED***elds.

For example:
```
apiVersion: v1
kind: Secret
metadata:
  name: your-aws-secret
data:
  aws-access-key-id: "dGVzdAo="
  aws-secret-access-key: "c2VjcmV0Cg=="
```

To store data in Amazon S3, the `aws-access-key-id` and `aws-secret-access-key` credentials must have read and write access to the bucket.
For an example of an IAM policy granting the required permissions, see the [aws/read-write.json](aws/read-write.json) ***REMOVED***le.
If you left `spec.storage.hive.s3.createBucket` set to true, or unset, then you should use [aws/read-write-create.json](aws/read-write-create.json) which contains permissions for creating and deleting buckets.

Please note that this must be done before installation.
Changing these settings after installation will result in broken and unexpected behavior.

## Using shared volumes for storage

Metering has no storage by default, but can use any ReadWriteMany PersistentVolume or [StorageClass][storage-classes] that provisions a ReadWriteMany PersistentVolume.

To use a ReadWriteMany PersistentVolume for storage, modify the [shared-storage.yaml][shared-storage-con***REMOVED***g] con***REMOVED***guration.

You have two options:

1) Set `storage.hive.sharedPVC.createPVC` to true and set the `storage.hive.sharedPVC.storageClass` to the name of a StorageClass with ReadWriteMany access mode. This will use dynamic volume provisioning to have a volume created automatically.
2) Set `storage.hive.sharedPVC.claimName` to the name of an existing ReadWriteMany PVC. This is necessary if you don't have dynamic volume provisioning, or wish to have more control over how the PersistentVolume is created.

> Note: NFS is not recommended to use with Metering.

## Using HDFS for storage (unsupported)

If you do not have access to S3, or storage provisioner that supports ReadWriteMany PVCs, you may also test using HDFS.

HDFS is currently unsupported.
We do not support running HDFS on Kubernetes as it's not very ef***REMOVED***cient, and has an increased complexity over using object storage.
However, because we historically have used HDFS for development, there are options available within Metering to deploy and use HDFS if you're willing to enable unsupported features.

For more details read [con***REMOVED***guring HDFS][con***REMOVED***guring-hdfs].

[storage-classes]: https://kubernetes.io/docs/concepts/storage/storage-classes/
[s3-storage-con***REMOVED***g]: ../manifests/metering-con***REMOVED***g/s3-storage.yaml
[shared-storage-con***REMOVED***g]: ../manifests/metering-con***REMOVED***g/shared-storage.yaml
[hdfs-storage-con***REMOVED***g]: ../manifests/metering-con***REMOVED***g/hdfs-storage.yaml
[con***REMOVED***guring-hive-metastore]: con***REMOVED***guring-hive-metastore.md
[con***REMOVED***guring-hdfs]: con***REMOVED***guring-hdfs.md
