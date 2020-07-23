# Debugging Metering

Debugging metering can be fairly difficult if you do not know how to directly interact with the components the operator is speaking to.
Below, we detail how you can connect and query Presto and Hive, as well as view the dashboards of the Presto and HDFS components.
For debugging issues surrounding the metering-ansible-operator, see the [ansible-operator](#metering-ansible-operator) section.

**Note**: All of the following commands assume you've set the `METERING_NAMESPACE` environment variable to the namespace your Metering installation is located in:

```bash
export METERING_NAMESPACE=<your-namespace>
```

## View the Reporting Operator Logs

The command below will follow the logs of the reporting-operator.

```bash
kubectl -n $METERING_NAMESPACE logs "$(kubectl -n $METERING_NAMESPACE get pods -l app=reporting-operator -o name | cut -c 5-)" -c reporting-operator
```

## Query Presto using presto-cli

The following will open up an interactive [presto-cli](https://prestosql.io/docs/current/installation/cli.html) session where you can interactively query Presto. One thing to note is that this runs in the same container as Presto and launches an additional Java instance, meaning you may run into memory limits for the pod. If this occurs, you should increase the memory request & limits of the Presto pod. By default, Presto is configured to communicate using TLS, and you would need to run the following command in order to run Presto queries:

```bash
kubectl -n $METERING_NAMESPACE exec -it "$(kubectl -n $METERING_NAMESPACE get pods -l app=presto,presto=coordinator -o name | cut -d/ -f2)"  -- /usr/local/bin/presto-cli --server https://presto:8080 --catalog hive --schema default --user root --keystore-path /opt/presto/tls/keystore.pem
```

In the case where you disabled the top-level `spec.tls.enabled` key, you would need to run the command below:

```bash
kubectl -n $METERING_NAMESPACE exec -it "$(kubectl -n $METERING_NAMESPACE get pods -l app=presto,presto=coordinator -o name | cut -d/ -f2)"  -- /usr/local/bin/presto-cli --server localhost:8080 --catalog hive --schema default --user root
```

After running the above command, you should be given a prompt where you can run queries. Use the `show tables from hive.metering;` query to view the list of available tables from the `metering` hive catalog:

```bash
presto:default> show tables from hive.metering;
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

The following will open up an interactive [beeline](https://cwiki.apache.org/confluence/display/Hive/HiveServer2+Clients#HiveServer2Clients-Beeline%E2%80%93CommandLineShell) session where you can interactively query Hive. One thing to note is that this runs in the same container as Hive and launches an additional Java instance, meaning you may run into memory limits for the pod. If this occurs, you should increase the memory request and limits of the Hive server Pod.

```bash
kubectl -n $METERING_NAMESPACE exec -it $(kubectl -n $METERING_NAMESPACE get pods -l app=hive,hive=server -o name | cut -d/ -f2) -c hiveserver2 -- beeline -u 'jdbc:hive2://127.0.0.1:10000/default;auth=noSasl'
```

After running the above command, you should be given a prompt where you can run queries. Use the `show tables;` query to view the list of tables:

```bash
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
It will show what queries are running, which have succeeded and which queries have failed.

**Note**: Due to client-side authentication being enabled in Presto by default, you won't be able to view the Presto web UI.

However, you can specify `spec.tls.enabled: false` and stop there to disable TLS/auth entirely, or only configure Presto to work with TLS (`spec.presto.tls`), and not client-side authentication.

```bash
kubectl -n $METERING_NAMESPACE get pods  -l app=presto,presto=coordinator -o name | cut -d/ -f2 | xargs -I{} kubectl -n $METERING_NAMESPACE port-forward {} 8080
```

You can now open <http://127.0.0.1:8080> in your browser window to view the Presto web interface.

## Port-forward to Hive web UI

```bash
kubectl -n $METERING_NAMESPACE port-forward hive-server-0 10002
```

You can now open <http://127.0.0.1:10002> in your browser window to view the Hive web interface.

## Port-forward to hdfs

### Namenode Pod

```bash
kubectl -n $METERING_NAMESPACE port-forward hdfs-namenode-0 9870
```

You can now open <http://127.0.0.1:9870> in your browser window to view the HDFS namenode web interface.

### Datanode Pod(s)

```bash
kubectl -n $METERING_NAMESPACE port-forward hdfs-datanode-0 9864
```

To check other datanodes, run the above command, and replace `hdfs-datanode-0` with the datanode pod you want to view more information on.

## Metering Ansible Operator

Metering uses the ansible-operator to watch and reconcile resources in a cluster environment.

When debugging a failed Metering install, it can be helpful to view the Ansible logs or status of your `MeteringConfig` custom resource.

### Accessing Ansible Logs

There are a couple of ways of accessing the Ansible logs depending on how you installed the Metering.

By default, the Ansible logs are merged with the internal ansible-operator logs, so it can be a difficult to parse at times.

#### Viewing Ansible logs from the metering operator Pod

In a typical install, the Metering operator is deployed as a pod. In this case, we can simply check the logs of the `ansible` container within this pod:

```bash
kubectl -n $METERING_NAMESPACE logs $(kubectl -n $METERING_NAMESPACE get pods -l app=metering-operator -o name | cut -d/ -f2)
```

To alleviate the increased verbosity and suppress the internal ansible-operator logs, you can edit the metering-operator deployment, and add the following argument to the `operator` container:

```yaml
...
spec:
  container:
  - name: operator
    args:
    - "--zap-level=error"
...
```

#### Viewing Ansible logs from the operator container

Alternatively, you can view the logs of the `operator` container (replace `-c ansible` with `-c operator`) for less verbose, condensed output.

If you are running the Metering operator locally (i.e. via `make run-metering-operator-local`), then there won't be a dedicated pod and you would need to check the local container logs. Run the following, replacing `docker` with the container runtime that created the metering container:

```bash
docker exec -it metering-operator bash -c 'tail -n +1 -f /tmp/ansible-operator/runner/metering.openshift.io/v1/MeteringConfig/*/*/artifacts/*/stdout'
```

When tracking down a failed task, you may encounter this output:

```yaml
changed: [localhost] => (item=None) => {"censored": "the output has been hidden due to the fact that 'no_log: true' was specified for this result", "changed": true}
```

This is because we use the Ansible module, `no_log`, on output-extensive tasks (running helm template, creation of resources, etc.) through the metering-ansible-operator.

If your installation continues to fail during the Helm templating-related tasks, you can specify `spec.logHelmTemplate: true` in your `MeteringConfig` custom resource, which will enable logging for those tasks. After applying that change to your custom resource, wait until the metering-operator has progressed far enough in the Ansible role for more information on why the installation failed.

### Checking the `MeteringConfig` Status

It can be helpful to view the `.status` field of your `MeteringConfig` custom resource to debug any recent failures. You can do this with the following command:

```bash
kubectl -n $METERING_NAMESPACE get meteringconfig operator-metering -o json | jq '.status'
```

### Checking the metering-ansible-operator events

In order to view the progress of the metering-ansible-operator, like where in the reconciliation process is the operator currently at, or checking if an error was encountered, you can run the following:

```bash
kubectl -n $METERING_NAMESPACE get events --field-selector involvedObject.kind=MeteringConfig --sort-by='.lastTimestamp'
```

If you're in the process of upgrading Metering, and you want to monitor those events more closely, you can prepend the `watch` command:

```bash
$ watch kubectl -n $METERING_NAMESPACE get events --field-selector involvedObject.kind=MeteringConfig --sort-by='.lastTimestamp'
Every 2.0s: kubectl -n tflannag get events --field-selector involvedObject.kind=MeteringConfig --sort-by=.lastTimestamp                   localhost.localdomain: Thu Jun  4 11:16:35 2020

LAST SEEN   TYPE     REASON       OBJECT                             MESSAGE
16s         Normal   Validating   meteringconfig/operator-metering   Validating the user-provided configuration
8s          Normal   Started      meteringconfig/operator-metering   Configuring storage for the metering-ansible-operator
5s          Normal   Started      meteringconfig/operator-metering   Configuring TLS for the metering-ansible-operator
```
