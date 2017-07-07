FROM alpine:3.6

RUN apk --update add ca-certi***REMOVED***cates bash

# add pod data collector binary
ADD https://s3.amazonaws.com/promsum/promsum /usr/local/bin/promsum
RUN chmod +x /usr/local/bin/promsum

ADD run.sh /usr/local/bin
ENTRYPOINT ["run.sh"]
