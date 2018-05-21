#!/bin/bash -e

ROOT_DIR=$(dirname "${BASH_SOURCE}")/..
source "${ROOT_DIR}/hack/common.sh"

: "${METERING_INSTALL_NAMESPACE_VIEWER_CLUSTERROLE:=true}"

if command -v oc; then
    oc new-project "${METERING_NAMESPACE}" || oc project "${METERING_NAMESPACE}"
***REMOVED***

if [ "$METERING_INSTALL_NAMESPACE_VIEWER_CLUSTERROLE" == "true" ]; then
    kubectl -n "${METERING_NAMESPACE}" \
        create clusterrole \
        metering-namespace-viewer \
        --verb=get \
        --resource=namespaces || true

    kubectl -n "${METERING_NAMESPACE}" \
        create clusterrolebinding \
        metering-namespace-viewer \
        --clusterrole \
        metering-namespace-viewer \
        --serviceaccount \
        "${METERING_NAMESPACE}:metering" || true
***REMOVED***

export DEPLOY_PLATFORM=openshift
"${ROOT_DIR}/hack/install.sh"
