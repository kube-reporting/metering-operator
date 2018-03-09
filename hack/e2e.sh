#!/bin/bash
set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source ${DIR}/util.sh

: "${KUBECONFIG?}"
: "${CHARGEBACK_NAMESPACE:=chargeback-e2e}"

if [[ -z "$NAMESPACE" && -z "$CHARGEBACK_NAMESPACE" ]]; then
    echo "One of \$NAMESPACE or \$CHARGEBACK_NAMESPACE must be set"
    exit 1
fi

: "${DEPLOY_CHARGEBACK:=true}"
: "${CLEANUP_CHARGEBACK:=true}"
: "${INSTALL_METHOD:=alm}"
: "${DEPLOY_SCRIPT:=deploy.sh}" # can be deploy-ci.sh
export INSTALL_METHOD
export CHARGEBACK_NAMESPACE="${CHARGEBACK_NAMESPACE:-NAMESPACE}"

# this script is run inside the container
echo "\$KUBECONFIG=$KUBECONFIG"
echo "\$CHARGEBACK_NAMESPACE=$CHARGEBACK_NAMESPACE"
echo "\$INSTALL_METHOD=$INSTALL_METHOD"

mkdir -p /out
touch /out/test.log
touch /out/deploy.log
touch /out/test-log.tap

# fail with the last non-zero exit code (preserves test fail exit code)
set -o pipefail

export SKIP_DELETE_CRDS=true
export DELETE_PVCS=true

function cleanup() {
    echo "Performing cleanup"
    uninstall_chargeback "$INSTALL_METHOD" >> /out/deploy.log
}


if [ "$DEPLOY_CHARGEBACK" == "true" ]; then
    TMP_DIR="$(mktemp -d)"
    export CHARGEBACK_CR_FILE="$TMP_DIR/custom-chargeback-cr-${DEPLOY_TAG}.yaml"
    export INSTALLER_MANIFEST_DIR="$TMP_DIR/installer_manifests-${DEPLOY_TAG}"
    export CUSTOM_VALUES_FILE="$TMP_DIR/helm-operator-values-${DEPLOY_TAG}.yaml"

    if [ "$CLEANUP_CHARGEBACK" == "true" ]; then
        trap cleanup EXIT SIGINT
    fi
    echo "Deploying Chargeback"
    "${DIR}/${DEPLOY_SCRIPT}" >> /out/deploy.log 2>&1
fi

echo "Running e2e tests"

"$DIR/run-e2e-tests.sh" 2>&1 \
    | tee -a /out/test.log \
    | go tool test2json \
    | tee -a /out/test.json \
    | jq -r -s -f "$DIR/tap-output.jq" \
    | tee -a /out/test-log.tap

if grep -q '^not' < /out/test.log; then
  exit 1
else
  exit 0
fi
