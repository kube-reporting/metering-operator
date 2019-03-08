#!/bin/bash
set -e

: "${KUBECONFIG?}"
export METERING_NAMESPACE="${METERING_NAMESPACE:=metering-integration}"

ROOT_DIR=$(dirname "${BASH_SOURCE}")/..
source "${ROOT_DIR}/hack/common.sh"
source "${ROOT_DIR}/hack/lib/tests.sh"

export DEPLOY_SCRIPT="${DEPLOY_SCRIPT:-deploy-e2e.sh}"
export TEST_SCRIPT="$ROOT_DIR/hack/run-integration-tests.sh"

export TEST_LOG_FILE="${TEST_LOG_FILE:-integration-tests.log}"
export DEPLOY_LOG_FILE="${DEPLOY_LOG_FILE:-integration-deploy.log}"
export TEST_TAP_FILE="${TEST_TAP_FILE:-integration-tests.tap}"

echo "\$KUBECONFIG=$KUBECONFIG"
echo "\$METERING_NAMESPACE=$METERING_NAMESPACE"
echo "\$METERING_OPERATOR_DEPLOY_REPO=$METERING_OPERATOR_DEPLOY_REPO"
echo "\$REPORTING_OPERATOR_DEPLOY_REPO=$REPORTING_OPERATOR_DEPLOY_REPO"
echo "\$METERING_OPERATOR_DEPLOY_TAG=$METERING_OPERATOR_DEPLOY_TAG"
echo "\$REPORTING_OPERATOR_DEPLOY_TAG=$REPORTING_OPERATOR_DEPLOY_TAG"

export DISABLE_PROMSUM=true

"$ROOT_DIR/hack/e2e-test-runner.sh"
