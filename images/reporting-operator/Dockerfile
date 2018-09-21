FROM quay.io/coreos/metering-builder:latest as build

COPY . $GOPATH/src/github.com/operator-framework/operator-metering
WORKDIR $GOPATH/src/github.com/operator-framework/operator-metering

RUN make reporting-operator-bin

FROM openshift/origin-base

RUN yum install ca-certificates bash

COPY --from=build /go/src/github.com/operator-framework/operator-metering/bin/reporting-operator /usr/local/bin/reporting-operator

LABEL io.k8s.display-name="OpenShift metering-reporting-operator" \
      io.k8s.description="This is a component of OpenShift Container Platform and manages collecting data from monitoring and running reports." \
      io.openshift.tags="openshift" \
      maintainer="Chance Zibolski <czibolsk@redhat.com>"

ENTRYPOINT ["reporting-operator"]
CMD ["start"]

USER 3001
