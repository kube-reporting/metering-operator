# Presto Tables

A `PrestoTable` is a custom resource that represents a database table accessible from within Presto.

When created, a PrestoTable resource causes the reporting-operator to create a table within Presto according to the con***REMOVED***guration provided or can expose an existing table that already exists to the reporting-operator.

Please read the [Presto concepts][presto-concepts] documentation to gain an understanding of any Presto speci***REMOVED***c terminology used below.

## Fields

- `unmanaged`: If true this indicates the PrestoTable resource is referencing an existing table, and should not attempt to create or manage the table within Presto.
- `catalog`: The catalog within Presto the table to be created in, or the catalog the table should exist within if unmanaged. In many cases, this will be `hive`.
- `schema`: The schema within the Presto catalog for the table to created in, or the schema the table should exist in if unmanaged. If the catalog is `hive` then there will always be at least the `default` schema.
- `tableName`: The name of the table to create in Presto, or the name of an existing table if unmanaged.
- `columns`: A list of columns that match the schema of the of the PrestoTable.
- `properties`: Optional: A map containing string keys and values, where each key value pair is a table property for con***REMOVED***guring the table. The available properties depends on the `catalog` used, see the Presto documentation for information on available properties.
- `comment`: Optional: If speci***REMOVED***ed, sets a comment on the table. Comments are just arbitrary strings that have no meaning to Presto, but can be used to store arbitrary information about a table.
- `view`: Optional: If true, then `query` must also be set. When `view` is true it causes reporting-operator to create a view within Presto using `query` as the SELECT statement for creating the view. `columns` must still be set in order for other Resources to determine the correct schema for the table, despite the schema being determined from the query used.
- `createTableAs`: Optional: If true, then `query` must also be set. When `createTableAs` is true it causes reporting-operator to create a table within Presto using `query` as the SELECT statement for creating the table, using the schema of the query and the results as the content of the table. `columns` must still be set in order for other Resources to determine the correct schema for the table, despite the schema being determined from the query used.
- `query`: Optional: If non-empty, you must set `view` or `createTableAs` to true. Must be a SELECT query to be used for creating the table or view.

## Example PrestoTables

```
apiVersion: metering.openshift.io/v1alpha1
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

[presto-concepts]: https://prestosql.io/docs/current/overview/concepts.html
[presto-select]: https://prestodb.io/docs/current/sql/select.html
[presto-types]: https://prestosql.io/docs/current/language/types.html
[presto-functions]: https://prestodb.io/docs/current/functions.html
