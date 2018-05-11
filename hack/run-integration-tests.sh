#!/bin/bash
set -e

export METERING_NAMESPACE=${METERING_NAMESPACE:-metering-ci-integration}
export METERING_SHORT_TESTS=${METERING_SHORT_TESTS:-false}

ROOT_DIR=$(dirname "${BASH_SOURCE}")/..
source "${ROOT_DIR}/hack/common.sh"

go test \
    -test.short="${METERING_SHORT_TESTS}" \
    -test.v \
    -timeout 20m \
    "./test/integration" \
    -namespace "${METERING_NAMESPACE}" \
    -kubecon***REMOVED***g "${KUBECONFIG}" "$@"


