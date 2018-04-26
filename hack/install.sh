#!/bin/bash -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source "${DIR}/default-env.sh"
source "${DIR}/util.sh"

MANIFESTS_DIR="$DIR/../manifests"
: "${CREATE_NAMESPACE:=true}"
: "${INSTALLER_MANIFEST_DIR:=$MANIFESTS_DIR/installer}"
: "${CHARGEBACK_CR_FILE:=$INSTALLER_MANIFEST_DIR/chargeback.yaml}"

if [ "$CREATE_NAMESPACE" == "true" ]; then
    echo "Creating namespace ${CHARGEBACK_NAMESPACE}"
    kubectl create namespace "${CHARGEBACK_NAMESPACE}" || true
elif ! kubectl get namespace ${CHARGEBACK_NAMESPACE} 2> /dev/null; then
    echo "Namespace '${CHARGEBACK_NAMESPACE}' does not exist, please create it before starting"
    exit 1
fi

msg "Installing Custom Resource Definitions"
kube-install \
    "$MANIFESTS_DIR/custom-resource-definitions"

msg "Installing chargeback-helm-operator service account and RBAC resources"
kube-install \
    "$INSTALLER_MANIFEST_DIR/chargeback-helm-operator-service-account.yaml" \
    "$INSTALLER_MANIFEST_DIR/chargeback-helm-operator-role.yaml" \
    "$INSTALLER_MANIFEST_DIR/chargeback-helm-operator-rolebinding.yaml"

msg "Installing chargeback-helm-operator"
kube-install \
    "$INSTALLER_MANIFEST_DIR/chargeback-helm-operator-deployment.yaml"

msg "Installing Chargeback Resource"
kube-install \
    "$CHARGEBACK_CR_FILE"
