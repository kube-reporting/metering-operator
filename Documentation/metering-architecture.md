# Operator Metering Architecture

Operator metering is composed of 3 major components:

- `reporting-operator`: A Kubernetes operator written in Go that uses custom resources as the method in which users request reports.
- [Presto][presto-overview]: A distributed SQL Database written in Java, designed for doing big data and analytical queries.
- [Hive][hive-overview]: A data warehousing application that facilitates reading, writing, and managing data already residing in a distributed storage location.
  - Presto has a dependency on Hive, and uses Hive for keeping metadata about the data Presto is working with.

## How it works

At a high-level view, it may be helpful to view the metering operator in terms of events and reactions. This is because the metering operator's main responsibility is to interact with custom resources in Kubernetes. We can then characterize any changes made to these custom resources as an event, and the metering operator  reacts to these events as appropriate. 

Internally, Operator Metering uses a database called [Presto][presto-overview] to do analytical querying on collected data using SQL.
When we use terms like `tables`, `views`, `SQL`, `statement`, or `query` in this document, we're referring to them in the context of the Presto database, and we're using SQL as the primary method of doing analysis and reporting on the data that Operator Metering collects.

To briefly summarize how metering works, it watches for a number of custom resources for changes.
Most of these custom resources are what you might consider con***REMOVED***guration, and allow for developers or end-users to extend what kind of data the operator can access, collect, report on, and in the case of reports, how to calculate the reports.
Upon being noti***REMOVED***ed that a particular resource has been created, the operator uses all of these resources to create tables or views in Presto, and eventually executes user-de***REMOVED***ned SQL queries on the data it has access to, or has already collected, and stores the data for retrieval as a report in a CSV or JSON ***REMOVED***le later.

The end goal of all of this is that Operator Metering provides building blocks for doing custom reporting and metering on any data that we can store into Presto.
Currently this is primarily focused on making Prometheus metrics and billing data accessible, but in the future we can expect other integrations to be added that allow accessing data stored in other locations.

The rest of this document covers how each custom resource that metering consumes operates individually, and how they're related to each other.

## How the operator handles custom resources

There are 6 custom resources that Operator Metering de***REMOVED***nes that you need to be aware of:

- `StorageLocations`: Provides a place to store data. Is used by `ReportDataSources`, `Reports`.
- `ReportDataSources`: Controls what data is collected (Prometheus data, AWS billing data), and where it's stored.
- `ReportPrometheusQueries`: Controls how Prometheus data is collected.
- `ReportGenerationQueries`: Controls how we query the data we've collected. This is referenced by a `Report` to manage what it will be reporting on.
- `Reports`: Causes reports to be generated using the con***REMOVED***gured `ReportGenerationQuery` resource. This is the primary resource an end-user of Operator Metering would interact with.

In the sections below, we will cover the resources described above in more detail.

### StorageLocation

For user-docs containing a description of the ***REMOVED***elds, and examples, see [StorageLocations][storagelocations].

A `StorageLocation` roughly maps to a [connector in Presto][presto-connector], which is effectively how Presto adapts to other datasources like Hive (and thus HDFS), S3 which we support. This also means we can eventually connect to other datasources too, such as PostgreSQL.

In terms of Operator Metering, a `StorageLocation` is intended to abstract some of those details away and expose the minimum con***REMOVED***guration required to expose where data is actually persisted at.
Today there is the concept of an `local` StorageLocation which is hard-coded to mean an HDFS cluster within the Kubernetes namespace, and it has a lot of hard-coded assumptions on how to communicate with this HDFS cluster. There is also an `s3` StorageLocation which allows us to connect Presto to an AWS S3 bucket. In both cases, the data is persisted as RCBinary ***REMOVED***les in either S3 or HDFS via Presto.

### ReportDataSource

For user-docs containing a description of the ***REMOVED***elds, and examples, see [ReportDataSources][reportdatasources].

A `ReportDataSource` instructs the metering operator to create a database table with the purpose of either storing Prometheus metric data (in a speci***REMOVED***ed StorageLocation), or a table pointing to an S3 bucket containing the AWS Cost and Usage reports.

#### Promsum ReportDataSources

A `promsum` ReportDataSource con***REMOVED***gures the reporting-operator to periodically poll Prometheus for metrics.

When the ReportDataSource is created, the metering operator does the following:

- Checks if this `ReportDataSource` has a table created for it yet, and if not, creates the table.
  - The underlying storage for this Presto table is controlled using the `StorageLocation` con***REMOVED***guration in `spec.promsum.storage`, which controls what options to use when creating the Presto table, such as what Presto connector is used.

Additionally, in the background the reporting-operator is periodically listing all `ReportDataSources` and if the `promsum` section is speci***REMOVED***ed, does the following to attempt to poll Prometheus metrics for each:

