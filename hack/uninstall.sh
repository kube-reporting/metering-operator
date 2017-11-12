#!/bin/bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source ${DIR}/util.sh

msg "Removing pull secrets"
kube-remove-non-***REMOVED***le secret coreos-pull-secret

msg "Removing query and collection layer"
kube-remove \
    manifests/hdfs \
    manifests/hive \
    manifests/presto \
    manifests/chargeback

msg "Removing Custom Resources"
kube-remove \
    manifests/custom-resources/prom-queries \
    manifests/custom-resources/datastores \
    manifests/custom-resources/report-queries

msg "Removing Custom Resource De***REMOVED***nitions"
kube-remove \
    manifests/custom-resource-de***REMOVED***nitions

