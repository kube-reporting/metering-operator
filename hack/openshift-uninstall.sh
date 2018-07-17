#!/bin/bash -e

ROOT_DIR=$(dirname "${BASH_SOURCE}")/..
source "${ROOT_DIR}/hack/common.sh"

export DEPLOY_PLATFORM=openshift
"${ROOT_DIR}/hack/uninstall.sh"

if [ "$METERING_UNINSTALL_NAMESPACE_VIEWER_CLUSTERROLE" == "true" ]; then
    kubectl -n "${METERING_NAMESPACE}" \
        delete clusterrolebinding \
        "${METERING_NAMESPACE_VIEWER_ROLEBINDING_NAME}"

    kubectl -n "${METERING_NAMESPACE}" \
        delete clusterrole \
        "${METERING_NAMESPACE_VIEWER_ROLE_NAME}"
***REMOVED***
