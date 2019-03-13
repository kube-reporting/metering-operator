#!/bin/bash

: "${KUBECONFIG:?}"
: "${METERING_NAMESPACE:?}"

: "${TEST_SCRIPT:?}"

: "${TEST_TAP_FILE:=tests.tap}"
: "${TEST_JUNIT_REPORT_FILE:=tests-junit.xml}"
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

LOG_DIR=$TEST_OUTPUT_PATH/logs
TEST_OUTPUT_DIR=$TEST_OUTPUT_PATH/tests
REPORT_RESULTS_DIR=$TEST_OUTPUT_PATH/report_results
REPORTS_DIR=$TEST_OUTPUT_PATH/reports
DATASOURCES_DIR=$TEST_OUTPUT_PATH/reportdatasources
REPORTGENERATIONQUERIES_DIR=$TEST_OUTPUT_PATH/reportgenerationqueries

TEST_LOG_FILE_PATH="${TEST_LOG_FILE_PATH:-$TEST_OUTPUT_DIR/$TEST_LOG_FILE}"
TEST_TAP_FILE_PATH="${TEST_TAP_FILE_PATH:-$TEST_OUTPUT_DIR/$TEST_TAP_FILE}"
TEST_JUNIT_REPORT_FILE_PATH="${TEST_JUNIT_REPORT_FILE_PATH:-$TEST_OUTPUT_DIR/$TEST_JUNIT_REPORT_FILE}"
DEPLOY_LOG_FILE_PATH="${DEPLOY_LOG_FILE_PATH:-$LOG_DIR/$DEPLOY_LOG_FILE}"
DEPLOY_POD_LOGS_LOG_FILE_PATH="${DEPLOY_POD_LOGS_LOG_FILE_PATH:-$LOG_DIR/$DEPLOY_POD_LOGS_LOG_FILE}"

mkdir -p "$LOG_DIR" "$TEST_OUTPUT_DIR" "$REPORT_RESULTS_DIR" "$REPORTS_DIR" "$DATASOURCES_DIR" "$REPORTGENERATIONQUERIES_DIR"

export SKIP_DELETE_CRDS=true
export DELETE_PVCS=true
export TEST_RESULT_REPORT_OUTPUT_DIRECTORY="$REPORT_RESULTS_DIR"

function cleanup() {
    exit_status=$?

    echo "Performing cleanup"

    echo "Storing pod descriptions and logs at $LOG_DIR"
    echo "Capturing pod descriptions"
    PODS="$(kubectl get pods --no-headers --namespace "$METERING_NAMESPACE" -o name | cut -d/ -f2)"
    while read -r pod; do
        if [[ -n "$pod" ]]; then
            echo "Capturing pod $pod description"
            if ! kubectl describe pod --namespace "$METERING_NAMESPACE" "$pod" > "$LOG_DIR/${pod}-description.txt"; then
                echo "Error capturing pod $pod description"
            ***REMOVED***
        ***REMOVED***
    done <<< "$PODS"

    echo "Capturing pod logs"
    while read -r pod; do
        if [[ -z "$pod" ]]; then
            continue
        ***REMOVED***
        # There can be multiple containers within a pod. We need to iterate
        # over each of those
        containers=$(kubectl get pods -o jsonpath="{.spec.containers[*].name}" --namespace "$METERING_NAMESPACE" "$pod")
        for container in $containers; do
            echo "Capturing pod $pod container $container logs"
            if ! kubectl logs --namespace "$METERING_NAMESPACE" -c "$container" "$pod" > "$LOG_DIR/${pod}-${container}.log"; then
                echo "Error capturing pod $pod container $container logs"
            ***REMOVED***
        done
    done <<< "$PODS"

    echo "Capturing Metering ReportDataSources"
    DATASOURCES="$(kubectl get reportdatasources --no-headers --namespace "$METERING_NAMESPACE" -o name | cut -d/ -f2)"
    while read -r datasource; do
        if [[ -n "$datasource" ]]; then
            echo "Capturing ReportDataSource $datasource as json"
            if ! kubectl get reportdatasource "$datasource" --namespace "$METERING_NAMESPACE" -o json > "$DATASOURCES_DIR/${datasource}.json"; then
                echo "Error getting $datasource as json"
            ***REMOVED***
        ***REMOVED***
    done <<< "$DATASOURCES"

    echo "Capturing Metering ReportGenerationQueries"
    RGQS="$(kubectl get reportgenerationqueries --no-headers --namespace "$METERING_NAMESPACE" -o name | cut -d/ -f2)"
    while read -r rgq; do
        if [[ -n "$rgq" ]]; then
            echo "Capturing ReportGenerationQuery $rgq as json"
            if ! kubectl get reportgenerationquery "$rgq" --namespace "$METERING_NAMESPACE" -o json > "$REPORTGENERATIONQUERIES_DIR/${rgq}.json"; then
                echo "Error getting $rgq as json"
            ***REMOVED***
        ***REMOVED***
    done <<< "$RGQS"

    echo "Capturing Metering Reports"
    REPORTS="$(kubectl get reports --no-headers --namespace "$METERING_NAMESPACE" -o name | cut -d/ -f2)"
    while read -r report; do
        if [[ -n "$report" ]]; then
            echo "Capturing Report $report as json"
            if ! kubectl get report "$report" --namespace "$METERING_NAMESPACE" -o json > "$REPORTS_DIR/${report}.json"; then
                echo "Error getting $report as json"
            ***REMOVED***
        ***REMOVED***
    done <<< "$REPORTS"

    if [ "$CLEANUP_METERING_NAMESPACE" == "true" ]; then
        echo "Deleting namespace"
        kubectl delete ns "$METERING_NAMESPACE" || true
    ***REMOVED***

    # kill any background jobs, such as stern
    kill $(jobs -rp)
    # Wait for any jobs
    wait 2>/dev/null

    exit "$exit_status"
}

