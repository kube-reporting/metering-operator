#!/bin/bash

set -e

: "${1?"Usage: $0 IMAGE_TAG"}"

cat <<EOF
chargeback-operator:
  image:
    tag: $1
presto:
  presto:
    image:
      tag: $1
  hive:
    image:
      tag: $1
hdfs:
  image:
    tag: $1
EOF
