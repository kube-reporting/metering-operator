# Configuring the Hive metastore

Hive metastore is responsible for storing all the [metadata](https://cwiki.apache.org/confluence/display/Hive/Design#Design-MetadataObjects) about the database tables we create in Presto and Hive.
By default, the metastore stores this information in a local embedded Derby database in a PersistentVolume attached to the pod.

Generally the default configuration of Hive metastore works for small clusters, but users may wish to improve performance or move storage requirements out of cluster by using a dedicated SQL database for storing the Hive metastore data.

## Configuring PersistentVolumes

Hive, by default requires one Persistent Volume to operate.

`hive-metastore-db-data` is the main PVC required by default.
This PVC is used by Hive metastore to store metadata about tables, such as table name, columns, and location.
Hive metastore is used by Presto and Hive server to lookup table metadata when processing queries.
In practice, it is possible to remove this requirement by using [MySQL](#use-mysql-for-the-hive-metastore-database) or [PostgreSQL](#use-postgresql-for-the-hive-metastore-database) for the Hive metastore database.

To install, Hive metastore requires that dynamic volume provisioning be enabled via a Storage Class, a persistent volume of the correct size must be manually pre-created, or that you use a pre-existing MySQL or PostgreSQL database.

## Configuring the Storage Class for Hive Metastore

To configure and specify a `StorageClass` for the hive-metastore-db-data PVC, specify the `StorageClass` in your MeteringConfig.
A example `StorageClass` section is included in [metastore-storage.yaml][metastore-storage-config].

Uncomment the `spec.hive.spec.metastore.storage.class` sections and replace the `null` in `class: null` value with the name of the StorageClass to use.
Leaving the value `null` will cause Metering to use the default StorageClass for the cluster.

## Configuring the Volume Sizes for Hive Metastore

Use [metastore-storage.yaml][metastore-storage-config] as a template and adjust the `size: "5Gi"` value to the desired capacity for the following sections:

- `spec.hive.spec.metastore.storage.size`

## Configuring the Database for Hive Metastore

By default, to make the installation easier, Metering configures Hive to use an embedded Java database called [Derby](https://db.apache.org/derby/#What+is+Apache+Derby%3F), however this is unsuitable for larger environments or metering installations where a lot of reports and metrics are being collected.

Currently two alternative options are available, MySQL and PostgreSQL, both of which have been tested with the metering-operator.

There are three configuration options you can use to control the database used by Hive metastore: `url` , `driver` , and `secretName`.

- `url`: the url of the MySQL or PostgreSQL instance. Examples are shown below.
- `driver`: configures the class name for the JDBC driver that will be used to store the hive metadata.
- `secretName`: the name of the secret which contains the base64 encrypted username and password for the database instance.

Before proceeding to following examples, you need to create a secret in the `$METERING_NAMESPACE` containing the base64 encrypted username and password combination to the database instance.

Through the command-line, you can replace the following commands wrapped in the `<...>` markers:

```bash
kubectl -n $METERING_NAMESPACE create secret generic <name of the secret> --from-literal=username=<database username> --from-literal=password=<database password>
```

### Using MySQL for the Hive Metastore database

**Note**: Metering cannot work with more recent versions of MySQL, which is being tracking in [BZ #1838802](https://bugzilla.redhat.com/show_bug.cgi?id=1838802). Instead, use the 5.7 version which has been tested.

```yaml
spec:
  hive:
    spec:
      config:
        db:
          url: "jdbc:mysql://mysql.example.com:3306/hive_metastore"
          driver: "com.mysql.jdbc.Driver"
          secretName: "REPLACEME"
```

You can pass additional JDBC parameters using the `spec.hive.spec.config.db.url`, for more details see [the MySQL Connector/J documentation](https://dev.mysql.com/doc/connector-j/5.1/en/connector-j-reference-configuration-properties.html).

## Using PostgreSQL for the Hive Metastore database

```yaml
spec:
  hive:
    spec:
      config:
        db:
          url: "jdbc:postgresql://postgresql.example.com:5432/hive_metastore"
          driver: "org.postgresql.Driver"
          secretName: "REPLACEME"
          autoCreateMetastoreSchema: false
```

You can pass additional JDBC parameters using the `url`, for more details see [the PostgreSQL JDBC driver documentation](https://jdbc.postgresql.org/documentation/head/connect.html#connection-parameters).

[metastore-storage-config]: ../manifests/metering-config/metastore-storage.yaml
