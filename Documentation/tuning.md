# Resource Tuning

Often times the default resource requests/limits set by Metering aren't suf***REMOVED***cient for larger clusters with more resources changing and producing more metrics.
Without proper resource limits, pods are unable to correctly operate without experiencing pod restarts or crashes (often due to the pod being OOMKilled).

Here's a common list of reasons you may need to increase your resource limits:

- Larger clusters. If you're running more than 10 nodes, it's likely the defaults will insuf***REMOVED***cient. The defaults are set low to make it easier to install onto smaller clusters.
- Pods being OOMKilled. Check if any pod has a restart count that is increasing over time. Use `kubectl describe pod` to determine why it restarted. If it's due to OOMKilled, then this pod needs more memory.
- High cluster activity. If you're running on a cluster with high activity in terms of pods being deleted and recreated, this will produce more metrics, and result in higher resource requirements than relatively idle/unchanging clusters.
- Storage. By default, if you're using HDFS, the storage requested is 5Gi, which will only store a few months of data on a smaller cluster.
- Performance. More memory for most of these components means they can do more work at once. Additionally, additional replicas of speci***REMOVED***c components can decrease the time it takes to perform speci***REMOVED***c tasks such as generating a report.


## Default resource requests and limits

An example Metering con***REMOVED***guration [default-resource-limits.yaml][default-resource-limits] contains the default resource request and limit values for CPU, memory, and storage for every component that Metering installs by default.
This can be a good starting point if you wish to experiment with different values starting with the defaults.

## Recommended values

Besides the default limits example, we also provide guidelines and examples of what a preferred set of values would look like on a larger cluster.

Here are a few guidelines:

- Presto, hive, and hdfs are all written in Java, and thus tend to consume more memory than other applications.
- Metering is naturally batch oriented and can be bursty in resources, and thus there are many idle periods. Without proper autoscaling, it may be necessary to over-provision resources if you want optimal performance, or ***REMOVED***nd that applications crash without more resources.
- Metering is designed to run at even the largest scale environments like Openshift online, which has clusters with over 5000 namespaces, and over 10,000 pods. It works well on smaller clusters too, but the cost of being able to running at these larger scales can mean the metering stack isn't as easily tuned for a smaller footprint.

In general, we can only provide guidelines based on the environments we run, so we always recommend using your monitoring stack in addition to our documentation to determine what resource limits make sense for you.
As we work to improve metering, over time we hope to address more of these concerns as part of the metering-operator by leveraging Kubernetes pod autoscaling, jobs, and newer features in the various components we leverage.

We recommend starting with our [recommended-resources-limit.yaml][recommended-resource-limits] example con***REMOVED***guration, and adjusting it based on monitoring data over the period of a few weeks. It's much easier to start with higher values and lower them until you see pods restarting or when monitoring shows pods reaching their resource limits.

## General Advice

Below is some generally useful information to know when trying to tune the various components and understand _why_ we suggest speci***REMOVED***c values in our [recommended-resources-limit.yaml][recommended-resource-limits] example con***REMOVED***guration.

### reporting-operator

The reporting-operator's most resource intensive task is querying Prometheus for metrics, storing the results in memory, converting them into a SQL insert statement, and batching those inserts to Presto.
Performing imports is mostly a memory and IO intensive task, and thus increasing memory beyond defaults can help when dealing with a lot of metrics on larger or highly active clusters.

### Presto

Presto is the main component that does real computation intensive work.
Additionally, it processes everything in memory when a query comes in, meaning its primary resource constraint will often be memory, and CPU can be a problem when there's a lot of garbage collections occurring while processing a query.

Presto itself doesn't store the data, and instead it accesses the data on demand when a query is executed.
This means network is often a bottleneck, and can be something you may need to consider.

When inserting data directly through Presto, this action is handled by the single Presto coordinator.
This means throughput of data collection is bottlenecked by the coordinator's ability to write these results to storage.
Reports on the other hand work on data already in storage, meaning Presto will split the work up across all available Presto pods.

### Hive

Hive metastore is often fairly idle, and is queried everytime Presto must interact with a table.
These interactions are all fairly light weight, and thus the amount of resources the metastore requires isn't very large.
However, it does retain quite a bit of information in memory, and can require more memory than the default values provide when it's been running for a longer period of time.

The Hive server component is very lightly used and is only interacted with when creating new tables, or modifying partitions of an AWS Billing ReportDataSource table, and thus requires fewer resources.

### HDFS

By default, Metering installs HDFS for storage. While the amount of data isn't large in most cases, you do want to consider running multiple HDFS datanode replicas for redundancy.
Because HDFS is commonly accessed by Presto when data is being stored and queried, it's also typically a component which consumes more resources over time.
As the amount of data stored in HDFS grows, the overhead for the hdfs-namenode is increased as it must maintain more metadata about all blocks stored in the HDFS cluster.

For this reason, we often support using Amazon S3 for storage to alleviate the need to scale HDFS.

[default-resource-limits]: ../manifests/metering-con***REMOVED***g/default-resource-limits.yaml
[recommended-resource-limits]: ../manifests/metering-con***REMOVED***g/recommended-resource-limits.yaml
