#!/bin/bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source ${DIR}/default-env.sh
source ${DIR}/util.sh

: "${CHARGEBACK_CR_FILE:=manifests/installer/chargeback-crd.yaml}"

if [ "$CHARGEBACK_NAMESPACE" != "tectonic-system" ]; then
    msg "Removing pull secrets"
    kube-remove-non-***REMOVED***le secret coreos-pull-secret
***REMOVED***

msg "Removing Chargeback"
kube-remove \
    "$CHARGEBACK_CR_FILE"

msg "Removing chargeback-helm-operator"
kube-remove \
    manifests/installer/chargeback-helm-operator-deployment.yaml

msg "Removing chargeback-helm-operator service account and RBAC resources"
kube-remove \
    manifests/installer/chargeback-helm-operator-service-account.yaml \
    manifests/installer/chargeback-helm-operator-rbac.yaml


if [ "$SKIP_DELETE_CRDS" == "true" ]; then
    echo "\$SKIP_DELETE_CRDS is true, skipping deletion of Custom Resource De***REMOVED***nitions"
***REMOVED***
    msg "Removing Chargeback CRD"
    kube-remove \
        manifests/installer/chargeback-crd.yaml

    msg "Removing Custom Resource De***REMOVED***nitions"
    kube-remove \
        manifests/custom-resource-de***REMOVED***nitions
***REMOVED***
