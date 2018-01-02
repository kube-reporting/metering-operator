# Chargeback CRD model

Chargeback uses CRDs for configuration. This document describes
the different CRDs, provides examples for each, and explains how they interact
with each other.

## ReportPrometheusQuery

The `ReportPrometheusQuery` object holds a Prometheus query and a name.

Example:
[manifests/custom-resources/prom-queries/pod-memory-usage.yaml][reportpromquery-object].

## ReportDataStore

The `ReportDataStore` object lists `ReportPrometheusQuerys` and a
location in which the results of these queries will be stored. When Chargeback
runs, it lists all `ReportDataStores`, runs all Prometheus queries listed by
each store, and saves the results into each location. This means that Prometheus
queries which are not pointed to by a `ReportDataStore` will not be run.

Example:
[manifests/custom-resources/datastores/pod-memory-usage.yaml][reportdatastore-object].

S3 storage is also supported. For more information, see [storing data in S3][storing-s3].

## ReportGenerationQuery

Each `ReportGenerationQuery` object is a different type of report that
Chargeback can generate. The object holds a Presto query that is used to convert
usage data (and potentially AWS billing data) into a report, and the data store
whose data will be used for the query. The `ReportGenerationQuery` object also
defines the columns that will be present in the produced report.

Example:
[manifests/custom-resources/report-queries/pod-memory-usage-by-node.yaml][reportgenquery-object].

## Report

The `Report` object is created by users to trigger report generation. The status
of a report, viewable through kubectl, will mark when a report is finished or
errors encountered while generating it.

The `Report` object holds a start and end time over which the report will be
generated, names a generation query to use, and provides a location for the
report to be written.

Example:
[manifests/custom-resources/reports/pod-memory-usage-by-node.yaml][report-object].


[reportpromquery-object]: ../manifests/custom-resources/prom-queries/pod-memory-usage.yaml
[reportdatastore-object]: ../manifests/custom-resources/datastores/pod-memory-usage.yaml
[storing-s3]: Storing-Data-In-S3.md
[reportgenquery-object]: ../manifests/custom-resources/reports/pod-memory-usage-by-node.yaml
[report-object]: ../manifests/custom-resources/reports/pod-memory-usage-by-node.yaml
