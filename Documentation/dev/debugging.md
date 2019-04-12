# Debugging Metering

Debugging metering can be fairly dif***REMOVED***cult if you do not know how to directly interact with the components the operator is speaking to.
Below, we detail how you can connect and query Presto and Hive, as well as view the dashboards of the Presto and HDFS components.

All of the follow commands assume you've set the `METERING_NAMESPACE` environment variable to the namespace your Metering installation is located in:

```
export METERING_NAMESPACE=your-namespace
```

## Get Reporting Operator Logs

The command below will follow the logs of the reporting-operator.

```
kubectl -n $METERING_NAMESPACE get pods  -l app=reporting-operator -o name | cut -d/ -f2 | xargs -o -I{} kubectl -n $METERING_NAMESPACE logs -f {}
```

## Query Presto using presto-cli

The following will open up an interactive presto-cli session where you can interactively query Presto. One thing to note is that this runs in the same container as Presto and launches an additional Java instance, meaning you may run into memory limits for the pod. If this occurs, you should increase the memory request & limits of the Presto pod.

```
kubectl -n $METERING_NAMESPACE exec -it "$(kubectl -n $METERING_NAMESPACE get pods -l app=presto,presto=coordinator -o name | cut -d/ -f2)"  -- /usr/local/bin/presto-cli --server localhost:8080 --catalog hive --schema default --user root
```

After the above command you should be given a prompt, where you can run queries. Use the `show tables;` query to view the list of tables:

```
presto:default> show tables;
                  Table
------------------------------------------
 operator_health_check
 datasource_aws_billing
 datasource_node_allocatable_cpu_cores
 datasource_node_allocatable_memory_bytes
 datasource_node_capacity_cpu_cores
 datasource_node_capacity_memory_bytes
 datasource_pod_limit_cpu_cores
 datasource_pod_limit_memory_bytes
 datasource_pod_request_cpu_cores
 datasource_pod_request_memory_bytes
 datasource_pod_usage_cpu_cores
 datasource_pod_usage_memory_bytes
 view_aws_ec2_billing_data
 view_node_cpu_allocatable
 view_node_cpu_capacity
 view_node_memory_allocatable
 view_node_memory_capacity
 view_pod_cpu_request_raw
 view_pod_cpu_usage_raw
 view_pod_memory_request_raw
 view_pod_memory_usage_raw
(22 rows)

Query 20180419_183245_12728_p64yz, FINISHED, 1 node
Splits: 18 total, 18 done (100.00%)
0:00 [22 rows, 986B] [110 rows/s, 4.83KB/s]

presto:default>
```

## Query Hive using beeline

The following will open up an interactive beeline session where you can interactively query Hive. One thing to note is that this runs in the same container as Hive and launches an additional Java instance, meaning you may run into memory limits for the pod. If this occurs, you should increase the memory request & limits of the Hive pod.

```
kubectl -n $METERING_NAMESPACE exec -it $(kubectl -n $METERING_NAMESPACE get pods -l app=hive,hive=server -o name | cut -d/ -f2) -c hiveserver2 -- beeline -u 'jdbc:hive2://127.0.0.1:10000/default;auth=noSasl'
```

After the above command you should be given a prompt, where you can run queries. Use the `show tables;` query to view the list of tables:

```
0: jdbc:hive2://127.0.0.1:10000/default> show tables;
+-------------------------------------------+
|                 tab_name                  |
+-------------------------------------------+
| operator_health_check                     |
| datasource_aws_billing                    |
| datasource_node_allocatable_cpu_cores     |
| datasource_node_allocatable_memory_bytes  |
| datasource_node_capacity_cpu_cores        |
| datasource_node_capacity_memory_bytes     |
| datasource_pod_limit_cpu_cores            |
| datasource_pod_limit_memory_bytes         |
| datasource_pod_request_cpu_cores          |
| datasource_pod_request_memory_bytes       |
| datasource_pod_usage_cpu_cores            |
| datasource_pod_usage_memory_bytes         |
| view_aws_ec2_billing_data                 |
| view_node_cpu_allocatable                 |
| view_node_cpu_capacity                    |
| view_node_memory_allocatable              |
| view_node_memory_capacity                 |
| view_pod_cpu_request_raw                  |
| view_pod_cpu_usage_raw                    |
| view_pod_memory_request_raw               |
| view_pod_memory_usage_raw                 |
+-------------------------------------------+
22 rows selected (1.725 seconds)
0: jdbc:hive2://127.0.0.1:10000/default>
```

## Port-forward to Presto web UI

The Presto web UI can be very useful when debugging.
It will show what queries are running, which have succeeded, and which queries have failed.

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
