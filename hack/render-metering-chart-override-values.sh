#!/bin/bash

set -e

: "${1?"Usage: $0 IMAGE_TAG"}"

cat <<EOF
metering-operator:
  spec:
    image:
      tag: $1
presto:
  spec:
    presto:
      image:
        tag: $1
    hive:
      image:
        tag: $1
hdfs:
  spec:
    image:
      tag: $1
EOF
