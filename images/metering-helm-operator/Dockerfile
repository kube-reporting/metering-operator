FROM quay.io/coreos/helm-operator:latest

COPY tectonic-metering-0.1.0.tgz /
COPY openshift-metering-0.1.0.tgz /
COPY operator-metering-0.1.0.tgz /

COPY metering-override-values.yaml /
ENV EXTRA_VALUES_FILE /metering-override-values.yaml

