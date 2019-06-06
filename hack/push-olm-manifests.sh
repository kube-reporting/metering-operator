#!/bin/bash
set -e

OLM_PACKAGE_ORG="${1:?}"
OLM_PACKAGE_NAME="${2:?}"
OLM_PACKAGE_VERSION="${3:?}"
: "${QUAY_AUTH_TOKEN:?}"

echo "operator-courier push manifests/deploy/openshift/olm/bundle ${OLM_PACKAGE_ORG} ${OLM_PACKAGE_NAME} ${OLM_PACKAGE_VERSION} <ommited>"
operator-courier push manifests/deploy/openshift/olm/bundle "${OLM_PACKAGE_ORG}" "${OLM_PACKAGE_NAME}" "${OLM_PACKAGE_VERSION}" "${QUAY_AUTH_TOKEN}"
