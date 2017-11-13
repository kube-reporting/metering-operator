#!/bin/bash

function addXMLProperty() {
  local path=$1
  local name=$2
  local value=$3

  local entry="<property><name>$name</name><value>${value}</value></property>"
  local escapedEntry=$(echo $entry | sed 's/\//\\\//g')
  sed -i "/<\/con***REMOVED***guration>/ s/.*/${escapedEntry}\n&/" $path
}

function con***REMOVED***gure() {
    local path=$1
    local module=$2
    local envPre***REMOVED***x=$3

    local var
    local value

    echo "Con***REMOVED***guring $module"
    for c in `printenv | perl -sne 'print "$1 " if m/^${envPre***REMOVED***x}_(.+?)=.*/' -- -envPre***REMOVED***x=$envPre***REMOVED***x`; do
        name=`echo ${c} | perl -pe 's/___/-/g; s/__/@/g; s/_/./g; s/@/_/g;'`
        var="${envPre***REMOVED***x}_${c}"
        value=${!var}

        echo " - Setting $name=$value"
        addXMLProperty $path $name "$value"
    done
}

# Hadoop (common to both Presto and Hive)
con***REMOVED***gure "${HADOOP_CONF_DIR}/core-site.xml" core CORE_CONF
con***REMOVED***gure "${HADOOP_CONF_DIR}/hdfs-site.xml" hdfs HDFS_CONF
con***REMOVED***gure "${HADOOP_CONF_DIR}/httpfs-site.xml" httpfs HTTPFS_CONF
con***REMOVED***gure "${HADOOP_CONF_DIR}/kms-site.xml" kms KMS_CONF

# Hive
con***REMOVED***gure "${HIVE_HOME}/conf/hive-site.xml" hive HIVE_SITE_CONF

max_memory() {
    local memory_limit=$1
    local ratio=${JAVA_MAX_MEM_RATIO:-50}
    echo "${memory_limit} ${ratio} 1048576" | awk '{printf "%d\n" , ($1*$2)/(100*$3) + 0.5}'
}

# Check for container memory limits/request and use it to set JVM Heap size.
# Defaults to 50% of the limit/request value.
if [ -n "$MY_MEM_LIMIT" ]; then
    export HADOOP_HEAPSIZE="$( max_memory $MY_MEM_LIMIT )"
elif [ -n "$MY_MEM_REQUEST" ]; then
    export HADOOP_HEAPSIZE="$( max_memory $MY_MEM_REQUEST )"
***REMOVED***

if [ -z "$HADOOP_HEAPSIZE" ]; then
    echo "Unable to automatically set HADOOP_HEAPSIZE"
***REMOVED***
    echo "Setting HADOOP_HEAPSIZE to ${HADOOP_HEAPSIZE}M"
***REMOVED***

exec $@

