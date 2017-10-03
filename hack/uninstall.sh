#!/bin/bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source ${DIR}/util.sh

msg "Removing pull secrets"
kube-remove-non-file secret coreos-pull-secret

msg "Removing query layer"
kube-remove manifests/hive manifests/presto manifests/chargeback

msg "Removing collection layer"
kube-remove manifests/promsum
