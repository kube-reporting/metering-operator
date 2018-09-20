FROM centos:7

RUN yum -y update && yum clean all

RUN INSTALL_PKGS="hg git golang make jq curl python PyYAML" && \
    mkdir -p /go && chmod -R 777 /go && \
    yum -y install epel-release && \
    yum -y install $INSTALL_PKGS && \
    yum clean all && \
    rm -rf /var/cache/yum

ENV GOPATH /go
