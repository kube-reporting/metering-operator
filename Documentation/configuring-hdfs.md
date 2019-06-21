# Con***REMOVED***guring HDFS

By default, HDFS is not used or installed with Metering.
However, if you set the `spec.unsupportedFeatures.enableHDFS` unsupported features toggle to true, then you can install a non-production grade HDFS cluster alongside metering for storage.

We are not HDFS experts, and thus we will not directly support HDFS, and it should only be used for testing or development.
Currently there is no way to secure the communications with HFDS, meaning all communications are done in plain text, without authentication.

Proceed at your own risk.

## Persistent Volumes created by HDFS

- `hdfs-namenode-data-hdfs-namenode-0`
  Used by the hdfs-namenode pod to store metadata about the ***REMOVED***les and blocks stored in the hdfs-datanodes.
- One `hdfs-datanode-data-hdfs-datanode-$i` PV per hdfs-datanode replica.
  Used by each hdfs-datanode pod to store blocks for ***REMOVED***les in the HDFS cluster.

## Con***REMOVED***guring the Storage Class for HDFS

Use [hdfs-storage.yaml][hdfs-storage-con***REMOVED***g] as a template and adjust the `class: null` value to name of the `StorageClass` to use.
Leaving the value `null` will cause Metering to use the default StorageClass for the cluster.

- `spec.hadoop.spec.hdfs.datanode.storage.class`
- `spec.hadoop.spec.hdfs.namenode.storage.class`

## Con***REMOVED***guring the volume sizes for HDFS

Use [hdfs-storage.yaml][hdfs-storage-con***REMOVED***g] as a template and adjust the `size: "5Gi"` value to the desired capacity for the following sections:

- `spec.hadoop.spec.hdfs.datanode.storage.size`
- `spec.hadoop.spec.hdfs.namenode.storage.size`

[hdfs-storage-con***REMOVED***g]: ../manifests/metering-con***REMOVED***g/hdfs-storage.yaml
