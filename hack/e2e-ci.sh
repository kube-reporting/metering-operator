#!/bin/bash
set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

: "${METERING_NAMESPACE:?}"
: "${KUBECONFIG:?}"
: "${DEPLOY_TAG:?}"

TMP_DIR="$(mktemp -d)"
export METERING_CR_FILE="$TMP_DIR/custom-metering-cr-${DEPLOY_TAG}.yaml"
export INSTALLER_MANIFESTS_DIR="$TMP_DIR/installer_manifests-${DEPLOY_TAG}"
export CUSTOM_VALUES_FILE="$TMP_DIR/helm-operator-values-${DEPLOY_TAG}.yaml"
export DEPLOY_SCRIPT="${DEPLOY_SCRIPT:-deploy-ci.sh}"

"$DIR/e2e.sh"
