#!/bin/bash
set -e

export METERING_NAMESPACE=${METERING_NAMESPACE:-metering-ci-e2e}
export METERING_SHORT_TESTS=${METERING_SHORT_TESTS:-false}
export METERING_HTTPS_API=${METERING_HTTPS_API:-false}

ROOT_DIR=$(dirname "${BASH_SOURCE}")/..
source "${ROOT_DIR}/hack/common.sh"

set -x
go test \
    -test.short="${METERING_SHORT_TESTS}" \
    -test.v \
    -timeout 20m \
    "./test/e2e" \
    -namespace "${METERING_NAMESPACE}" \
    -kubeconfig "${KUBECONFIG}" \
    -https-api="${METERING_HTTPS_API}" \
    "$@"
