FROM openjdk:8
ENV DEBIAN_FRONTEND noninteractive
ENV PRESTO_VERSION 0.166

RUN apt-get update && apt-get install -y --no-install-recommends \
        less \
        && rm -rf /var/lib/apt/lists/*

RUN set -x \
    && curl -fSLs -o /tmp/presto.tar.gz http://repo1.maven.org/maven2/com/facebook/presto/presto-server/$PRESTO_VERSION/presto-server-$PRESTO_VERSION.tar.gz \
    && mkdir -p /opt/presto \
    && tar -zxf /tmp/presto.tar.gz -C /opt/presto \
    && rm /tmp/presto.tar.gz

ENV PRESTO_HOME /opt/presto/presto-server-$PRESTO_VERSION
RUN curl -fSLs -o /opt/presto/presto-server-$PRESTO_VERSION/presto-cli https://repo1.maven.org/maven2/com/facebook/presto/presto-cli/$PRESTO_VERSION/presto-cli-$PRESTO_VERSION-executable.jar \
        && chmod +x /opt/presto/presto-server-$PRESTO_VERSION/presto-cli \
        && ln -s /opt/presto/presto-server-$PRESTO_VERSION/presto-cli /usr/local/bin/presto-cli
ENV TERM=linux
COPY etc/ $PRESTO_HOME/etc/

# Configure JSON serialization
ADD json-serde-1.3.8-jar-with-dependencies.jar ${PRESTO_HOME}/plugin/hive-hadoop2/

EXPOSE 8080

WORKDIR $PRESTO_HOME

ADD entrypoint.sh /usr/local/bin
ENTRYPOINT ["entrypoint.sh"]
CMD ["bin/launcher", "run"]
