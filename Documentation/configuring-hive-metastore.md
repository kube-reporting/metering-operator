# Configuring the Hive metastore

Generally the default configuration of Hive metastore works for small clusters, but users may wish to improve performance or move storage requirements out of cluster by using a dedicated database for storing table metadata for Presto and Hive server.

## Use MySQL or Postgresql for the Hive Metastore database

By default to make installation easier Metering configures Hive to use an embedded Java database called [Derby](https://db.apache.org/derby/#What+is+Apache+Derby%3F), however this is unsuited for larger environments or metering installations with a lot of reports and metrics being collected.
Currently two alternative options are available, MySQL and Postgresql, both of which have been tested with operator metering.

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

Using Postgresql:

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

You can pass additional JDBC parameters using the `dbConnectionURL`, for more details see [the Postgresql JDBC driver documentation](https://jdbc.postgresql.org/documentation/head/connect.html#connection-parameters).

