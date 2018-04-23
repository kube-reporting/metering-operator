#!/bin/bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source ${DIR}/default-env.sh
source ${DIR}/util.sh

: "${INSTALLER_MANIFEST_DIR:=$DIR/../manifests/installer}"
: "${CHARGEBACK_CR_FILE:=$INSTALLER_MANIFEST_DIR/chargeback.yaml}"
: "${SKIP_DELETE_CRDS:=true}"

msg "Removing Chargeback Resource"
kube-remove \
    "$CHARGEBACK_CR_FILE"

msg "Removing Chargeback Cluster Service Version"
kube-remove \
    manifests/alm/chargeback.clusterserviceversion.yaml

if [ "$SKIP_DELETE_CRDS" == "true" ]; then
    echo "\$SKIP_DELETE_CRDS is true, skipping deletion of Custom Resource Definitions"
else
    msg "Removing Custom Resource Definitions"
    kube-remove \
        manifests/custom-resource-definitions
fi
