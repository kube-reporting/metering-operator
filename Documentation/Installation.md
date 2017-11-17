# Installing Chargeback

Chargeback consists of a few components:

- A chargeback pod which aggregates Prometheus data and generates reports based
  on the collected usage information.
- Hive and Presto clusters, used by the chargeback pod to perform queries on the
  collected usage data.

## Prerequisites

In order to install and use chargeback the following components will be
necessary:

- A Tectonic installed Kubernetes cluster, with the following components
  (Tectonic 1.8.0 meets these requirements):
  - Tectonic Prometheus Operator of version 1.6.0 or greater (Prometheus
    Operator v0.13)
  - ALM installed
- A properly configured kubectl to access the Kubernetes cluster.

To alter the version of the Tectonic Prometheus operator to be 1.6.0, run the
following command:

```
$ kubectl -n tectonic-system patch deploy tectonic-prometheus-operator -p '{"spec":{"template":{"spec":{"containers":[{"name":"tectonic-prometheus-operator","image":"quay.io/coreos/tectonic-prometheus-operator:v1.6.0"}]}}}}'
```

Once the operator changes the version of the `kube-state-metrics` pod to 1.0.1,
chargeback installation may proceed.

## Installation

Chargeback can be installed via the Tectonic console, but a couple commands must
be run first to make it available via the web UI.

```
$ kubectl -n tectonic-system create -f manifests/alm/chargeback-alm-install-resources.configmap.yaml
$ kubectl -n tectonic-system patch deploy catalog-operator -p '{"spec":{"template":{"spec":{"containers":[{"name":"catalog-operator","volumeMounts":[{"name":"chargeback-alm-install-resources","mountPath":"/var/catalog_resources/chargeback"}]}],"volumes":[{"name":"chargeback-alm-install-resources","configMap":{"name":"chargeback-alm-install-resources"}}]}}}}'
```

Once these commands are run the `catalog-operator` pod should restart, after
which chargeback can be enabled via the Tectonic console.

Chargeback will only function in a namespace that has the `coreos-pull-secret`
installed, which by default is only `tectonic-system`. There is a script
available to help with copying the `coreos-pull-secret` into any other
namespace, which can be run with the following command:

```
./hack/copy-pull-secret.sh
```

## Verifying operation

Check the logs of the "chargeback" deployment, there should be no errors:

```
$ kubectl get pods -n $CHARGEBACK_NAMESPACE -l app=chargeback -o name | cut -d/ -f2 | xargs -I{} kubectl -n $CHARGEBACK_NAMESPACE logs {} -f
```

## Using chargeback

For instructions on using chargeback, please read the documentation on [using
chargeback](Using-chargeback.md)

## Storing data in S3

By default the data that chargeback collects and generates is ephemeral, and
will not survive restarts of the HDFS pod it deploys. To make this data
persistent by storing it in S3, follow the instructions in the [storing data in
S3 document](Storing-Data-In-S3.md) before proceeding with these instructions.
