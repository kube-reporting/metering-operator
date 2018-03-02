#!/bin/bash

export HIVE_LOGLEVEL="${HIVE_LOGLEVEL:-INFO}"
export HIVE_METASTORE_HADOOP_OPTS=" -Dhive.log.level=${HIVE_LOGLEVEL} "
export HIVE_OPTS="$HIVE_OPTS --hiveconf hive.root.logger=${HIVE_LOGLEVEL},console "

exec /usr/local/bin/entrypoint.sh /opt/hive/bin/hive "$@"
