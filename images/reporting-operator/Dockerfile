FROM quay.io/coreos/metering-builder:latest as build

COPY . $GOPATH/src/github.com/operator-framework/operator-metering
WORKDIR $GOPATH/src/github.com/operator-framework/operator-metering

RUN make reporting-operator-bin

FROM openshift/origin-base

RUN yum install ca-certi***REMOVED***cates bash

COPY --from=build /go/src/github.com/operator-framework/operator-metering/bin/reporting-operator /usr/local/bin/reporting-operator

ENTRYPOINT ["reporting-operator"]
CMD ["start"]

USER 3001
