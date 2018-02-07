<br>
<div class=“alert alert-info” role=“alert”>
<i class=“fa fa-exclamation-triangle”></i><b> Note:</b> This documentation is for an alpha feature. For questions and feedback on the Metering and Chargeback Alpha program, email <a href="mailto:tectonic-alpha-feedback@coreos.com">tectonic-alpha-feedback@coreos.com</a>.
</div>

# Installing Chargeback

Chargeback consists of a few components:

- A Chargeback Pod which aggregates Prometheus data and generates reports based
  on the collected usage information.
- Hive and Presto clusters, used by the Chargeback Pod to perform queries on the
  collected usage data.

## Prerequisites

Chargeback requires the following components:

- A Tectonic 1.8 cluster.
- A StorageClass for dynamic volume provisioning. ([See con***REMOVED***guring chargeback][con***REMOVED***guring-chargeback] for more information.)
- 3.5GB Memory and 1.15 CPU Cores (1150 Millicores).
- At least 1 node with 1.5GB available memory (the highest memory request for a single Chargeback Pod)
    - Memory and CPU consumption may often be lower, but will spike when running reports, or collecting data for larger clusters.
- A properly con***REMOVED***gured kubectl to access the Kubernetes cluster.

## Installation

First, set up the namespace by creating it and copying the `coreos-pull-secret` into it:

```
export CHARGEBACK_NAMESPACE=chargeback
kubectl create ns $CHARGEBACK_NAMESPACE
kubectl get secret -n tectonic-system coreos-pull-secret --export -o json | kubectl create -n $CHARGEBACK_NAMESPACE -f -
```

### Con***REMOVED***guration

Before continuing with the installation, please read [Con***REMOVED***guring Chargeback][con***REMOVED***guring-chargeback].
Some options may not be changed post-install. Be certain to con***REMOVED***gure these options, if desired, before installation.

If you do not wish to modify the Chargeback con***REMOVED***guration, a minimal con***REMOVED***guration example that doesn't override anything can be found in [manifests/chargeback-con***REMOVED***g/default.yaml][default-con***REMOVED***g].

### Install Chargeback with Con***REMOVED***guration

Installation is a two step process. First, install the Chargeback Helm operator. Then, install the `Chargeback` resource that de***REMOVED***nes the con***REMOVED***guration.

To start, download the [Chargeback install plan][chargeback-installplan] and save it as `chargeback.installplan.yaml`, and download your `Chargeback` resource and save it as `chargeback.yaml`.

The install plan is used by the Tectonic Application Lifecycle Management and Catalog operators to install CRDs and the Chargeback Helm operator.

Install the install plan into the cluster:

```
kubect create -n $CHARGEBACK_NAMESPACE -f chargeback.installplan.yaml
```

Finally, install the `Chargeback` resource, which causes the Chargeback Helm operator to install and con***REMOVED***gure Chargeback and its dependencies.

```
kubectl create -n $CHARGEBACK_NAMESPACE -f chargeback.yaml
```

## Verifying operation

First, wait until the Chargeback Helm operator deploys all of the Chargeback components:

```
kubectl get pods -n $CHARGEBACK_NAMESPACE -l app=chargeback-helm-operator -o name | cut -d/ -f2 | xargs -I{} kubectl -n $CHARGEBACK_NAMESPACE logs -f {} -c chargeback-helm-operator
```

When output similar to the following appears, the rest of the Pods should be initializing:

```
Waiting for Tiller to become ready
Waiting for Tiller to become ready
Getting pod chargeback-helm-operator-b5f86788c-ks4zq owner information
Querying for Deployment chargeback-helm-operator
No values, using default values
Running helm upgrade for release tectonic-chargeback
Release "tectonic-chargeback" has been upgraded. Happy Helming!
LAST DEPLOYED: Fri Jan 26 19:18:34 2018
NAMESPACE: chargeback
STATUS: DEPLOYED

RESOURCES:

... the rest is omitted for brevity ...
```

Check the logs of the `chargeback` deployment for errors:

```
$ kubectl get pods -n $CHARGEBACK_NAMESPACE -l app=chargeback -o name | cut -d/ -f2 | xargs -I{} kubectl -n $CHARGEBACK_NAMESPACE logs {} -f
```

## Using Chargeback

For instructions on using Chargeback, please see [Using Chargeback][using-chargeback].


[chargeback-installplan]: ../manifests/alm/chargeback.installplan.yaml
[default-con***REMOVED***g]: ../manifests/chargeback-con***REMOVED***g/default.yaml
[using-chargeback]: using-chargeback.md
[con***REMOVED***guring-chargeback]: chargeback-con***REMOVED***g.md
