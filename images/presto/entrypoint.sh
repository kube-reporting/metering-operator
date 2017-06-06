#!/bin/bash

function addProperty() {
  local path=$1
  local name=$2
  local value=$3

  echo "${name}=${value}" >> ${path}
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


con***REMOVED***gure ${PRESTO_HOME}/etc/catalog/hive.properties hive-catalog HIVE_CATALOG

exec $@
