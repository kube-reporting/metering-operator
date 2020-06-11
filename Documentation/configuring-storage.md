# Configuring Storage

By Default, Metering requires persistent storage in two main ways.
The primary storage requirement is to persist data collected by the reporting-operator and store the results of reports. This is usually some form of object storage or distributed file system.

Additionally, Hive metastore requires storage for it's database containing metadata about database tables managed by Presto or Hive. By default, this information is stored in an embedded database called Derby, which keeps it's data on disk in a PersistentVolume, but metastore can also be configured to use an existing Mysql or Postgresql database, instead of Derby. Read the [configuring the Hive metastore documentation][configuring-hive-metastore] for more details.

> **Note**: This **must be done before installation**.
> Changing these settings after installation will result in broken and unexpected behavior.

## Storing data in Amazon S3
### Create a S3 bucket
To use Amazon S3 for storage, edit the `spec.storage` section in the example [s3-storage.yaml][s3-storage-config] configuration.
Set the `spec.storage.hive.s3.bucket`, `spec.storage.hive.s3.region` and `spec.storage.hive.s3.secretName` values.

The `bucket` field defines the bucket name, and if desired, the path for storing Metering data.
`region` is the AWS region in which you wish to create your bucket in.

> **Note**: The values of the `aws-access-key-id` and `aws-secret-access-key` must be base64 encoded.

For example, a file to be loaded with `kubectl -n $METERING_NAMESPACE create -f <aws-secret-file.yaml>`:
```
apiVersion: v1
kind: Secret
metadata:
  name: your-aws-secret
data:
  aws-access-key-id: "dGVzdAo="
  aws-secret-access-key: "c2VjcmV0Cg=="
```
`spec.storage.hive.s3.secretName` must then be set to: `metadata.name` from the Secret.
### Using an existing S3 bucket

If you want to provide an existing S3 bucket, or do not want to provide IAM credentials that have CreateBucket permissions, set `spec.storage.hive.s3.createBucket` to `false` and provide the name of a pre-existing bucket for `spec.storage.hive.s3.bucket`.
The `spec.storage.hive.s3.secretName` should be the name of a secret in the Metering namespace containing the AWS credentials in the `data.aws-access-key-id` and `data.aws-secret-access-key` fields.

To store data in Amazon S3, the `aws-access-key-id` and `aws-secret-access-key` credentials must have read and write access to the bucket.
For an example of an IAM policy granting the required permissions, see the [aws/read-write.json](aws/read-write.json) file.
If you left `spec.storage.hive.s3.createBucket` set to true, or unset, then you should use [aws/read-write-create.json](aws/read-write-create.json) which contains permissions for creating and deleting buckets.

## Storing data in S3 Compatible Storage

To use S3 compatible storage such as Noobaa, edit the `spec.storage` section in the example [s3-compatible-storage.yaml][s3-compatible-storage-config] configuration.
Set the `spec.storage.hive.s3Compatible.bucket`, `spec.storage.hive.s3Compatible.endpoint` and `spec.storage.hive.s3Compatible.secretName` values.

You must provide an existing S3 Compatible bucket, under the field `spec.storage.hive.s3Compatible.bucket`.
You must also provide the endpoint for your storage, under the field `spec.storage.hive.s3Compatible.endpoint`.
If you want to use HTTPS with a self-signed certificate on your s3Compatible endpoint, you need to set the options in the `spec.storage.hive.s3Compatible.ca` section.
Set `spec.storage.hive.s3Compatible.ca.createSecret` to true and `spec.storage.hive.s3Compatible.ca.content` to your PEM encoded CA bundle.
The `spec.storage.hive.s3Compatible.secretName` should be the name of a secret in the Metering namespace containing the AWS credentials in the `data.aws-access-key-id` and `data.aws-secret-access-key` fields.

> **Note**: The values of the `aws-access-key-id` and `aws-secret-access-key` must be base64 encoded.

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

To store data in S3 compatible storage, the `aws-access-key-id` and `aws-secret-access-key` credentials must have read and write access to the bucket.

## Storing data in Azure

You can also store your data in Azure blob storage, and to do so, you must use an existing container.
Edit the `spec.storage` section in the example [azure-blob-storage.yaml][azure-blob-storage-config] configuration.
Set the `spec.storage.hive.azure.container`, and `spec.storage.hive.azure.secretName` values.

