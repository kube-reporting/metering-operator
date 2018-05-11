#!/bin/bash
ROOT_DIR=$(dirname "${BASH_SOURCE}")/..
source "${ROOT_DIR}/hack/common.sh"

msg "Removing Metering Resource"
kube-remove \
    "$METERING_CR_FILE"

msg "Removing Metering Cluster Service Version"
kube-remove \
    "$ALM_MANIFESTS_DIR/metering.${METERING_VERSION}.clusterserviceversion.yaml"

if [ "$SKIP_DELETE_CRDS" == "true" ]; then
    echo "\$SKIP_DELETE_CRDS is true, skipping deletion of Custom Resource Definitions"
else
    msg "Removing Custom Resource Definitions"
    kube-remove \
    "$MANIFESTS_DIR/custom-resource-definitions"
fi
