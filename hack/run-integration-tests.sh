#!/bin/bash
set -e

export METERING_NAMESPACE=${METERING_NAMESPACE:-metering-ci-integration}
export METERING_SHORT_TESTS=${METERING_SHORT_TESTS:-false}
export METERING_HTTPS_API=${METERING_HTTPS_API:-true}
export METERING_USE_KUBE_PROXY_FOR_REPORTING_API=${METERING_USE_KUBE_PROXY_FOR_REPORTING_API:-true}
export METERING_REPORTING_API_URL=${METERING_REPORTING_API_URL:-""}

ROOT_DIR=$(dirname "${BASH_SOURCE}")/..
source "${ROOT_DIR}/hack/common.sh"
source "${ROOT_DIR}/hack/lib/tests.sh"

set -x
go test \
    -test.short="${METERING_SHORT_TESTS}" \
    -test.v \
    -parallel 10 \
    -timeout 15m \
    "./test/integration" \
    -namespace "${METERING_NAMESPACE}" \
    -kubeconfig "${KUBECONFIG}" \
    -https-api="${METERING_HTTPS_API}" \
    -use-kube-proxy-for-reporting-api="${METERING_USE_KUBE_PROXY_FOR_REPORTING_API}" \
    -reporting-api-url="${METERING_REPORTING_API_URL}" \
    "$@"
