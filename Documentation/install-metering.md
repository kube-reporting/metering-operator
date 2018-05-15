# Installing Operator Metering

Operator Metering is a collection of a few components:

- A Metering Operator Pod which aggregates Prometheus data and generates reports based
  on the collected usage information.
- Hive and Presto clusters, used by the Operator Metering Pod to perform queries on the
  collected usage data.

## Prerequisites

Operator Metering requires the following components:

- A Kubernetes 1.8 cluster.
- A StorageClass for dynamic volume provisioning. ([See configuring metering][configuring-metering] for more information.)
- A Prometheus installation within the cluster configured to do Kubernetes cluster-monitoring.
    - The prometheus-operator repository's [kube-prometheus instructions][kube-prometheus] are the standard way of achieving Prometheus cluster-monitoring.
    - At a minimum, we require kube-state-metrics, node-exporter, and built-in Kubernetes target metrics.
- 3.5GB Memory and 1.5 CPU Cores (1500 Millicores).
- At least 1 node with 1.5GB available memory (the highest memory request for a single Operator Metering Pod)
    - Memory and CPU consumption may often be lower, but will spike when running reports, or collecting data for larger clusters.
- A properly configured kubectl to access the Kubernetes cluster.

## Installation

First, start by creating your namespace:

```
export METERING_NAMESPACE=metering
kubectl create ns $METERING_NAMESPACE
```

### Configuration

Before continuing with the installation, please read [Configuring Operator Metering][configuring-metering].
Some options may not be changed post-install. Be certain to configure these options, if desired, before installation.

If you're not using [kube-prometheus][kube-prometheus] installation, or your Prometheus service is not named `prometheus-k8s` and in the `monitoring` namespace, then you must customize the [prometheus URL config option][configure-prometheus-url] before proceeding.

If you do not wish to modify the Operator Metering configuration, a minimal configuration example that doesn't override anything can be found in [default.yaml][default-config].

### Install Operator Metering with Configuration

Installation is a two step process. First, install the Metering Helm operator. Then, install the `Metering` resource that defines the configuration.

To start, download the [Metering subscription][metering-subscription] and save it as `metering.subscription.yaml`, and download your `Metering` resource and save it as `metering.yaml`.

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
subscription-v1.app.coreos.com/metering-helm-operator.v0.6.0   4m

NAME                                                                        AGE
installplan-v1.app.coreos.com/install-metering-helm-operator.v0.6.0-jhrr4   4m

NAME                                                                    AGE
clusterserviceversion-v1.app.coreos.com/metering-helm-operator.v0.6.0   4m
```

**Note: The Subscription, and InstallPlan resources declare an intent to perform an installation once. This means they do not ensure the ClusterServiceVersion exists after creating it the first time, and deleting them will not result in the operator being uninstalled. For details on uninstall, see [Uninstalling Metering](#uninstalling-metering).**

Finally, install the `Metering` resource, which causes the Metering Helm operator to install and configure Metering and its dependencies.

```
kubectl create -n $METERING_NAMESPACE -f metering.yaml
```

## Verifying operation

First, wait until the Metering Helm operator deploys all of the Metering components:

```
kubectl get pods -n $METERING_NAMESPACE -l app=metering-helm-operator -o name | cut -d/ -f2 | xargs -I{} kubectl -n $METERING_NAMESPACE logs -f {} -c metering-helm-operator
```

When output similar to the following appears, the rest of the Pods should be initializing:

```
Waiting for Tiller to become ready
Waiting for Tiller to become ready
Getting pod metering-helm-operator-b5f86788c-ks4zq owner information
Querying for Deployment metering-helm-operator
No values, using default values
Running helm upgrade for release operator-metering
Release "operator-metering" has been upgraded. Happy Helming!
LAST DEPLOYED: Fri Jan 26 19:18:34 2018
NAMESPACE: metering
STATUS: DEPLOYED

RESOURCES:

... the rest is omitted for brevity ...
```

Check the logs of the `metering` deployment for errors:

```
$ kubectl get pods -n $METERING_NAMESPACE -l app=metering -o name | cut -d/ -f2 | xargs -I{} kubectl -n $METERING_NAMESPACE logs {} -f
```

## Uninstalling Metering

The operator-lifecycle-manager (OLM) operator does not automatically uninstall the operator deployment when you delete a `Subscription` or `InstallPlan` in order to avoid accidental deletions of components when removing a subscription, such as if you no longer want automatic updates.
This means subscriptions orphan their `ClusterServiceVersions` when deleted, and that we must explicitly delete the `ClusterServiceVersions` it created to do an uninstall.

So perform an uninstall, you must first delete the subscription, and then delete the related `ClusterServiceVersions` as the commands below demonstrate:

```
kubectl delete -n $METERING_NAMESPACE -f metering.subscription.yaml
kubectl delete -n $METERING_NAMESPACE clusterserviceversion-v1s -l operator-metering=true
```

## Using Operator Metering

For instructions on using Operator Metering, please see [Using Operator Metering][using-metering].

[metering-subscription]: ../manifests/alm/metering.subscription.yaml
[default-config]: ../manifests/metering-config/default.yaml
[using-metering]: using-metering.md
[configuring-metering]: metering-config.md
[configure-prometheus-url]: metering-config.md#Prometheus-URL
[kube-prometheus]: https://github.com/coreos/prometheus-operator/tree/master/contrib/kube-prometheus
