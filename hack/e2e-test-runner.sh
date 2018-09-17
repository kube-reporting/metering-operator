#!/bin/bash
set -e

: "${KUBECONFIG:?}"
: "${METERING_NAMESPACE:?}"

: "${TEST_SCRIPT:?}"

: "${TEST_LOG_FILE:?}"
: "${DEPLOY_LOG_FILE:?}"
: "${TEST_TAP_FILE:?}"

: "${DEPLOY_POD_LOGS_LOG_FILE:=""}"
: "${FINAL_POD_LOGS_LOG_FILE:=""}"

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

mkdir -p $TEST_OUTPUT_PATH
touch "$TEST_OUTPUT_PATH/$TEST_LOG_FILE"
touch "$TEST_OUTPUT_PATH/$DEPLOY_LOG_FILE"
touch "$TEST_OUTPUT_PATH/$TEST_TAP_FILE"
touch "$TEST_OUTPUT_PATH/$DEPLOY_POD_LOGS_LOG_FILE"
touch "$TEST_OUTPUT_PATH/$FINAL_POD_LOGS_LOG_FILE"

# fail with the last non-zero exit code (preserves test fail exit code)
set -o pipefail

export SKIP_DELETE_CRDS=true
export DELETE_PVCS=true

function cleanup() {
    echo "Performing cleanup"

    if [ -n "$FINAL_POD_LOGS_LOG_FILE" ]; then
        echo "Capturing ***REMOVED***nal pod logs"
        capture_pod_logs "$METERING_NAMESPACE" >> "$TEST_OUTPUT_PATH/$FINAL_POD_LOGS_LOG_FILE"
        echo "Finished capturing ***REMOVED***nal pod logs"
    ***REMOVED***

    echo "Running uninstall"
    if [ "$OUTPUT_DEPLOY_LOG_STDOUT" == "true" ]; then
        uninstall_metering "$INSTALL_METHOD" | tee -a "$TEST_OUTPUT_PATH/$DEPLOY_LOG_FILE"
    ***REMOVED***
        uninstall_metering "$INSTALL_METHOD" >> "$TEST_OUTPUT_PATH/$DEPLOY_LOG_FILE"
    ***REMOVED***

    # kill any background jobs, such as stern
    jobs -p | xargs kill
    # Wait for any jobs
    wait
}

if [ "$DEPLOY_METERING" == "true" ]; then
    if [ -n "$DEPLOY_POD_LOGS_LOG_FILE" ]; then
        echo "Capturing pod logs"
        if [ "$OUTPUT_POD_LOG_STDOUT" == "true" ]; then
            stern --timestamps -n "$METERING_NAMESPACE" '.*' | tee -a "$TEST_OUTPUT_PATH/$DEPLOY_POD_LOGS_LOG_FILE" &
        ***REMOVED***
            stern --timestamps -n "$METERING_NAMESPACE" '.*' >> "$TEST_OUTPUT_PATH/$DEPLOY_POD_LOGS_LOG_FILE" &
        ***REMOVED***
    ***REMOVED***

    if [ "$CLEANUP_METERING" == "true" ]; then
        trap cleanup EXIT SIGINT
    ***REMOVED***

    echo "Deploying Metering"
    if [ "$OUTPUT_DEPLOY_LOG_STDOUT" == "true" ]; then
        "$ROOT_DIR/hack/${DEPLOY_SCRIPT}" | tee -a "$TEST_OUTPUT_PATH/$DEPLOY_LOG_FILE" 2>&1
    ***REMOVED***
        "$ROOT_DIR/hack/${DEPLOY_SCRIPT}" >> "$TEST_OUTPUT_PATH/$DEPLOY_LOG_FILE" 2>&1
    ***REMOVED***
***REMOVED***

if [ "$TEST_METERING" == "true" ]; then
    echo "Running tests"


    if [ "$OUTPUT_TEST_LOG_STDOUT" == "true" ]; then
        tail -f "$TEST_OUTPUT_PATH/$TEST_LOG_FILE" &
    ***REMOVED***

    "$TEST_SCRIPT" 2>&1 \
        | tee -a "$TEST_OUTPUT_PATH/$TEST_LOG_FILE" \
        | "$ROOT_DIR/bin/test2json" \
        | tee -a "$TEST_OUTPUT_PATH/${TEST_LOG_FILE}.json" \
        | jq -r -s -f "$ROOT_DIR/hack/tap-output.jq" \
        | tee -a "$TEST_OUTPUT_PATH/$TEST_TAP_FILE"

    if grep -q '^not' < "$TEST_OUTPUT_PATH/$TEST_LOG_FILE"; then
      exit 1
    ***REMOVED***
      exit 0
    ***REMOVED***
***REMOVED***
