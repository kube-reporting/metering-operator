# Manual Installation using Operator Lifecycle Manager (OLM)

Currently Metering via OLM is only supported on Openshift 4.2 or newer via the Openshift Marketplace.
If you want to install metering into a non-Openshift Kubernetes cluster, please use the [manual installation documentation][manual-install].

## Install

This will cover installing the metering-operator via the openshift-marketplace using `kubectl`/`oc` and will then create a Metering resource de***REMOVED***ning the con***REMOVED***guration for the metering-operator to use to install the rest of the Metering stack.

### Install Metering Operator

Installing the metering-operator is done by creating a `Subscription` resource in a namespace with a `CatalogSource` containing the `metering` package.

Currently we require that metering is installed into its own namespace which requires some setup.

First, start by creating the `openshift-metering` namespace:

```
kubectl create ns openshift-metering
```

Next, you will create an `OperatorGroup` in your namespace that restricts the namespaces the operator will monitor to the `openshift-metering` namespace.

Download the `metering-operators` [metering.operatorgroup.yaml][metering-operatorgroup] and install it into the `openshift-metering` namespace:

```
kubectl -n openshift-metering apply -f metering.operatorgroup.yaml
```

Lastly, create a `Subscription` to install the metering-operator.

Download the [metering.subscription.yaml][metering-subscription] and install it into the `openshift-metering` namespace:


```
kubectl -n openshift-metering apply -f metering.subscription.yaml
```

Once the subscription is created, OLM will create all the required resources for the metering-operator to run.
This step takes a bit longer than others, within a few minutes the metering-operator pod should be created.
Verify the metering-operator has been created and is running:

```
kubectl -n openshift-metering get pods
NAME                                READY   STATUS    RESTARTS   AGE
metering-operator-c7545d555-h5m6x   2/2     Running   0          32s
```

### Install Metering

Once the metering-operator is installed, we can now use it to install the rest of the Metering stack by con***REMOVED***guring a `MeteringCon***REMOVED***g` CR.

#### Con***REMOVED***guration

All of the supported con***REMOVED***guration options are documented in [con***REMOVED***guring metering][con***REMOVED***guring-metering].
In this document, we will refer to your con***REMOVED***guration as your `metering.yaml`.

To start, download the example [default.yaml][default-con***REMOVED***g] `MeteringCon***REMOVED***g` resource and save it as `metering.yaml`, and make any additional customizations you require.

#### Install Metering Custom Resource

To install all the components that make up Metering, install your `metering.yaml` into the cluster:

```
kubectl -n openshift-metering apply -f metering.yaml
```

Within a minute you should see resources being created in your namespace:

```
kubectl -n openshift-metering get pods
NAME                                  READY   STATUS              RESTARTS   AGE
hive-metastore-0                      1/2     Running             0          52s
hive-server-0                         2/3     Running             0          52s
metering-operator-68dd64cfb6-pxh8v    2/2     Running             0          2m49s
presto-coordinator-0                  2/2     Running             0          31s
reporting-operator-56c6c878fb-2zbhp   0/2     ContainerCreating   0          4s
```

It can take several minutes for all the pods to become "Ready".
Many pods rely on other components to function before they themselves can be considered ready.
Some pods may restart if other pods take too long to start, this is okay and can be expected during installation.

Eventually your pod output should look like this:

```
NAME                                  READY   STATUS    RESTARTS   AGE
hive-metastore-0                      2/2     Running   0          3m28s
hive-server-0                         3/3     Running   0          3m28s
metering-operator-68dd64cfb6-2k7d9    2/2     Running   0          5m17s
presto-coordinator-0                  2/2     Running   0          3m9s
reporting-operator-5588964bf8-x2tkn   2/2     Running   0          2m40s
```

Once all pods are ready, you can begin using Metering to collect and Report on your cluster.
For further reading on using metering, see the [using metering documentation][using-metering]

[manual-install]: manual-install.md
[metering-catalogsourcecon***REMOVED***g]: ../manifests/deploy/openshift/olm/metering.catalogsourcecon***REMOVED***g.yaml
[metering-operatorgroup]: ../manifests/deploy/openshift/olm/metering.operatorgroup.yaml
[metering-subscription]: ../manifests/deploy/openshift/olm/metering.subscription.yaml
[con***REMOVED***guring-metering]: metering-con***REMOVED***g.md
[default-con***REMOVED***g]: ../manifests/metering-con***REMOVED***g/default.yaml
[using-metering]: using-metering.md
