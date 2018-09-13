FROM centos:7

RUN yum -y update && yum clean all

RUN mkdir -p /go && chmod -R 777 /go && \
    yum -y install epel-release && \
    yum -y install hg git golang make jq curl python python-setuptools
RUN yum clean all && rm -rf /var/cache/yum

RUN easy_install pyyaml

ENV GOPATH /go
