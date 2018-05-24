FROM openjdk:8-slim

ENV DEBIAN_FRONTEND noninteractive

RUN \
    apt-get update \
    && apt-get install -y --no-install-recommends \
        curl \
        less  \
        perl \
        procps \
        net-tools \
        dnsutils \
        jq \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

# ENV JAVA_HOME=/usr/lib/jvm/java-8-openjdk-amd64/

ENV HADOOP_VERSION 2.9.1
ENV HADOOP_URL http://apache.osuosl.org/hadoop/common/hadoop-$HADOOP_VERSION/hadoop-$HADOOP_VERSION.tar.gz
# The upstream .mds checksum files aren't supported by sha256sum or shasum so
# this checksum is taken from that file and put here manually.
ENV HADOOP_SHA256_CHECKSUM eed6015a123644d3b4247bac58770e4a8b31340fa62721987430e15a0dd942fc

ENV HADOOP_CLASSPATH=/opt/hadoop-$HADOOP_VERSION/share/hadoop/tools/lib/*
ENV HADOOP_PREFIX=/opt/hadoop-$HADOOP_VERSION
ENV HADOOP_CONF_DIR=/etc/hadoop
ENV HADOOP_HOME=/opt/hadoop-$HADOOP_VERSION
ENV PATH=$HADOOP_PREFIX/bin/:$PATH

RUN set -x \
    && curl -fSLs "$HADOOP_URL" -o /tmp/hadoop-$HADOOP_VERSION.tar.gz \
    && echo $HADOOP_SHA256_CHECKSUM /tmp/hadoop-$HADOOP_VERSION.tar.gz | sha256sum -c \
    && tar \
        -xzf \
        /tmp/hadoop-$HADOOP_VERSION.tar.gz \
        -C /opt \
    # cleanup
    # - remove unnecessary doc/src files
    && rm -rf ${HADOOP_HOME}/share/doc \
    && for dir in common hdfs mapreduce tools yarn; do \
         rm -rf ${HADOOP_HOME}/share/hadoop/${dir}/sources; \
       done \
    && rm -rf ${HADOOP_HOME}/share/hadoop/common/jdiff \
    && rm -rf ${HADOOP_HOME}/share/hadoop/mapreduce/lib-examples \
    && rm -rf ${HADOOP_HOME}/share/hadoop/yarn/test \
    && find ${HADOOP_HOME}/share/hadoop -name *test*.jar | xargs rm -rf \
    && rm /tmp/hadoop-$HADOOP_VERSION.tar.gz

RUN ln -s /opt/hadoop-$HADOOP_VERSION/etc/hadoop /etc/hadoop
RUN cp /etc/hadoop/mapred-site.xml.template /etc/hadoop/mapred-site.xml
RUN mkdir -p /opt/hadoop-$HADOOP_VERSION/logs

ADD entrypoint.sh /usr/local/bin
ADD check-datanode-healthy.sh /usr/local/bin
ADD datanode-entrypoint.sh /usr/local/bin
ADD namenode-entrypoint.sh /usr/local/bin
ADD start-name.sh /usr/local/bin
ADD etc /etc/hadoop

# to allow running as non-root
RUN \
    mkdir -p /hadoop/dfs/data /hadoop/dfs/name && \
    chown -R 1002:0 /opt /hadoop /hadoop /etc/hadoop && \
    chmod -R 770 /opt /hadoop /etc/hadoop /etc/passwd

VOLUME /hadoop/dfs/data /hadoop/dfs/name
ENTRYPOINT ["entrypoint.sh"]

USER 1002
