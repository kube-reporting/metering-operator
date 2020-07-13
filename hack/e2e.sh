#!/bin/bash

set -e

: "${KUBECONFIG:?}"

ROOT_DIR=$(dirname "${BASH_SOURCE[0]}")/..
source "${ROOT_DIR}/hack/common.sh"
source "${ROOT_DIR}/hack/lib/tests.sh"

function cleanup() {
    exit_status=$?

    if [[ "${METERING_RUN_DEV_TEST_SETUP}" = false ]]; then
        echo "Removing namespaces with the 'name=${METERING_NAMESPACE}-metering-testing-ns' label"
        kubectl delete ns -l "name=${METERING_NAMESPACE}-metering-testing-ns" --wait=false || true

        # Note: the `openshift-marketplace` namespace is hardcoded for now until we have the need
        # to create the registry-related resources in another namespace (e.g. testing upstream manifests).
        echo "Removing the local registry resources with the 'name=${METERING_NAMESPACE}-metering-testing-ns' label"
        kubectl -n openshift-marketplace delete catsrc,deployment,service -l "name=${METERING_NAMESPACE}-metering-testing-ns" --wait=false || true
    else
        echo "Skipping the deletion of the metering testing namespaces and the local registry resources"
    fi

    echo "Exiting hack/e2e.sh"
    exit "$exit_status"
}

trap cleanup EXIT

EXTRA_TEST_FLAGS="${EXTRA_TEST_FLAGS:=""}"
MANIFESTS_BASE_DIR="${MANIFESTS_BASE_DIR:=manifests/deploy}"
METERING_SHORT_TESTS="${METERING_SHORT_TESTS:=false}"
METERING_REPO_PATH="${METERING_REPO_PATH:=${ROOT_DIR}}"
METERING_REPO_VERSION="$(basename $(find $MANIFESTS_BASE_DIR -type d | sort -r | head -n 1))"
METERING_RUN_TESTS_LOCALLY="${METERING_RUN_TESTS_LOCALLY:=false}"
METERING_OLM_SUBSCRIPTION_CHANNEL="${METERING_OLM_SUBSCRIPTION_CHANNEL:="4.6"}"
METERING_ANSIBLE_OPERATOR_IMAGE_REGISTRY="${METERING_ANSIBLE_OPERATOR_IMAGE_REGISTRY:=""}"

TEST_OUTPUT_PATH="${TEST_OUTPUT_PATH:="$(mktemp -d)/${METERING_NAMESPACE}"}"
TEST_LOG_LEVEL="${TEST_LOG_LEVEL:="debug"}"
TEST_OUTPUT_DIR="${TEST_OUTPUT_PATH}/tests"
TEST_LOG_FILE="${TEST_LOG_FILE:-e2e-tests.log}"
TEST_LOG_FILE_PATH="${TEST_LOG_FILE_PATH:-$TEST_OUTPUT_DIR/$TEST_LOG_FILE}"
TEST_JUNIT_REPORT_FILE="${TEST_JUNIT_REPORT_FILE:-junit-e2e-tests.xml}"
TEST_JUNIT_REPORT_FILE_PATH="${TEST_JUNIT_REPORT_FILE_PATH:-$TEST_OUTPUT_DIR/$TEST_JUNIT_REPORT_FILE}"
TEST_RUN_ALL_INSTALL_TESTS="${TEST_RUN_ALL_INSTALL_TESTS:-false}"
METERING_RUN_DEV_TEST_SETUP="${METERING_RUN_DEV_TEST_SETUP:-false}"

mkdir -p "$TEST_OUTPUT_DIR"

echo "\$KUBECONFIG=$KUBECONFIG"
echo "\$METERING_NAMESPACE=$METERING_NAMESPACE"
echo "\$METERING_RUN_TESTS_LOCALLY=$METERING_RUN_TESTS_LOCALLY"
echo "\$METERING_RUN_DEV_TEST_SETUP=$METERING_RUN_DEV_TEST_SETUP"
echo "\$METERING_REPO_VERSION=$METERING_REPO_VERSION"
echo "\$METERING_OLM_SUBSCRIPTION_CHANNEL=$METERING_OLM_SUBSCRIPTION_CHANNEL"
echo "\$METERING_OPERATOR_IMAGE_REPO=$METERING_OPERATOR_IMAGE_REPO"
echo "\$REPORTING_OPERATOR_IMAGE_REPO=$REPORTING_OPERATOR_IMAGE_REPO"
echo "\$METERING_OPERATOR_IMAGE_TAG=$METERING_OPERATOR_IMAGE_TAG"
echo "\$REPORTING_OPERATOR_IMAGE_TAG=$REPORTING_OPERATOR_IMAGE_TAG"
echo "\$METERING_ANSIBLE_OPERATOR_IMAGE_REGISTRY=$METERING_ANSIBLE_OPERATOR_IMAGE_REGISTRY"
echo "\$TEST_OUTPUT_PATH=$TEST_OUTPUT_PATH"
echo "\$TEST_RUN_ALL_INSTALL_TESTS=$TEST_RUN_ALL_INSTALL_TESTS"

if [ -n "${IMAGE_FORMAT:-}" ]; then
    echo "\$IMAGE_FORMAT=$IMAGE_FORMAT"
fi

set +e
set +o pipefail
set -x
echo "Running E2E Tests"

go test \
    -test.short="${METERING_SHORT_TESTS}" \
    -test.v \
    -parallel 10 \
    -timeout 90m \
    "./test/e2e" \
    -kubeconfig="${KUBECONFIG}" \
    -namespace-prefix="${METERING_NAMESPACE}" \
    -test-output-path="${TEST_OUTPUT_PATH}" \
    -test-run-all-install-tests="${TEST_RUN_ALL_INSTALL_TESTS}" \
    -log-level="${TEST_LOG_LEVEL}" \
    -run-tests-local="${METERING_RUN_TESTS_LOCALLY}" \
    -repo-path="${METERING_REPO_PATH}" \
    -repo-version="${METERING_REPO_VERSION}" \
    -run-dev-setup="${METERING_RUN_DEV_TEST_SETUP}" \
    -registry-image="${METERING_ANSIBLE_OPERATOR_IMAGE_REGISTRY}" \
    -subscription-channel="${METERING_OLM_SUBSCRIPTION_CHANNEL}" \
    ${EXTRA_TEST_FLAGS} \
    2>&1 | tee "$TEST_LOG_FILE_PATH" ; TEST_EXIT_CODE=${PIPESTATUS[0]}

# if go-junit-report is installed, create a junit report also
if command -v go-junit-report >/dev/null 2>&1; then
    go-junit-report < "$TEST_LOG_FILE_PATH" > "${TEST_JUNIT_REPORT_FILE_PATH}"
fi

exit "$TEST_EXIT_CODE"
