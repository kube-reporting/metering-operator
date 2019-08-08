# Presto Tables

A `PrestoTable` is a custom resource that represents a database table accessible from within Presto.

When a `PrestoTable` resource is created, the reporting-operator creates a table within Presto according to the con***REMOVED***guration provided, or can expose an existing table that already exists to the reporting-operator.

Please read the [Presto concepts][presto-concepts] documentation to gain an understanding of any Presto speci***REMOVED***c terminology used below.

## Fields
For more information on the speci***REMOVED***c ***REMOVED***elds listed below, follow the respective links to the Presto concepts documentation.
##### Required Fields:
- `unmanaged`: Indicates whether a PrestoTable resource is referencing an existing table, and if set to true, the reporting-operator should not attempt to create or manage the table within Presto.
- `catalog`: The [catalog](https://prestosql.io/docs/current/overview/concepts.html#catalog) the Presto table is created within, or the catalog the table should exist within if unmanaged. In many cases, this will be `hive`.
- `schema`: The [schema](https://prestosql.io/docs/current/overview/concepts.html#schema) within the Presto catalog for the table to created in, or the schema the table should exist in if unmanaged. If the catalog is `hive` then there will always be at least the `default` schema.
- `tableName`: The desired name of the [table](https://prestosql.io/docs/current/overview/concepts.html#table) to be created in Presto, or the name of an existing table if unmanaged.
- `columns`: A list of columns that match the schema of the PrestoTable. For each list item, you must specify a `name` ***REMOVED***eld, which is the name of an individual column for the Presto table, and a `type` ***REMOVED***eld, which corresponds to a valid [Presto type.](https://prestosql.io/docs/current/language/types.html)
##### Optional Fields:
- `query`: The SELECT [query](https://prestosql.io/docs/current/overview/concepts.html#query) used for creating the table or view. If `query` is non-empty, you must set either `view` or `createTableAs` to true. Continue onto the next section for [examples](#example-prestotables) that use the `query` ***REMOVED***eld.
- `view`: Controls whether the reporting-operator needs to create a view within Presto. If true, the reporting-operator uses the `query` ***REMOVED***eld as the SELECT statement used to create the table, using both the schemas of the query and results for the content of the table.
- `createTableAs`: Controls whether the reporting-operator needs to create a table within Presto using the `query` ***REMOVED***elds as the SELECT statement for creating the table.
- `properties`: A map containing string keys and values, where each key-value pair is a table property for con***REMOVED***guring the table. See the [Presto connector documentation](https://prestosql.io/docs/current/connector.html) to ***REMOVED***nd the available properties for a speci***REMOVED***c connector's catalog.
- `comment`: Sets a comment on the Presto table. Comments are just arbitrary strings that have no meaning to Presto, but can be used to store arbitrary information about a table.

## Example PrestoTables
##### Creating a table in Presto using a `SELECT` query:
```
apiVersion: metering.openshift.io/v1
kind: PrestoTable
metadata:
  name: example-baremetal-cost
spec:
  catalog: "hive"
  schema: "default"
  tableName: "example_baremetal_cost"
  columns:
  - name: "cost_per_gigabyte_hour"
    type: "double"
  - name: "cost_per_cpu_hour"
    type: "double"
  - name: "currency"
    type: "varchar"
  createTableAs: true
  query: |
    SELECT * FROM (
      VALUES (10.00, 50.00, 'USD')
    ) AS t (cost_per_gigabyte_hour, cost_per_cpu_hour, currency)
```

##### Creating a view in Presto:
```
apiVersion: metering.openshift.io/v1
kind: PrestoTable
metadata:
  name: example_cluster_cpu_capacity_view
spec:
  catalog: hive
  schema: metering
  tableName: "example_cluster_cpu_capacity_view"
  columns:
  - name: timestamp
    type: timestamp
  - name: dt
    type: varchar
  - name: cpu_cores
    type: double
  - name: cpu_core_seconds
    type: double
  - name: node_count
    type: double
  comment: '""'
  unmanaged: false
  view: true
  query: |
    SELECT
      "timestamp",
      dt,
      sum(node_capacity_cpu_cores) as cpu_cores,
      sum(node_capacity_cpu_core_seconds) as cpu_core_seconds,
      count(*) AS node_count
    FROM hive.metering.example_node_cpu_capacity_view
    GROUP BY "timestamp", dt
```

[presto-concepts]: https://prestosql.io/docs/current/overview/concepts.html
[presto-select]: https://prestodb.io/docs/current/sql/select.html
[presto-types]: https://prestosql.io/docs/current/language/types.html
[presto-functions]: https://prestodb.io/docs/current/functions.html
