#!/bin/bash

datadir=/hadoop/dfs/data
if [ ! -d $datadir ]; then
  echo "Datanode data directory not found: $dataedir"
  exit 2
***REMOVED***

$HADOOP_PREFIX/bin/hdfs --con***REMOVED***g $HADOOP_CONF_DIR datanode
