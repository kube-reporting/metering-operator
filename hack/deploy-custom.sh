#!/bin/bash
set -e

export DELETE_PVCS=${DELETE_PVCS:-true}

: "${CUSTOM_METERING_CR_FILE:?Must set \$CUSTOM_METERING_CR_FILE}"

TMP_DIR="$(mktemp -d)"
export CUSTOM_DEPLOY_MANIFESTS_DIR=${CUSTOM_DEPLOY_MANIFESTS_DIR:-"$TMP_DIR/custom-deploy-manifests"}
export CUSTOM_INSTALLER_MANIFESTS_DIR="$CUSTOM_DEPLOY_MANIFESTS_DIR/openshift/metering-operator"
export CUSTOM_OLM_MANIFESTS_DIR="$CUSTOM_DEPLOY_MANIFESTS_DIR/openshift/olm"

ROOT_DIR=$(dirname "${BASH_SOURCE}")/..
source "${ROOT_DIR}/hack/common.sh"

echo "Creating metering manifests"

mkdir -p "$CUSTOM_DEPLOY_MANIFESTS_DIR" "$CUSTOM_OLM_MANIFESTS_DIR"

export INSTALLER_MANIFESTS_DIR="$CUSTOM_INSTALLER_MANIFESTS_DIR"
export OLM_MANIFESTS_DIR="$CUSTOM_OLM_MANIFESTS_DIR"
export DEPLOY_MANIFESTS_DIR="$CUSTOM_DEPLOY_MANIFESTS_DIR"

"$ROOT_DIR/hack/create-metering-manifests.sh" \
    "$INSTALLER_MANIFESTS_DIR" \
    "$OLM_MANIFESTS_DIR" \
    "$OCP_TELEMETER_MANIFESTS_DIR"

echo "Deploying"
"${ROOT_DIR}/hack/deploy.sh"
