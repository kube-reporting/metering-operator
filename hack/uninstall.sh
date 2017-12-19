#!/bin/bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source ${DIR}/default-env.sh
source ${DIR}/util.sh

if [ "$CHARGEBACK_NAMESPACE" != "tectonic-system" ]; then
    msg "Removing pull secrets"
    kube-remove-non-file secret coreos-pull-secret
fi

msg "Removing chargeback-helm-operator"
kube-remove \
    manifests/installer


if [ "$SKIP_DELETE_CRDS" == "true" ]; then
    echo "\$SKIP_DELETE_CRDS is true, skipping deletion of Custom Resource Definitions"
else
    msg "Removing Custom Resource Definitions"
    kube-remove \
        manifests/custom-resource-definitions
fi
