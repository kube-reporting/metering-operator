# Operator Metering Architecture

Operator metering is composed of 3 major components:

- `reporting-operator`: A Kubernetes operator written in Go that uses custom resources as the way users request reports.
- [Presto][presto-overview]: A distributed SQL Database written in Java, designed for doing big data and analytical queries.
- [Hive][hive-overview]: A data warehousing application the facilitates reading, writing, and managing data already residing in a distributed storage location.
  - Presto has a dependency on Hive, and uses Hive for keeping metadata about the data Presto is working with.

## How it works

Internally, Operator Metering uses a database called [Presto][presto-overview] to do analytical querying on collected data using SQL.
This means that in this document when terms like `tables`, `views`, `SQL`, `statement`, or `query` are used, we're referring to these in the context of the Presto database, and we're using SQL as the primary method of doing analysis and reporting on the data that Operator Metering collects.

The best way to describe how metering works is to describe it in terms of events and reactions, because the primary interaction with Operator Metering is through Kubernetes custom resources, and the metering operator's job is to react to changes in these resources.

To briefly summarize how metering works, it watches for a number of custom resources for changes.
Most of these custom resources are what you might consider con***REMOVED***guration, and allow for developers or end-users to extend what kind of data the operator can access, collect, report on, and in the case of reports, how to calculate the reports.
Upon being noti***REMOVED***ed that a particular resource has been created, the operator uses all of these resources to create tables or views in Presto, and eventually execute user-de***REMOVED***ned SQL queries on the data it has access to, or has collected, storing the data for retrieval as a report CSV or JSON ***REMOVED***le later.

The end goal of all of this is that Operator Metering provides building blocks for doing custom reporting and metering on any data that we can store into Presto.
Currently this is primarily focused on making Prometheus metrics and billing data accessible, but in the future we can expect other integrations to be added that allow accessing data stored in other locations.

The rest of this document covers how each custom resource that metering consumes operates individually, and how they're related to each other.

## How the operator handles custom resources

There are 6 custom resources that Operator Metering de***REMOVED***nes that you need to be aware of:

- `StorageLocations`: Provides a place to store data. Is used by `ReportDataSources`, `Reports`.
- `ReportDataSources`: Controls what data is collected (Prometheus data, AWS billing data), and where it's stored
- `ReportPrometheusQueries`: Controls how Prometheus data is collected
- `ReportGenerationQueries`: Controls how we query the data we've collected and is what a `Report` references to control what it's reporting on.
- `Reports`: Causes reports to be generated using the `ReportGenerationQuery` con***REMOVED***gured. This is the primary resource an end-user of Operator Metering would interact with.

In the sections below, we will cover these resources in more detail.

### StorageLocation

For user-docs containing a description of the ***REMOVED***elds, and examples, see [StorageLocations][storagelocations].

A `StorageLocation` roughly maps to a [connector in Presto][presto-connector], which is effectively how Presto adapts to other datasources like Hive (and thus HDFS), S3 which we support. This also means we can eventually others too, such as PostgreSQL.

In terms of Operator Metering, a `StorageLocation` is intended to abstract some of those details away and expose the minimum con***REMOVED***guration required to expose where data is actually persisted at.
Today there is the concept of an `local` StorageLocation which is hard-coded to mean an HDFS cluster within the Kubernetes namespace, and it has a lot of hard-coded assumptions to how to communicate with this HDFS cluster. There is also an `s3` StorageLocation which allows connecting Presto to an S3 bucket. In both cases, the data is persisted as RCBinary ***REMOVED***les in either S3 or HDFS via Presto.

### ReportDataSource

For user-docs containing a description of the ***REMOVED***elds, and examples, see [ReportDataSources][reportdatasources].

A `ReportDataSource` instructs the metering operator to create a database table for either storing Prometheus metric data (in a speci***REMOVED***ed StorageLocation), or a table pointing at an S3 bucket containing AWS Cost and Usage reports.

#### Promsum ReportDataSources

A `promsum` ReportDataSource con***REMOVED***gures the reporting-operator to periodically poll Prometheus for metrics.

When the ReportDataSource is created, the metering operator does the following:

- Checks if this `ReportDataSource` has a table created for it yet, and if not, create the table.
  - The underlying storage for this Presto table is controlled using the `StorageLocation` con***REMOVED***guration in `spec.promsum.storage`, which controls what options to use when creating the Presto table, such what Presto connector is used.

Additionally, in the background the reporting-operator is periodically, listing all `ReportDataSources` and if the `promsum` section is speci***REMOVED***ed, does the following to attempt to poll Prometheus metrics for each:

- Check if the table for this ReportDataSource exists, if it doesn't, it will skip collecting any data, until the next poll for the `ReportDataSource` again.
- Retrieve the query speci***REMOVED***ed in the `ReportPrometheusQuery` named by the ReportDataSources `spec.promsum.query` ***REMOVED***eld.
- Execute the Prometheus query against the Prometheus server.
- After receiving the metrics results, it then stores the data into a Presto table.
  - Currently this is done using an `INSERT` query using Presto, but this is subject to change as other `StorageLocations` are added.

