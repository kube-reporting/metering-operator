#!/bin/bash
set -e

# TODO: handle `make e2e-local`
# TODO: handle uninstall metering before running tests
# TODO: add a metering-poll-timeout for the pod readiness?

: "${KUBECONFIG:?}"

ROOT_DIR=$(dirname "${BASH_SOURCE}")/..
source "${ROOT_DIR}/hack/common.sh"
source "${ROOT_DIR}/hack/lib/tests.sh"

METERING_NAMESPACE="${METERING_E2E_NAMESPACE:=${METERING_NAMESPACE}-e2e}"
TEST_LOG_FILE="${TEST_LOG_FILE:-e2e-tests.log}"

METERING_SHORT_TESTS=${METERING_SHORT_TESTS:-false}
METERING_DEPLOY_MANIFESTS_PATH=${METERING_DEPLOY_MANIFESTS_PATH:-${ROOT_DIR}/manifests/deploy}
METERING_CLEANUP_SCRIPT_PATH=${METERING_CLEANUP_SCRIPT_PATH:-${ROOT_DIR}/hack/run-test-cleanup.sh}
METERING_HTTPS_API=${METERING_HTTPS_API:-true}
METERING_USE_KUBE_PROXY_FOR_REPORTING_API=${METERING_USE_KUBE_PROXY_FOR_REPORTING_API:-false}
METERING_USE_ROUTE_FOR_REPORTING_API=${METERING_USE_ROUTE_FOR_REPORTING_API:-true}
METERING_REPORTING_API_URL=${METERING_REPORTING_API_URL:-""}
TEST_OUTPUT_PATH=${TEST_OUTPUT_PATH:-""}

echo "\$KUBECONFIG=$KUBECONFIG"
echo "\$METERING_NAMESPACE=$METERING_NAMESPACE"
echo "\$METERING_OPERATOR_IMAGE_REPO=$METERING_OPERATOR_IMAGE_REPO"
echo "\$REPORTING_OPERATOR_IMAGE_REPO=$REPORTING_OPERATOR_IMAGE_REPO"
echo "\$METERING_OPERATOR_IMAGE_TAG=$METERING_OPERATOR_IMAGE_TAG"
echo "\$REPORTING_OPERATOR_IMAGE_TAG=$REPORTING_OPERATOR_IMAGE_TAG"
echo "\$TEST_OUTPUT_PATH=$TEST_OUTPUT_PATH"

function remove_namespaces() {
    echo "Removing namespaces with the 'name=e2e-testing' label"
    kubectl delete ns -l "name=e2e-testing" || true
}

trap remove_namespaces SIGINT

set -x
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
    -log-level debug \
    "$@"
