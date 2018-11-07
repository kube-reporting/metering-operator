#!/bin/bash
set -e

: "${KUBECONFIG:?}"
: "${METERING_NAMESPACE:?}"

: "${TEST_SCRIPT:?}"

: "${TEST_TAP_FILE:=tests.tap}"
: "${TEST_LOG_FILE:=tests.txt}"
: "${DEPLOY_LOG_FILE:=deploy.log}"
: "${DEPLOY_POD_LOGS_LOG_FILE:=pod-logs.log}"
: "${FINAL_POD_LOGS_LOG_FILE:=final-pod-descriptions-logs.log}"

: "${DEPLOY_METERING:=true}"
: "${TEST_METERING:=true}"
: "${CLEANUP_METERING:=true}"
: "${INSTALL_METHOD:=direct}"
: "${DEPLOY_SCRIPT:=deploy.sh}" # can be deploy-ci.sh or deploy-openshift-ci.sh
: "${TEST_OUTPUT_PATH:=/out}"
: "${OUTPUT_TEST_LOG_STDOUT:=false}"
: "${OUTPUT_DEPLOY_LOG_STDOUT:=false}"
: "${OUTPUT_POD_LOG_STDOUT:=false}"
: "${ENABLE_AWS_BILLING:=false}"
: "${ENABLE_AWS_BILLING_TEST:=false}"

ROOT_DIR=$(dirname "${BASH_SOURCE}")/..
source "${ROOT_DIR}/hack/common.sh"

export INSTALL_METHOD
export METERING_NAMESPACE
export KUBECONFIG

# this script is run inside the container
echo "\$KUBECONFIG=$KUBECONFIG"
echo "\$METERING_NAMESPACE=$METERING_NAMESPACE"
echo "\$INSTALL_METHOD=$INSTALL_METHOD"

REPORTS_DIR=$TEST_OUTPUT_PATH/reports
LOG_DIR=$TEST_OUTPUT_PATH/logs
TEST_OUT_DIR=$TEST_OUTPUT_PATH/tests

TEST_LOG_FILE_PATH="${TEST_LOG_FILE_PATH:-$TEST_OUT_DIR/$TEST_LOG_FILE}"
TEST_TAP_FILE_PATH="${TEST_TAP_FILE_PATH:-$TEST_OUT_DIR/$TEST_TAP_FILE}"
DEPLOY_LOG_FILE_PATH="${DEPLOY_LOG_FILE_PATH:-$LOG_DIR/$DEPLOY_LOG_FILE}"
DEPLOY_POD_LOGS_LOG_FILE_PATH="${DEPLOY_POD_LOGS_LOG_FILE_PATH:-$LOG_DIR/$DEPLOY_POD_LOGS_LOG_FILE}"
FINAL_POD_LOGS_LOG_FILE_PATH="${FINAL_POD_LOGS_LOG_FILE_PATH:-$LOG_DIR/$FINAL_POD_LOGS_LOG_FILE}"

mkdir -p $TEST_OUT_DIR $LOG_DIR $REPORTS_DIR
touch "$TEST_LOG_FILE_PATH"
touch "$TEST_TAP_FILE_PATH"
touch "$DEPLOY_LOG_FILE_PATH"
touch "$DEPLOY_POD_LOGS_LOG_FILE_PATH"
touch "$FINAL_POD_LOGS_LOG_FILE_PATH"

# fail with the last non-zero exit code (preserves test fail exit code)
set -o pipefail

export SKIP_DELETE_CRDS=true
export DELETE_PVCS=true
export TEST_RESULT_REPORT_OUTPUT_DIRECTORY="$REPORTS_DIR"

function cleanup() {
    echo "Performing cleanup"

    if [ -n "$FINAL_POD_LOGS_LOG_FILE" ]; then
        echo "Capturing final pod logs"
        capture_pod_logs "$METERING_NAMESPACE" >> "$FINAL_POD_LOGS_LOG_FILE_PATH"
        echo "Finished capturing final pod logs"
    fi

    echo "Running uninstall"
    if [ "$OUTPUT_DEPLOY_LOG_STDOUT" == "true" ]; then
        uninstall_metering "$INSTALL_METHOD" | tee -a "$DEPLOY_LOG_FILE_PATH"
    else
        uninstall_metering "$INSTALL_METHOD" >> "$DEPLOY_LOG_FILE_PATH"
    fi

    # kill any background jobs, such as stern
    jobs -p | xargs kill
    # Wait for any jobs
    wait
}

if [ "$DEPLOY_METERING" == "true" ]; then
    if [ -n "$DEPLOY_POD_LOGS_LOG_FILE" ]; then
        echo "Capturing pod logs"
        if [ "$OUTPUT_POD_LOG_STDOUT" == "true" ]; then
            stern --timestamps -n "$METERING_NAMESPACE" '.*' | tee -a "$DEPLOY_POD_LOGS_LOG_FILE_PATH" &
        else
            stern --timestamps -n "$METERING_NAMESPACE" '.*' >> "$DEPLOY_POD_LOGS_LOG_FILE_PATH" &
        fi
    fi

    if [ "$CLEANUP_METERING" == "true" ]; then
        trap cleanup EXIT SIGINT
    fi

    echo "Deploying Metering"
    if [ "$OUTPUT_DEPLOY_LOG_STDOUT" == "true" ]; then
        "$ROOT_DIR/hack/${DEPLOY_SCRIPT}" | tee -a "$DEPLOY_LOG_FILE_PATH" 2>&1
    else
        "$ROOT_DIR/hack/${DEPLOY_SCRIPT}" >> "$DEPLOY_LOG_FILE_PATH" 2>&1
    fi
fi

if [ "$TEST_METERING" == "true" ]; then
    echo "Running tests"


    if [ "$OUTPUT_TEST_LOG_STDOUT" == "true" ]; then
        tail -f "$TEST_LOG_FILE_PATH" &
    fi

    "$TEST_SCRIPT" 2>&1 \
        | tee -a "$TEST_LOG_FILE_PATH" \
        | "$ROOT_DIR/bin/test2json" \
        | tee -a "${TEST_LOG_FILE}.json_PATH" \
        | jq -r -s -f "$ROOT_DIR/hack/tap-output.jq" \
        | tee -a "$TEST_TAP_FILE_PATH"

    if grep -q '^not' < "$TEST_LOG_FILE_PATH"; then
      exit 1
    else
      exit 0
    fi
fi
