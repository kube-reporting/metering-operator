#!/bin/bash

CUR_DIR=$(dirname "${BASH_SOURCE[0]}")

set -e
set -u

: "${ENABLE_DEBUG:=false}"

if [ "$ENABLE_DEBUG" == "true" ]; then
    set -x
***REMOVED***

: "${KUBECTL_BIN:=kubectl}"
: "${HELM_BIN:=helm}"

RELEASE_NAME=${1:?}
CHART=${2:?}
NAMESPACE=${3:?}

EXTRA_ARGS=( "${@:4}" )

helmTemplate() {
    "$HELM_BIN" template \
        "$CHART" \
        --name "$RELEASE_NAME" \
        --namespace "$NAMESPACE" \
        "${EXTRA_ARGS[@]+"${EXTRA_ARGS[@]}"}"
}

helmTemplate
