#!/bin/bash
set -e


DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source ${DIR}/util.sh

export CHARGEBACK_NAMESPACE=${CHARGEBACK_NAMESPACE:-chargeback-ci}
export CHARGEBACK_SHORT_TESTS=${CHARGEBACK_SHORT_TESTS:-false}

# lowercase the value, since namespaces must be lowercase values
CHARGEBACK_NAMESPACE=$(echo -n "$CHARGEBACK_NAMESPACE" | tr '[:upper:]' '[:lower:]')

set -x
go test \
    -test.short=${CHARGEBACK_SHORT_TESTS} \
    -v \
    -timeout 20m \
    "./test/integration" \
    -namespace "${CHARGEBACK_NAMESPACE}" \
    -kubecon***REMOVED***g "${KUBECONFIG}"

