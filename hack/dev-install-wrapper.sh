#!/bin/bash
set -o errexit
set -o nounset
set -o pipefail

ROOT_DIR=$(dirname "${BASH_SOURCE[0]}")/..

IMAGE_TAG=""
if [ -n "$1" ]; then
    IMAGE_TAG="$1"
    echo "using $1 as metering-helm-operator image tag"
    shift
***REMOVED***

export METERING_OPERATOR_IMAGE="${METERING_OPERATOR_IMAGE:-"quay.io/coreos/metering-helm-operator"}"
export METERING_OPERATOR_IMAGE_TAG="${METERING_OPERATOR_IMAGE_TAG:-$IMAGE_TAG}"

if [ -z "$METERING_OPERATOR_IMAGE_TAG" ]; then
    echo "Must pass IMAGE_TAG as ***REMOVED***rst argument, or set \$METERING_OPERATOR_IMAGE_TAG"
    exit 1
***REMOVED***

TMPDIR="$(mktemp -d)"
trap "rm -rf $TMPDIR" EXIT SIGINT

export CUSTOM_HELM_OPERATOR_OVERRIDE_VALUES="$TMPDIR/override-helm-operator-values.yaml"
"$ROOT_DIR/hack/render-helm-operator-override-values.sh" > "$CUSTOM_HELM_OPERATOR_OVERRIDE_VALUES"

export CUSTOM_ALM_OVERRIDE_VALUES="$TMPDIR/override-alm-values.yaml"
"$ROOT_DIR/hack/render-alm-override-values.sh" > "$CUSTOM_ALM_OVERRIDE_VALUES"

export MANIFEST_OUTPUT_DIR="$TMPDIR"
"$ROOT_DIR/hack/create-metering-manifests.sh"

export DEPLOY_MANIFESTS_DIR="$TMPDIR"
"$@"
