#!/bin/bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source ${DIR}/default-env.sh
source ${DIR}/util.sh

if [ "$CHARGEBACK_NAMESPACE" != "tectonic-system" ]; then
    msg "Removing pull secrets"
    kube-remove-non-file secret coreos-pull-secret
fi

msg "Removing Chargeback"
kube-remove \
    manifests/installer/chargeback.yaml

msg "Removing chargeback-helm-operator"
kube-remove \
    manifests/installer/chargeback-helm-operator-deployment.yaml

msg "Removing chargeback-helm-operator service account and RBAC resources"
kube-remove \
    manifests/installer/chargeback-helm-operator-service-account.yaml \
    manifests/installer/chargeback-helm-operator-rbac.yaml


if [ "$SKIP_DELETE_CRDS" == "true" ]; then
    echo "\$SKIP_DELETE_CRDS is true, skipping deletion of Custom Resource Definitions"
else
    msg "Removing Chargeback CRD"
    kube-remove \
        manifests/installer/chargeback-crd.yaml

    msg "Removing Custom Resource Definitions"
    kube-remove \
        manifests/custom-resource-definitions
fi
