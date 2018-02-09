FROM openjdk:8-slim

ENV DEBIAN_FRONTEND noninteractive

RUN \
    apt-get update \
    && apt-get install -y --no-install-recommends \
        curl \
        less  \
        gpg \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

# ENV JAVA_HOME=/usr/lib/jvm/java-8-openjdk-amd64/

ENV HADOOP_VERSION 2.8.3
ENV HADOOP_URL http://apache.osuosl.org/hadoop/common/hadoop-$HADOOP_VERSION/hadoop-$HADOOP_VERSION.tar.gz
ENV HADOOP_SIGNATURE_URL https://dist.apache.org/repos/dist/release/hadoop/common/hadoop-$HADOOP_VERSION/hadoop-$HADOOP_VERSION.tar.gz.asc
ENV HADOOP_GPG_KEYS_URL https://dist.apache.org/repos/dist/release/hadoop/common/KEYS

ENV HADOOP_CLASSPATH=/opt/hadoop-$HADOOP_VERSION/share/hadoop/tools/lib/*
ENV HADOOP_PREFIX=/opt/hadoop-$HADOOP_VERSION
ENV HADOOP_CONF_DIR=/etc/hadoop
ENV HADOOP_HOME=/opt/hadoop-$HADOOP_VERSION
ENV PATH=$HADOOP_PREFIX/bin/:$PATH

RUN set -x \
    && curl -fSLs "$HADOOP_URL" -o /tmp/hadoop-$HADOOP_VERSION.tar.gz \
    && curl -fSLs "$HADOOP_SIGNATURE_URL" -o /tmp/hadoop-$HADOOP_VERSION.tar.gz.asc \
    && curl -fSLs "$HADOOP_GPG_KEYS_URL" -o /tmp/hadoop-gpg-keys \
    && gpg --import /tmp/hadoop-gpg-keys \
    && gpg --verify /tmp/hadoop-$HADOOP_VERSION.tar.gz.asc \
    && tar \
        -xzf \
        /tmp/hadoop-$HADOOP_VERSION.tar.gz \
        -C /opt \
    # cleanup
    # - remove unnecessary doc/src ***REMOVED***les
    && rm -rf ${HADOOP_HOME}/share/doc \
    && for dir in common hdfs mapreduce tools yarn; do \
         rm -rf ${HADOOP_HOME}/share/hadoop/${dir}/sources; \
       done \
    && rm -rf ${HADOOP_HOME}/share/hadoop/common/jdiff \
    && rm -rf ${HADOOP_HOME}/share/hadoop/mapreduce/lib-examples \
    && rm -rf ${HADOOP_HOME}/share/hadoop/yarn/test \
    && ***REMOVED***nd ${HADOOP_HOME}/share/hadoop -name *test*.jar | xargs rm -rf \
    && rm /tmp/hadoop-$HADOOP_VERSION.tar.gz

RUN ln -s /opt/hadoop-$HADOOP_VERSION/etc/hadoop /etc/hadoop
RUN cp /etc/hadoop/mapred-site.xml.template /etc/hadoop/mapred-site.xml
RUN mkdir /opt/hadoop-$HADOOP_VERSION/logs

RUN mkdir /hadoop-data
ADD entrypoint.sh /usr/local/bin
ADD datanode-entrypoint.sh /usr/local/bin
ADD namenode-entrypoint.sh /usr/local/bin
ADD start-name.sh /usr/local/bin
ADD etc /etc/hadoop

ENTRYPOINT ["entrypoint.sh"]

