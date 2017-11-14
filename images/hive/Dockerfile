FROM quay.io/coreos/chargeback-hadoop:latest

ENV HIVE_VERSION=2.3.2
ENV HIVE_HOME=/opt/hive
ENV PATH=$HIVE_HOME/bin:$PATH

WORKDIR /opt

# Install Hive
RUN set -x \
    && curl -fSLs http://apache.osuosl.org/hive/hive-$HIVE_VERSION/apache-hive-$HIVE_VERSION-bin.tar.gz | tar -xz && mv apache-hive-$HIVE_VERSION-bin hive

ENV HADOOP_CLASSPATH /opt/hive/hcatalog/share/hcatalog/*:${HADOOP_CLASSPATH}
ENV HIVE_AUX_JARS_PATH /usr/hdp/current/hive-server2/auxlib

# Install PostgreSQL JDBC
RUN set -x \
    && curl -fSLs -o $HIVE_HOME/lib/postgresql-jdbc.jar https://jdbc.postgresql.org/download/postgresql-9.4.1212.jar

# Configure JSON SerDe and AWS Jars
RUN mkdir -p /usr/hdp/current/hive-server2/auxlib && ln -s ${HADOOP_HOME}/share/hadoop/tools/lib/*aws* /opt/hive/lib
ADD json-serde-1.3.8-jar-with-dependencies.jar /usr/hdp/current/hive-server2/auxlib

ADD conf/* $HIVE_HOME/conf/

VOLUME /user/hive/warehouse
VOLUME /var/lib/hive

CMD ["/opt/hive/bin/hive", "--service", "metastore"]
