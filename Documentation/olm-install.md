# Installing Metering using Operator Lifecycle Manager (OLM)

Currently, installing Metering via OLM is only supported on OpenShift 4.2 and newer from the OpenShift Marketplace.
If you want to install Metering into a non-OpenShift Kubernetes cluster, use the [manual installation documentation][manual-install].

This procedure covers:
- Installing the metering-operator using the OperatorHub within the OpenShift web console
- Creating a Metering resource that defines the installation configuration for the rest of the Metering stack

## Installing the Metering Operator

Create a dedicated OpenShift project for Metering, and then install the Metering Operator:

1. Create a new project/namespace called *openshift-metering* using the OpenShift web console (navigate to **Administration > Namespaces > Create Namespace**) or the `oc` command:

```
oc create namespace openshift-metering
```

2. From the web console, click **Catalog > OperatorHub**, and search for *metering-ocp* to find the Metering Operator.

3. Click the Metering card to open its package description, then click **Install**.

4. In the **Create Operator Subscription** screen, select the *openshift-metering* namespace in the **A specific namespace on the cluster** drop-down, and specify your update channel and approval strategy. Click **Subscribe** to install the metering-operator into your selected namespace.

5. On the **Subscription Overview** screen, the **Upgrade status** indicates *1 installed* when the Metering Operator has finished installing. Click the *1 installed* (or *installed version*) link to view the ClusterServiceVersion overview for the metering-operator.

From the ClusterServiceVersion overview, you can create different resources related to Metering.

## Installing the Metering stack

Next, create a Metering resource that instructs the metering-operator to install the Metering stack in the namespace.

This resource holds all the top level configuration for each component (such as requests, limits, storage, etc.).

**IMPORTANT:**
There can only be one Metering resource in the namespace containing the metering-operator. Any other configuration is not supported.

1. From the web console, ensure you are on the ClusterServiceVersion overview page for the Metering project.
You can navigate to this page from **Catalog > Installed Operators**, then selecting the *Metering* operator.

2. Under **Provided APIs**, click **Create New** on the *Metering* card. This opens a YAML editor where you can define your Metering installation configuration.

3. Download the example [default.yaml][default-config] Metering resource and customize the YAML as desired. Enter your configuration into the YAML editor and click **Create**.

**NOTE:**
All supported configuration options are documented in [configuring metering][configuring-metering].

4. Navigate to **Workloads > Pods** and wait for your resources to be created and become ready.

5. Next, verify that the ReportDataSources are beginning to import data, indicated by a valid timestamp in the `EARLIEST METRIC` column (this may take a few minutes). We filter out the "-raw" ReportDataSources which don't import data:

```
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

**NOTE:**
The Metering documentation refers to `$METERING_NAMESPACE` in most examples; this value will be `openshift-metering` if you followed the above instructions to create the Metering project/namespace.

## Manual/CLI based OLM install

To learn more about how the OLM installation process works under the hood, or to use the CLI to install Metering via OLM, see the [manual OLM install documentation][manual-olm-install].

[manual-install]: manual-install.md
[manual-olm-install]: manual-olm-install.md
[configuring-metering]: metering-config.md
[default-config]: ../manifests/metering-config/default.yaml
[using-metering]: using-metering.md
