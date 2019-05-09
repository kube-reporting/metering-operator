#!/bin/bash

set -e
ROOT_DIR=$(dirname "${BASH_SOURCE}")/..
source "${ROOT_DIR}/hack/common.sh"

TMPDIR="$(mktemp -d)"
trap "rm -rf $TMPDIR" EXIT

export METERING_OPERATOR_IMAGE_REPO="${METERING_OPERATOR_IMAGE_REPO:?}"
export METERING_OPERATOR_IMAGE_TAG="${METERING_OPERATOR_IMAGE_TAG:?}"

export CUSTOM_HELM_OPERATOR_OVERRIDE_VALUES="$TMPDIR/override-helm-operator-values.yaml"
"$ROOT_DIR/hack/render-helm-operator-override-values.sh" > "$CUSTOM_HELM_OPERATOR_OVERRIDE_VALUES"

"$ROOT_DIR/hack/create-metering-manifests.sh" \
    "$INSTALLER_MANIFESTS_DIR" \
    "$OLM_MANIFESTS_DIR" \
    "$CUSTOM_HELM_OPERATOR_OVERRIDE_VALUES"
