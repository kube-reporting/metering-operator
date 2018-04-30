#!/bin/bash -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source "${DIR}/default-env.sh"
source "${DIR}/util.sh"

if command -v oc; then
    oc new-project "${METERING_NAMESPACE}" || oc project "${METERING_NAMESPACE}"
fi

export INSTALLER_MANIFEST_DIR="${INSTALLER_MANIFEST_DIR:-$MANIFESTS_DIR/deploy/openshift/helm-operator}"
export METERING_CR_FILE="${METERING_CR_FILE:-$INSTALLER_MANIFEST_DIR/metering.yaml}"

"${DIR}/install.sh"
