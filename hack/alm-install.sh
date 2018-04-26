#!/bin/bash -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source "${DIR}/default-env.sh"
source "${DIR}/util.sh"

MANIFESTS_DIR="$DIR/../manifests"
: "${INSTALLER_MANIFEST_DIR:=$MANIFESTS_DIR/deploy/tectonic/helm-operator}"
: "${ALM_MANIFEST_DIR:=$MANIFESTS_DIR/deploy/tectonic/alm}"
: "${CHARGEBACK_CR_FILE:=$INSTALLER_MANIFEST_DIR/metering.yaml}"

kubectl create namespace "${CHARGEBACK_NAMESPACE}" || true

if [ "$CHARGEBACK_NAMESPACE" != "tectonic-system" ]; then
    msg "Con***REMOVED***guring pull secrets"
    copy-tectonic-pull
***REMOVED***

msg "Installing Custom Resource De***REMOVED***nitions"
kube-install \
    "$MANIFESTS_DIR/custom-resource-de***REMOVED***nitions"

msg "Installing Metering Cluster Service Version"
kube-install \
    "$ALM_MANIFEST_DIR/metering.clusterserviceversion.yaml"

msg "Installing Chargeback Resource"
kube-install \
    "$CHARGEBACK_CR_FILE"

