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
==> v1beta1/RoleBinding
NAME        AGE
chargeback  18s

==> v1/Service
NAME                 CLUSTER-IP    EXTERNAL-IP  PORT(S)                       AGE
chargeback           10.3.236.62   <none>       8080/TCP                      18s
hdfs-datanode        None          <none>       50010/TCP                     18s
hdfs-namenode        None          <none>       8020/TCP                      18s
hdfs-namenode-proxy  10.3.25.89    <none>       8020/TCP                      18s
hive                 10.3.132.150  <none>       9083/TCP,10000/TCP,10002/TCP  18s
presto               10.3.165.92   <none>       8080/TCP                      18s

==> v1beta1/Deployment
NAME        DESIRED  CURRENT  UP-TO-DATE  AVAILABLE  AGE
chargeback  1        1        1           0          18s
presto      1        1        1           0          18s

==> v1alpha1/ReportGenerationQuery
NAME                                         KIND
node-capacity-cpu                            ReportGenerationQuery.v1alpha1.chargeback.coreos.com
node-allocatable-cpu                         ReportGenerationQuery.v1alpha1.chargeback.coreos.com
node-cpu-usage                               ReportGenerationQuery.v1alpha1.chargeback.coreos.com
node-memory-usage                            ReportGenerationQuery.v1alpha1.chargeback.coreos.com
node-allocatable-memory                      ReportGenerationQuery.v1alpha1.chargeback.coreos.com
node-capacity-memory                         ReportGenerationQuery.v1alpha1.chargeback.coreos.com
pod-request-cpu-usage                        ReportGenerationQuery.v1alpha1.chargeback.coreos.com
pod-cpu-usage-by-namespace                   ReportGenerationQuery.v1alpha1.chargeback.coreos.com
pod-cpu-usage-by-node                        ReportGenerationQuery.v1alpha1.chargeback.coreos.com
pod-request-memory-usage                     ReportGenerationQuery.v1alpha1.chargeback.coreos.com
pod-memory-usage-by-node-with-usage-percent  ReportGenerationQuery.v1alpha1.chargeback.coreos.com
pod-memory-usage-by-namespace                ReportGenerationQuery.v1alpha1.chargeback.coreos.com
pod-memory-usage-by-node                     ReportGenerationQuery.v1alpha1.chargeback.coreos.com

==> v1alpha1/StorageLocation
local  StorageLocation.v1alpha1.chargeback.coreos.com

==> v1/ConfigMap
NAME               DATA  AGE
chargeback-config  8     18s

==> v1/PersistentVolumeClaim
NAME                    STATUS  VOLUME                                    CAPACITY  ACCESSMODES  STORAGECLASS  AGE
hive-metastore-db-data  Bound   pvc-dbe5006b-f263-11e7-bdc3-06a45d7816a8  5Gi       RWO          gp2           18s

==> v1/ServiceAccount
NAME        SECRETS  AGE
chargeback  1        18s

==> v1beta1/Role
NAME              AGE
chargeback-admin  18s

==> v1beta1/StatefulSet
NAME           DESIRED  CURRENT  AGE
hdfs-datanode  1        1        18s
hdfs-namenode  1        1        18s
hive           1        1        18s

==> v1alpha1/ReportDataSource
NAME                           KIND
node-allocatable-memory-bytes  ReportDataSource.v1alpha1.chargeback.coreos.com
node-allocatable-cpu-cores     ReportDataSource.v1alpha1.chargeback.coreos.com
node-capacity-memory-bytes     ReportDataSource.v1alpha1.chargeback.coreos.com
node-capacity-cpu-cores        ReportDataSource.v1alpha1.chargeback.coreos.com
pod-request-cpu-cores          ReportDataSource.v1alpha1.chargeback.coreos.com
pod-limit-cpu-cores            ReportDataSource.v1alpha1.chargeback.coreos.com
pod-limit-memory-bytes         ReportDataSource.v1alpha1.chargeback.coreos.com
pod-request-memory-bytes       ReportDataSource.v1alpha1.chargeback.coreos.com

==> v1alpha1/ReportPrometheusQuery
node-allocatable-cpu-cores     ReportPrometheusQuery.v1alpha1.chargeback.coreos.com
node-allocatable-memory-bytes  ReportPrometheusQuery.v1alpha1.chargeback.coreos.com
node-capacity-cpu-cores        ReportPrometheusQuery.v1alpha1.chargeback.coreos.com
node-capacity-memory-bytes     ReportPrometheusQuery.v1alpha1.chargeback.coreos.com
pod-limit-cpu-cores            ReportPrometheusQuery.v1alpha1.chargeback.coreos.com
pod-request-cpu-cores          ReportPrometheusQuery.v1alpha1.chargeback.coreos.com
pod-limit-memory-bytes         ReportPrometheusQuery.v1alpha1.chargeback.coreos.com
pod-request-memory-bytes       ReportPrometheusQuery.v1alpha1.chargeback.coreos.com

==> v1/Secret
NAME                TYPE    DATA  AGE
chargeback-secrets  Opaque  2     18s
presto-secrets      Opaque  2     18s
```

Next check the logs of the `chargeback` deployment for errors:

```
$ kubectl get pods -n $CHARGEBACK_NAMESPACE -l app=chargeback -o name | cut -d/ -f2 | xargs -I{} kubectl -n $CHARGEBACK_NAMESPACE logs {} -f
```

## Using Chargeback

For instructions on using Chargeback, please see [Using Chargeback][using-chargeback].


[using-chargeback]: using-chargeback.md
[configuring-chargeback]: configuration.md
