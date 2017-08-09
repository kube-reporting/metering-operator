#!/bin/bash

function addProperty() {
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
        addProperty $path $name "$value"
    done
}

export HADOOP_CLASSPATH="/opt/hive/hcatalog/share/hcatalog/*:/opt/hadoop-2.8.0/share/hadoop/tools/lib/*"
export HIVE_AUX_JARS_PATH=/usr/hdp/current/hive-server2/auxlib

con***REMOVED***gure /opt/hive/conf/hive-site.xml hive HIVE_SITE_CONF
con***REMOVED***gure /etc/hadoop/core-site.xml core CORE_CONF
con***REMOVED***gure /etc/hadoop/hdfs-site.xml hdfs HDFS_CONF
con***REMOVED***gure /etc/hadoop/httpfs-site.xml httpfs HTTPFS_CONF
con***REMOVED***gure /etc/hadoop/kms-site.xml kms KMS_CONF

exec $@
