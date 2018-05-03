#!/bin/bash -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source "${DIR}/default-env.sh"
source "${DIR}/util.sh"

MANIFESTS_DIR="$DIR/../manifests"
: "${CREATE_NAMESPACE:=true}"
: "${DEPLOY_PLATFORM:=generic}"
: "${DEPLOY_MANIFESTS_DIR:=$MANIFESTS_DIR/deploy}"
: "${INSTALLER_MANIFESTS_DIR:=$DEPLOY_MANIFESTS_DIR/$DEPLOY_PLATFORM/helm-operator}"
: "${METERING_CR_FILE:=$INSTALLER_MANIFESTS_DIR/metering.yaml}"

if [ "$CREATE_NAMESPACE" == "true" ]; then
    echo "Creating namespace ${METERING_NAMESPACE}"
    kubectl create namespace "${METERING_NAMESPACE}" || true
elif ! kubectl get namespace ${METERING_NAMESPACE} 2> /dev/null; then
    echo "Namespace '${METERING_NAMESPACE}' does not exist, please create it before starting"
    exit 1
***REMOVED***

msg "Installing Custom Resource De***REMOVED***nitions"
kube-install \
    "$MANIFESTS_DIR/custom-resource-de***REMOVED***nitions"

msg "Installing metering-helm-operator service account and RBAC resources"
kube-install \
    "$INSTALLER_MANIFESTS_DIR/metering-helm-operator-service-account.yaml" \
    "$INSTALLER_MANIFESTS_DIR/metering-helm-operator-role.yaml" \
    "$INSTALLER_MANIFESTS_DIR/metering-helm-operator-rolebinding.yaml"

msg "Installing metering-helm-operator"
kube-install \
    "$INSTALLER_MANIFESTS_DIR/metering-helm-operator-deployment.yaml"

msg "Installing Metering Resource"
kube-install \
    "$METERING_CR_FILE"
