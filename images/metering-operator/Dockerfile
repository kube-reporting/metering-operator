FROM quay.io/coreos/helm-operator:latest as helm-operator
FROM quay.io/coreos/metering-builder:latest as build

# we need the helm CLI to build the charts
COPY --from=helm-operator /usr/local/bin/helm /usr/local/bin/helm

COPY . $GOPATH/src/github.com/operator-framework/operator-metering
WORKDIR $GOPATH/src/github.com/operator-framework/operator-metering

RUN make \
    bin/tectonic-metering-0.1.0.tgz \
    bin/openshift-metering-0.1.0.tgz \
    bin/operator-metering-0.1.0.tgz \
    bin/metering-override-values.yaml

FROM quay.io/coreos/helm-operator:latest

COPY --from=build /go/src/github.com/operator-framework/operator-metering/bin/tectonic-metering-0.1.0.tgz /
COPY --from=build /go/src/github.com/operator-framework/operator-metering/bin/openshift-metering-0.1.0.tgz /
COPY --from=build /go/src/github.com/operator-framework/operator-metering/bin/operator-metering-0.1.0.tgz /
COPY --from=build /go/src/github.com/operator-framework/operator-metering/bin/metering-override-values.yaml /

ENV EXTRA_VALUES_FILE /metering-override-values.yaml

