FROM openjdk:8-slim

ENV DEBIAN_FRONTEND noninteractive

RUN \
    apt-get update \
    && apt-get install -y --no-install-recommends \
        curl \
        less  \
        gpg \
        python \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

# ENV JAVA_HOME=/usr/lib/jvm/java-8-openjdk-amd64/

ENV PRESTO_VERSION 0.188
ENV PRESTO_HOME /opt/presto/presto-server-$PRESTO_VERSION
ENV TERM linux

RUN set -x \
    && curl -fSLs -o /tmp/presto.tar.gz http://repo1.maven.org/maven2/com/facebook/presto/presto-server/$PRESTO_VERSION/presto-server-$PRESTO_VERSION.tar.gz \
    && mkdir -p /opt/presto \
    && tar -zxf /tmp/presto.tar.gz -C /opt/presto \
    && rm /tmp/presto.tar.gz

WORKDIR $PRESTO_HOME

RUN set -x \
        && curl -fSLs -o /opt/presto/presto-server-$PRESTO_VERSION/presto-cli https://repo1.maven.org/maven2/com/facebook/presto/presto-cli/$PRESTO_VERSION/presto-cli-$PRESTO_VERSION-executable.jar \
        && chmod +x /opt/presto/presto-server-$PRESTO_VERSION/presto-cli \
        && ln -s /opt/presto/presto-server-$PRESTO_VERSION/presto-cli /usr/local/bin/presto-cli
# Con***REMOVED***gure JSON serialization
ADD json-serde-1.3.8-jar-with-dependencies.jar ${PRESTO_HOME}/plugin/hive-hadoop2/

COPY etc/ $PRESTO_HOME/etc/
COPY entrypoint.sh /usr/local/bin

# Default to using localhost for discovery
ENV PRESTO_CONF_discovery_uri http://localhost:8080

EXPOSE 8080

RUN \
    mkdir -p /var/presto/data && \
    chown -R 1003:0 /opt /var/presto/data && \
    chmod -R 770 /opt /var/presto/data /etc/passwd

VOLUME /var/presto/data
ENTRYPOINT ["entrypoint.sh"]
CMD ["bin/launcher", "run"]

USER 1003
