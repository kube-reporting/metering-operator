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
Most of these custom resources are what you might consider configuration, and allow for developers or end-users to extend what kind of data the operator can access, collect, report on, and in the case of reports, how to calculate the reports.
Upon being notified that a particular resource has been created, the operator uses all of these resources to create tables or views in Presto, and eventually executes user-defined SQL queries on the data it has access to, or has already collected, and stores the data for retrieval as a report in a CSV or JSON file later.

The end goal of all of this is that Operator Metering provides building blocks for doing custom reporting and metering on any data that we can store into Presto.
Currently this is primarily focused on making Prometheus metrics and billing data accessible, but in the future we can expect other integrations to be added that allow accessing data stored in other locations.

The rest of this document covers how each custom resource that metering consumes operates individually, and how they're related to each other.

## How the operator handles custom resources

There are 6 custom resources that Operator Metering defines that you need to be aware of:

- `StorageLocations`: Provides a place to store data. Is used by `ReportDataSources`, `Reports`, `HiveTables` and `PrestoTables`.
- `PrestoTables`: Defines a table in Presto. A PrestoTable can be "unmanaged" to expose a table that already exists, or "managed" to instruct metering to create a table as a result of the resource being created.
- `HiveTables`: Defines a table in Hive. When created, it instructs metering to create the table in Hive which causes the table to be available to Presto.
- `ReportDataSources`: Controls what data is available (Prometheus data, AWS billing data, Presto tables, views into other tables).
- `ReportQueries`: Controls how we query the data available within ReportDataSources. If referenced by a `Report` it will manage what it will be reporting on when the report is run. If it's referenced by a `ReportDataSource` it will instruct metering to create a view within Presto based on the rendered query.
- `Reports`: Causes reports to be generated using the configured `ReportQuery` resource. This is the primary resource an end-user of Operator Metering would interact with. Can be configured to run on a schedule.

In the sections below, we will cover the resources described above in more detail.

### StorageLocation

For user-docs containing a description of the fields, and examples, see [StorageLocations][storagelocations].

A `StorageLocation` roughly maps to a [connector in Presto][presto-connector], which is effectively how Presto adapts to other datasources like Hive (and thus HDFS), S3 which we support. This also means we can eventually support others too, such as PostgreSQL.

In terms of Operator Metering, a `StorageLocation` is intended to abstract some of those details away and expose the minimum configuration required to expose where data is actually persisted at.
Today there is the concept of an default StorageLocation which is uses an HDFS cluster within the metering namespace. You can also define a custom storage location to default to S3 or a ReadWriteMany PVC.
In both cases, the data is persisted as ORC files in either S3, HDFS, or a ReadWriteMany PVC via Presto.

### ReportDataSource

For user-docs containing a description of the fields, and examples, see [ReportDataSources][reportdatasources].

A `ReportDataSource` represents a database table data lives.
There are many types of ReportDataSources, with the most common being a "PrometheusMetricsImporter" ReportDataSource.

A PrometheusMetricsImporter ReportDataSource instructs the reporting-operator to create a database table for storing Prometheus metric data and to being the process of importing Prometheus metrics.
You can also define a AWSBilling ReportDataSource to create table pointing at an existing S3 bucket containing AWS Cost and Usage reports.
Additionally, there are ReportQueryView ReportDataSource's which create views in Presto, and PrestoTable ReportDataSource's which just expose an existing PrestoTable as a ReportDataSource.


#### PrometheusMetricsImporter ReportDataSources

A `PrometheusMetricsImporter` ReportDataSource configures the reporting-operator to periodically poll Prometheus for metrics.

When the ReportDataSource is created, the metering operator does the following:

- Checks if this `ReportDataSource` has a table created for it yet by checking the `status.tableRef` field.
- If the field is empty it creates the table by creating a `HiveTable` resource, and waiting for it's `status.tableName` to be set, indicating the table has been created, then records the HiveTable name as the `status.tableRef`.
  - The underlying storage for this table is controlled using the `StorageLocationRef` configuration in `spec.prometheusMetricsImporter.storage`, which controls what options to use when creating the Presto table, such as what Presto connector is used. Currently only Hive StorageLocation's are supported.

Additionally, in the background the reporting-operator is periodically listing all `ReportDataSources` and if the `spec.prometheusMetricsImporter` section is specified, does the following to attempt to poll Prometheus metrics for each:

- Checks if the table for this ReportDataSource exists, and if it doesn't, it will skip collecting any data until the next poll for the `ReportDataSource` again.
- Executes the Prometheus query contained in `spec.prometheusMetricsImporter.query` against the Prometheus server.
- Stores the data received from the metric results into a Presto table.
  - Currently this is done using an `INSERT` query using Presto, but this is subject to change as other `StorageLocations` are added.

