FROM busybox:1.26.2

# add pod data collector binary
ADD https://s3.amazonaws.com/promsum/promsum /usr/local/bin/promsum
RUN chmod +x /usr/local/bin/promsum

ADD run.sh /usr/local/bin
ENTRYPOINT ["run.sh"]
