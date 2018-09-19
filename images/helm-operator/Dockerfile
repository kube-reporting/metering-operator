FROM centos:7 as build
# CentOS golang build environment based on
# https://github.com/CentOS/CentOS-Docker***REMOVED***les/blob/master/golang/centos7/Docker***REMOVED***le

RUN yum -y update && yum clean all

RUN mkdir -p /go && chmod -R 777 /go && \
    yum -y install epel-release && yum -y install protobuf hg git golang glide make
RUN yum clean all && rm -rf /var/cache/yum

ENV HELM_VERSION 2.6.2
ENV GOPATH /go

RUN mkdir -p /go/src/k8s.io/helm
WORKDIR /go/src/k8s.io/helm
RUN git clone https://github.com/helm/helm.git . && git checkout v${HELM_VERSION}

RUN glide install --strip-vendor \
    && go build -o bin/protoc-gen-go ./vendor/github.com/golang/protobuf/protoc-gen-go \
    && scripts/setup-apimachinery.sh
# build tiller (in later Helm versions, this will build both binaries)
RUN make docker-binary
# build helm
RUN make build

FROM centos:7

COPY --from=build /go/src/k8s.io/helm/rootfs/tiller .
COPY --from=build /go/src/k8s.io/helm/bin/helm /usr/local/bin

ENV KUBERNETES_VERSION 1.8.3
ENV HELM_VERSION 2.6.2

USER root

RUN yum install -y epel-release \
    && yum install -y curl bash jq ca-certi***REMOVED***cates socat

EXPOSE 44134

RUN curl \
    --silent \
    --show-error \
    --location \
    "https://storage.googleapis.com/kubernetes-release/release/v${KUBERNETES_VERSION}/bin/linux/amd64/kubectl" \
    -o /usr/local/bin/kubectl \
     && chmod +x /usr/local/bin/kubectl

COPY run-operator.sh /usr/local/bin/run-operator.sh
COPY get_owner.sh /usr/local/bin/get_owner.sh

ENV HELM_HOST 127.0.0.1:44134

ENV HOME /tmp

CMD ["run-operator.sh"]

USER 2001
