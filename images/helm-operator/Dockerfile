FROM centos:7 as build
# CentOS golang build environment based on
# https://github.com/CentOS/CentOS-Dockerfiles/blob/master/golang/centos7/Dockerfile

RUN yum -y update && yum clean all

RUN mkdir -p /go && chmod -R 777 /go && \
    yum -y install epel-release && yum -y install hg git golang glide && yum group install -y "Development Tools"
RUN yum clean all && rm -rf /var/cache/yum

ENV HELM_VERSION 2.6.2
ENV GOPATH /go

WORKDIR /go/src/k8s.io/helm
RUN git clone https://github.com/helm/helm.git . && git checkout v${HELM_VERSION}

RUN go get github.com/golang/protobuf/proto
RUN make bootstrap build
RUN make docker-binary

FROM centos:7

COPY --from=build /go/src/k8s.io/helm/rootfs/tiller .

ENV KUBERNETES_VERSION 1.8.3
ENV HELM_VERSION 2.6.2

USER root

RUN yum install -y epel-release \
    && yum install -y curl bash jq ca-certificates socat

EXPOSE 44134

RUN curl \
    --silent \
    --show-error \
    --location \
    "https://storage.googleapis.com/kubernetes-release/release/v${KUBERNETES_VERSION}/bin/linux/amd64/kubectl" \
    -o /usr/local/bin/kubectl \
     && chmod +x /usr/local/bin/kubectl

RUN mkdir /tmp/helm \
    && curl \
    --silent \
    --show-error \
    --location \
    "https://storage.googleapis.com/kubernetes-helm/helm-v${HELM_VERSION}-linux-amd64.tar.gz" \
    | tar xz -C /tmp/helm \
    && mv /tmp/helm/linux-amd64/helm /usr/local/bin/helm \
    && rm -rf /tmp/helm \
    && chmod +x /usr/local/bin/helm

COPY run-operator.sh /usr/local/bin/run-operator.sh
COPY get_owner.sh /usr/local/bin/get_owner.sh

ENV HELM_HOST 127.0.0.1:44134

ENV HOME /tmp

CMD ["run-operator.sh"]

USER 2001
