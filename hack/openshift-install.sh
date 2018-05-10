#!/bin/bash -e

ROOT_DIR=$(dirname "${BASH_SOURCE}")/..
source "${ROOT_DIR}/hack/common.sh"

if command -v oc; then
    oc new-project "${METERING_NAMESPACE}" || oc project "${METERING_NAMESPACE}"
***REMOVED***

export DEPLOY_PLATFORM=openshift
"${ROOT_DIR}/install.sh"
