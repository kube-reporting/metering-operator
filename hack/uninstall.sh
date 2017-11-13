#!/bin/bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source ${DIR}/util.sh

msg "Removing pull secrets"
kube-remove-non-file secret coreos-pull-secret

msg "Removing query and collection layer"
kube-remove \
    manifests/chargeback \
    manifests/presto \
    manifests/hive \
    manifests/hdfs

msg "Removing Custom Resources"
kube-remove \
    manifests/custom-resources/prom-queries \
    manifests/custom-resources/datastores \
    manifests/custom-resources/report-queries

msg "Removing Custom Resource Definitions"
kube-remove \
    manifests/custom-resource-definitions

