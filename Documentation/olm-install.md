# Installation using Operator Lifecycle Manager (OLM)

Before you begin please make sure that the [Operator Lifecycle Manager][install-olm] version 0.4.0 or greater is installed in the cluster before running this guide.
It must be installed using the `upstream` method.

## Install

First, start by creating your namespace:

```
export METERING_NAMESPACE=metering
kubectl create ns $METERING_NAMESPACE
```

### Configuration

All of the supported configuration options are documented in [configuring metering][configuring-metering].
In this document, we will refer to your configuration as your `metering.yaml`.

### Install Operator Metering with Configuration

Installation is a two step process. First, install the Metering Helm operator. Then, install the `Metering` resource that defines the configuration.

To start, download the [Metering subscription][metering-subscription] and save it as `metering.subscription.yaml`, and download your [Metering][example-config] resource and save it as `metering.yaml`.

The subscription is used by the Operator Lifecycle Management and Catalog operators to install CRDs and the Metering Helm operator.

Install the subscription into the cluster:

```
kubectl create -n $METERING_NAMESPACE -f metering.subscription.yaml
```

This causes the operator-lifecycle-manager (OLM) to create an `InstallPlan` resource named after the `Subscription`.
The `InstallPlan` references a `ClusterServiceVersion` within it's catalog which describes the Metering Helm operator Deployment, Role, ServiceAccount, and CRD resources that need to be created.
The OLM operator will read the `ClusterServiceVersion` from it's catalog, and then create a `ClusterServiceVersion` named after our `InstallPlan`.
Once the `ClusterServiceVersion` exists, the operator creates the deployment containing our Metering Helm operator.


To verify this, run the following command:

```
kubectl get -n $METERING_NAMESPACE subscription-v1s,installplan-v1s,clusterserviceversion-v1s
```

You should see something like this in the output:

```
NAME                                                           AGE
subscription-v1.app.coreos.com/metering-operator.v0.6.0        4m

NAME                                                                        AGE
installplan-v1.app.coreos.com/install-metering-operator.v0.6.0-jhrr4        4m

NAME                                                                    AGE
clusterserviceversion-v1.app.coreos.com/metering-operator.v0.6.0        4m
```

**Note: The Subscription, and InstallPlan resources declare an intent to perform an installation once. This means they do not ensure the ClusterServiceVersion exists after creating it the first time, and deleting them will not result in the operator being uninstalled. For details on uninstall, see [Uninstalling Metering](#uninstalling-metering).**

Finally, install the `Metering` resource, which causes the Metering Helm operator to install and configure Metering and its dependencies.

```
kubectl create -n $METERING_NAMESPACE -f metering.yaml
```

## Uninstall

The operator-lifecycle-manager (OLM) operator does not automatically uninstall the operator deployment when you delete a `Subscription` or `InstallPlan` in order to avoid accidental deletions of components when removing a subscription, such as if you no longer want automatic updates.
This means subscriptions orphan their `ClusterServiceVersions` when deleted, and that we must explicitly delete the `ClusterServiceVersions` it created to do an uninstall.

To perform an uninstall, you must first delete the subscription, and then delete the related `ClusterServiceVersions` as the commands below demonstrate:

```
kubectl delete -n $METERING_NAMESPACE -f metering.subscription.yaml
kubectl delete -n $METERING_NAMESPACE clusterserviceversion-v1s -l operator-metering=true
```

[install-olm]: https://github.com/operator-framework/operator-lifecycle-manager/blob/master/Documentation/install/install.md#install-the-latest-released-version-of-olm-for-upstream-kubernetes
[metering-subscription]: ../manifests/deploy/generic/alm/metering.subscription.yaml
[configuring-metering]: metering-config.md
[example-config]: ../manifests/metering-config/default.yaml
