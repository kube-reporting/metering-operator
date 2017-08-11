#!/bin/bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source ${DIR}/util.sh

msg "Creating chargeback namespace"
kube-install manifests/chargeback/namespace.yaml

msg "Installing collection layer (with build of kube-state-metrics with Node info)"
kube-install manifests/kube-state-metrics manifests/promsum manifests/prom-operator

msg "Installing query layer"
kube-install manifests/hive manifests/presto manifests/chargeback
