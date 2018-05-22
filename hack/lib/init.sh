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
: "${DELETE_PVCS:=false}"

: "${DEPLOY_PLATFORM:=generic}"
METERING_NAMESPACE=$(sanetize_namespace "${METERING_NAMESPACE:-metering}")

: "${DEPLOY_MANIFESTS_DIR:=$MANIFESTS_DIR/deploy}"
: "${INSTALLER_MANIFESTS_DIR:=$DEPLOY_MANIFESTS_DIR/$DEPLOY_PLATFORM/helm-operator}"
: "${ALM_MANIFESTS_DIR:=$MANIFESTS_DIR/deploy/$DEPLOY_PLATFORM/alm}"
: "${METERING_CR_FILE:=$INSTALLER_MANIFESTS_DIR/metering.yaml}"
: "${CRD_DIR:=$MANIFESTS_DIR/custom-resource-de***REMOVED***nitions}"

# These are currently openshift speci***REMOVED***c con***REMOVED***g options, controlling if a
# clusterrole and clusterrolebinding are created granting access to GET
# namespaces. This is for granting access to querying the Prometheus API.
: "${METERING_INSTALL_NAMESPACE_VIEWER_CLUSTERROLE:=true}"
: "${METERING_UNINSTALL_NAMESPACE_VIEWER_CLUSTERROLE:=false}"
: "${METERING_NAMESPACE_VIEWER_ROLEBINDING_NAME:=${METERING_NAMESPACE}-metering-namespace-viewer}"
