#!/bin/bash
set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source ${DIR}/util.sh

: "${KUBECONFIG?}"
: "${CHARGEBACK_NAMESPACE:=chargeback-e2e}"

export TEST_SCRIPT="${TEST_SCRIPT:-$DIR/run-e2e-tests.sh}"
export TEST_LOG_FILE="${TEST_LOG_FILE:-e2e-tests.log}"
export DEPLOY_LOG_FILE="${DEPLOY_LOG_FILE:-e2e-deploy.log}"
export TEST_TAP_FILE="${TEST_TAP_FILE:-e2e-tests.tap}"

"$DIR/e2e-test-runner.sh"
