#!/bin/bash
set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source ${DIR}/util.sh

: "${KUBECONFIG:?}"
: "${CHARGEBACK_NAMESPACE:?}"

: "${TEST_SCRIPT:?}"

: "${TEST_LOG_FILE:?}"
: "${DEPLOY_LOG_FILE:?}"
: "${TEST_TAP_FILE:?}"

: "${DEPLOY_POD_LOGS_LOG_FILE:=""}"
: "${DEPLOY_CHARGEBACK:=true}"
: "${CLEANUP_CHARGEBACK:=true}"
: "${INSTALL_METHOD:=direct}"
: "${DEPLOY_SCRIPT:=deploy.sh}" # can be deploy-ci.sh or deploy-openshift-ci.sh
: "${TEST_OUTPUT_DIRECTORY:=/out}"
: "${TEST_OUTPUT_LOG_STDOUT:=false}"

export INSTALL_METHOD
export CHARGEBACK_NAMESPACE

# this script is run inside the container
echo "\$KUBECONFIG=$KUBECONFIG"
echo "\$CHARGEBACK_NAMESPACE=$CHARGEBACK_NAMESPACE"
echo "\$INSTALL_METHOD=$INSTALL_METHOD"

mkdir -p $TEST_OUTPUT_DIRECTORY
touch "$TEST_OUTPUT_DIRECTORY/$TEST_LOG_FILE"
touch "$TEST_OUTPUT_DIRECTORY/$DEPLOY_LOG_FILE"
touch "$TEST_OUTPUT_DIRECTORY/$TEST_TAP_FILE"
touch "$TEST_OUTPUT_DIRECTORY/$DEPLOY_POD_LOGS_LOG_FILE"

# fail with the last non-zero exit code (preserves test fail exit code)
set -o pipefail

export SKIP_DELETE_CRDS=true
export DELETE_PVCS=true

function cleanup() {
    echo "Performing cleanup"
    if [ "$TEST_OUTPUT_LOG_STDOUT" == "true" ]; then
        uninstall_chargeback "$INSTALL_METHOD" | tee "$TEST_OUTPUT_DIRECTORY/$DEPLOY_LOG_FILE"
    else
        uninstall_chargeback "$INSTALL_METHOD" >> "$TEST_OUTPUT_DIRECTORY/$DEPLOY_LOG_FILE"
    fi

    # kill any background jobs, such as stern
    jobs -p | xargs kill
    # Wait for any jobs
    wait
}

if [ "$DEPLOY_CHARGEBACK" == "true" ]; then
    if [ -n "$DEPLOY_POD_LOGS_LOG_FILE" ]; then
        echo "Capturing pod logs"
        if [ "$TEST_OUTPUT_LOG_STDOUT" == "true" ]; then
            stern --timestamps -n "$CHARGEBACK_NAMESPACE" '.*' | tee "$TEST_OUTPUT_DIRECTORY/$DEPLOY_POD_LOGS_LOG_FILE" &
        else
            stern --timestamps -n "$CHARGEBACK_NAMESPACE" '.*' >> "$TEST_OUTPUT_DIRECTORY/$DEPLOY_POD_LOGS_LOG_FILE" &
        fi
    fi

    if [ "$CLEANUP_CHARGEBACK" == "true" ]; then
        trap cleanup EXIT SIGINT
    fi

    echo "Deploying Chargeback"
    if [ "$TEST_OUTPUT_LOG_STDOUT" == "true" ]; then
        "${DIR}/${DEPLOY_SCRIPT}" | tee "$TEST_OUTPUT_DIRECTORY/$DEPLOY_LOG_FILE" 2>&1
    else
        "${DIR}/${DEPLOY_SCRIPT}" >> "$TEST_OUTPUT_DIRECTORY/$DEPLOY_LOG_FILE" 2>&1
    fi
fi

echo "Running tests"

"$TEST_SCRIPT" 2>&1 \
    | tee -a "$TEST_OUTPUT_DIRECTORY/$TEST_LOG_FILE" \
    | go tool test2json \
    | tee -a "$TEST_OUTPUT_DIRECTORY/${TEST_LOG_FILE}.json" \
    | jq -r -s -f "$DIR/tap-output.jq" \
    | tee -a "$TEST_OUTPUT_DIRECTORY/$TEST_TAP_FILE"

if grep -q '^not' < "$TEST_OUTPUT_DIRECTORY/$TEST_LOG_FILE"; then
  exit 1
else
  exit 0
fi
