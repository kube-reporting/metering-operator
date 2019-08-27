#!/bin/bash

ROOT_DIR=$(dirname "${BASH_SOURCE}")/..
source "${ROOT_DIR}/hack/common.sh"

export INSTALLER_MANIFESTS_DIR="$UPSTREAM_INSTALLER_MANIFESTS_DIR"

"${ROOT_DIR}/hack/uninstall.sh"
