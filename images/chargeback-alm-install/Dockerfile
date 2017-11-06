FROM alpine:3.6

RUN apk add --no-cache curl bash jq
RUN curl -L -o /bin/kubectl https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl \
 && chmod +x /bin/kubectl

COPY ./hack /opt/hack
COPY ./manifests /opt/manifests

ENV PRIVILEGED_INSTALL false

WORKDIR /opt

ENTRYPOINT /opt/hack/alm-install.sh
