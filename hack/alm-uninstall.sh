#!/bin/bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source ${DIR}/default-env.sh
source ${DIR}/util.sh

MANIFESTS_DIR="$DIR/../manifests"
: "${INSTALLER_MANIFEST_DIR:=$MANIFESTS_DIR/deploy/tectonic/helm-operator}"
: "${ALM_MANIFEST_DIR:=$MANIFESTS_DIR/deploy/tectonic/alm}"
: "${CHARGEBACK_CR_FILE:=$INSTALLER_MANIFEST_DIR/metering.yaml}"
: "${SKIP_DELETE_CRDS:=true}"

msg "Removing Chargeback Resource"
kube-remove \
    "$CHARGEBACK_CR_FILE"

msg "Removing Metering Cluster Service Version"
kube-remove \
    "$ALM_MANIFEST_DIR/metering.clusterserviceversion.yaml"

if [ "$SKIP_DELETE_CRDS" == "true" ]; then
    echo "\$SKIP_DELETE_CRDS is true, skipping deletion of Custom Resource De***REMOVED***nitions"
***REMOVED***
    msg "Removing Custom Resource De***REMOVED***nitions"
    kube-remove \
    "$MANIFESTS_DIR/custom-resource-de***REMOVED***nitions"
***REMOVED***
