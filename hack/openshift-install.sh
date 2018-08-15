#!/bin/bash -e

ROOT_DIR=$(dirname "${BASH_SOURCE}")/..
source "${ROOT_DIR}/hack/common.sh"

if command -v oc; then
    oc new-project "${METERING_NAMESPACE}" || oc project "${METERING_NAMESPACE}"
***REMOVED***

if [ "$METERING_INSTALL_NAMESPACE_VIEWER_CLUSTERROLE" == "true" ]; then
    kubectl \
        create clusterrole \
        "${METERING_NAMESPACE_VIEWER_ROLE_NAME}" \
        --verb=get \
        --resource=namespaces || true

    kubectl \
        create clusterrolebinding \
        "${METERING_NAMESPACE_VIEWER_ROLEBINDING_NAME}" \
        --clusterrole \
        "${METERING_NAMESPACE_VIEWER_ROLE_NAME}" \
        --serviceaccount \
        "${METERING_NAMESPACE}:reporting-operator" || true
***REMOVED***

export DEPLOY_PLATFORM=openshift
"${ROOT_DIR}/hack/install.sh"
