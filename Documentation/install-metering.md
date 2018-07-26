# Installing Operator Metering

Operator Metering is a collection of a few components:

- A Metering Operator Pod which aggregates Prometheus data and generates reports based
  on the collected usage information.
- Hive and Presto clusters, used by the Operator Metering Pod to perform queries on the
  collected usage data.

## Prerequisites

Operator Metering requires the following components:

- A Kubernetes v1.8 or newer cluster
- A StorageClass for dynamic volume provisioning. ([See configuring metering][configuring-metering] for more information.)
- A Prometheus installation within the cluster configured to do Kubernetes cluster-monitoring.
    - The prometheus-operator repository's [kube-prometheus instructions][kube-prometheus] are the standard way of achieving Prometheus cluster-monitoring.
    - At a minimum, we require kube-state-metrics, node-exporter, and built-in Kubernetes target metrics.
    - Openshift 3.10 or newer includes monitoring via the [openshift-montoring playbook](https://github.com/openshift/openshift-ansible/tree/master/playbooks/openshift-monitoring).
- 4GB Memory and 2 CPU Cores available cluster capacity.
- At least 1 node with 2GB available memory (the highest memory request for a single Operator Metering Pod)
    - Memory and CPU consumption may often be lower, but will spike when running reports, or collecting data for larger clusters.
- A properly configured kubectl to access the Kubernetes cluster.

## Configuration

Before continuing with the installation, please read [Configuring Operator Metering][configuring-metering].
Some options may not be changed post-install. Be certain to configure these options, if desired, before installation.

If you do not wish to modify the Operator Metering configuration, a minimal configuration example that doesn't override anything can be found in [default.yaml][default-config].

### Prometheus Monitoring Configuration

For installs into Openshift, ensuring Prometheus is installed can be done using Ansible. For installs into Tectonic, the manual installation method configures Metering to use the Prometheus that's installed by default into the tectonic-system namespace.

If you're not using Openshift or Tectonic, then you will need to use OLM or the manual install method. In this case if you are not using a [kube-prometheus][kube-prometheus] installation, or your Prometheus service is not named `prometheus-k8s` and in the `monitoring` namespace, then you must customize the [prometheus URL config option][configure-prometheus-url] before proceeding.

## Install Methods

There are multiple installation methods depending on your Kubernetes platform and the version of Operator Metering you want.

### Ansible via openshift-ansible

Using Ansible is the currently recommended approach for installing onto Openshift.
The [openshift-metering playbook][metering-playbook] is included in the [openshift-ansible repository][openshift-ansible] and can be used to install and configure operator metering on top of Openshift.
It properly handles installing the Metering with the correct configuration for communicating to Openshift Cluster Monitoring, as well as enables other options for better integration with Openshift.

Read the [Ansible installation guide][ansible-install] for full details on how to use the playbook and what the available parameters are.

### Operator Lifecycle Manager (OLM)

Using OLM for installation is the recommended approach for installing the most recent packaged release for Kubernetes.
In the future, using OLM on Openshift will be a supported option.

For instructions on installing using OLM follow the [OLM installation guide][olm-install].

### Manual install scripts

Manual installation is generally not recommended as it is always changing and relies on a local checkout of the operator-metering git repository.
The primary use for manual installation is for doing development work or installing an unreleased versions of components.

For instructions on installing using our manual install scripts follow the [manual installation guide][manual-install].

## Verifying operation

First, wait until the Metering Helm operator deploys all of the Metering components:

```
kubectl get pods -n $METERING_NAMESPACE -l app=metering-helm-operator -o name | cut -d/ -f2 | xargs -I{} kubectl -n $METERING_NAMESPACE logs -f {} -c metering-helm-operator
```

When output similar to the following appears, the rest of the Pods should be initializing:

```
Waiting for Tiller to become ready
Waiting for Tiller to become ready
Getting pod metering-helm-operator-b5f86788c-ks4zq owner information
Querying for Deployment metering-helm-operator
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
NAME                                      READY     STATUS    RESTARTS   AGE
hdfs-datanode-0                           1/1       Running   0          10m
hdfs-namenode-0                           1/1       Running   0          10m
hive-metastore-0                          1/1       Running   0          10m
hive-server-0                             1/1       Running   0          10m
metering-5c6c9d6cc5-7pzwv                 1/1       Running   1          10m
metering-helm-operator-79666787c5-z4d2h   2/2       Running   0          10m
presto-coordinator-54469ccb68-jfblb       1/1       Running   0          10m
```

Check the logs of the `metering` deployment for errors:

```
$ kubectl get pods -n $METERING_NAMESPACE -l app=metering -o name | cut -d/ -f2 | xargs -I{} kubectl -n $METERING_NAMESPACE logs {} -f
```

## Using Operator Metering

For instructions on using Operator Metering, please see [Using Operator Metering][using-metering].

[default-config]: ../manifests/metering-config/default.yaml
[using-metering]: using-metering.md
[configuring-metering]: metering-config.md
[configure-prometheus-url]: metering-config.md#prometheus-url
[kube-prometheus]: https://github.com/coreos/prometheus-operator/tree/master/contrib/kube-prometheus
[olm-install]: olm-install.md
[ansible-install]: ansible-install.md
[manual-install]: manual-install.md
[metering-playbook]: https://github.com/openshift/openshift-ansible/tree/master/playbooks/openshift-metering
[openshift-ansible]: https://github.com/openshift/openshift-ansible
