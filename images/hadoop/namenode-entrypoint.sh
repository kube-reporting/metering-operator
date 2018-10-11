#!/bin/bash

namedir=/hadoop/dfs/name
if [ ! -d "$namedir" ]; then
  echo "Namenode name directory not found: $namedir"
  exit 2
fi

if [ -z "$CLUSTER_NAME" ]; then
  echo "Cluster name not specified"
  exit 2
fi

if [ "$(ls -A $namedir)" == "" ]; then
  echo "Formatting namenode name directory: $namedir"
  hdfs --config "$HADOOP_CONF_DIR" namenode -format "$CLUSTER_NAME"
fi

exec hdfs --config "$HADOOP_CONF_DIR" namenode "$@"
