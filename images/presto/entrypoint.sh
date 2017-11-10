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
        if [[ $path = *.xml ]]; then
            addXMLProperty $path $name "$value"
        elif [[ $path = *.properties ]]; then
            addSimpleProperty $path $name "$value"
        ***REMOVED***
            echo "unsupported ***REMOVED***le extension for $***REMOVED***le, must end in .properties or .xml"
        ***REMOVED***
    done
}

# Presto
con***REMOVED***gure "${PRESTO_HOME}/etc/catalog/hive.properties" hive-catalog HIVE_CATALOG
con***REMOVED***gure "${PRESTO_HOME}/etc/con***REMOVED***g.properties" presto-conf PRESTO_CONF

exec $@

