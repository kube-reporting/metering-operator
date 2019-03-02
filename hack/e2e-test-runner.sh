#!/bin/bash
set -e

: "${KUBECONFIG:?}"
: "${METERING_NAMESPACE:?}"

: "${TEST_SCRIPT:?}"

: "${TEST_TAP_FILE:=tests.tap}"
: "${TEST_LOG_FILE:=tests.txt}"
: "${DEPLOY_LOG_FILE:=deploy.log}"
: "${DEPLOY_POD_LOGS_LOG_FILE:=pod-logs.log}"

: "${DEPLOY_METERING:=true}"
: "${TEST_METERING:=true}"
: "${CLEANUP_METERING_NAMESPACE:=true}"
# can be deploy.sh, deploy-custom.sh, deploy-e2e.sh, deploy-integration.sh
: "${DEPLOY_SCRIPT:=deploy.sh}"
: "${TEST_OUTPUT_PATH:="$(mktemp -d)"}"
: "${OUTPUT_TEST_LOG_STDOUT:=true}"
: "${OUTPUT_DEPLOY_LOG_STDOUT:=true}"
: "${OUTPUT_POD_LOG_STDOUT:=false}"
: "${ENABLE_AWS_BILLING:=false}"
: "${ENABLE_AWS_BILLING_TEST:=false}"

ROOT_DIR=$(dirname "${BASH_SOURCE}")/..
source "${ROOT_DIR}/hack/common.sh"

export METERING_NAMESPACE
export KUBECONFIG

# this script is run inside the container
echo "\$KUBECONFIG=$KUBECONFIG"
echo "\$METERING_NAMESPACE=$METERING_NAMESPACE"
echo "\$DEPLOY_SCRIPT=$DEPLOY_SCRIPT"
echo "\$TEST_OUTPUT_PATH=$TEST_OUTPUT_PATH"

REPORTS_DIR=$TEST_OUTPUT_PATH/reports
LOG_DIR=$TEST_OUTPUT_PATH/logs
TEST_OUTPUT_DIR=$TEST_OUTPUT_PATH/tests

TEST_LOG_FILE_PATH="${TEST_LOG_FILE_PATH:-$TEST_OUTPUT_DIR/$TEST_LOG_FILE}"
TEST_TAP_FILE_PATH="${TEST_TAP_FILE_PATH:-$TEST_OUTPUT_DIR/$TEST_TAP_FILE}"
DEPLOY_LOG_FILE_PATH="${DEPLOY_LOG_FILE_PATH:-$LOG_DIR/$DEPLOY_LOG_FILE}"
DEPLOY_POD_LOGS_LOG_FILE_PATH="${DEPLOY_POD_LOGS_LOG_FILE_PATH:-$LOG_DIR/$DEPLOY_POD_LOGS_LOG_FILE}"

mkdir -p "$TEST_OUTPUT_DIR" "$LOG_DIR" "$REPORTS_DIR"
touch "$TEST_LOG_FILE_PATH"
touch "$TEST_TAP_FILE_PATH"
touch "$DEPLOY_LOG_FILE_PATH"
touch "$DEPLOY_POD_LOGS_LOG_FILE_PATH"

# fail with the last non-zero exit code (preserves test fail exit code)
set -o pipefail

export SKIP_DELETE_CRDS=true
export DELETE_PVCS=true
export TEST_RESULT_REPORT_OUTPUT_DIRECTORY="$REPORTS_DIR"

function cleanup() {
    exit_status=$?

    echo "Performing cleanup"

    echo "Storing pod descriptions and logs at $LOG_DIR"
    echo "Capturing pod descriptions"
    PODS="$(kubectl get pods --no-headers --namespace "$METERING_NAMESPACE")"
    echo "$PODS" | awk '{ print $1 }' | while read -r pod; do
        if [[ -n "$pod" ]]; then
            echo "Capturing pod $pod description"
            if ! kubectl describe pod --namespace "$METERING_NAMESPACE" "$pod" >> "$LOG_DIR/${pod}-description.txt"; then
                echo "Error capturing pod $pod description"
            ***REMOVED***
        ***REMOVED***
    done

    echo "Capturing pod logs"
    echo "$PODS" | awk '{ print $1 }' | while read -r pod; do
        # There can be multiple containers within a pod. We need to iterate
        # over each of those
        containers=$(kubectl get pods -o jsonpath="{.spec.containers[*].name}" --namespace "$METERING_NAMESPACE" "$pod")
        for container in $containers; do
            echo "Capturing pod $pod container $container logs"
            if ! kubectl logs --namespace "$METERING_NAMESPACE" -c "$container" "$pod" >> "$LOG_DIR/${pod}-${container}.log"; then
                echo "Error capturing pod $pod container $container logs"
            ***REMOVED***
        done
    done


    if [ "$CLEANUP_METERING_NAMESPACE" == "true" ]; then
        echo "Deleting namespace"
        kubectl delete ns "$METERING_NAMESPACE" || true
    ***REMOVED***

    # kill any background jobs, such as stern
    jobs -p | xargs kill
    # Wait for any jobs
    wait

    exit "$exit_status"
}

if [ "$DEPLOY_METERING" == "true" ]; then
    if [ -n "$DEPLOY_POD_LOGS_LOG_FILE" ]; then
        echo "Streaming pod logs"
        echo "Storing logs at $DEPLOY_POD_LOGS_LOG_FILE_PATH"
        if [ "$OUTPUT_POD_LOG_STDOUT" == "true" ]; then
            stern --timestamps -n "$METERING_NAMESPACE" '.*' | tee -a "$DEPLOY_POD_LOGS_LOG_FILE_PATH" &
        ***REMOVED***
            stern --timestamps -n "$METERING_NAMESPACE" '.*' >> "$DEPLOY_POD_LOGS_LOG_FILE_PATH" &
        ***REMOVED***
    ***REMOVED***

    trap cleanup EXIT

    echo "Deploying Metering"
    echo "Storing deploy logs at $DEPLOY_LOG_FILE_PATH"
    if [ "$OUTPUT_DEPLOY_LOG_STDOUT" == "true" ]; then
        "$ROOT_DIR/hack/${DEPLOY_SCRIPT}" | tee -a "$DEPLOY_LOG_FILE_PATH" 2>&1
    ***REMOVED***
        "$ROOT_DIR/hack/${DEPLOY_SCRIPT}" >> "$DEPLOY_LOG_FILE_PATH" 2>&1
    ***REMOVED***
***REMOVED***

if [ "$TEST_METERING" == "true" ]; then
    echo "Running tests"

    echo "Storing test results at $TEST_OUTPUT_DIR"

    if [ "$OUTPUT_TEST_LOG_STDOUT" == "true" ]; then
        tail -f "$TEST_LOG_FILE_PATH" &
    ***REMOVED***

    "$TEST_SCRIPT" 2>&1 \
        | tee -a "$TEST_LOG_FILE_PATH" \
        | "$ROOT_DIR/bin/test2json" \
        | tee -a "${TEST_LOG_FILE_PATH}.json" \
        | "$FAQ_BIN" -f json -o json -M -c -r -s -F "$ROOT_DIR/hack/tap-output.jq" \
        | tee -a "$TEST_TAP_FILE_PATH"

    if grep -q '^not' < "$TEST_LOG_FILE_PATH"; then
      exit 1
    ***REMOVED***
      exit 0
    ***REMOVED***
***REMOVED***
