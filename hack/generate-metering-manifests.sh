#!/bin/bash

set -e
ROOT_DIR=$(dirname "${BASH_SOURCE}")/..
source "${ROOT_DIR}/hack/common.sh"

TMPDIR="$(mktemp -d)"
trap "rm -rf $TMPDIR" EXIT

"$ROOT_DIR/hack/create-metering-manifests.sh" \
    "$INSTALLER_MANIFESTS_DIR" \
    "$OLM_MANIFESTS_DIR"
