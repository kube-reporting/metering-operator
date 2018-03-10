FROM gcr.io/kubernetes-helm/tiller:v2.6.2

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

CMD ["run-operator.sh"]

USER 2001
