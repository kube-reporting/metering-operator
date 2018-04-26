FROM quay.io/coreos/helm-operator:latest

COPY tectonic-chargeback-0.1.0.tgz /
COPY openshift-chargeback-0.1.0.tgz /

COPY chargeback-override-values.yaml /
ENV EXTRA_VALUES_FILE /chargeback-override-values.yaml

