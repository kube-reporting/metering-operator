FROM quay.io/coreos/presto:metering-0.212

RUN yum -y install perl \
    && yum clean all \
    && rm -rf /var/cache/yum

COPY jmx-exporter-config.yml /opt/jmx_exporter/config.yml
COPY etc/ $PRESTO_HOME/etc/
COPY entrypoint.sh /usr/local/bin

# Default to using localhost for discovery
ENV PRESTO_CONF_discovery_uri http://localhost:8080

RUN \
    mkdir -p /var/presto/data && \
    chown -R 1003:0 /opt/presto /var/presto/data && \
    chmod -R 770 /var/presto/data && \
    chmod -R 774 /etc/passwd && \
    chmod -R 775 /opt/presto

VOLUME /var/presto/data
USER 1003
ENTRYPOINT ["entrypoint.sh"]
CMD ["bin/launcher", "run"]
