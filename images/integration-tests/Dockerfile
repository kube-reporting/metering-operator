FROM golang:1.10-alpine

RUN apk update
RUN apk add libc-dev git bash jq zip curl make
VOLUME /out

ENV HELM_VERSION 2.8.0
RUN curl \
    --silent \
    --show-error \
    --location \
    "https://storage.googleapis.com/kubernetes-helm/helm-v${HELM_VERSION}-linux-amd64.tar.gz" \
    | tar xz --strip-components=1 -C /usr/local/bin linux-amd64/helm \
    && chmod +x /usr/local/bin/helm

env KUBERNETES_VERSION 1.8.3
RUN curl \
    --silent \
    --show-error \
    --location \
    "https://storage.googleapis.com/kubernetes-release/release/v${KUBERNETES_VERSION}/bin/linux/amd64/kubectl" \
    -o /usr/local/bin/kubectl \
    && chmod +x /usr/local/bin/kubectl

RUN go get -u github.com/kardianos/govendor
RUN \
    set -x \
    && mkdir -p $GOPATH/src/github.com/wercker \
    && git clone \
        https://github.com/wercker/stern \
        $GOPATH/src/github.com/wercker/stern \
    && cd $GOPATH/src/github.com/wercker/stern \
    && git remote add coreos https://github.com/coreos/stern \
    && git fetch coreos remove_terminated_containers \
    && git checkout coreos/remove_terminated_containers \
    && govendor sync -v \
    && go install

RUN helm init --client-only --skip-refresh && helm repo remove stable || true

RUN mkdir -p /go/src/github.com/coreos-inc
COPY . /go/src/github.com/coreos-inc/kube-chargeback
WORKDIR /go/src/github.com/coreos-inc/kube-chargeback

CMD ["hack/e2e.sh"]
