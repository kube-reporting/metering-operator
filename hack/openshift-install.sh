#!/bin/bash -e

ROOT_DIR=$(dirname "${BASH_SOURCE}")/..
source "${ROOT_DIR}/hack/common.sh"

if command -v oc; then
    oc new-project "${METERING_NAMESPACE}" || oc project "${METERING_NAMESPACE}"
fi

echo "Labeling namespace ${METERING_NAMESPACE} with 'openshift.io/cluster-monitoring=true'"
kubectl label \
    --overwrite \
    namespace "${METERING_NAMESPACE}" \
    "openshift.io/cluster-monitoring=true"

export DEPLOY_PLATFORM=openshift
"${ROOT_DIR}/hack/install.sh" "$@"
