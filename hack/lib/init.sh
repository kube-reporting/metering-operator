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

: "${CREATE_NAMESPACE:=true}"
: "${SKIP_DELETE_CRDS:=true}"
: "${SKIP_METERING_OPERATOR_DEPLOYMENT:=false}"
: "${DELETE_PVCS:=false}"

: "${DEPLOY_PLATFORM:=openshift}"
METERING_NAMESPACE=$(sanetize_namespace "${METERING_NAMESPACE:-metering}")

: "${DEPLOY_MANIFESTS_DIR:=$MANIFESTS_DIR/deploy}"
: "${RBAC_MANIFESTS_DIR:=$MANIFESTS_DIR/rbac}"
: "${INSTALLER_MANIFESTS_DIR:=$DEPLOY_MANIFESTS_DIR/$DEPLOY_PLATFORM/metering-operator}"
: "${OLM_MANIFESTS_DIR:=$DEPLOY_MANIFESTS_DIR/$DEPLOY_PLATFORM/olm}"
: "${METERING_CR_FILE:=$INSTALLER_MANIFESTS_DIR/metering.yaml}"
: "${CRD_DIR:=$MANIFESTS_DIR/custom-resource-definitions}"
: "${METERING_UNINSTALL_CLUSTERROLEBINDING:=true}"
: "${METERING_INSTALL_CLUSTERROLEBINDING:=true}"
: "${METERING_OPERATOR_IMAGE_REPO:="quay.io/openshift/origin-metering-helm-operator"}"
: "${METERING_OPERATOR_IMAGE_TAG:="latest"}"
: "${REPORTING_OPERATOR_IMAGE_REPO:="quay.io/openshift/origin-metering-reporting-operator"}"
: "${REPORTING_OPERATOR_IMAGE_TAG:="latest"}"
: "${FAQ_BIN:=faq}"
: "${DEPLOY_REPORTING_OPERATOR_LOCAL:=false}"
: "${DEPLOY_METERING_OPERATOR_LOCAL:=false}"
: "${REPORTING_OPERATOR_PID_FILE:="/tmp/${METERING_NAMESPACE}-reporting-operator.pid"}"
: "${METERING_OPERATOR_PID_FILE:="/tmp/${METERING_NAMESPACE}-metering-operator.pid"}"
: "${REPORTING_OPERATOR_LOG_FILE:="/tmp/${METERING_NAMESPACE}-reporting-operator.log"}"
: "${METERING_OPERATOR_LOG_FILE:="/tmp/${METERING_NAMESPACE}-metering-operator.log"}"
