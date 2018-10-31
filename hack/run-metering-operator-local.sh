#!/bin/bash

ROOT_DIR=$(dirname "${BASH_SOURCE}")/..
source "${ROOT_DIR}/hack/common.sh"

load_version_vars

: "${METERING_OPERATOR_IMAGE_REPO:=quay.io/coreos/metering-helm-operator}"
: "${METERING_OPERATOR_IMAGE_TAG:=$METERING_VERSION}"
: "${METERING_OPERATOR_IMAGE:=${METERING_OPERATOR_IMAGE_REPO}:${METERING_OPERATOR_IMAGE_TAG}}"
: "${METERING_CHART:=/openshift-metering-0.1.0.tgz}"
: "${LOCAL_METERING_OPERATOR_RUN_INSTALL:=true}"
: "${METERING_INSTALL_SCRIPT:=./hack/openshift-install.sh}"

set -ex

if [ "$LOCAL_METERING_OPERATOR_RUN_INSTALL" == "true" ]; then
    export SKIP_METERING_OPERATOR_DEPLOYMENT=true
    "$METERING_INSTALL_SCRIPT"
fi

docker run \
    -it \
    --rm \
    -v "$KUBECONFIG:/kubeconfig" \
    -e KUBECONFIG=/kubeconfig \
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
    bash -c 'tiller & sleep 2 && run-operator.sh'
