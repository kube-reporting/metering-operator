#!/bin/bash

ROOT_DIR=$(dirname "${BASH_SOURCE}")/..
source "${ROOT_DIR}/hack/common.sh"


: "${METERING_OPERATOR_IMAGE:=${METERING_OPERATOR_IMAGE_REPO}:${METERING_OPERATOR_IMAGE_TAG}}"
: "${LOCAL_METERING_OPERATOR_RUN_INSTALL:=true}"
: "${METERING_INSTALL_SCRIPT:=./hack/openshift-install.sh}"
: "${METERING_OPERATOR_CONTAINER_NAME:=metering-operator}"
: "${ENABLE_DEBUG:=false}"

set -ex

if [ "$LOCAL_METERING_OPERATOR_RUN_INSTALL" == "true" ]; then
    export SKIP_METERING_OPERATOR_DEPLOYMENT=true
    "$METERING_INSTALL_SCRIPT"
***REMOVED***

VOLUMES=(\
    -v "$KUBECONFIG:/kubecon***REMOVED***g" \
    -v /tmp/ansible-operator/runner \
)
if [ -d "$HOME/.minikube" ]; then
    VOLUMES+=(-v "$HOME/.minikube")
***REMOVED***

docker run \
    --name "${METERING_OPERATOR_CONTAINER_NAME}" \
    --rm \
    -u 0:0 \
    "${VOLUMES[@]}" \
    -e KUBECONFIG=/kubecon***REMOVED***g \
    -e OPERATOR_NAME="metering-ansible-operator" \
    -e POD_NAME="metering-ansible-operator" \
    -e WATCH_NAMESPACE="$METERING_NAMESPACE" \
    -e ENABLE_DEBUG="$ENABLE_DEBUG" \
    "${METERING_OPERATOR_IMAGE}"
