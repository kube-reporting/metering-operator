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
: "${INSTALL_METHOD:=alm}"
: "${DEPLOY_SCRIPT:=deploy.sh}" # can be deploy-ci.sh

export INSTALL_METHOD
export CHARGEBACK_NAMESPACE

# this script is run inside the container
echo "\$KUBECONFIG=$KUBECONFIG"
echo "\$CHARGEBACK_NAMESPACE=$CHARGEBACK_NAMESPACE"
echo "\$INSTALL_METHOD=$INSTALL_METHOD"

mkdir -p /out
touch "/out/$TEST_LOG_FILE"
touch "/out/$DEPLOY_LOG_FILE"
touch "/out/$TEST_TAP_FILE"
touch "/out/$DEPLOY_POD_LOGS_LOG_FILE"

# fail with the last non-zero exit code (preserves test fail exit code)
set -o pipefail

export SKIP_DELETE_CRDS=true
export DELETE_PVCS=true

function cleanup() {
    echo "Performing cleanup"
    uninstall_chargeback "$INSTALL_METHOD" >> "/out/$DEPLOY_LOG_FILE"

    # kill any background jobs, such as stern
    jobs -p | xargs kill
    # Wait for any jobs
    wait
}

if [ "$DEPLOY_CHARGEBACK" == "true" ]; then
    if [ -n "$DEPLOY_POD_LOGS_LOG_FILE" ]; then
        echo "Capturing pod logs"
        stern -n "$CHARGEBACK_NAMESPACE" '.*' >> "/out/$DEPLOY_POD_LOGS_LOG_FILE" &
    fi

    TMP_DIR="$(mktemp -d)"
    export CHARGEBACK_CR_FILE="$TMP_DIR/custom-chargeback-cr-${DEPLOY_TAG}.yaml"
    export INSTALLER_MANIFEST_DIR="$TMP_DIR/installer_manifests-${DEPLOY_TAG}"
    export CUSTOM_VALUES_FILE="$TMP_DIR/helm-operator-values-${DEPLOY_TAG}.yaml"

    if [ "$CLEANUP_CHARGEBACK" == "true" ]; then
        trap cleanup EXIT SIGINT
    fi
    echo "Deploying Chargeback"
    "${DIR}/${DEPLOY_SCRIPT}" >> "/out/$DEPLOY_LOG_FILE" 2>&1
fi

echo "Running tests"

"$TEST_SCRIPT" 2>&1 \
    | tee -a "/out/$TEST_LOG_FILE" \
    | go tool test2json \
    | tee -a "/out/${TEST_LOG_FILE}.json" \
    | jq -r -s -f "$DIR/tap-output.jq" \
    | tee -a "/out/$TEST_TAP_FILE"

if grep -q '^not' < "/out/$TEST_LOG_FILE"; then
  exit 1
else
  exit 0
fi
