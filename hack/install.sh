#!/bin/bash -e

ROOT_DIR=$(dirname "${BASH_SOURCE}")/..
source "${ROOT_DIR}/hack/common.sh"

if [ "$CREATE_NAMESPACE" == "true" ]; then
    echo "Creating namespace ${METERING_NAMESPACE}"
    kubectl create namespace "${METERING_NAMESPACE}" || true
elif ! kubectl get namespace "${METERING_NAMESPACE}" 2> /dev/null; then
    echo "Namespace '${METERING_NAMESPACE}' does not exist, please create it before starting"
    exit 1
fi

msg "Installing Custom Resource Definitions"
kube-install \
    "$MANIFESTS_DIR/custom-resource-definitions"

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
