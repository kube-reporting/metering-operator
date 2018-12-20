#!/bin/bash
set -e

: "${DEPLOY_PLATFORM:?must be set to openshift}"

TMP_DIR="$(mktemp -d)"

export INSTALL_METHOD="${DEPLOY_PLATFORM}-direct"
export CUSTOM_DEPLOY_MANIFESTS_DIR=${CUSTOM_DEPLOY_MANIFESTS_DIR:-"$TMP_DIR/custom-deploy-manifests"}
export DELETE_PVCS=${DELETE_PVCS:-true}

: "${CUSTOM_METERING_CR_FILE:?Must set \$CUSTOM_METERING_CR_FILE}"
: "${CUSTOM_HELM_OPERATOR_OVERRIDE_VALUES:?Must set \$CUSTOM_HELM_OPERATOR_OVERRIDE_VALUES}"
: "${CUSTOM_ALM_OVERRIDE_VALUES:?Must set \$CUSTOM_ALM_OVERRIDE_VALUES}"

export CUSTOM_METERING_CR_FILE
export CUSTOM_HELM_OPERATOR_OVERRIDE_VALUES
export CUSTOM_ALM_OVERRIDE_VALUES

ROOT_DIR=$(dirname "${BASH_SOURCE}")/..
source "${ROOT_DIR}/hack/common.sh"

echo "Creating metering manifests"
export MANIFEST_OUTPUT_DIR="$CUSTOM_DEPLOY_MANIFESTS_DIR"
"$ROOT_DIR/hack/create-metering-manifests.sh"

echo "Deploying"
export DEPLOY_MANIFESTS_DIR="$CUSTOM_DEPLOY_MANIFESTS_DIR"
"${ROOT_DIR}/hack/deploy.sh"
