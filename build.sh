#!/bin/bash
# Builds all images for demo
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

function build (){
  ${DIR}/images/${@}/build.sh
}

# build hadoop base first
build hadoop
build hive
build presto
build pod-data
