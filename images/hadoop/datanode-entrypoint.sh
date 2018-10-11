#!/bin/bash

datadir=/hadoop/dfs/data
if [ ! -d "$datadir" ]; then
  echo "Datanode data directory not found: $datadir"
  exit 2
***REMOVED***

exec hdfs --con***REMOVED***g "$HADOOP_CONF_DIR" datanode "$@"
