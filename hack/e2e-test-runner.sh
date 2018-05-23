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
: "${CLEANUP_METERING:=true}"
: "${INSTALL_METHOD:=direct}"
: "${DEPLOY_SCRIPT:=deploy.sh}" # can be deploy-ci.sh or deploy-openshift-ci.sh
: "${TEST_OUTPUT_DIRECTORY:=/out}"
: "${TEST_OUTPUT_LOG_STDOUT:=false}"
: "${ENABLE_AWS_BILLING:=false}"
: "${ENABLE_AWS_BILLING_TEST:=false}"

ROOT_DIR=$(dirname "${BASH_SOURCE}")/..
source "${ROOT_DIR}/hack/common.sh"

export INSTALL_METHOD
export METERING_NAMESPACE

# this script is run inside the container
echo "\$KUBECONFIG=$KUBECONFIG"
echo "\$METERING_NAMESPACE=$METERING_NAMESPACE"
echo "\$INSTALL_METHOD=$INSTALL_METHOD"

mkdir -p $TEST_OUTPUT_DIRECTORY
touch "$TEST_OUTPUT_DIRECTORY/$TEST_LOG_FILE"
touch "$TEST_OUTPUT_DIRECTORY/$DEPLOY_LOG_FILE"
touch "$TEST_OUTPUT_DIRECTORY/$TEST_TAP_FILE"
touch "$TEST_OUTPUT_DIRECTORY/$DEPLOY_POD_LOGS_LOG_FILE"
touch "$TEST_OUTPUT_DIRECTORY/$FINAL_POD_LOGS_LOG_FILE"

# fail with the last non-zero exit code (preserves test fail exit code)
set -o pipefail

export SKIP_DELETE_CRDS=true
export DELETE_PVCS=true

# Take and modified slightly from https://github.com/kubernetes/charts/blob/f1711c220988b69e530263dc924eaed0a759e441/test/changed.sh#L42
capture_pod_logs() {
    # List all logs for all containers in all pods for the namespace which was
    kubectl get pods --show-all --no-headers --namespace "$METERING_NAMESPACE" | awk '{ print $1 }' | while read -r pod; do
        if [[ -n "$pod" ]]; then
            printf '===Details from pod %s:===\n' "$pod"

            printf '...Description of pod %s:...\n' "$pod"
            kubectl describe pod --namespace "$METERING_NAMESPACE" "$pod" || true
            printf '...End of description for pod %s...\n\n' "$pod"

            # There can be multiple containers within a pod. We need to iterate
            # over each of those
            containers=$(kubectl get pods --show-all -o jsonpath="{.spec.containers[*].name}" --namespace "$METERING_NAMESPACE" "$pod")
            for container in $containers; do
                printf -- '---Logs from container %s in pod %s:---\n' "$container" "$pod"
                kubectl logs --namespace "$METERING_NAMESPACE" -c "$container" "$pod" || true
                printf -- '---End of logs for container %s in pod %s---\n\n' "$container" "$pod"
            done

            printf '===End of details for pod %s===\n' "$pod"
        fi
    done
}

function cleanup() {
    echo "Performing cleanup"

    if [ -n "$FINAL_POD_LOGS_LOG_FILE" ]; then
        echo "Capturing final pod logs"
        capture_pod_logs >> "$TEST_OUTPUT_DIRECTORY/$FINAL_POD_LOGS_LOG_FILE"
        echo "Finished capturing final pod logs"
    fi

    echo "Running uninstall"
    if [ "$TEST_OUTPUT_LOG_STDOUT" == "true" ]; then
        uninstall_metering "$INSTALL_METHOD" | tee "$TEST_OUTPUT_DIRECTORY/$DEPLOY_LOG_FILE"
    else
        uninstall_metering "$INSTALL_METHOD" >> "$TEST_OUTPUT_DIRECTORY/$DEPLOY_LOG_FILE"
    fi

    # kill any background jobs, such as stern
    jobs -p | xargs kill
    # Wait for any jobs
    wait
}

if [ "$DEPLOY_METERING" == "true" ]; then
    if [ -n "$DEPLOY_POD_LOGS_LOG_FILE" ]; then
        echo "Capturing pod logs"
        if [ "$TEST_OUTPUT_LOG_STDOUT" == "true" ]; then
            stern --timestamps -n "$METERING_NAMESPACE" '.*' | tee "$TEST_OUTPUT_DIRECTORY/$DEPLOY_POD_LOGS_LOG_FILE" &
        else
            stern --timestamps -n "$METERING_NAMESPACE" '.*' >> "$TEST_OUTPUT_DIRECTORY/$DEPLOY_POD_LOGS_LOG_FILE" &
        fi
    fi

    if [ "$CLEANUP_METERING" == "true" ]; then
        trap cleanup EXIT SIGINT
    fi

    echo "Deploying Metering"
    if [ "$TEST_OUTPUT_LOG_STDOUT" == "true" ]; then
        "$ROOT_DIR/hack/${DEPLOY_SCRIPT}" | tee "$TEST_OUTPUT_DIRECTORY/$DEPLOY_LOG_FILE" 2>&1
    else
        "$ROOT_DIR/hack/${DEPLOY_SCRIPT}" >> "$TEST_OUTPUT_DIRECTORY/$DEPLOY_LOG_FILE" 2>&1
    fi
fi

echo "Running tests"

"$TEST_SCRIPT" 2>&1 \
    | tee -a "$TEST_OUTPUT_DIRECTORY/$TEST_LOG_FILE" \
    | go tool test2json \
    | tee -a "$TEST_OUTPUT_DIRECTORY/${TEST_LOG_FILE}.json" \
    | jq -r -s -f "$ROOT_DIR/hack/tap-output.jq" \
    | tee -a "$TEST_OUTPUT_DIRECTORY/$TEST_TAP_FILE"

if grep -q '^not' < "$TEST_OUTPUT_DIRECTORY/$TEST_LOG_FILE"; then
  exit 1
else
  exit 0
fi
