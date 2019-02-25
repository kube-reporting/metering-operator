#!/bin/bash

set -e
set -u

cat <<EOF
reporting-operator:
  spec:
    image:
      repository: ${REPORTING_OPERATOR_IMAGE_REPO}
      tag: ${REPORTING_OPERATOR_IMAGE_TAG}
presto:
  spec:
    presto:
      image:
        repository: ${PRESTO_IMAGE_REPO}
        tag: ${PRESTO_IMAGE_TAG}
    hive:
      image:
        repository: ${HIVE_IMAGE_REPO}
        tag: ${HIVE_IMAGE_TAG}
hdfs:
  spec:
    image:
      repository: ${HDFS_IMAGE_REPO}
      tag: ${HDFS_IMAGE_TAG}
EOF
