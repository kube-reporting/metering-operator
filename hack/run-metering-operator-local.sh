#!/bin/bash

ROOT_DIR=$(dirname "${BASH_SOURCE}")/..
source "${ROOT_DIR}/hack/common.sh"


: "${METERING_OPERATOR_IMAGE:=${METERING_OPERATOR_IMAGE_REPO}:${METERING_OPERATOR_IMAGE_TAG}}"
: "${METERING_CHART:=/openshift-metering}"
: "${LOCAL_METERING_OPERATOR_RUN_INSTALL:=true}"
: "${METERING_INSTALL_SCRIPT:=./hack/openshift-install.sh}"
: "${METERING_OPERATOR_CONTAINER_NAME:=metering-operator}"
: "${ENABLE_DEBUG:=false}"

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
    -v /tmp/ansible-operator/runner \
    -e KUBECONFIG=/kubecon***REMOVED***g \
    -e HELM_CHART_PATH="$METERING_CHART" \
    -e OPERATOR_NAME="metering-ansible-operator" \
    -e POD_NAME="metering-ansible-operator" \
    -e WATCH_NAMESPACE="$METERING_NAMESPACE" \
    -e ENABLE_DEBUG="$ENABLE_DEBUG" \
    "${METERING_OPERATOR_IMAGE}"
