#!/bin/bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source ${DIR}/util.sh

msg "Removing pull secrets"
kube-remove-non-file secret coreos-pull-secret

msg "Removing chargeback-helm-operator"
kube-remove \
    manifests/installer

msg "Removing Custom Resource Definitions"
kube-remove \
    manifests/custom-resource-definitions
