#!/bin/bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source ${DIR}/util.sh

msg "Removing pull secrets"
kube-remove-non-***REMOVED***le secret coreos-pull-secret

msg "Removing alm-install-operator"
kube-remove \
    manifests/installer

msg "Removing Custom Resource De***REMOVED***nitions"
kube-remove \
    manifests/custom-resource-de***REMOVED***nitions
