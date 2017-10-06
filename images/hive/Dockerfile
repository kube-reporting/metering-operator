FROM quay.io/coreos/chargeback-hadoop:latest

ENV HIVE_VERSION=2.3.0
ENV HIVE_HOME=/opt/hive
ENV PATH=$HIVE_HOME/bin:$PATH

WORKDIR /opt

# Install Hive
RUN curl -fSLs http://www-eu.apache.org/dist/hive/hive-$HIVE_VERSION/apache-hive-$HIVE_VERSION-bin.tar.gz | tar -xz && mv apache-hive-$HIVE_VERSION-bin hive

# Install PostgreSQL JDBC
RUN curl -fSLs -o $HIVE_HOME/lib/postgresql-jdbc.jar https://jdbc.postgresql.org/download/postgresql-9.4.1212.jar

# Con***REMOVED***gure JSON SerDe and AWS Jars
RUN mkdir -p /usr/hdp/current/hive-server2/auxlib && ln -s ${HADOOP_HOME}/share/hadoop/tools/lib/*aws* /opt/hive/lib
ADD json-serde-1.3.8-jar-with-dependencies.jar /usr/hdp/current/hive-server2/auxlib

ADD conf/* $HIVE_HOME/conf/
