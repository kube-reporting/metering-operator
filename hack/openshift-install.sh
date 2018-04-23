#!/bin/bash -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source "${DIR}/default-env.sh"
source "${DIR}/util.sh"

if command -v oc; then
    oc new-project "${CHARGEBACK_NAMESPACE}" || oc project "${CHARGEBACK_NAMESPACE}"
fi

"${DIR}/install.sh"
