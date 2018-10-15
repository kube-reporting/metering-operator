FROM centos:7

RUN \
    yum install epel-release -y \
    && yum install -y \
        java-1.8.0-openjdk \
        curl \
        less  \
        perl \
        procps \
        net-tools \
        bind-utils \
        which \
        jq \
    && yum clean all \
    && rm -rf /tmp/* /var/tmp/*

ENV JAVA_HOME=/etc/alternatives/jre

ENV HADOOP_VERSION 3.1.1
ENV HADOOP_URL https://apache.osuosl.org/hadoop/common/hadoop-$HADOOP_VERSION/hadoop-$HADOOP_VERSION.tar.gz
# The upstream .mds checksum files aren't supported by sha256sum or shasum so
# this checksum is taken from that file and put here manually.
ENV HADOOP_SHA256_CHECKSUM f837fe260587f71629aad1f4fb6719274e948111dc96ffc5a8e26f27deac5602

ENV HADOOP_CLASSPATH=/opt/hadoop-$HADOOP_VERSION/share/hadoop/tools/lib/*
ENV HADOOP_CONF_DIR=/etc/hadoop
ENV HADOOP_HOME=/opt/hadoop-$HADOOP_VERSION
ENV PATH=$HADOOP_HOME/bin:$PATH

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
RUN mkdir -p /opt/hadoop-$HADOOP_VERSION/logs

ADD entrypoint.sh /usr/local/bin
ADD check-datanode-healthy.sh /usr/local/bin
ADD datanode-entrypoint.sh /usr/local/bin
ADD namenode-entrypoint.sh /usr/local/bin
ADD etc /etc/hadoop

# to allow running as non-root
RUN \
    mkdir -p /hadoop/dfs/data /hadoop/dfs/name && \
    chown -R 1002:0 /opt /hadoop /hadoop /etc/hadoop && \
    chmod -R 770 /opt /hadoop /etc/hadoop /etc/passwd

VOLUME /hadoop/dfs/data /hadoop/dfs/name
ENTRYPOINT ["entrypoint.sh"]

USER 1002
