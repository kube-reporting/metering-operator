FROM quay.io/coreos/chargeback-hadoop:latest

ENV HIVE_VERSION=2.3.3
ENV HIVE_HOME=/opt/hive
ENV PATH=$HIVE_HOME/bin:$PATH

WORKDIR /opt

USER root
# Install Hive
RUN set -x \
    && curl -fSLs http://apache.osuosl.org/hive/hive-$HIVE_VERSION/apache-hive-$HIVE_VERSION-bin.tar.gz | tar -xz && mv apache-hive-$HIVE_VERSION-bin hive

ENV HADOOP_CLASSPATH /opt/hive/hcatalog/share/hcatalog/*:${HADOOP_CLASSPATH}
ENV HIVE_AUX_JARS_PATH /usr/hdp/current/hive-server2/auxlib

ENV POSTGRESQL_JDBC_JAR postgresql-42.2.2.jar
# Using mysql-connector-java-8.0.11 resulted in hive schema creation failing due to incorrect syntax.
ENV MYSQL_JDBC_VERSION mysql-connector-java-5.1.46
ENV MYSQL_JDBC_JAR $MYSQL_JDBC_VERSION.jar

# Install PostgreSQL JDBC
RUN set -x \
    && curl -fSLs -o "$HIVE_HOME/lib/$POSTGRESQL_JDBC_JAR" "https://jdbc.postgresql.org/download/$POSTGRESQL_JDBC_JAR"

# Install MySQL JDBC
RUN set -x \
    && curl -fSLs "https://dev.mysql.com/get/Downloads/Connector-J/$MYSQL_JDBC_VERSION.tar.gz" | tar -zx --strip-components=1 -C "$HIVE_HOME/lib" "$MYSQL_JDBC_VERSION/$MYSQL_JDBC_JAR"

# Configure JSON SerDe and AWS Jars
RUN mkdir -p /usr/hdp/current/hive-server2/auxlib && ln -s ${HADOOP_HOME}/share/hadoop/tools/lib/*aws* /opt/hive/lib
COPY json-serde-1.3.8-jar-with-dependencies.jar /usr/hdp/current/hive-server2/auxlib

COPY metastore.sh /opt/hive/bin/ext/metastore.sh
COPY hive-entrypoint.sh /usr/local/bin/hive-entrypoint.sh
COPY conf/* $HIVE_HOME/conf/

RUN \
    mkdir -p /var/lib/hive /user/hive/warehouse && \
    chown -R 1002:0 /opt /var/lib/hive /user/hive/warehouse && \
    chmod -R 770 /opt /var/lib/hive /user/hive/warehouse /etc/passwd

VOLUME /user/hive/warehouse /var/lib/hive
ENTRYPOINT ["hive-entrypoint.sh"]

USER 1002
