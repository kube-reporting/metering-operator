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

: "${CREATE_NAMESPACE:=true}"
: "${SKIP_DELETE_CRDS:=true}"
: "${SKIP_METERING_OPERATOR_DEPLOYMENT:=false}"
: "${DELETE_PVCS:=false}"

: "${DEPLOY_PLATFORM:=generic}"
METERING_NAMESPACE=$(sanetize_namespace "${METERING_NAMESPACE:-metering}")

: "${DEPLOY_MANIFESTS_DIR:=$MANIFESTS_DIR/deploy}"
: "${RBAC_MANIFESTS_DIR:=$MANIFESTS_DIR/rbac}"
: "${INSTALLER_MANIFESTS_DIR:=$DEPLOY_MANIFESTS_DIR/$DEPLOY_PLATFORM/helm-operator}"
: "${ALM_MANIFESTS_DIR:=$MANIFESTS_DIR/deploy/$DEPLOY_PLATFORM/alm}"
: "${METERING_CR_FILE:=$INSTALLER_MANIFESTS_DIR/metering.yaml}"
: "${CRD_DIR:=$MANIFESTS_DIR/custom-resource-de***REMOVED***nitions}"

# These are currently openshift speci***REMOVED***c con***REMOVED***g options, controlling if a
# clusterrole and clusterrolebinding are created granting access to GET
# namespaces , create subject access reviews and token reviews
# This is for granting access to querying the Prometheus API and checking users
# permissions with the auth proxy.
: "${METERING_REPORTING_OPERATOR_EXTRA_ROLE_NAME:=openshift-reporting-operator-extra}"
: "${METERING_REPORTING_OPERATOR_EXTRA_ROLEBINDING_NAME:=${METERING_NAMESPACE}-openshift-reporting-operator-extra}"
: "${METERING_REPORTING_OPERATOR_EXTRA_ROLE_NAME:=openshift-reporting-operator-extra}"
: "${METERING_UNINSTALL_REPORTING_OPERATOR_EXTRA_CLUSTERROLEBINDING:=false}"
: "${METERING_INSTALL_REPORTING_OPERATOR_EXTRA_CLUSTERROLEBINDING:=true}"
