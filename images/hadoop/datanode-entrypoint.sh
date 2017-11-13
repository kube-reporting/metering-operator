#!/bin/bash

datadir=/hadoop/dfs/data
if [ ! -d $datadir ]; then
  echo "Datanode data directory not found: $dataedir"
  exit 2
fi

$HADOOP_PREFIX/bin/hdfs --config $HADOOP_CONF_DIR datanode
