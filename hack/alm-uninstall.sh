#!/bin/bash
ROOT_DIR=$(dirname "${BASH_SOURCE}")/..
source "${ROOT_DIR}/hack/common.sh"

load_version_vars

set +e

msg "Removing Metering Resource"
kube-remove \
    "$METERING_CR_FILE"

msg "Removing Metering Cluster Service Version"
kube-remove \
    "$ALM_MANIFESTS_DIR/metering.${METERING_VERSION}.clusterserviceversion.yaml"

if [ "$SKIP_DELETE_CRDS" == "true" ]; then
    echo "\$SKIP_DELETE_CRDS is true, skipping deletion of Custom Resource De***REMOVED***nitions"
***REMOVED***
    msg "Removing Custom Resource De***REMOVED***nitions"
    kube-remove \
    "$MANIFESTS_DIR/custom-resource-de***REMOVED***nitions"
***REMOVED***