if [ "$DEPLOY_METERING" == "true" ]; then
    if [ -n "$DEPLOY_POD_LOGS_LOG_FILE" ]; then
        echo "Streaming pod logs"
        echo "Storing logs at $DEPLOY_POD_LOGS_LOG_FILE_PATH"
        if [ "$OUTPUT_POD_LOG_STDOUT" == "true" ]; then
            stern --timestamps -n "$METERING_NAMESPACE" '.*' | tee -a "$DEPLOY_POD_LOGS_LOG_FILE_PATH" &
        ***REMOVED***
            stern --timestamps -n "$METERING_NAMESPACE" '.*' > "$DEPLOY_POD_LOGS_LOG_FILE_PATH" &
        ***REMOVED***
    ***REMOVED***

    trap cleanup EXIT

    echo "Deploying Metering"
    echo "Storing deploy logs at $DEPLOY_LOG_FILE_PATH"
    if [ "$OUTPUT_DEPLOY_LOG_STDOUT" == "true" ]; then
        time "$ROOT_DIR/hack/${DEPLOY_SCRIPT}" | tee -a "$DEPLOY_LOG_FILE_PATH" 2>&1
    ***REMOVED***
        time "$ROOT_DIR/hack/${DEPLOY_SCRIPT}" > "$DEPLOY_LOG_FILE_PATH" 2>&1
    ***REMOVED***
***REMOVED***

if [ "$TEST_METERING" == "true" ]; then
    echo "Storing test results at $TEST_OUTPUT_DIR"

    set +e
    set +o pipefail
    echo "Running tests"
    # Log the results and store in a ***REMOVED***le
    time "$TEST_SCRIPT" 2>&1 | tee "$TEST_LOG_FILE_PATH"; TEST_EXIT_CODE=${PIPESTATUS[0]}

    echo "Converting test results"
    # turn the results into json
    "$ROOT_DIR/bin/test2json" < "$TEST_LOG_FILE_PATH"  > "${TEST_LOG_FILE_PATH}.json"

    # turn the json results into tap output
    "$FAQ_BIN" -f json -o json -M -c -r -s -F "$ROOT_DIR/hack/tap-output.jq" < "${TEST_LOG_FILE_PATH}.json" > "$TEST_TAP_FILE_PATH"

    # Log the tap results too
    cat "$TEST_TAP_FILE_PATH"

    # if go-junit-report is installed, create a junit report also
    if command -v go-junit-report >/dev/null 2>&1; then
        go-junit-report < "$TEST_LOG_FILE_PATH" > "${TEST_JUNIT_REPORT_FILE_PATH}"
    ***REMOVED***

    exit "$TEST_EXIT_CODE"
***REMOVED***
