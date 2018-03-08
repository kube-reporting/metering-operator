#!/bin/bash -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source "${DIR}/default-env.sh"
source "${DIR}/util.sh"

: ${CHARGEBACK_PULL_SECRET_PATH:=""}

if [ ! -s "$CHARGEBACK_PULL_SECRET_PATH" ]; then
    echo "\$CHARGEBACK_PULL_SECRET_PATH must be set to a dockercon***REMOVED***gjson ***REMOVED***le"
    exit 1
***REMOVED***

export CHARGEBACK_CR_FILE="${CHARGEBACK_CR_FILE:-"$DIR/../manifests/chargeback-con***REMOVED***g/openshift.yaml"}"
export SKIP_COPY_PULL_SECRET=true

oc new-project "${CHARGEBACK_NAMESPACE}" || oc project "${CHARGEBACK_NAMESPACE}"
oc create secret generic coreos-pull-secret --from-***REMOVED***le=.dockercon***REMOVED***gjson="${CHARGEBACK_PULL_SECRET_PATH}" --type='kubernetes.io/dockercon***REMOVED***gjson'

"${DIR}/install.sh"
