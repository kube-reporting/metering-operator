#!/bin/bash
set -e

export METERING_NAMESPACE=${METERING_NAMESPACE:-metering-ci-integration}
export METERING_SHORT_TESTS=${METERING_SHORT_TESTS:-false}
export METERING_HTTPS_API=${METERING_HTTPS_API:-true}

ROOT_DIR=$(dirname "${BASH_SOURCE}")/..
source "${ROOT_DIR}/hack/common.sh"

set -x
go test \
    -test.short="${METERING_SHORT_TESTS}" \
    -test.v \
    -timeout 15m \
    "./test/integration" \
    -namespace "${METERING_NAMESPACE}" \
    -kubecon***REMOVED***g "${KUBECONFIG}" \
    -https-api="${METERING_HTTPS_API}" \
    "$@"
