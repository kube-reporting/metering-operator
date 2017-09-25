FROM gliderlabs/alpine:3.6

RUN apk-install ca-certi***REMOVED***cates bash
# add pod data collector binary
ADD ./chargeback /usr/local/bin/chargeback

ENTRYPOINT ["chargeback"]
