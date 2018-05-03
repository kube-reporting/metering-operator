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

To start, download the [Metering install plan][metering-installplan] and save it as `metering.installplan.yaml`, and download your `Metering` resource and save it as `metering.yaml`.

The install plan is used by the Operator Lifecycle Management and Catalog operators to install CRDs and the Metering Helm operator.

Install the install plan into the cluster:

```
kubectl create -n $METERING_NAMESPACE -f metering.installplan.yaml
```

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

## Using Operator Metering

For instructions on using Operator Metering, please see [Using Operator Metering][using-metering].


[metering-installplan]: ../manifests/alm/metering.installplan.yaml
[default-config]: ../manifests/metering-config/default.yaml
[using-metering]: using-metering.md
[configuring-metering]: metering-config.md
[configure-prometheus-url]: metering-config.md#Prometheus-URL
[kube-prometheus]: https://github.com/coreos/prometheus-operator/tree/master/contrib/kube-prometheus
