#!/bin/bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source ${DIR}/util.sh

msg "Removing pull secrets"
kube-remove-non-***REMOVED***le secret coreos-pull-secret

msg "Removing Chargeback Cluster Service Version"
kube-remove \
    manifests/alm/chargeback-clusterserviceversion.yaml

msg "Removing chargeback-helm-operator"
kube-remove-non-***REMOVED***le deployment -l alm-owner-name=chargeback-helm-operator.v0.5.0

msg "Removing Custom Resource De***REMOVED***nitions"
kube-remove \
    manifests/custom-resource-de***REMOVED***nitions
