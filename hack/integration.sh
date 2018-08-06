#!/bin/bash
set -e

: "${KUBECONFIG?}"
export METERING_NAMESPACE=${METERING_NAMESPACE:-metering-integration}

ROOT_DIR=$(dirname "${BASH_SOURCE}")/..
source "${ROOT_DIR}/hack/common.sh"

export TEST_SCRIPT="${TEST_SCRIPT:-$ROOT_DIR/hack/run-integration-tests.sh}"
export TEST_LOG_FILE="${TEST_LOG_FILE:-integration-tests.log}"
export DEPLOY_LOG_FILE="${DEPLOY_LOG_FILE:-integration-deploy.log}"
export TEST_TAP_FILE="${TEST_TAP_FILE:-integration-tests.tap}"

"$ROOT_DIR/hack/e2e-test-runner.sh"
