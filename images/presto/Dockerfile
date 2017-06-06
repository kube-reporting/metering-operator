FROM java:8
ENV DEBIAN_FRONTEND noninteractive
ENV PRESTO_VERSION 0.166

RUN set -x \
    && wget -O /tmp/presto.tar.gz http://repo1.maven.org/maven2/com/facebook/presto/presto-server/$PRESTO_VERSION/presto-server-$PRESTO_VERSION.tar.gz \
    && mkdir -p /opt/presto \
    && tar -zxvf /tmp/presto.tar.gz -C /opt/presto \
    && rm /tmp/presto.tar.gz

ENV PRESTO_HOME /opt/presto/presto-server-$PRESTO_VERSION

COPY etc/ $PRESTO_HOME/etc/

# Configure JSON serialization
ADD json-serde-1.3.8-jar-with-dependencies.jar ${PRESTO_HOME}/plugin/hive-hadoop2/

EXPOSE 8080

WORKDIR $PRESTO_HOME

ADD entrypoint.sh /usr/local/bin
ENTRYPOINT ["entrypoint.sh"]
CMD ["bin/launcher", "run"]
