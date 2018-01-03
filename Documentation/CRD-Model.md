# Chargeback CRD Model

Chargeback uses a few different CRDs for configuration. This document describes
the different CRDs, provides examples for each, and explains how they interact
with each other.

## `ReportPrometheusQuery`

The `ReportPrometheusQuery` object simply holds a Prometheus query and a name.

Example File: [manifests/custom-resources/prom-queries/pod-memory-usage.yaml](../manifests/custom-resources/prom-queries/pod-memory-usage.yaml)

## `ReportDataSource`

The `ReportDataSource` object lists `ReportPrometheusQuery`s, and lists a
location and for the results of these queries to be stored. When `chargeback`
runs, it lists all `ReportDataSource`s, runs all Prometheus queries listed by
each store, and saves the results into each location. This means that Prometheus
queries which are not pointed to by a `ReportDataSource` will not be run.

Example File: [manifests/custom-resources/datasources/pod-memory-usage.yaml](../manifests/custom-resources/datasources/pod-memory-usage.yaml)

S3 storage is also supported. An example of using it can be found in the [storing data in s3](Storing-Data-In-S3.md) document.

## `ReportGenerationQuery`

Each `ReportGenerationQuery` object is a different type of report that
Chargeback can generate. The object holds a Presto query that is used to convert
usage data (and potentially AWS billing data) into a report, and the data store
whose data shall be used for the query. Additionally the `ReportGenerationQuery`
object defines the columns that will be present in the produced report.

Example File: [manifests/custom-resources/report-queries/pod-memory-usage-by-node.yaml](../manifests/custom-resources/report-queries/pod-memory-usage-by-node.yaml)

## `Report`

The `Report` object is created by users to trigger reports being generated. The
status of a report, viewable through kubectl, will mark when a report is
finished or errors encountered while generating it.

The `Report` object holds a start and end time over which the report should be
generated, names a generation query to use, and provides a location for the
report to be written to.

Example File: [manifests/custom-resources/reports/pod-memory-usage-by-node.yaml](../manifests/custom-resources/reports/pod-memory-usage-by-node.yaml)
