#!/bin/bash

function addSimpleProperty() {
  local path=$1
  local name=$2
  local value=$3

  echo "${name}=${value}" >> ${path}
}

function addXMLProperty() {
  local path=$1
  local name=$2
  local value=$3

  local entry="<property><name>$name</name><value>${value}</value></property>"
  local escapedEntry=$(echo $entry | sed -e 's/\//\\\//g' -e 's/&/\\&/g' )
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
        if [[ $path = *.xml ]]; then
            addXMLProperty $path $name "$value"
        elif [[ $path = *.properties ]]; then
            addSimpleProperty $path $name "$value"
        ***REMOVED***
            echo "unsupported ***REMOVED***le extension for $***REMOVED***le, must end in .properties or .xml"
        ***REMOVED***
    done
}

max_memory() {
    local memory_limit=$1
    local ratio=${JAVA_MAX_MEM_RATIO:-50}
    echo "${memory_limit} ${ratio} 1048576" | awk '{printf "%d\n" , ($1*$2)/(100*$3) + 0.5}'
}

# Check for container memory limits/request and use it to set JVM Heap size.
# Defaults to 50% of the limit/request value.
if [ -n "$MY_MEM_LIMIT" ]; then
    export MAX_HEAPSIZE="$( max_memory $MY_MEM_LIMIT )"
elif [ -n "$MY_MEM_REQUEST" ]; then
    export MAX_HEAPSIZE="$( max_memory $MY_MEM_REQUEST )"
***REMOVED***

if [ -z "$MAX_HEAPSIZE" ]; then
    echo "Unable to automatically set Presto JVM Max Heap Size based on pod request/limits"
    export MAX_HEAPSIZE=1024
    echo "Setting Presto JVM Max Heap Size to ${MAX_HEAPSIZE}M"
***REMOVED***
    echo "Setting Presto JVM Max Heap Size to ${MAX_HEAPSIZE}M"
***REMOVED***

echo "-Xmx${MAX_HEAPSIZE}M" >> "${PRESTO_HOME}/etc/jvm.con***REMOVED***g"

# Presto
con***REMOVED***gure "${PRESTO_HOME}/etc/catalog/hive.properties" hive-catalog HIVE_CATALOG
con***REMOVED***gure "${PRESTO_HOME}/etc/con***REMOVED***g.properties" presto-conf PRESTO_CONF
con***REMOVED***gure "${PRESTO_HOME}/etc/log.properties" presto-log PRESTO_LOG
con***REMOVED***gure "${PRESTO_HOME}/etc/node.properties" presto-node PRESTO_NODE

# add UID to /etc/passwd if missing
if ! whoami &> /dev/null; then
  if [ -w /etc/passwd ]; then
    echo "${USER_NAME:-presto}:x:$(id -u):0:${USER_NAME:-presto} user:${HOME}:/sbin/nologin" >> /etc/passwd
  ***REMOVED***
***REMOVED***

exec $@

