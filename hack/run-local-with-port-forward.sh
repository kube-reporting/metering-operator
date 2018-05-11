#!/bin/bash -e

ROOT_DIR=$(dirname "${BASH_SOURCE}")/..
source "${ROOT_DIR}/hack/common.sh"

: "${CHARGEBACK_BIN:=$ROOT_DIR/chargeback-local}"
: "${METERING_NAMESPACE:?}"

: "${METERING_PROMETHEUS_NAMESPACE:=tectonic-system}"
: "${METERING_PROMTHEUS_LABEL_SELECTOR:=app=prometheus}"

trap 'kill $(jobs -p)' SIGINT SIGTERM EXIT

echo Starting presto port-forward
kubectl get pods -n "$METERING_NAMESPACE" -l app=presto -o name \
    | cut -d/ -f2 \
    | xargs -I{} kubectl -n "$METERING_NAMESPACE" port-forward {} 9991:8080 &

echo Starting hive port-forward
kubectl get pods -n "$METERING_NAMESPACE" -l app=hive -o name \
    | cut -d/ -f2 \
    | xargs -I{} kubectl -n "$METERING_NAMESPACE" port-forward {} 9992:10000 &

echo Starting Prometheus port-forward
kubectl -n "$METERING_PROMETHEUS_NAMESPACE" get pods -l "$METERING_PROMTHEUS_LABEL_SELECTOR" -o name \
    | cut -d/ -f2 \
    | xargs -I{}  kubectl -n "$METERING_PROMETHEUS_NAMESPACE" port-forward {} 9993:9090 &

sleep 6

echo Starting chargeback
set -x
"$CHARGEBACK_BIN" \
    start \
    --namespace "$METERING_NAMESPACE" \
    --presto-host "127.0.0.1:9991" \
    --hive-host "127.0.0.1:9992" \
    --prometheus-host "http://127.0.0.1:9993"
