#!/bin/bash

ROOT_DIR=$(dirname "${BASH_SOURCE}")/..
source "${ROOT_DIR}/hack/common.sh"


: "${METERING_OPERATOR_IMAGE:=${METERING_OPERATOR_IMAGE_REPO}:${METERING_OPERATOR_IMAGE_TAG}}"
: "${METERING_CHART:=/openshift-metering-0.1.0.tgz}"
: "${LOCAL_METERING_OPERATOR_RUN_INSTALL:=true}"
: "${METERING_INSTALL_SCRIPT:=./hack/openshift-install.sh}"
: "${METERING_OPERATOR_CONTAINER_NAME:=metering-operator}"

set -ex

if [ "$LOCAL_METERING_OPERATOR_RUN_INSTALL" == "true" ]; then
    export SKIP_METERING_OPERATOR_DEPLOYMENT=true
    "$METERING_INSTALL_SCRIPT"
***REMOVED***

docker run \
    --name "${METERING_OPERATOR_CONTAINER_NAME}" \
    --rm \
    -u 0:0 \
    -v "$KUBECONFIG:/kubecon***REMOVED***g" \
    -e KUBECONFIG=/kubecon***REMOVED***g \
    -e HELM_RELEASE_CRD_NAME="Metering" \
    -e HELM_RELEASE_CRD_API_GROUP="metering.openshift.io" \
    -e HELM_CHART_PATH="$METERING_CHART" \
    -e MY_POD_NAME="local-pod" \
    -e MY_POD_NAMESPACE="$METERING_NAMESPACE" \
    -e HELM_HOST="127.0.0.1:44134" \
    -e HELM_WAIT="false" \
    -e HELM_RECONCILE_INTERVAL_SECONDS="5" \
    -e RELEASE_HISTORY_LIMIT="3" \
    -e TILLER_NAMESPACE="$METERING_NAMESPACE" \
    -e TILLER_HISTORY_MAX="3" \
    "${METERING_OPERATOR_IMAGE}" \
    bash -c '
    cleanup() {
        echo "Got signal!"
        trap - SIGINT SIGTERM # clear the trap
        kill -- -$$ # Sends SIGTERM to child/sub processes
    }
    trap cleanup SIGINT SIGTERM
    tiller &
    run-operator.sh &
    wait
    '
