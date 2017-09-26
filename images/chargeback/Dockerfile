FROM gliderlabs/alpine:3.6

RUN apk-install ca-certificates bash
# add pod data collector binary
ADD ./bin/chargeback /usr/local/bin/chargeback

ENTRYPOINT ["chargeback"]
