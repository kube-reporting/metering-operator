FROM quay.io/coreos/helm:metering-v2.8.2

USER root
# epel-release is for `jq`
RUN INSTALL_PKGS="curl bash jq ca-certi***REMOVED***cates socat kubernetes-client" \
    && yum install -y epel-release \
    && yum install -y $INSTALL_PKGS  \
    && yum clean all \
    && rm -rf /var/cache/yum

EXPOSE 44134

COPY run-operator.sh /usr/local/bin/run-operator.sh
COPY get_owner.sh /usr/local/bin/get_owner.sh

ENV HELM_HOST 127.0.0.1:44134

ENV HOME /tmp

CMD ["run-operator.sh"]

USER 2001
