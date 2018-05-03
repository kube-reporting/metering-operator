# Troubleshooting Metering

The most likely issue to occur with Metering is that it's not starting all the pods.
Pods may fail to start due to lack of resources, or if they have a dependency on a resource that does not exist, such as a StorageClass or Secret.

This guide will help determine the cause.

## Not enough compute resources

The most common issue when installing or running Metering is lack of compute resources. Ensure that Metering has been allocated the minimum resource requirements described in the [installation prerequisites][prerequisites].

To determine if the issue is with resources or scheduling, follow the troubleshooting instructions included in the Kubernetes document [Managing Compute Resources for Containers][resource-troubleshooting].

If a container's status is `pending`, the issue is most likely with scheduling.

## Storage Class not con***REMOVED***gured

Metering requires that a default Storage Class be con***REMOVED***gured for dynamic provisioning.

See [con***REMOVED***guring metering][con***REMOVED***guring-metering-storage] for information on how to check if there are any StorageClasses con***REMOVED***gured for the cluster, how to set the default, and how to con***REMOVED***gure Metering to use a StorageClass other than the default.


[resource-troubleshooting]: https://kubernetes.io/docs/concepts/con***REMOVED***guration/manage-compute-resources-container/#troubleshooting
[prerequisites]: install-metering.md#prerequisites
[con***REMOVED***guring-metering-storage]: metering-con***REMOVED***g.md#dynamically-provisioning-persistent-volumes-using-storage-classes
