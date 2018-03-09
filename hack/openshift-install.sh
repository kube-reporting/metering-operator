#!/bin/bash -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source "${DIR}/default-env.sh"
source "${DIR}/util.sh"

: ${CHARGEBACK_PULL_SECRET_PATH:=""}

export CHARGEBACK_CR_FILE="${CHARGEBACK_CR_FILE:-"$DIR/../manifests/chargeback-config/openshift.yaml"}"
export SKIP_COPY_PULL_SECRET=${SKIP_COPY_PULL_SECRET:=true}

if command -v oc; then
    oc new-project "${CHARGEBACK_NAMESPACE}" || oc project "${CHARGEBACK_NAMESPACE}"
fi

"${DIR}/install.sh"
