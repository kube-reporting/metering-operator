# Manual Installation using Operator Lifecycle Manager (OLM)

Currently Metering via OLM is only supported on Openshift 4.3 or newer via the Openshift Marketplace.
If you want to install metering into a non-Openshift Kubernetes cluster, please use the [manual installation documentation][manual-install].

## Install

This will cover installing the metering-operator via the openshift-marketplace using `kubectl`/`oc` and will then create a Metering resource defining the configuration for the metering-operator to use to install the rest of the Metering stack.

### Install Metering Operator

Installing the metering-operator is done by creating a `Subscription` resource in a namespace with a `CatalogSource` containing the `metering` package.

Currently we require that metering is installed into its own namespace which requires some setup.

First, start by creating the `openshift-metering` namespace:

```bash
kubectl create ns openshift-metering
```

Next a `CatalogSourceConfig` needs to be added to the `openshift-marketplace` namespace.
This results in a `CatalogSource` containing the `metering` OLM package being created in the `openshift-metering` namespace.

Download the `metering-operators` [metering.catalogsourceconfig.yaml][metering-catalogsourceconfig] and install it into the `openshift-marketplace` namespace:

```bash
kubectl apply -n openshift-marketplace -f metering.catalogsourceconfig.yaml
```

After it is created, confirm a new `CatalogSource` is created in the `openshift-metering` namespace:

```bash
$ kubectl -n openshift-metering get catalogsources
NAME                                                     NAME     TYPE   PUBLISHER   AGE
installed-redhat-metering-operators-openshift-metering   Custom   grpc   Custom      2m56s
```

You should also see a pod with a name resembling `metering-operators-12345` in the `openshift-marketplace` namespace, this pod is the package registry pod OLM will use to get the `metering` package contents:

```bash
$ kubectl -n openshift-marketplace get pods
NAME                                                              READY   STATUS    RESTARTS   AGE
certified-operators-7f89948b85-mpzw6                              1/1     Running   0          3h36m
community-operators-7c7b9447cf-gzp78                              1/1     Running   0          3h36m
installed-redhat-metering-operators-openshift-metering-6d6hhfmg   1/1     Running   0          3m34s
marketplace-operator-7df66dbf67-99zql                             1/1     Running   2          3h38m
redhat-operators-7f6b7fd9d9-hffnv                                 1/1     Running   0          3h36m
```

Next, you will create an `OperatorGroup` in your namespace that restricts the namespaces the operator will monitor to the `openshift-metering` namespace.

Download the `metering-operators` [metering.operatorgroup.yaml][metering-operatorgroup] and install it into the `openshift-metering` namespace:

```bash
kubectl -n openshift-metering apply -f metering.operatorgroup.yaml
```

Lastly, create a `Subscription` to install the metering-operator.

Download the [metering.subscription.yaml][metering-subscription] and install it into the `openshift-metering` namespace:

```bash
kubectl -n openshift-metering apply -f metering.subscription.yaml
```

Once the subscription is created, OLM will create all the required resources for the metering-operator to run.
This step takes a bit longer than others, within a few minutes the metering-operator pod should be created.
Verify the metering-operator has been created and is running:

```bash
$ kubectl -n openshift-metering get pods
NAME                                READY   STATUS    RESTARTS   AGE
metering-operator-c7545d555-h5m6x   2/2     Running   0          32s
```

### Install Metering

Once the metering-operator is installed, we can now use it to install the rest of the Metering stack by configuring a `MeteringConfig` CR.

#### Configuration

All of the supported configuration options are documented in [configuring metering][configuring-metering].
In this document, we will refer to your configuration as your `metering.yaml`.

To start, download the example [default.yaml][default-config] `MeteringConfig` resource and save it as `metering.yaml`, and make any additional customizations you require.

#### Install Metering Custom Resource

To install all the components that make up Metering, install your `metering.yaml` into the cluster:

```bash
kubectl -n openshift-metering apply -f metering.yaml
```

Within a minute you should see resources being created in your namespace:

