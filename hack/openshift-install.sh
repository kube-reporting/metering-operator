#!/bin/bash -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source "${DIR}/default-env.sh"
source "${DIR}/util.sh"

if command -v oc; then
    oc new-project "${CHARGEBACK_NAMESPACE}" || oc project "${CHARGEBACK_NAMESPACE}"
fi

export INSTALLER_MANIFEST_DIR="${INSTALLER_MANIFEST_DIR:-$MANIFESTS_DIR/deploy/openshift/helm-operator}"
export CHARGEBACK_CR_FILE="${CHARGEBACK_CR_FILE:-$INSTALLER_MANIFEST_DIR/metering.yaml}"

"${DIR}/install.sh"
