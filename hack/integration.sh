#!/bin/bash
set -e
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source ${DIR}/util.sh

: "${KUBECONFIG?}"
: "${METERING_NAMESPACE:=metering-integration}"

export TEST_SCRIPT="${TEST_SCRIPT:-$DIR/run-integration-tests.sh}"
export TEST_LOG_FILE="${TEST_LOG_FILE:-integration-tests.log}"
export DEPLOY_LOG_FILE="${DEPLOY_LOG_FILE:-integration-deploy.log}"
export TEST_TAP_FILE="${TEST_TAP_FILE:-integration-tests.tap}"

"$DIR/e2e-test-runner.sh"

