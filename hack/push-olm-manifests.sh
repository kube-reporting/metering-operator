#!/bin/bash
set -e
set -o pipefail

ROOT_DIR=$(dirname "${BASH_SOURCE}")/..
source "${ROOT_DIR}/hack/common.sh"

OLM_PACKAGE_ORG="${1:?}"
OLM_PACKAGE_NAME="${2:?}"
MANIFEST_BUNDLE="${3:-$OCP_OLM_MANIFESTS_DIR/bundle}"
OLM_PACKAGE_VERSION="${OLM_PACKAGE_VERSION:-"4.2.0-$(date +'%Y%m%d%H%M%S')"}"
: "${QUAY_AUTH_TOKEN:?}"

echo "operator-courier push $MANIFEST_BUNDLE ${OLM_PACKAGE_ORG} ${OLM_PACKAGE_NAME} ${OLM_PACKAGE_VERSION} <ommited>"
operator-courier push "$MANIFEST_BUNDLE" "${OLM_PACKAGE_ORG}" "${OLM_PACKAGE_NAME}" "${OLM_PACKAGE_VERSION}" "${QUAY_AUTH_TOKEN}"
