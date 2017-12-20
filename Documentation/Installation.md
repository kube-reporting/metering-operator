# Installing Chargeback

Chargeback consists of a few components:

- A chargeback pod which aggregates Prometheus data and generates reports based
  on the collected usage information.
- Hive and Presto clusters, used by the chargeback pod to perform queries on the
  collected usage data.

## Prerequisites

In order to install and use chargeback the following components will be
necessary:

- A Tectonic 1.8 cluster
- A StorageClass for dynamic volume provisioning ([see con***REMOVED***guring chargeback](con***REMOVED***guration.md) for details.)
- A properly con***REMOVED***gured kubectl to access the Kubernetes cluster.

## Installation

To install Chargeback you can run our installation script.
Before running the script, you can customize the installation if you want to
customize where Chargeback is installed, or if you want to change where it
stores data, etc.

### Modifying default values

Chargeback will install into an existing namespace. Without con***REMOVED***guration, the
default is currently `chargeback`.

Chargeback also assumes it needs a docker pull secret to pull images, which
defaults to a secret named `coreos-pull-secret` in the `tectonic-system`
namespace.

To change either of these, override the following environment variables
(defaults are used in the example):

```
$ export CHARGEBACK_NAMESPACE=chargeback
$ export PULL_SECRET_NAMESPACE=tectonic-system
$ export PULL_SECRET=coreos-pull-secret
```

### Con***REMOVED***guration

Before installing, please read about [con***REMOVED***guring chargeback](con***REMOVED***guration.md).
Some options do not support being changed post-install, so you may wish to
adjust some con***REMOVED***guration options, before continuing with the install.

### Run the install script

Chargeback can now be installed with the following command:

```
$ ./hack/alm-install.sh
```

### Uninstall

To uninstall chargeback, and it's related resources:

```
$ ./hack/alm-uninstall.sh
```

## Verifying operation

Check the logs of the "chargeback" deployment, there should be no errors:

```
$ kubectl get pods -n $CHARGEBACK_NAMESPACE -l app=chargeback -o name | cut -d/ -f2 | xargs -I{} kubectl -n $CHARGEBACK_NAMESPACE logs {} -f
```

## Using chargeback

For instructions on using chargeback, please read the documentation on [using chargeback](Using-chargeback.md).
