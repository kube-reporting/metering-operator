<br>
<div class="alert alert-info" role="alert">
    <i class="fa fa-exclamation-triangle"></i><b> Note:</b> This documentation is for a pre-alpha feature. To register for the Chargeback Alpha program, email <a href="mailto:tectonic-alpha-feedback@coreos.com">tectonic-alpha-feedback@coreos.com</a>.
</div>

# Troubleshooting

The most likely issue to occur with Chargeback is that it's not starting all the pods.
Pods not starting is typically due to lack of resources, but it can also be caused when the pods have a dependency on another resource that doesn't exist, such as a StorageClass or Secret.
The sections below provide some references for determining if this is the cause.

## Not enough compute resources

The most common issue when installing or running Chargeback is often lack of compute resources.
Chargeback's minimum resource requirements are described in the [installation prerequisites][prerequisites].

To determine if you're running into issues with resources or scheduling, follow the instructions from the Kubernetes upstream documentation on [troubleshooting for compute resources][resource-troubleshooting].
The key diagnostic steps are to check if a container's status is `pending` then it's likely a issue with scheduling.

## Storage Class Not Configured

Another common issue is not having a default StorageClass configured, which is used for dynamic provisioning.
[See configuring chargeback][configuring-chargeback-storage] for information on how to check if you've got any StorageClasses configured for your cluster, as well as how to set the default, or configure Chargeback to use a StorageClass other than the default.


[resource-troubleshooting]: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/#troubleshooting
[prerequisites]: install-chargeback.md#prerequisites
[configuring-chargeback-storage]: chargeback-config.md#dynamically-provisioning-persistent-volumes-using-storage-classes
