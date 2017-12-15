#!/bin/bash
set -e


DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source ${DIR}/util.sh

export CHARGEBACK_NAMESPACE=${CHARGEBACK_NAMESPACE:-chargeback-ci}

set -x
go test \
    -v "./test/integration" \
    -namespace "${CHARGEBACK_NAMESPACE}" \
    -kubeconfig "${KUBECONFIG}"