Currently all promsum ReportDataSources are collected at the same time in parallel.
Metric resolution, and poll interval is controlled at a global level on the metering operator via the `Metering` resource's `spec.reporting-operator.con***REMOVED***g` section.

#### AWSBilling ReportDataSources

An `awsBilling` ReportDataSource con***REMOVED***gures the reporting-operator to periodically scan the speci***REMOVED***ed AWS S3 bucket for [AWS Cost and Usage reports][AWS-billing].

When the ReportDataSource is created, the metering operator does the following:

- Checks if this `ReportDataSource` has a table created for it yet, and if not, create the table.
  - In the case of an `awsBilling` ReportDataSource, the operator creates the table using Hive. When creating the table, it is con***REMOVED***gured to point at the S3 Bucket con***REMOVED***gured in the `spec.awsBilling` section and to read gzipped CSV ***REMOVED***les (`.csv.gz`, the ***REMOVED***le format the cost and usage reports are in).

Additionally, in the background the reporting-operator is periodically, listing all `ReportDataSources` and if the `awsBilling` section is speci***REMOVED***ed, does the following to con***REMOVED***gure the table partitions for each:

- Check if the table for this ReportDataSource exists, if it doesn't, it will skip con***REMOVED***guring the partitions of the table ReportDataSource, until the next poll for all `ReportDataSources` again.
- Scan the con***REMOVED***gured S3 bucket for the cost usage report JSON manifests. There is a manifest for each billing period, and each manifest contains information about what reports are the most up-to-date for a given billing period.
- For each manifest, the operator determines which S3 directory contains the most up to date report ***REMOVED***les for the billing period, and then creates a partition in the table, pointing at the directory containing the cost & usage reports. If the partition already exists for the speci***REMOVED***ed billing period, the operator will remove it and replace it with the more up to date S3 directory.

This results in a table that has multiple partitions in it, one partition per AWS billing period, and each partition points to an S3 directory containing the most up to date billing information for that billing period.

By default, Operator Metering has an section in the `Metering` resource for con***REMOVED***guring an awsBilling `ReportDataSource`, so you generally shouldn't need to create one directly.
For more details on con***REMOVED***guring this read the [AWS billing correlation section in the Metering Con***REMOVED***guration doc][metering-aws-billing-conf].

### ReportPrometheusQuery

For user-docs containing a description of the ***REMOVED***elds, and examples, see [ReportPrometheusQueries][reportprometheusqueries].

A `ReportPrometheusQuery` is basically just a Prometheus query. All `ReportPrometheusQueries` in the namespace of the operator are available for use in a `ReportDataSource`.

### ReportGenerationQuery

For user-docs containing a description of the ***REMOVED***elds, and examples, see [ReportGenerationQueries][reportgenerationqueries].

When the metering operator sees a new `ReportGenerationQuery` in it's namespace, it will check if the `spec.view.disabled` ***REMOVED***eld is true, if it is, it doesn't do anything with these queries on creation.
If it's false, then it will create a database view.

### Report

For user-docs containing a description of the ***REMOVED***elds, and examples, see [Reports][reports].

When a `Report` is created, and it sees the creation event, it does the following to generate the results:

- Retrieve the `ReportGenerationQuery` for the Report.
- For each `ReportGenerationQuery` from the `spec.reportQueries` and `spec.dynamicReportQueries` ***REMOVED***elds, retrieve them, and then do the same for each of those queries until we have all the required `ReportGenerationQueries`
- Validate the `ReportDataSources` that each query depends on has it's table created.
- Update the Report status to indicate we're beginning to generate the report.
- Evaluate the `ReportGenerationQuery` template, passing the StartPeriod & EndPeriod into the template context.
- Create a database table using Hive named after the Report This table is either con***REMOVED***gured to use HDFS or S3 based on the Report's StorageLocation.
- Execute the query using Presto.
- Update the Report status that everything succeeded.

[presto-overview]: https://prestodb.io/docs/current/overview/use-cases.html
[hive-overview]: https://cwiki.apache.org/confluence/display/Hive/Home#Home-ApacheHive
[presto-connector]: https://prestodb.io/docs/current/overview/concepts.html#connector
[AWS-billing]: https://docs.aws.amazon.com/awsaccountbilling/latest/aboutv2/billing-reports-costusage.html
[metering-aws-billing-conf]: metering-con***REMOVED***g.md#aws-billing-correlation
[storagelocations]: storagelocations.md
[reportdatasources]: reportdatasources.md
[reportprometheusqueries]: reportprometheusqueries.md
[reportgenerationqueries]: reportgenerationqueries.md
[reports]: report.md
