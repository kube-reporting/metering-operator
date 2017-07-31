ARG BASE_IMAGE
FROM ${BASE_IMAGE}

# add pod data collector binary
ADD promsum /usr/local/bin/promsum

ADD run.sh /usr/local/bin
ENTRYPOINT ["run.sh"]
