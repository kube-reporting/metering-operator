#!/bin/bash -e

ROOT_DIR=$(dirname "${BASH_SOURCE}")/..
source "${ROOT_DIR}/hack/common.sh"

kubectl create namespace "${METERING_NAMESPACE}" || true

echo "Labeling namespace ${METERING_NAMESPACE} with 'openshift.io/cluster-monitoring=true'"
kubectl label \
    --overwrite \
    namespace "${METERING_NAMESPACE}" \
    "openshift.io/cluster-monitoring=true"

export INSTALLER_MANIFESTS_DIR="${INSTALLER_MANIFESTS_DIR:-"$OCP_INSTALLER_MANIFESTS_DIR"}"

"${ROOT_DIR}/hack/install.sh" "$@"
