#!/bin/bash
set -e


DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source ${DIR}/util.sh

export CHARGEBACK_NAMESPACE=${CHARGEBACK_NAMESPACE:-chargeback-ci}

# lowercase the value, since namespaces must be lowercase values
CHARGEBACK_NAMESPACE=$(echo -n "$CHARGEBACK_NAMESPACE" | tr '[:upper:]' '[:lower:]')

set -x
go test \
    -v "./test/integration" \
    -timeout 20m \
    -namespace "${CHARGEBACK_NAMESPACE}" \
    -kubecon***REMOVED***g "${KUBECONFIG}"

