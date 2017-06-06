#!/bin/bash

function addProperty() {
  local path=$1
  local name=$2
  local value=$3

  echo "${name}=${value}" >> ${path}
}

function configure() {
    local path=$1
    local module=$2
    local envPrefix=$3

    local var
    local value

    echo "Configuring $module"
    for c in `printenv | perl -sne 'print "$1 " if m/^${envPrefix}_(.+?)=.*/' -- -envPrefix=$envPrefix`; do
        name=`echo ${c} | perl -pe 's/___/-/g; s/__/@/g; s/_/./g; s/@/_/g;'`
        var="${envPrefix}_${c}"
        value=${!var}
        echo " - Setting $name=$value"
        addProperty $path $name "$value"
    done
}


configure ${PRESTO_HOME}/etc/catalog/hive.properties hive-catalog HIVE_CATALOG

exec $@
