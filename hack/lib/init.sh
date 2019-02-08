#!/bin/bash
# shellcheck disable=SC2034
set -o errexit
set -o nounset
set -o pipefail

# Unset CDPATH so that path interpolation can work correctly
# https://github.com/kubernetes/kubernetes/issues/52255
unset CDPATH

# The root of the build/dist directory
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd -P)"
MANIFESTS_DIR="$ROOT_DIR/manifests"

source "${ROOT_DIR}/hack/lib/util.sh"
source "${ROOT_DIR}/hack/lib/version.sh"

load_version_vars

: "${CREATE_NAMESPACE:=true}"
: "${SKIP_DELETE_CRDS:=true}"
: "${SKIP_METERING_OPERATOR_DEPLOYMENT:=false}"
: "${DELETE_PVCS:=false}"

: "${DEPLOY_PLATFORM:=openshift}"
METERING_NAMESPACE=$(sanetize_namespace "${METERING_NAMESPACE:-metering}")

: "${DEPLOY_MANIFESTS_DIR:=$MANIFESTS_DIR/deploy}"
: "${RBAC_MANIFESTS_DIR:=$MANIFESTS_DIR/rbac}"
: "${INSTALLER_MANIFESTS_DIR:=$DEPLOY_MANIFESTS_DIR/$DEPLOY_PLATFORM/helm-operator}"
: "${OLM_MANIFESTS_DIR:=$MANIFESTS_DIR/deploy/$DEPLOY_PLATFORM/olm}"
: "${METERING_CR_FILE:=$INSTALLER_MANIFESTS_DIR/metering.yaml}"
: "${CRD_DIR:=$MANIFESTS_DIR/custom-resource-definitions}"
: "${METERING_UNINSTALL_REPORTING_OPERATOR_CLUSTERROLEBINDING:=false}"
: "${METERING_INSTALL_REPORTING_OPERATOR_CLUSTERROLEBINDING:=true}"
: "${USE_CUSTOM_METERING_OPERATOR:=false}"
: "${CUSTOM_METERING_OPERATOR_IMAGE:="quay.io/coreos/metering-helm-operator"}"
: "${CUSTOM_METERING_OPERATOR_IMAGE_TAG:="$METERING_VERSION"}"
: "${FAQ_BIN:=faq}"
