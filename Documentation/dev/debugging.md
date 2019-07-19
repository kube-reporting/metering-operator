# Debugging Metering

Debugging metering can be fairly dif***REMOVED***cult if you do not know how to directly interact with the components the operator is speaking to.
Below, we detail how you can connect and query Presto and Hive, as well as view the dashboards of the Presto and HDFS components.
For debugging issues surrounding the metering-ansible-operator, see the [ansible-operator](#metering-ansible-operator) section.

All of the follow commands assume you've set the `METERING_NAMESPACE` environment variable to the namespace your Metering installation is located in:

```
export METERING_NAMESPACE=your-namespace
```

## Get Reporting Operator Logs

The command below will follow the logs of the reporting-operator.

```
kubectl -n $METERING_NAMESPACE logs "$(kubectl -n $METERING_NAMESPACE get pods -l app=reporting-operator -o name | cut -c 5-)" -c reporting-operator
```

## Query Presto using presto-cli

The following will open up an interactive presto-cli session where you can interactively query Presto. One thing to note is that this runs in the same container as Presto and launches an additional Java instance, meaning you may run into memory limits for the pod. If this occurs, you should increase the memory request & limits of the Presto pod. By default, Presto is con***REMOVED***gured to communicate using TLS, and you would need to run the following command in order to run Presto queries:
```
kubectl -n $METERING_NAMESPACE exec -it "$(kubectl -n $METERING_NAMESPACE get pods -l app=presto,presto=coordinator -o name | cut -d/ -f2)"  -- /usr/local/bin/presto-cli --server https://presto:8080 --catalog hive --schema default --user root --keystore-path /opt/presto/tls/keystore.pem
```

In the case where you disabled the top-level `spec.tls.enabled` key, you would instead need to run the command below:
```
kubectl -n $METERING_NAMESPACE exec -it "$(kubectl -n $METERING_NAMESPACE get pods -l app=presto,presto=coordinator -o name | cut -d/ -f2)"  -- /usr/local/bin/presto-cli --server localhost:8080 --catalog hive --schema default --user root
```

After the above command you should be given a prompt, where you can run queries. Use the `show tables from metering;` query to view the list of tables:

```
presto:default> show tables from metering;
                                 Table
------------------------------------------------------------------------
 datasource_your_namespace_cluster_cpu_capacity_raw
 datasource_your_namespace_cluster_cpu_usage_raw
 datasource_your_namespace_cluster_memory_capacity_raw
 datasource_your_namespace_cluster_memory_usage_raw
 datasource_your_namespace_node_allocatable_cpu_cores
 datasource_your_namespace_node_allocatable_memory_bytes
 datasource_your_namespace_node_capacity_cpu_cores
 datasource_your_namespace_node_capacity_memory_bytes
 datasource_your_namespace_node_cpu_allocatable_raw
 datasource_your_namespace_node_cpu_capacity_raw
 datasource_your_namespace_node_memory_allocatable_raw
 datasource_your_namespace_node_memory_capacity_raw
 datasource_your_namespace_persistentvolumeclaim_capacity_bytes
 datasource_your_namespace_persistentvolumeclaim_capacity_raw
 datasource_your_namespace_persistentvolumeclaim_phase
 datasource_your_namespace_persistentvolumeclaim_phase_raw
 datasource_your_namespace_persistentvolumeclaim_request_bytes
 datasource_your_namespace_persistentvolumeclaim_request_raw
 datasource_your_namespace_persistentvolumeclaim_usage_bytes
 datasource_your_namespace_persistentvolumeclaim_usage_raw
 datasource_your_namespace_persistentvolumeclaim_usage_with_phase_raw
 datasource_your_namespace_pod_cpu_request_raw
 datasource_your_namespace_pod_cpu_usage_raw
 datasource_your_namespace_pod_limit_cpu_cores
 datasource_your_namespace_pod_limit_memory_bytes
 datasource_your_namespace_pod_memory_request_raw
 datasource_your_namespace_pod_memory_usage_raw
 datasource_your_namespace_pod_persistentvolumeclaim_request_info
 datasource_your_namespace_pod_request_cpu_cores
 datasource_your_namespace_pod_request_memory_bytes
 datasource_your_namespace_pod_usage_cpu_cores
 datasource_your_namespace_pod_usage_memory_bytes
(32 rows)

Query 20190503_175727_00107_3venm, FINISHED, 1 node
Splits: 19 total, 19 done (100.00%)
0:02 [32 rows, 2.23KB] [19 rows/s, 1.37KB/s]

presto:default>
```

## Query Hive using beeline

The following will open up an interactive beeline session where you can interactively query Hive. One thing to note is that this runs in the same container as Hive and launches an additional Java instance, meaning you may run into memory limits for the pod. If this occurs, you should increase the memory request & limits of the Hive pod.

```
kubectl -n $METERING_NAMESPACE exec -it $(kubectl -n $METERING_NAMESPACE get pods -l app=hive,hive=server -o name | cut -d/ -f2) -c hiveserver2 -- beeline -u 'jdbc:hive2://127.0.0.1:10000/default;auth=noSasl'
```

After the above command you should be given a prompt, where you can run queries. Use the `show tables;` query to view the list of tables:

```
0: jdbc:hive2://127.0.0.1:10000/default> show tables from metering;
+----------------------------------------------------+
|                      tab_name                      |
+----------------------------------------------------+
| datasource_your_namespace_cluster_cpu_capacity_raw |
| datasource_your_namespace_cluster_cpu_usage_raw  |
| datasource_your_namespace_cluster_memory_capacity_raw |
| datasource_your_namespace_cluster_memory_usage_raw |
| datasource_your_namespace_node_allocatable_cpu_cores |
| datasource_your_namespace_node_allocatable_memory_bytes |
| datasource_your_namespace_node_capacity_cpu_cores |
| datasource_your_namespace_node_capacity_memory_bytes |
| datasource_your_namespace_node_cpu_allocatable_raw |
| datasource_your_namespace_node_cpu_capacity_raw  |
| datasource_your_namespace_node_memory_allocatable_raw |
| datasource_your_namespace_node_memory_capacity_raw |
| datasource_your_namespace_persistentvolumeclaim_capacity_bytes |
| datasource_your_namespace_persistentvolumeclaim_capacity_raw |
| datasource_your_namespace_persistentvolumeclaim_phase |
| datasource_your_namespace_persistentvolumeclaim_phase_raw |
| datasource_your_namespace_persistentvolumeclaim_request_bytes |
| datasource_your_namespace_persistentvolumeclaim_request_raw |
| datasource_your_namespace_persistentvolumeclaim_usage_bytes |
| datasource_your_namespace_persistentvolumeclaim_usage_raw |
| datasource_your_namespace_persistentvolumeclaim_usage_with_phase_raw |
| datasource_your_namespace_pod_cpu_request_raw    |
| datasource_your_namespace_pod_cpu_usage_raw      |
| datasource_your_namespace_pod_limit_cpu_cores    |
| datasource_your_namespace_pod_limit_memory_bytes |
| datasource_your_namespace_pod_memory_request_raw |
| datasource_your_namespace_pod_memory_usage_raw   |
| datasource_your_namespace_pod_persistentvolumeclaim_request_info |
| datasource_your_namespace_pod_request_cpu_cores  |
| datasource_your_namespace_pod_request_memory_bytes |
| datasource_your_namespace_pod_usage_cpu_cores    |
| datasource_your_namespace_pod_usage_memory_bytes |
+----------------------------------------------------+
32 rows selected (13.101 seconds)
0: jdbc:hive2://127.0.0.1:10000/default>
```

## Port-forward to Presto web UI

The Presto web UI can be very useful when debugging.
It will show what queries are running, which have succeeded, and which queries have failed.

Note: Due to client-side authentication being enabled in Presto by default, you won't be able to view the Presto web UI at this time.
However, you can specify `spec.tls.enabled: false` and stop there to disable TLS/auth entirely, or only con***REMOVED***gure Presto to work with TLS (`spec.presto.tls`) and not client-side authentication.

```
kubectl -n $METERING_NAMESPACE get pods  -l app=presto,presto=coordinator -o name | cut -d/ -f2 | xargs -I{} kubectl -n $METERING_NAMESPACE port-forward {} 8080
```

You can now open http://127.0.0.1:8080 in your browser window to view the Presto web interface.

## Port-forward to Hive web UI

```
kubectl -n $METERING_NAMESPACE port-forward hive-server-0 10002
```

You can now open http://127.0.0.1:10002 in your browser window to view the Hive web interface.


## Port-forward to hdfs

To the namenode:

```
kubectl -n $METERING_NAMESPACE port-forward hdfs-namenode-0 9870
```

You can now open http://127.0.0.1:9870 in your browser window to view the HDFS web interface.


To the ***REMOVED***rst datanode:

```
kubectl -n $METERING_NAMESPACE port-forward hdfs-datanode-0 9864
```

To check other datanodes, run the above command, replacing `hdfs-datanode-0` with the pod you want to view information on.

## Metering Ansible Operator
Metering uses the ansible operator to watch and reconcile resources in a cluster environment. When debugging a failed Metering install, it can be helpful to view the Ansible logs or status of your `MeteringCon***REMOVED***g` custom resource.

##### Accessing Ansible Logs
There are a couple of ways of accessing the Ansible logs depending on how you installed the Metering resources.

In a typical install, the Metering operator is deployed as a pod. In this case, we can simply check the logs of the `ansible` container within this pod:
```
kubectl -n $METERING_NAMESPACE logs $(kubectl -n $METERING_NAMESPACE get pods -l app=metering-operator -o name | cut -d/ -f2) -c ansible
```

Alternatively, you can view the logs of the `operator` container (replace `-c ansible` with `-c operator`) for condensed output.

If you are running the Metering operator locally (i.e. `make run-metering-operator-local`), there won't be a dedicated pod, and so you would need to run the following command ():
```
docker exec -it metering-operator bash -c 'tail -n +1 -f /tmp/ansible-operator/runner/metering.openshift.io/v1/MeteringCon***REMOVED***g/*/*/artifacts/*/stdout'
```

When tracking down a failed task, you may encounter this output:
```
changed: [localhost] => (item=None) => {"censored": "the output has been hidden due to the fact that 'no_log: true' was speci***REMOVED***ed for this result", "changed": true}
```

This is because we use the Ansible module, `no_log`, on output-extensive tasks (running helm template, creation of resources, etc.). If your install fails during the Helm templating task, you can specify `spec.logHelmTemplate: true` in your `MeteringCon***REMOVED***g` CR, which will enable logging for that task, and you can re-run your installation for more information on why it failed.

##### Checking the `MeteringCon***REMOVED***g` Status
It can be helpful to view the `.status` ***REMOVED***eld of your `MeteringCon***REMOVED***g` custom resource to debug any recent failures. You can do this with the following command:
```
kubectl -n $METERING_NAMESPACE get meteringcon***REMOVED***g operator-metering -o json | jq '.status'
```
