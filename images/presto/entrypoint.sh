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
  sed -i "/<\/configuration>/ s/.*/${escapedEntry}\n&/" $path
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
        if [[ $path = *.xml ]]; then
            addXMLProperty $path $name "$value"
        elif [[ $path = *.properties ]]; then
            addSimpleProperty $path $name "$value"
        else
            echo "unsupported file extension for $file, must end in .properties or .xml"
        fi
    done
}

# Presto
configure "${PRESTO_HOME}/etc/catalog/hive.properties" hive-catalog HIVE_CATALOG
configure "${PRESTO_HOME}/etc/config.properties" presto-conf PRESTO_CONF

exec $@

