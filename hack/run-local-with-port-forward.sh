#!/bin/bash

ROOT_DIR=$(dirname "${BASH_SOURCE}")/..
source "${ROOT_DIR}/hack/common.sh"

: "${REPORTING_OP_BIN:=$ROOT_DIR/bin/reporting-operator-local}"
: "${METERING_NAMESPACE:?}"

: "${METERING_PROMETHEUS_NAMESPACE:=openshift-monitoring}"
: "${METERING_PROMETHEUS_SVC:=prometheus-k8s}"
: "${METERING_PROMETHEUS_SVC_PORT:=9091}"
: "${METERING_PROMETHEUS_SCHEME:=https}"

set -e -o pipefail
trap 'jobs -p | xargs kill' EXIT


echo Starting presto port-forward
kubectl -n "$METERING_NAMESPACE" \
    port-forward "svc/presto" 9991:8080 &

echo Starting hive port-forward
kubectl -n "$METERING_NAMESPACE" \
    port-forward "svc/hive-server" 9992:10000 &

echo Starting Prometheus port-forward
kubectl -n "$METERING_PROMETHEUS_NAMESPACE" \
    port-forward "svc/${METERING_PROMETHEUS_SVC}" \
    9993:"${METERING_PROMETHEUS_SVC_PORT}" &

sleep 6

ARGS=("$@")

if [ "$METERING_PROMETHEUS_SCHEME" == "https" ]; then
    ARGS+=(--prometheus-skip-tls-verify)
fi

echo Starting reporting-operator
set -x

"$REPORTING_OP_BIN" \
    start \
    --namespace "$METERING_NAMESPACE" \
    --presto-host "127.0.0.1:9991" \
    --hive-host "127.0.0.1:9992" \
    --prometheus-host "${METERING_PROMETHEUS_SCHEME}://127.0.0.1:9993" \
    "${ARGS[@]}" &

wait
