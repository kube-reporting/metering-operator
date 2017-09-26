FROM quay.io/coreos/chargeback-hadoop:0.2

ENV HIVE_VERSION=2.1.1
ENV HIVE_HOME=/opt/hive
ENV PATH=$HIVE_HOME/bin:$PATH HADOOP_HOME=/opt/hadoop-$HADOOP_VERSION

WORKDIR /opt

# Install Hive
RUN curl -fSLs http://www-eu.apache.org/dist/hive/hive-$HIVE_VERSION/apache-hive-$HIVE_VERSION-bin.tar.gz | tar -xz && mv apache-hive-$HIVE_VERSION-bin hive

# Install PostgreSQL JDBC
RUN curl -fSLs -o $HIVE_HOME/lib/postgresql-jdbc.jar https://jdbc.postgresql.org/download/postgresql-9.4.1212.jar

# Add Serde JSON Serialization Jar
RUN mkdir -p /usr/hdp/current/hive-server2/auxlib
ADD json-serde-1.3.8-jar-with-dependencies.jar /usr/hdp/current/hive-server2/auxlib

ADD conf/* $HIVE_HOME/conf/
