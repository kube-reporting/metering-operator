FROM quay.io/coreos/helm-operator:latest

COPY tectonic-chargeback-0.1.0.tgz /
COPY tectonic-chargeback-override-values.yaml /
COPY openshift-chargeback-0.1.0.tgz /

ENV HELM_CHART_PATH /tectonic-chargeback-0.1.0.tgz
ENV HELM_RELEASE_NAME tectonic-chargeback
ENV EXTRA_VALUES_FILE /tectonic-chargeback-override-values.yaml