```bash
$ kubectl -n openshift-metering get pods
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

```bash
NAME                                  READY   STATUS    RESTARTS   AGE
hive-metastore-0                      2/2     Running   0          3m28s
hive-server-0                         3/3     Running   0          3m28s
metering-operator-68dd64cfb6-2k7d9    2/2     Running   0          5m17s
presto-coordinator-0                  2/2     Running   0          3m9s
reporting-operator-5588964bf8-x2tkn   2/2     Running   0          2m40s
```

Next, verify that the ReportDataSources are beginning to import data, indicated by a valid timestamp in the `EARLIEST METRIC` column (this may take a few minutes).
We filter out the "-raw" ReportDataSources which don't import data:

```bash
$ kubectl get reportdatasources -n $METERING_NAMESPACE | grep -v raw
NAME                                         EARLIEST METRIC        NEWEST METRIC          IMPORT START           IMPORT END             LAST IMPORT TIME       AGE
node-allocatable-cpu-cores                   2019-08-05T16:52:00Z   2019-08-05T18:52:00Z   2019-08-05T16:52:00Z   2019-08-05T18:52:00Z   2019-08-05T18:54:45Z   9m50s
node-allocatable-memory-bytes                2019-08-05T16:51:00Z   2019-08-05T18:51:00Z   2019-08-05T16:51:00Z   2019-08-05T18:51:00Z   2019-08-05T18:54:45Z   9m50s
node-capacity-cpu-cores                      2019-08-05T16:51:00Z   2019-08-05T18:29:00Z   2019-08-05T16:51:00Z   2019-08-05T18:29:00Z   2019-08-05T18:54:39Z   9m50s
node-capacity-memory-bytes                   2019-08-05T16:52:00Z   2019-08-05T18:41:00Z   2019-08-05T16:52:00Z   2019-08-05T18:41:00Z   2019-08-05T18:54:44Z   9m50s
persistentvolumeclaim-capacity-bytes         2019-08-05T16:51:00Z   2019-08-05T18:29:00Z   2019-08-05T16:51:00Z   2019-08-05T18:29:00Z   2019-08-05T18:54:43Z   9m50s
persistentvolumeclaim-phase                  2019-08-05T16:51:00Z   2019-08-05T18:29:00Z   2019-08-05T16:51:00Z   2019-08-05T18:29:00Z   2019-08-05T18:54:28Z   9m50s
persistentvolumeclaim-request-bytes          2019-08-05T16:52:00Z   2019-08-05T18:30:00Z   2019-08-05T16:52:00Z   2019-08-05T18:30:00Z   2019-08-05T18:54:34Z   9m50s
persistentvolumeclaim-usage-bytes            2019-08-05T16:52:00Z   2019-08-05T18:30:00Z   2019-08-05T16:52:00Z   2019-08-05T18:30:00Z   2019-08-05T18:54:36Z   9m49s
pod-limit-cpu-cores                          2019-08-05T16:52:00Z   2019-08-05T18:30:00Z   2019-08-05T16:52:00Z   2019-08-05T18:30:00Z   2019-08-05T18:54:26Z   9m49s
pod-limit-memory-bytes                       2019-08-05T16:51:00Z   2019-08-05T18:40:00Z   2019-08-05T16:51:00Z   2019-08-05T18:40:00Z   2019-08-05T18:54:30Z   9m49s
pod-persistentvolumeclaim-request-info       2019-08-05T16:51:00Z   2019-08-05T18:40:00Z   2019-08-05T16:51:00Z   2019-08-05T18:40:00Z   2019-08-05T18:54:37Z   9m49s
pod-request-cpu-cores                        2019-08-05T16:51:00Z   2019-08-05T18:18:00Z   2019-08-05T16:51:00Z   2019-08-05T18:18:00Z   2019-08-05T18:54:24Z   9m49s
pod-request-memory-bytes                     2019-08-05T16:52:00Z   2019-08-05T18:08:00Z   2019-08-05T16:52:00Z   2019-08-05T18:08:00Z   2019-08-05T18:54:32Z   9m49s
pod-usage-cpu-cores                          2019-08-05T16:52:00Z   2019-08-05T17:57:00Z   2019-08-05T16:52:00Z   2019-08-05T17:57:00Z   2019-08-05T18:54:10Z   9m49s
pod-usage-memory-bytes                       2019-08-05T16:52:00Z   2019-08-05T18:08:00Z   2019-08-05T16:52:00Z   2019-08-05T18:08:00Z   2019-08-05T18:54:20Z   9m49s
```

Once all pods are ready and you have verified that data is being imported, you can begin using Metering to collect and Report on your cluster.
For further reading on using metering, see the [using metering documentation][using-metering].

[manual-install]: manual-install.md
[metering-catalogsourceconfig]: ../manifests/deploy/openshift/olm/metering.catalogsourceconfig.yaml
[metering-operatorgroup]: ../manifests/deploy/openshift/olm/metering.operatorgroup.yaml
[metering-subscription]: ../manifests/deploy/openshift/olm/metering.subscription.yaml
[configuring-metering]: metering-config.md
[default-config]: ../manifests/metering-config/default.yaml
[using-metering]: using-metering.md
