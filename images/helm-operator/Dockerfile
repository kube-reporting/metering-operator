FROM alpine:3.6

ENV KUBERNETES_VERSION 1.8.3
ENV HELM_VERSION 2.6.2

RUN apk add --no-cache curl bash jq
RUN curl \
    --silent \
    --show-error \
    --location \
    "https://storage.googleapis.com/kubernetes-release/release/v${KUBERNETES_VERSION}/bin/linux/amd64/kubectl" \
    -o /usr/local/bin/kubectl \
     && chmod +x /usr/local/bin/kubectl

RUN curl \
    --silent \
    --show-error \
    --location \
    "https://storage.googleapis.com/kubernetes-helm/helm-v${HELM_VERSION}-linux-amd64.tar.gz" \
    | tar xz --strip-components=1 -C /usr/local/bin linux-amd64/helm \
    && chmod +x /usr/local/bin/helm

COPY entrypoint.sh /usr/local/bin/entrypoint.sh
COPY get_owner.sh /usr/local/bin/get_owner.sh

ENV HELM_HOST 127.0.0.1:44134

ENTRYPOINT ["entrypoint.sh"]
