#!/bin/bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source ${DIR}/util.sh

msg "Removing pull secrets"
kube-remove-non-file secret coreos-pull-secret

msg "Removing Chargeback Cluster Service Version"
kube-remove \
    manifests/alm/chargeback-clusterserviceversion.yaml

msg "Removing chargeback-helm-operator"
kube-remove-non-file deployment -l alm-owner-name=chargeback-helm-operator.v0.5.0

msg "Removing Custom Resource Definitions"
kube-remove \
    manifests/custom-resource-definitions
