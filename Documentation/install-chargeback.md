<br>
<div class="alert alert-info" role="alert">
    <i class="fa fa-exclamation-triangle"></i><b> Note:</b> This documentation is for a pre-alpha feature. To register for the Chargeback Alpha program, email [tectonic-alpha-feedback@coreos.com](mailto:tectonic-alpha-feedback@coreos.com).
</div>

# Installing Chargeback

Chargeback consists of a few components:

- A Chargeback pod which aggregates Prometheus data and generates reports based
  on the collected usage information.
- Hive and Presto clusters, used by the Chargeback pod to perform queries on the
  collected usage data.

## Prerequisites

Chargeback requires the following components:

- A Tectonic 1.8 cluster.
- A StorageClass for dynamic volume provisioning. ([See configuring chargeback][configuring-chargeback] for more information.)
- A properly configured kubectl to access the Kubernetes cluster.

## Installation

Use the installation script to install Chargeback. Before running the script, customize the installation to define installation or data storage location.

### Modifying default values

Chargeback will install into an existing namespace. Without configuration, the
default is `chargeback`.

Chargeback also assumes it needs a docker pull secret to pull images, which
defaults to a secret named `coreos-pull-secret` in the `tectonic-system`
namespace.

To change either of these, override the following environment variables:

```
$ export CHARGEBACK_NAMESPACE=chargeback
$ export PULL_SECRET_NAMESPACE=tectonic-system
$ export PULL_SECRET=coreos-pull-secret
```

### Configuration

Before installing, please read [Configuring Chargeback][configuring-chargeback].
Some options may not be changed post-install. Be certain to configure these options, if desired, before installation.

### Run the install script

Chargeback can be installed with the following command:

```
$ ./hack/alm-install.sh
```

### Uninstall

To uninstall Chargeback and its related resources:

```
$ ./hack/alm-uninstall.sh
```

## Verifying operation

First, wait until the Chargeback Helm operator deploys all of the Chargeback components:

```
kubectl get pods -n $CHARGEBACK_NAMESPACE -l app=chargeback-helm-operator -o name | cut -d/ -f2 | xargs -I{} kubectl -n $CHARGEBACK_NAMESPACE logs -f {} -c chargeback-helm-operator
```

Once you see output like the following, the rest of the pods should be initializing:

```
Waiting for Tiller to become ready
Waiting for Tiller to become ready
Waiting for Tiller to become ready
Getting list of helm release configmaps to delete
No release configmaps to delete yet
Getting pod chargeback-helm-operator-7c4cf9849c-846g5 owner information
Owner references:
global:
  ownerReferences:
  - apiVersion: "apps/v1beta1"
    blockOwnerDeletion: false
    controller: true
    kind: "Deployment"
    name: chargeback-helm-operator
    uid: b2b9e446-f263-11e7-bdc3-06a45d7816a8
Setting ownerReferences for Helm release configmaps
No release configmaps to patch ownership of yet
Fetching helm values from secret chargeback-settings
Secret chargeback-settings does not exist, default values will be used
Running helm upgrade
Release "tectonic-chargeback" does not exist. Installing it now.
NAME:   tectonic-chargeback
LAST DEPLOYED: Fri Jan  5 22:00:01 2018
NAMESPACE: chargeback
STATUS: DEPLOYED

RESOURCES:

... the rest is omitted for brevity ...
```

Next check the logs of the `chargeback` deployment for errors:

```
$ kubectl get pods -n $CHARGEBACK_NAMESPACE -l app=chargeback -o name | cut -d/ -f2 | xargs -I{} kubectl -n $CHARGEBACK_NAMESPACE logs {} -f
```

## Using Chargeback

For instructions on using Chargeback, please see [Using Chargeback][using-chargeback].


[using-chargeback]: using-chargeback.md
[configuring-chargeback]: chargeback-config.md
