FROM java:8

RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends net-tools curl && \
   apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*
ENV JAVA_HOME=/usr/lib/jvm/java-8-openjdk-amd64/

RUN gpg --keyserver pgp.mit.edu --recv-keys \
    A210C095 \
    1F27E622 \
    A7239D59 \
    E964B5FF \
    B8F47547 \
    3D0C92B9 \
    30CD0996 \
    C36C5F0F \
    7CC7352A \
    56D8D31A \
    AEC77EAF \
    ECB31663 \
    7869A2FB \
    3DD51430 \
    E4E11856 \
    7340E1CB \
    F0B000F0 \
    B021B57F \
    1EFF7770 \
    E72E74D3 \
    9475BD5D \
    EB8C6EFA \
    DE5868E6 \
    C99785DE \
    0DCED714 \
    7501105C \
    90348D47 \
    526633F3 \
    D33D195C

ENV HADOOP_VERSION 2.8.0
ENV HADOOP_URL https://www.apache.org/dist/hadoop/common/hadoop-$HADOOP_VERSION/hadoop-$HADOOP_VERSION.tar.gz
RUN set -x \
    && curl -fSL "$HADOOP_URL" -o /tmp/hadoop.tar.gz \
    && curl -fSL "$HADOOP_URL.asc" -o /tmp/hadoop.tar.gz.asc \
    && gpg --verify /tmp/hadoop.tar.gz.asc \
    && tar -xvf /tmp/hadoop.tar.gz -C /opt/ \
    && rm /tmp/hadoop.tar.gz*

RUN ln -s /opt/hadoop-$HADOOP_VERSION/etc/hadoop /etc/hadoop
RUN cp /etc/hadoop/mapred-site.xml.template /etc/hadoop/mapred-site.xml
RUN mkdir /opt/hadoop-$HADOOP_VERSION/logs

RUN mkdir /hadoop-data
ADD entrypoint.sh /usr/local/bin
ADD start-name.sh /usr/local/bin
ADD etc /etc/hadoop

ENV HADOOP_CLASSPATH=/opt/hadoop-2.8.0/share/hadoop/tools/lib/*:$HADOOP_CLASSPATH
ENV HADOOP_PREFIX=/opt/hadoop-$HADOOP_VERSION
ENV HADOOP_CONF_DIR=/etc/hadoop
ENV PATH=$HADOOP_PREFIX/bin/:$PATH

ENTRYPOINT ["entrypoint.sh"]

