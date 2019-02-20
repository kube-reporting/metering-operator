# Installing Operator Metering

Operator Metering is a collection of a few components:

- A Metering Operator Pod which aggregates Prometheus data and generates reports based
  on the collected usage information.
- Hive and Presto clusters, used by the Operator Metering Pod to perform queries on the
  collected usage data.

## Prerequisites

Operator Metering requires the following components:

- A Kubernetes v1.8 or newer cluster
- A StorageClass for dynamic volume provisioning. ([See con***REMOVED***guring metering][con***REMOVED***guring-metering] for more information.)
- A Prometheus installation within the cluster con***REMOVED***gured to do Kubernetes cluster-monitoring.
    - The prometheus-operator repository's [kube-prometheus instructions][kube-prometheus] are the standard way of achieving Prometheus cluster-monitoring.
    - At a minimum, we require kube-state-metrics, node-exporter, and built-in Kubernetes target metrics.
    - Openshift 3.10 or newer includes monitoring via the [openshift-montoring playbook](https://github.com/openshift/openshift-ansible/tree/master/playbooks/openshift-monitoring).
- 4GB Memory and 4 CPU Cores available cluster capacity.
- Minimum resources needed for the largest single pod is 2 GB of memory and 2 CPU cores.
    - Memory and CPU consumption may often be lower, but will spike when running reports, or collecting data for larger clusters.
- A properly con***REMOVED***gured kubectl to access the Kubernetes cluster.

## Con***REMOVED***guration

Before continuing with the installation, please read [Con***REMOVED***guring Operator Metering][con***REMOVED***guring-metering].
Some options may not be changed post-install. Be certain to con***REMOVED***gure these options, if desired, before installation.

If you do not wish to modify the Operator Metering con***REMOVED***guration, a minimal con***REMOVED***guration example that doesn't override anything can be found in [default.yaml][default-con***REMOVED***g].

### Prometheus Monitoring Con***REMOVED***guration

For Openshift 3.11 or later, Prometheus is installed by default through cluster monitoring in the openshift-monitoring namespace.

If you're not using Openshift, then you will need to use the manual install method.
In this case you must customize the [prometheus URL con***REMOVED***g option][con***REMOVED***gure-prometheus-url] before proceeding.

## Install Methods

There are multiple installation methods depending on your Kubernetes platform and the version of Operator Metering you want.

### Manual install scripts

Manual installation is generally the recommended install method unless OLM can be used, or if you need to run a custom version of metering rather than what OLM has available.
For instructions on installing using our manual install scripts follow the [manual installation guide][manual-install].

### Operator Lifecycle Manager

OLM is currently only supported on Openshift 4.0.
For instructions on installing using OLM follow the [OLM install guide][olm-install].

## Verifying operation

First, wait until the Metering Helm operator deploys all of the Metering components:

```
kubectl get pods -n $METERING_NAMESPACE -l app=metering-operator -o name | cut -d/ -f2 | xargs -I{} kubectl -n $METERING_NAMESPACE logs -f {} -c metering-operator
```

When output similar to the following appears, the rest of the Pods should be initializing:

```
Waiting for Tiller to become ready
Waiting for Tiller to become ready
Getting pod metering-operator-b5f86788c-ks4zq owner information
Querying for Deployment metering-operator
No values, using default values
Running helm upgrade for release operator-metering
Release "operator-metering" has been upgraded. Happy Helming!
LAST DEPLOYED: Fri Jan 26 19:18:34 2018
NAMESPACE: metering
STATUS: DEPLOYED

RESOURCES:

... the rest is omitted for brevity ...
```

Next, get the list of pods:

```
kubectl -n $METERING_NAMESPACE get pods
```

It may take a 2-3 minutes, but eventually all pods should have a status of `Running`:

```
NAME                                  READY     STATUS    RESTARTS   AGE
hdfs-datanode-0                       1/1       Running   0          9m
hdfs-namenode-0                       1/1       Running   0          9m
hive-metastore-0                      1/1       Running   0          9m
hive-server-0                         1/1       Running   0          9m
metering-operator-df67bb6cb-6d7vh     2/2       Running   1          11m
presto-coordinator-7b7b87ff49-bhzgg   1/1       Running   0          9m
reporting-operator-7cf77b68f9-l6jrd   1/1       Running   1          9m
```

Check the logs of the `reporting-operator` pod for errors:

```
$ kubectl get pods -n $METERING_NAMESPACE -l app=reporting-operator -o name | cut -d/ -f2 | xargs -I{} kubectl -n $METERING_NAMESPACE logs {} -f
```

## Using Operator Metering

For instructions on using Operator Metering, please see [using Operator Metering][using-metering].

[default-con***REMOVED***g]: ../manifests/metering-con***REMOVED***g/default.yaml
[using-metering]: using-metering.md
[con***REMOVED***guring-metering]: metering-con***REMOVED***g.md
[con***REMOVED***gure-prometheus-url]: metering-con***REMOVED***g.md#prometheus-url
[kube-prometheus]: https://github.com/coreos/prometheus-operator/tree/master/contrib/kube-prometheus
[olm-install]: olm-install.md
[ansible-install]: ansible-install.md
[manual-install]: manual-install.md
[metering-playbook]: https://github.com/openshift/openshift-ansible/tree/master/playbooks/openshift-metering
[openshift-ansible]: https://github.com/openshift/openshift-ansible
