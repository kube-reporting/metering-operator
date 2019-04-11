# Installation using Operator Lifecycle Manager (OLM)

Currently Metering via OLM is only supported on Openshift 4.x via the Openshift Marketplace.
If you want to install metering into a non-Openshift Kubernetes cluster, please use the [manual installation documentation][manual-install].

## Install

This will cover installing the metering-operator using the OperatorHub within the Openshift admin console and will then create a Metering resource defining the configuration for the metering-operator to use to install the rest of the Metering stack.

### Install Metering Operator

The first step to installing metering is creating a dedicated Openshift Project for it.

Either using the Openshift admin console, or `oc` create a project/namespace called `metering`:

```
oc create project metering
```

Next, from the admin console, go to the Operator Hub, and search for `metering`, then click on the Metering card.

Next you should see a description of the Metering package, and at the top left an `Install` button should be available, click it.

In the following screen, specify the `metering` namespace in the `A specific namespace on the cluster` drop-down, adjust your update channel and update approval strategy and click `Subscribe` to install the metering-operator in to your selected namespace.

Next, you will wait until the Upgrade status under the Subscription Overview indicates `1 installed`.
Once the status indicates the operator is installed, click on the `1 installed` link or the `installed version` link to view the ClusterServiceVersion for the operator.

From the ClusterServiceVersion overview page, you can create different resources related to Metering.

### Install metering

Next we need to create a `Metering` resource which will instruct the metering-operator, to install the metering stack in the namespace.
This resource holds all the top level configuration for each component (requests, limits, storage, etc).
There must only be one Metering resource in the namespace containing metering-operator – any other configuration is not supported.

From the admin console, ensure your on on the ClusterServiceVersion overview page for Metering.
This can be reached by going to the installed operators page within the catalog, then clicking on Metering in the list.

In the provided APIs section of the ClusterServiceVersion overview page, click `Create New` on the `Metering` card.

From here, you will be prompted with a yaml editor to define your Metering installation configuration.
All of the supported configuration options are documented in [configuring metering][configuring-metering].
To start, download the example [default.yaml][default-config] Metering resource and make any additional customizations you require.

Once satisfied with your configuration, put it into the yaml editor within the admin console, and click the `create` button.

From there, go to the Pods page under the Workloads section in the sidebar, and wait for everything to get created and become ready.
Once all pods are ready, you can begin using Metering to collect and Report on your cluster.

For further reading on using metering, see the [using metering documentation][using-metering].
*Note:* The metering documentation will refer to `$METERING_NAMESPACE` in most of it's examples, this will be `metering` if you followed the above instructions to create the metering project/namespace.

## Manual/CLI based OLM install

If you want to know more about how the OLM installation process works under the hood, or want to use the CLI to install metering via OLM, then you can read the [manual OLM install documentation][manual-olm-install].

[manual-install]: manual-install.md
[manual-olm-install]: manual-olm-install.md
[configuring-metering]: metering-config.md
[default-config]: ../manifests/metering-config/default.yaml
[using-metering]: using-metering.md
