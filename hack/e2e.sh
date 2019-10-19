#!/bin/bash

set -e

: "${KUBECONFIG:?}"

ROOT_DIR=$(dirname "${BASH_SOURCE[0]}")/..
source "${ROOT_DIR}/hack/common.sh"
source "${ROOT_DIR}/hack/lib/tests.sh"

function remove_namespaces() {
    echo "Removing namespaces with the 'name=metering-testing-ns' label"
    kubectl delete ns -l "name=metering-testing-ns" || true
}

trap remove_namespaces SIGINT

export METERING_NAMESPACE="${METERING_E2E_NAMESPACE:=${METERING_NAMESPACE}-e2e}"
export METERING_SHORT_TESTS=${METERING_SHORT_TESTS:=false}
export METERING_DEPLOY_MANIFESTS_PATH=${METERING_DEPLOY_MANIFESTS_PATH:=${ROOT_DIR}/manifests/deploy}
export METERING_CLEANUP_SCRIPT_PATH=${METERING_CLEANUP_SCRIPT_PATH:=${ROOT_DIR}/hack/run-test-cleanup.sh}
export TEST_OUTPUT_PATH=${TEST_OUTPUT_PATH:="$(mktemp -d)/${METERING_E2E_NAMESPACE}"}
export TEST_LOG_LEVEL=${TEST_LOG_LEVEL:="debug"}

TEST_OUTPUT_DIR="${TEST_OUTPUT_PATH}/tests"
TEST_LOG_FILE="${TEST_LOG_FILE:-e2e-tests.log}"
TEST_LOG_FILE_PATH="${TEST_LOG_FILE_PATH:-$TEST_OUTPUT_DIR/$TEST_LOG_FILE}"
TEST_JUNIT_REPORT_FILE="${TEST_JUNIT_REPORT_FILE:-junit-e2e-tests.xml}"
TEST_JUNIT_REPORT_FILE_PATH="${TEST_JUNIT_REPORT_FILE_PATH:-$TEST_OUTPUT_DIR/$TEST_JUNIT_REPORT_FILE}"

mkdir -p "$TEST_OUTPUT_DIR"

echo "\$KUBECONFIG=$KUBECONFIG"
echo "\$METERING_NAMESPACE=$METERING_NAMESPACE"
echo "\$METERING_OPERATOR_IMAGE_REPO=$METERING_OPERATOR_IMAGE_REPO"
echo "\$REPORTING_OPERATOR_IMAGE_REPO=$REPORTING_OPERATOR_IMAGE_REPO"
echo "\$METERING_OPERATOR_IMAGE_TAG=$METERING_OPERATOR_IMAGE_TAG"
echo "\$REPORTING_OPERATOR_IMAGE_TAG=$REPORTING_OPERATOR_IMAGE_TAG"
echo "\$TEST_OUTPUT_PATH=$TEST_OUTPUT_PATH"

set +e
set +o pipefail
set -x
echo "Running E2E Tests"

go test \
    -test.short="${METERING_SHORT_TESTS}" \
    -test.v \
    -parallel 10 \
    -timeout 30m \
    "./test/e2e" \
    -kubeconfig="${KUBECONFIG}" \
    -namespace-prefix="${METERING_NAMESPACE}" \
    -deploy-manifests-dir="${METERING_DEPLOY_MANIFESTS_PATH}" \
    -cleanup-script-path="${METERING_CLEANUP_SCRIPT_PATH}" \
    -test-output-path="${TEST_OUTPUT_PATH}" \
    -log-level="${TEST_LOG_LEVEL}" \
    |& tee "$TEST_LOG_FILE_PATH" ; TEST_EXIT_CODE=${PIPESTATUS[0]}

# if go-junit-report is installed, create a junit report also
if command -v go-junit-report >/dev/null 2>&1; then
    go-junit-report < "$TEST_LOG_FILE_PATH" > "${TEST_JUNIT_REPORT_FILE_PATH}"
fi

exit "$TEST_EXIT_CODE"
