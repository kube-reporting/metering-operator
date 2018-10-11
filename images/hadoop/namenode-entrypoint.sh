#!/bin/bash

namedir=/hadoop/dfs/name
if [ ! -d "$namedir" ]; then
  echo "Namenode name directory not found: $namedir"
  exit 2
***REMOVED***

if [ -z "$CLUSTER_NAME" ]; then
  echo "Cluster name not speci***REMOVED***ed"
  exit 2
***REMOVED***

if [ "$(ls -A $namedir)" == "" ]; then
  echo "Formatting namenode name directory: $namedir"
  hdfs --con***REMOVED***g "$HADOOP_CONF_DIR" namenode -format "$CLUSTER_NAME"
***REMOVED***

exec hdfs --con***REMOVED***g "$HADOOP_CONF_DIR" namenode "$@"
