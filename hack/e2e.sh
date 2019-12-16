#!/bin/bash

set -e

: "${KUBECONFIG:?}"

ROOT_DIR=$(dirname "${BASH_SOURCE[0]}")/..
source "${ROOT_DIR}/hack/common.sh"
source "${ROOT_DIR}/hack/lib/tests.sh"

function cleanup() {
    exit_status=$?

    if [[ "${METERING_RUN_DEV_TEST_SETUP}" = false ]]; then
        echo "Removing namespaces with the 'name=metering-testing-ns' label"
        kubectl delete ns -l "name=metering-testing-ns" || true
    ***REMOVED***
        echo "Skipping the namespace deletion"
    ***REMOVED***

    echo "Exiting hack/e2e.sh"
    exit "$exit_status"
}

trap cleanup EXIT

METERING_SHORT_TESTS="${METERING_SHORT_TESTS:=false}"
METERING_REPO_PATH="${METERING_REPO_PATH:=${ROOT_DIR}}"
METERING_RUN_TESTS_LOCALLY="${METERING_RUN_TESTS_LOCALLY:=false}"

TEST_OUTPUT_PATH="${TEST_OUTPUT_PATH:="$(mktemp -d)/${METERING_NAMESPACE}"}"
TEST_LOG_LEVEL="${TEST_LOG_LEVEL:="debug"}"
TEST_OUTPUT_DIR="${TEST_OUTPUT_PATH}/tests"
TEST_LOG_FILE="${TEST_LOG_FILE:-e2e-tests.log}"
TEST_LOG_FILE_PATH="${TEST_LOG_FILE_PATH:-$TEST_OUTPUT_DIR/$TEST_LOG_FILE}"
TEST_JUNIT_REPORT_FILE="${TEST_JUNIT_REPORT_FILE:-junit-e2e-tests.xml}"
TEST_JUNIT_REPORT_FILE_PATH="${TEST_JUNIT_REPORT_FILE_PATH:-$TEST_OUTPUT_DIR/$TEST_JUNIT_REPORT_FILE}"
METERING_RUN_DEV_TEST_SETUP="${METERING_RUN_DEV_TEST_SETUP:-false}"


mkdir -p "$TEST_OUTPUT_DIR"

echo "\$KUBECONFIG=$KUBECONFIG"
echo "\$METERING_NAMESPACE=$METERING_NAMESPACE"
echo "\$METERING_RUN_TESTS_LOCALLY=$METERING_RUN_TESTS_LOCALLY"
echo "\$METERING_RUN_DEV_TEST_SETUP=$METERING_RUN_DEV_TEST_SETUP"
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
    -timeout 60m \
    "./test/e2e" \
    -kubecon***REMOVED***g="${KUBECONFIG}" \
    -namespace-pre***REMOVED***x="${METERING_NAMESPACE}" \
    -test-output-path="${TEST_OUTPUT_PATH}" \
    -log-level="${TEST_LOG_LEVEL}" \
    -run-tests-local="${METERING_RUN_TESTS_LOCALLY}" \
    -repo-path="${METERING_REPO_PATH}" \
    -run-dev-setup="${METERING_RUN_DEV_TEST_SETUP}" \
    2>&1 | tee "$TEST_LOG_FILE_PATH" ; TEST_EXIT_CODE=${PIPESTATUS[0]}

# if go-junit-report is installed, create a junit report also
if command -v go-junit-report >/dev/null 2>&1; then
    go-junit-report < "$TEST_LOG_FILE_PATH" > "${TEST_JUNIT_REPORT_FILE_PATH}"
***REMOVED***

exit "$TEST_EXIT_CODE"
