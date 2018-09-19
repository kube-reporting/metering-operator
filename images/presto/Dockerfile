FROM centos:7

RUN \
    yum install -y \
        java-1.8.0-openjdk \
        curl \
        less  \
        gpg \
        python \
        perl \
    && yum clean all \
    && rm -rf /tmp/* /var/tmp/*

ENV JAVA_HOME=/etc/alternatives/jre/

ENV PRESTO_VERSION 0.202
ENV PRESTO_HOME /opt/presto/presto-server-$PRESTO_VERSION
ENV TERM linux
ENV HOME /opt/presto

ENV PROMETHEUS_JMX_EXPORTER_VERSION 0.3.1
ENV PROMETHEUS_JMX_EXPORTER_JAR jmx_prometheus_javaagent-$PROMETHEUS_JMX_EXPORTER_VERSION.jar

RUN mkdir -p $PRESTO_HOME
RUN useradd presto -u 1003 -d /opt/presto
WORKDIR $PRESTO_HOME

# install presto server
RUN set -x \
    && curl -fSLs -o /tmp/presto.tar.gz https://repo1.maven.org/maven2/com/facebook/presto/presto-server/$PRESTO_VERSION/presto-server-$PRESTO_VERSION.tar.gz \
    && mkdir -p /opt/presto \
    && tar -zxf /tmp/presto.tar.gz -C /opt/presto \
    && rm /tmp/presto.tar.gz

# install presto cli
RUN set -x \
        && curl -fSLs -o /opt/presto/presto-server-$PRESTO_VERSION/presto-cli https://repo1.maven.org/maven2/com/facebook/presto/presto-cli/$PRESTO_VERSION/presto-cli-$PRESTO_VERSION-executable.jar \
        && chmod +x /opt/presto/presto-server-$PRESTO_VERSION/presto-cli \
        && ln /opt/presto/presto-server-$PRESTO_VERSION/presto-cli /usr/local/bin/presto-cli \
        && chmod 755 /usr/local/bin/presto-cli

# install jmx exporter
RUN set -x \
RUN set -x \
    && mkdir -p /opt/jmx_exporter \
    && curl -fSLs -o /opt/jmx_exporter/jmx_exporter.jar https://repo1.maven.org/maven2/io/prometheus/jmx/jmx_prometheus_javaagent/$PROMETHEUS_JMX_EXPORTER_VERSION/$PROMETHEUS_JMX_EXPORTER_JAR

COPY jmx-exporter-con***REMOVED***g.yml /opt/jmx_exporter/con***REMOVED***g.yml

# Con***REMOVED***gure JSON serialization
ADD json-serde-1.3.8-jar-with-dependencies.jar ${PRESTO_HOME}/plugin/hive-hadoop2/

COPY etc/ $PRESTO_HOME/etc/
COPY entrypoint.sh /usr/local/bin

# Default to using localhost for discovery
ENV PRESTO_CONF_discovery_uri http://localhost:8080

EXPOSE 8080

RUN \
    mkdir -p /var/presto/data && \
    chown -R 1003:0 /opt/presto /var/presto/data && \
    chmod -R 770 /var/presto/data /etc/passwd && \
    chmod -R 775 /opt/presto

VOLUME /var/presto/data
ENTRYPOINT ["entrypoint.sh"]
CMD ["bin/launcher", "run"]

USER 1003
