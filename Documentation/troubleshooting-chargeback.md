<br>
<div class="alert alert-info" role="alert">
    <i class="fa fa-exclamation-triangle"></i><b> Note:</b> This documentation is for a pre-alpha feature. To register for the Chargeback Alpha program, email <a href="mailto:tectonic-alpha-feedback@coreos.com">tectonic-alpha-feedback@coreos.com</a>.
</div>

# Troubleshooting Chargeback

The most likely issue to occur with Chargeback is that it's not starting all the pods.
Pods may fail to start due to lack of resources, or if they have a dependency on a resource that does not exist, such as a StorageClass or Secret.

This guide will help determine the cause.

## Not enough compute resources

The most common issue when installing or running Chargeback is lack of compute resources. Ensure that
Chargeback has been allocated the minimum resource requirements described in the [installation prerequisites][prerequisites].

To determine if the issue is with resources or scheduling, follow the troubleshooting instructions included in the Kubernetes document [Managing Compute Resources for Containers][resource-troubleshooting].

If a container's status is `pending`, the issue is most likely with scheduling.

## Storage Class not con***REMOVED***gured

Chargeback requires that a default Storage Class be con***REMOVED***gured for dynamic provisioning. 

See [con***REMOVED***guring chargeback][con***REMOVED***guring-chargeback-storage] for information on how to check if there are any StorageClasses con***REMOVED***gured for the cluster, how to set the default, and how to con***REMOVED***gure Chargeback to use a StorageClass other than the default.


[resource-troubleshooting]: https://kubernetes.io/docs/concepts/con***REMOVED***guration/manage-compute-resources-container/#troubleshooting
[prerequisites]: install-chargeback.md#prerequisites
[con***REMOVED***guring-chargeback-storage]: chargeback-con***REMOVED***g.md#dynamically-provisioning-persistent-volumes-using-storage-classes
