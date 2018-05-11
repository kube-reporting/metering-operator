#!/bin/bash -e

ROOT_DIR=$(dirname "${BASH_SOURCE}")/..
source "${ROOT_DIR}/hack/common.sh"

load_version_vars

kubectl create namespace "${METERING_NAMESPACE}" || true

echo "Deploying Metering ${METERING_VERSION}"

msg "Installing Custom Resource De***REMOVED***nitions"
kube-install \
    "$MANIFESTS_DIR/custom-resource-de***REMOVED***nitions"

msg "Installing Metering Cluster Service Versions"
kube-install \
    "$ALM_MANIFESTS_DIR/metering.${METERING_VERSION}.clusterserviceversion.yaml"

msg "Installing Metering Resource"
kube-install \
    "$METERING_CR_FILE"

