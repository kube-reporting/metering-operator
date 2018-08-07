#!/bin/bash
set -e

: "${KUBECONFIG:?}"
export METERING_NAMESPACE=${METERING_NAMESPACE:-metering-e2e}

ROOT_DIR=$(dirname "${BASH_SOURCE}")/..
source "${ROOT_DIR}/hack/common.sh"

export TEST_SCRIPT="${TEST_SCRIPT:-$ROOT_DIR/hack/run-e2e-tests.sh}"
export TEST_LOG_FILE="${TEST_LOG_FILE:-e2e-tests.log}"
export DEPLOY_LOG_FILE="${DEPLOY_LOG_FILE:-e2e-deploy.log}"
export TEST_TAP_FILE="${TEST_TAP_FILE:-e2e-tests.tap}"

"$ROOT_DIR/hack/e2e-test-runner.sh"