- Checks if the table for this ReportDataSource exists, and if it doesn't, it will skip collecting any data until the next poll for the `ReportDataSource` again.
- Retrieves the query speci***REMOVED***ed in the `ReportPrometheusQuery` named by the ReportDataSources `spec.promsum.query` ***REMOVED***eld.
- Executes the Prometheus query against the Prometheus server.
- Stores the data received from the metric results into a Presto table.
  - Currently this is done using an `INSERT` query using Presto, but this is subject to change as other `StorageLocations` are added.

Currently, all `promsum` ReportDataSources are collected at the same time in parallel.
Metric resolution, and poll intervals are controlled at a global level on the metering operator via the `Metering` resource's `spec.reporting-operator.con***REMOVED***g` section.

#### AWSBilling ReportDataSources

An `awsBilling` ReportDataSource con***REMOVED***gures the reporting-operator to periodically scan the speci***REMOVED***ed AWS S3 bucket for [AWS Cost and Usage reports][AWS-billing].

When the ReportDataSource is created, the metering operator does the following:

- Checks if this `ReportDataSource` has a table created for it yet, and if not, creates the table.
  - In the case of an `awsBilling` ReportDataSource, the operator creates the table using Hive. When creating the table, it's con***REMOVED***gured to point at the S3 Bucket con***REMOVED***gured in the `spec.awsBilling` section and reads the gzipped CSV ***REMOVED***les (`.csv.gz`, the ***REMOVED***le format the cost and usage reports are in).

Additionally, in the background the reporting-operator is periodically listing all `ReportDataSources` and if the `awsBilling` section is speci***REMOVED***ed, does the following to con***REMOVED***gure the table partitions for each:

- Checks if the table for this ReportDataSource exists, and if it doesn't, it will skip con***REMOVED***guring the partitions of the table ReportDataSource, until the next poll for all `ReportDataSources` again.
- Scans the con***REMOVED***gured S3 bucket for the cost usage report JSON manifests. There is a manifest for each billing period, and each manifest contains information about what reports are the most up-to-date for a given billing period.
- For each manifest, the operator determines which S3 directory contains the most up to date report ***REMOVED***les for the billing period, and then creates a partition in the table, pointing at the directory containing the cost & usage reports. If the partition already exists for the speci***REMOVED***ed billing period, the operator will remove it and replace it with the more up-to-date S3 directory.

This results in a table that has multiple partitions in it. There will be one partition per AWS billing period, and each partition points to an S3 directory containing the most up-to-date billing information for that billing period.

By default, Operator Metering has an section in the `Metering` resource for con***REMOVED***guring an awsBilling `ReportDataSource`, so you generally shouldn't need to create one directly.
For more details on con***REMOVED***guring this read the [AWS billing correlation section in the Metering Con***REMOVED***guration doc][metering-aws-billing-conf].

### ReportPrometheusQuery

For user-docs containing a description of the ***REMOVED***elds, and examples, see [ReportPrometheusQueries][reportprometheusqueries].

A `ReportPrometheusQuery` is basically just a Prometheus query. All `ReportPrometheusQueries` in the namespace of the operator are available for use in a `ReportDataSource`.

### ReportGenerationQuery

For user-docs containing a description of the ***REMOVED***elds, and examples, see [ReportGenerationQueries][reportgenerationqueries].

When the metering operator sees a new `ReportGenerationQuery` in its namespace, it will check if the `spec.view.disabled` ***REMOVED***eld is true, and if it is, it doesn't do anything with these queries on creation.
If it's false, then it will create a database view.

### Report

For user-docs containing a description of the ***REMOVED***elds, and examples, see [Reports][reports].

When a `Report` is created, and it sees the creation event, it does the following to generate the results:

- Retrieves the `ReportGenerationQuery` for the Report.
- For each `ReportGenerationQuery` from the `spec.reportQueries` and `spec.dynamicReportQueries` ***REMOVED***elds, retrieve them, and then do the same for each of those queries until we have all the required `ReportGenerationQueries`.
- Validates the `ReportDataSources` that each query depends on had its table created.
- Updates the Report status to indicate we're beginning to generate the report.
- Evaluates the `ReportGenerationQuery` template, passing the StartPeriod & EndPeriod into the template context.
- Creates a database table using Hive named after the `Report`. This table is either con***REMOVED***gured to use HDFS or S3 based on the Report's StorageLocation.
- Executes the query using Presto.
- Updates the `Report` status that everything succeeded.

[presto-overview]: https://prestosql.io/docs/current/overview/use-cases.html
[hive-overview]: https://cwiki.apache.org/confluence/display/Hive/Home#Home-ApacheHive
[presto-connector]: https://prestosql.io/docs/current/overview/concepts.html#connector
[AWS-billing]: https://docs.aws.amazon.com/awsaccountbilling/latest/aboutv2/billing-reports-costusage.html
[metering-aws-billing-conf]: metering-con***REMOVED***g.md#aws-billing-correlation
[storagelocations]: storagelocations.md
[reportdatasources]: reportdatasources.md
[reportprometheusqueries]: reportprometheusqueries.md
[reportgenerationqueries]: reportgenerationqueries.md
[reports]: report.md
