# Installing Operator Metering

Operator Metering is a collection of a few components:

- A Metering Operator Pod which aggregates Prometheus data and generates reports based on the collected usage information.
- Presto and Hive, used by the Operator Metering Pod to perform queries on the collected usage data.

## Prerequisites

Operator Metering requires the following components:

- A Kubernetes v1.11 or newer cluster
- A StorageClass for dynamic volume provisioning. ([See configuring metering][configuring-metering] for more information.)
- A Prometheus installation within the cluster configured to do Kubernetes cluster-monitoring.
    - The [kube-prometheus instructions][kube-prometheus] are the standard way of achieving Prometheus cluster-monitoring.
    - At a minimum, we require [kube-state-metrics][kube-state-metrics], node-exporter, and built-in Kubernetes target metrics.
- 4GB Memory and 4 CPU Cores available cluster capacity.
- Minimum resources needed for the largest single pod is 2 GB of memory and 2 CPU cores.
    - Memory and CPU consumption may often be lower, but will spike when running reports, or collecting data for larger clusters.
- A properly configured [kubectl][kubectl-install] to access the Kubernetes cluster.

In addition, Metering **storage must be configured before proceeding**.
Available storage options are listed in the [configuring storage documentation][configuring-storage].

## Configuration

Before continuing with the installation, please read [Configuring Operator Metering][configuring-metering].
Some options may not be changed post-install, such as storage.
Be certain to configure these options, if desired, before installation.

### Prometheus Monitoring Configuration

For Openshift 3.11, 4.x and later, Prometheus is installed by default through cluster monitoring in the openshift-monitoring namespace, and the default configuration is already setup to use Openshift cluster monitoring.

If you're not using Openshift, then you will need to use the manual install method and customize the [prometheus URL config option][configure-prometheus-url] before proceeding.

## Install Methods

There are multiple installation methods depending on your Kubernetes platform and the version of Operator Metering you want.

### Operator Lifecycle Manager

Using OLM is the recommended option as it ensures you are getting a stable release that we have tested.

OLM is currently only supported on Openshift 4.3 or newer.
For instructions on installing using OLM follow the [OLM install guide][olm-install].

### Manual install scripts

Manual installation is generally not recommended unless OLM is unavailable, or if you need to run a custom version of metering rather than what OLM has available.
Please remember that the manual installation method does not guarantee any consistent version or upgrade experience.
For instructions on installing using our manual install scripts follow the [manual installation guide][manual-install].

## Verifying operation

First, wait until the Metering Ansible operator deploys all of the Metering components:

```
kubectl get pods -n $METERING_NAMESPACE -l app=metering-operator -o name | cut -d/ -f2 | xargs -I{} kubectl -n $METERING_NAMESPACE logs -f {}
```

This can potentially take a minute or two to complete, but when it's done, you should see log output similar the following:

```
{"level":"info","ts":1560984641.7900484,"logger":"runner","msg":"Ansible-runner exited successfully","job":"7911455193145848968","name":"operator-metering","namespace":"metering"}
```

If you see any failures, check out [debugging the Metering Ansible operator][ansible-debugging] section for troubleshooting tips.

Next, get the list of pods:

```
kubectl -n $METERING_NAMESPACE get pods
```

It may take a couple of minutes, but eventually all pods should have a status of `Running`:

```
NAME                                  READY   STATUS    RESTARTS   AGE
hive-metastore-0                      2/2     Running   0          3m
hive-server-0                         3/3     Running   0          3m
metering-operator-df67bb6cb-6d7vh     2/2     Running   0          4m
presto-coordinator-0                  2/2     Running   0          3m
reporting-operator-6fd758c9c7-crjsw   2/2     Running   0          3m
```

Check the logs of the `reporting-operator` pod for signs of any persistent errors:

```
$ kubectl get pods -n $METERING_NAMESPACE -l app=reporting-operator -o name | cut -d/ -f2 | xargs -I{} kubectl -n $METERING_NAMESPACE logs {} -f -c reporting-operator
```

Next, verify that the ReportDataSources are beginning to import data, indicated by a valid timestamp in the `EARLIEST METRIC` column (this may take a few minutes).
We filter out the "-raw" ReportDataSources which don't import data:

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

[default-config]: ../manifests/metering-config/default.yaml
[using-metering]: using-metering.md
[configuring-metering]: metering-config.md
[configuring-storage]: configuring-storage.md
[configure-prometheus-url]: configuring-reporting-operator.md#prometheus-connection
[kube-prometheus]: https://github.com/coreos/kube-prometheus
[olm-install]: olm-install.md
[manual-install]: manual-install.md
[storage-classes]: https://kubernetes.io/docs/concepts/storage/storage-classes/
[ansible-debugging]: dev/debugging.md#metering-ansible-operator
[kubectl-install]: https://kubernetes.io/docs/tasks/tools/install-kubectl/
[kube-state-metrics]: https://github.com/kubernetes/kube-state-metrics
