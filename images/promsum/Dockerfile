FROM alpine

RUN apk update
RUN apk add ca-certi***REMOVED***cates bash

# add pod data collector binary
ADD bin/promsum /usr/local/bin/promsum

ADD run.sh /usr/local/bin
ENTRYPOINT ["run.sh"]
