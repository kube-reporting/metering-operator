# Configuring the Hive metastore

Hive metastore is responsible for storing all the metadata about the database tables we create in Presto and Hive.
By default, the metastore stores this information in a local embedded Derby database in a PersistentVolume attached to the pod.

Generally the default configuration of Hive metastore works for small clusters, but users may wish to improve performance or move storage requirements out of cluster by using a dedicated SQL database for storing the Hive metastore data.

## Configuring PersistentVolumes

Hive, by default requires one Persistent Volume to operate.

`hive-metastore-db-data` is the main PVC required by default.
This PVC is used by Hive metastore to store metadata about tables, such as table name, columns, and location.
Hive metastore is used by Presto and Hive server to lookup table metadata when processing queries.
In practice, it is possible to remove this requirement by using [MySQL](#use-mysql-for-the-hive-metastore-database) or [PostgreSQL](#use-postgresql-for-the-hive-metastore-database) for the Hive metastore database.

To install, Hive metastore requires that dynamic volume provisioning be enabled via a Storage Class, a persistent volume of the correct size must be manually pre-created, or that you use a pre-existing MySQL or PostgreSQL database.

### Configuring the Storage Class for Hive Metastore

To configure and specify a `StorageClass` for the hive-metastore-db-data PVC, specify the `StorageClass` in your MeteringConfig.
A example `StorageClass` section is included in [metastore-storage.yaml][metastore-storage-config].

Uncomment the `spec.hive.spec.metastore.storage.class` sections and replace the `null` in `class: null` value with the name of the StorageClass to use.
Leaving the value `null` will cause Metering to use the default StorageClass for the cluster.

### Configuring the volume sizes for Hive Metastore

Use [metastore-storage.yaml][metastore-storage-config] as a template and adjust the `size: "5Gi"` value to the desired capacity for the following sections:

- `spec.hive.spec.metastore.storage.size`

## Use MySQL for the Hive Metastore database

By default to make installation easier Metering configures Hive to use an embedded Java database called [Derby](https://db.apache.org/derby/#What+is+Apache+Derby%3F), however this is unsuited for larger environments or metering installations with a lot of reports and metrics being collected.
Currently two alternative options are available, MySQL and PostgreSQL, both of which have been tested with operator metering.

There are 4 configuration options you can use to control the database used by Hive metastore: `dbConnectionURL` , `dbConnectionDriver` , `dbConnectionUsername` , and `dbConnectionPassword`.

Using MySQL:

```
spec:
  presto:
    spec:
      hive:
        config:
          dbConnectionURL: "jdbc:mysql://mysql.example.com:3306/hive_metastore"
          dbConnectionDriver: "com.mysql.jdbc.Driver"
          dbConnectionUsername: "REPLACEME"
          dbConnectionPassword: "REPLACEME"
```

You can pass additional JDBC parameters using the `dbConnectionURL`, for more details see [the MySQL Connector/J documentation](https://dev.mysql.com/doc/connector-j/5.1/en/connector-j-reference-configuration-properties.html).

## Use PostgreSQL for the Hive Metastore database

```
spec:
  presto:
    spec:
      hive:
        config:
          dbConnectionURL: "jdbc:postgresql://postgresql.example.com:5432/hive_metastore"
          dbConnectionDriver: "org.postgresql.Driver"
          dbConnectionUsername: "REPLACEME"
          dbConnectionPassword: "REPLACEME"
```

You can pass additional JDBC parameters using the `dbConnectionURL`, for more details see [the PostgreSQL JDBC driver documentation](https://jdbc.postgresql.org/documentation/head/connect.html#connection-parameters).

[metastore-storage-config]: ../manifests/metering-config/metastore-storage.yaml
