FROM quay.io/fest-data-demo/hadoop:0.1

# add pod data collector binary
ADD https://s3-us-west-2.amazonaws.com/kube-cluster-data/linux /usr/local/bin/pod-data
RUN chmod +x /usr/local/bin/pod-data

ADD startup.sh /usr/local/bin
ENTRYPOINT ["startup.sh"]
