#!/bin/bash -e

ROOT_DIR=$(dirname "${BASH_SOURCE}")/..
source "${ROOT_DIR}/hack/common.sh"

export DEPLOY_PLATFORM=openshift
"${ROOT_DIR}/hack/uninstall.sh"

if [ "$METERING_UNINSTALL_REPORTING_OPERATOR_EXTRA_CLUSTERROLEBINDING" == "true" ]; then
    kubectl -n "${METERING_NAMESPACE}" \
        delete clusterrolebinding \
        "${METERING_REPORTING_OPERATOR_EXTRA_ROLEBINDING_NAME}"
***REMOVED***
