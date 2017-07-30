FROM gliderlabs/alpine:3.6

RUN apk-install ca-certi***REMOVED***cates

# add pod data collector binary
ADD promsum /usr/local/bin/promsum

ADD run.sh /usr/local/bin
ENTRYPOINT ["run.sh"]
