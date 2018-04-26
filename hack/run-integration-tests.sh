#!/bin/bash
set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source ${DIR}/util.sh

export CHARGEBACK_NAMESPACE=${CHARGEBACK_NAMESPACE:-chargeback-ci-integration}
export CHARGEBACK_SHORT_TESTS=${CHARGEBACK_SHORT_TESTS:-false}

# lowercase the value, since namespaces must be lowercase values
CHARGEBACK_NAMESPACE="$(sanetize_namespace "$CHARGEBACK_NAMESPACE")"

go test \
    -test.short="${CHARGEBACK_SHORT_TESTS}" \
    -test.v \
    -timeout 20m \
    "github.com/coreos-inc/kube-chargeback/test/integration" \
    -namespace "${CHARGEBACK_NAMESPACE}" \
    -kubeconfig "${KUBECONFIG}" "$@"


