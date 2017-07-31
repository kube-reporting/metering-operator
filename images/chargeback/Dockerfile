ARG BASE_IMAGE
FROM ${BASE_IMAGE}

# add pod data collector binary
ADD ./chargeback /usr/local/bin/chargeback

ENTRYPOINT ["chargeback"]
