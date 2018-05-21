#!/bin/bash -e

ROOT_DIR=$(dirname "${BASH_SOURCE}")/..
source "${ROOT_DIR}/hack/common.sh"

: "${METERING_UNINSTALL_NAMESPACE_VIEWER_CLUSTERROLE:=false}"

export DEPLOY_PLATFORM=openshift
"${ROOT_DIR}/hack/uninstall.sh"


if [ "$METERING_UNINSTALL_NAMESPACE_VIEWER_CLUSTERROLE" == "true" ]; then
    kubectl -n "${METERING_NAMESPACE}" \
        delete clusterrolebinding \
        metering-namespace-viewer

    kubectl -n "${METERING_NAMESPACE}" \
        delete clusterrole \
        metering-namespace-viewer
***REMOVED***
