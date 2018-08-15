FROM centos:7

RUN yum install ca-certificates bash
# add pod data collector binary
ADD ./bin/reporting-operator /usr/local/bin/reporting-operator

ENTRYPOINT ["reporting-operator"]
CMD ["start"]

USER 3001
