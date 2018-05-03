#!/bin/bash
set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source ${DIR}/util.sh

export METERING_NAMESPACE=${METERING_NAMESPACE:-metering-ci-e2e}
export METERING_SHORT_TESTS=${METERING_SHORT_TESTS:-false}

# lowercase the value, since namespaces must be lowercase values
METERING_NAMESPACE="$(sanetize_namespace "$METERING_NAMESPACE")"

go test \
    -test.short="${METERING_SHORT_TESTS}" \
    -test.v \
    -timeout 20m \
    "./test/e2e" \
    -namespace "${METERING_NAMESPACE}" \
    -kubecon***REMOVED***g "${KUBECONFIG}" "$@"


