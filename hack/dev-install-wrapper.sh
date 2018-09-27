#!/bin/bash
set -o errexit
set -o nounset
set -o pipefail

ROOT_DIR=$(dirname "${BASH_SOURCE[0]}")/..

TMP="$(mktemp -d)"
trap "rm -rf $TMP" EXIT SIGINT

export CUSTOM_MANIFEST_OUTPUT_DIR="$TMP"
if [[ -n "${CUSTOM_HELM_OPERATOR_OVERRIDE_VALUES-}" && -n "${CUSTOM_ALM_OVERRIDE_VALUES-}" ]]; then
    "$ROOT_DIR/hack/create-metering-manifests.sh" "$CUSTOM_MANIFEST_OUTPUT_DIR"
else
    if [ $# -lt 2 ]; then
        echo "Usage: $0 [metering-operator-image-tag] [install-script] [args]"
        echo "Must specify at least two args, metering-operator image tag, and the install script to run"
        exit 1
    fi

    HELM_OPERATOR_IMAGE_TAG="$1"
    shift
    "$ROOT_DIR/hack/create-metering-manifests.sh" "$HELM_OPERATOR_IMAGE_TAG" "$CUSTOM_MANIFEST_OUTPUT_DIR"
fi

export DEPLOY_MANIFESTS_DIR="$CUSTOM_MANIFEST_OUTPUT_DIR"
"$@"
