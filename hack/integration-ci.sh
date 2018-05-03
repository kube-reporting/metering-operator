#!/bin/bash
set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

: "${METERING_NAMESPACE:?}"
: "${KUBECONFIG:?}"
: "${DEPLOY_TAG:?}"

TMP_DIR="$(mktemp -d)"
# We set these in here so that when we run integration.sh, which runs e2e-test-runner.sh,
# they're set to the same files for both the deploy-ci.sh script, and the
# cleanup after deploy-ci.sh finishes. Without this, the cleanup would use the
# default files, for uninstall rather than the same files used for deploy.
export METERING_CR_FILE=${METERING_CR_FILE:-"$TMP_DIR/custom-metering-cr-${DEPLOY_TAG}.yaml"}
export CUSTOM_DEPLOY_MANIFESTS_DIR=${CUSTOM_DEPLOY_MANIFESTS_DIR:-"$TMP_DIR/custom-deploy-manifests-${DEPLOY_TAG}"}
export CUSTOM_HELM_OPERATOR_OVERRIDE_VALUES=${CUSTOM_HELM_OPERATOR_OVERRIDE_VALUES:-"$TMP_DIR/custom-helm-operator-values-${DEPLOY_TAG}.yaml"}
export CUSTOM_ALM_OVERRIDE_VALUES=${CUSTOM_ALM_OVERRIDE_VALUES:-"$TMP_DIR/custom-alm-values-${DEPLOY_TAG}.yaml"}

export DEPLOY_SCRIPT="${DEPLOY_SCRIPT:-deploy-ci.sh}"

"$DIR/integration.sh"
