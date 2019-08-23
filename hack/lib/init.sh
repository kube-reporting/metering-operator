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
: "${DELETE_PVCS:=true}"

: "${OCP_DEPLOY_PLATFORM:=openshift}"
: "${UPSTREAM_DEPLOY_PLATFORM:=upstream}"
METERING_NAMESPACE=$(sanetize_namespace "${METERING_NAMESPACE:-metering}")

: "${DEPLOY_PLATFORM:=$OCP_DEPLOY_PLATFORM}"
: "${DEPLOY_MANIFESTS_DIR:=$MANIFESTS_DIR/deploy}"
: "${RBAC_MANIFESTS_DIR:=$MANIFESTS_DIR/rbac}"

: "${OCP_INSTALLER_MANIFESTS_DIR:=$DEPLOY_MANIFESTS_DIR/$OCP_DEPLOY_PLATFORM/metering-ansible-operator}"
: "${UPSTREAM_INSTALLER_MANIFESTS_DIR:=$DEPLOY_MANIFESTS_DIR/$UPSTREAM_DEPLOY_PLATFORM/metering-ansible-operator}"
: "${INSTALLER_MANIFESTS_DIR=$OCP_INSTALLER_MANIFESTS_DIR}"

: "${OCP_OLM_MANIFESTS_DIR:=$DEPLOY_MANIFESTS_DIR/$OCP_DEPLOY_PLATFORM/olm}"
: "${UPSTREAM_OLM_MANIFESTS_DIR:=$DEPLOY_MANIFESTS_DIR/$UPSTREAM_DEPLOY_PLATFORM/olm}"
: "${OLM_MANIFESTS_DIR=$OCP_OLM_MANIFESTS_DIR}"

: "${METERING_CR_FILE:=$INSTALLER_MANIFESTS_DIR/meteringcon***REMOVED***g.yaml}"
: "${METERING_UNINSTALL_CLUSTERROLEBINDING:=true}"
: "${METERING_INSTALL_CLUSTERROLEBINDING:=true}"

: "${FAQ_BIN:=faq}"
: "${DEPLOY_REPORTING_OPERATOR_LOCAL:=false}"
: "${DEPLOY_METERING_OPERATOR_LOCAL:=false}"
: "${REPORTING_OPERATOR_PID_FILE:="/tmp/${METERING_NAMESPACE}-reporting-operator.pid"}"
: "${METERING_OPERATOR_PID_FILE:="/tmp/${METERING_NAMESPACE}-metering-operator.pid"}"
: "${REPORTING_OPERATOR_LOG_FILE:="/tmp/${METERING_NAMESPACE}-reporting-operator.log"}"
: "${METERING_OPERATOR_LOG_FILE:="/tmp/${METERING_NAMESPACE}-metering-operator.log"}"
