FROM centos:7

RUN yum install ca-certificates bash
# add pod data collector binary
ADD ./bin/chargeback /usr/local/bin/chargeback

ENTRYPOINT ["chargeback"]
CMD ["start"]

USER 3001