You must provide an existing Azure blob storage container under the field `spec.storage.hive.azure.container`.
The `spec.storage.hive.azure.secretName` should be the name of a secret in the Metering namespace containing the AWS credentials in the `data.azure-storage-account-name` and `data.azure-secret-access-key` fields.

> **Note**: The values of the `azure-storage-account-name` and `azure-secret-access-key` must be base64 encoded.

For example:

```
apiVersion: v1
kind: Secret
metadata:
  name: your-azure-secret
data:
  azure-storage-account-name: "dGVzdAo="
  azure-secret-access-key: "c2VjcmV0Cg=="
```

The `spec.storage.hive.azure.container` should be the container you wish to store metering data at, and the optional `spec.storage.hive.azure.rootDirectory` should be the folder you want your data in, inside the container.

## Storing data in Google Cloud Storage

You can also store your data in Google Cloud Storage, and to do so, you must use an existing bucket.
Edit the `spec.storage` section in the example [gcs-storage.yaml][gcs-config] configuration.
Set the `spec.storage.hive.gcs.bucket` and `spec.storage.hive.gcs.secretName`.

The `spec.storage.hive.gcs.bucket` field should be the name of an existing bucket you wish to hold your metering data.
You can also specify a sub-directory within this bucket by appending the name of the directory after the bucket name.
For example: `bucket: metering-gcs/test`, where the bucket's name is `metering-gcs` and `test` is the name of the sub-directory.
The `spec.storage.hive.gcs.secretName` should be the name of a secret in the Metering namespace containing the GCS service account JSON file in the `data.service-account.json` field.

> **Note**: The values of the `gcs-service-account.json` must be base64 encoded.

For example:

```
apiVersion: v1
kind: Secret
metadata:
  name: your-gcs-secret
data:
  gcs-service-account.json: "c2VjcmV0Cg=="
```

To create this secret from a file using `kubectl`, ensure that `$METERING_NAMESPACE` is set to your Metering namespace and run:

```
kubectl -n $METERING_NAMESPACE create secret generic your-gcs-secret --from-file gcs-service-account.json=/path/to/your/service-account-key.json
```

## Using shared volumes for storage

Metering has no storage by default, but can use any ReadWriteMany PersistentVolume or [StorageClass][storage-classes] that provisions a ReadWriteMany PersistentVolume.

To use a ReadWriteMany PersistentVolume for storage, modify the [shared-storage.yaml][shared-storage-config] configuration.

You have two options:

1) Set `storage.hive.sharedPVC.createPVC` to true and set the `storage.hive.sharedPVC.storageClass` to the name of a StorageClass with ReadWriteMany access mode. This will use dynamic volume provisioning to have a volume created automatically.
2) Set `storage.hive.sharedPVC.claimName` to the name of an existing ReadWriteMany PVC. This is necessary if you don't have dynamic volume provisioning, or wish to have more control over how the PersistentVolume is created.

## Using HDFS for storage (unsupported)

If you do not have access to S3, or storage provisioner that supports ReadWriteMany PVCs, you may also test using HDFS.

HDFS is currently unsupported.
We do not support running HDFS on Kubernetes as it's not very efficient, and has an increased complexity over using object storage.
However, because we historically have used HDFS for development, there are options available within Metering to deploy and use HDFS if you're willing to enable unsupported features.

For more details read [configuring HDFS][configuring-hdfs].

[storage-classes]: https://kubernetes.io/docs/concepts/storage/storage-classes/
[s3-storage-config]: ../manifests/metering-config/s3-storage.yaml
[s3-compatible-storage-config]: ../manifests/metering-config/s3-compatible-storage.yaml
[azure-blob-storage-config]: ../manifests/metering-config/azure-blob-storage.yaml
[gcs-config]: ../manifests/metering-config/gcs-storage.yaml
[shared-storage-config]: ../manifests/metering-config/shared-storage.yaml
[hdfs-storage-config]: ../manifests/metering-config/hdfs-storage.yaml
[configuring-hive-metastore]: configuring-hive-metastore.md
[configuring-hdfs]: configuring-hdfs.md