Currently, multiple `PrometheusMetricsImporter` ReportDataSources are collected at the same time concurrently.
Metric resolution, and poll intervals are controlled at a global level on the metering operator via the `Metering` resource's `spec.reporting-operator.config` section.

#### AWSBilling ReportDataSources

An `awsBilling` ReportDataSource configures the reporting-operator to periodically scan the specified AWS S3 bucket for [AWS Cost and Usage reports][AWS-billing].

When the ReportDataSource is created, the reporting-operator does the following:

- Checks if this `ReportDataSource` has a table created for it yet, and if not, creates the table.
  - In the case of an `awsBilling` ReportDataSource, the operator creates the table using Hive. When creating the table, it's configured to point at the S3 Bucket configured in the `spec.awsBilling` section and reads the gzipped CSV files (`.csv.gz`, the file format the cost and usage reports are in).

Additionally, in the background the reporting-operator is periodically listing all `ReportDataSources` and if the `awsBilling` section is specified, does the following to configure the table partitions for each:

- Checks if the table for this ReportDataSource exists, and if it doesn't, it will skip configuring the partitions of the table ReportDataSource, until the next poll for all `ReportDataSources` again.
- Scans the configured S3 bucket for the cost usage report JSON manifests. There is a manifest for each billing period, and each manifest contains information about what reports are the most up-to-date for a given billing period.
- For each manifest, the operator determines which S3 directory contains the most up to date report files for the billing period, and then creates a partition in the table, pointing at the directory containing the cost & usage reports. If the partition already exists for the specified billing period, the operator will remove it and replace it with the more up-to-date S3 directory.

This results in a table that has multiple partitions in it. There will be one partition per AWS billing period, and each partition points to an S3 directory containing the most up-to-date billing information for that billing period.

By default, Operator Metering has an section in the `Metering` resource for configuring an awsBilling `ReportDataSource`, so you generally shouldn't need to create one directly.
For more details on configuring this read the [AWS billing correlation section in the Metering Configuration doc][metering-aws-billing-conf].


#### ReportReportQueryView View ReportDataSources

A `ReportQuery` ReportDataSource configures the reporting-operator to create a Presto view based on the query specified.

When the ReportDataSource is created, the reporting-operator:

- Checks if the `ReportDataSource` has a view created for it yet by checking the `status.tableRef` field.
- If it field is empty, it creates a view by creating a `PrestoTable` resource with `spec.view` set to true, and `spec.query` set the rendered value of the ReportQuery's spec.query field.
- It then waits for the PrestoTable's `spec.tableName` to be set, and updates the ReportDataSource's `spec.tableRef` to indicate the view was created successfully.

#### PrestoTable ReportDataSources

A `PrestoTable` ReportDataSource configures the reporting-operator to simply lookup the specified `PrestoTable` and associate it with the ReportDataSource.

When the ReportDataSource is created, the reporting-operator:

- Lookups the `PrestoTable` resource and verifies it's `status.tableName` is set.
- If the `status.tableName` is set then it will update the ReportDataSource's `spec.tableRef` to the tableName.

### ReportQuery

For user-docs containing a description of the fields, and examples, see [ReportQueries][reportqueries].

When the metering operator sees a new `ReportQuery` in its namespace, it will reprocess anything that depends on it (ReportDataSources or Reports).

### Report

For user-docs containing a description of the fields, and examples, see [Reports][reports].

When a `Report` is created, and it sees the creation event, it does the following to generate the results:

- Retrieve the `ReportQuery` for the Report.
- For each `ReportQuery`, `ReportDataSource` and `Report` input in the `spec.inputs`, retrieve and validate them, and any of their dependencies.
- Update the Report status to indicate we're beginning to generate the report.
- Evaluate the `ReportQuery` template, passing the StartPeriod & EndPeriod into the template context.
- Create a database table using Hive named after the Report. This table is either configured to use HDFS or S3 based on the Report's StorageLocation.
- Execute the query using Presto.
- Update the Report status that everything succeeded.

[presto-overview]: https://prestosql.io/docs/current/overview/use-cases.html
[hive-overview]: https://cwiki.apache.org/confluence/display/Hive/Home#Home-ApacheHive
[presto-connector]: https://prestosql.io/docs/current/overview/concepts.html#connector
[AWS-billing]: https://docs.aws.amazon.com/awsaccountbilling/latest/aboutv2/billing-reports-costusage.html
[metering-aws-billing-conf]: configuring-aws-billing.md
[storagelocations]: storagelocations.md
[reportdatasources]: reportdatasources.md
[reportqueries]: reportqueries.md
[reports]: reports.md
