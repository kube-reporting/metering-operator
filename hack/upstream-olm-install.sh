#! /bin/bash

ROOT_DIR=$(dirname "${BASH_SOURCE[0]}")/..
# shellcheck disable=SC1090
source "${ROOT_DIR}/hack/common.sh"

export INSTALLER_MANIFESTS_DIR="$UPSTREAM_INSTALLER_MANIFESTS_DIR"
export OLM_MANIFESTS_DIR="$UPSTREAM_OLM_MANIFESTS_DIR"

"${ROOT_DIR}/hack/olm-install.sh"
