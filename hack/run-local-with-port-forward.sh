#!/bin/bash

ROOT_DIR=$(dirname "${BASH_SOURCE}")/..
source "${ROOT_DIR}/hack/common.sh"

: "${CHARGEBACK_BIN:=$ROOT_DIR/chargeback-local}"
: "${METERING_NAMESPACE:?}"

: "${METERING_PROMETHEUS_NAMESPACE:=tectonic-system}"
: "${METERING_PROMTHEUS_LABEL_SELECTOR:=app=prometheus}"

set -e -o pipefail
trap 'jobs -p | xargs kill' EXIT

PRESTO="$(kubectl get pods -n "$METERING_NAMESPACE" -l app=presto,presto=coordinator -o name | cut -d/ -f2)"
HIVE="$(kubectl get pods -n "$METERING_NAMESPACE" -l app=hive,hive=server -o name | cut -d/ -f2)"
PROM="$(kubectl get pods -n "$METERING_PROMETHEUS_NAMESPACE" -l "$METERING_PROMTHEUS_LABEL_SELECTOR" -o name | cut -d/ -f2)"

echo Starting presto port-forward
kubectl -n "$METERING_NAMESPACE" port-forward "$PRESTO" 9991:8080 &

echo Starting hive port-forward
kubectl -n "$METERING_NAMESPACE" port-forward "$HIVE" 9992:10000 &

echo Starting Prometheus port-forward
kubectl -n "$METERING_PROMETHEUS_NAMESPACE" port-forward "$PROM" 9993:9090 &

sleep 6

echo Starting chargeback
set -x
"$CHARGEBACK_BIN" \
    start \
    --namespace "$METERING_NAMESPACE" \
    --presto-host "127.0.0.1:9991" \
    --hive-host "127.0.0.1:9992" \
    --prometheus-host "http://127.0.0.1:9993" \
    "$@"
