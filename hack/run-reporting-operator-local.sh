#!/bin/bash

ROOT_DIR=$(dirname "${BASH_SOURCE}")/..
source "${ROOT_DIR}/hack/common.sh"

: "${REPORTING_OPERATOR_BIN_OUT:=$ROOT_DIR/bin/reporting-operator-local}"
: "${METERING_NAMESPACE:?}"

: "${METERING_PROMETHEUS_NAMESPACE:=openshift-monitoring}"
: "${METERING_PROMETHEUS_SVC:=prometheus-k8s}"
: "${METERING_PROMETHEUS_SVC_PORT:=9091}"
: "${METERING_PROMETHEUS_SCHEME:=https}"
: "${METERING_PROMETHEUS_PORT_FORWARD:=true}"

: "${METERING_PRESTO_PORT_FORWARD_PORT:=9991}"
: "${METERING_HIVE_PORT_FORWARD_PORT:=9992}"
: "${METERING_PROMETHEUS_PORT_FORWARD_PORT:=9993}"

: "${METERING_PRESTO_HOST:="127.0.0.1:${METERING_PRESTO_PORT_FORWARD_PORT}"}"
: "${METERING_HIVE_HOST:="127.0.0.1:${METERING_HIVE_PORT_FORWARD_PORT}"}"
: "${METERING_PROMETHEUS_HOST:="127.0.0.1:${METERING_PROMETHEUS_PORT_FORWARD_PORT}"}"

set -e -o pipefail
trap 'jobs -p | xargs kill' EXIT

echo Starting presto port-forward
kubectl -n "$METERING_NAMESPACE" \
    port-forward "svc/presto" ${METERING_PRESTO_PORT_FORWARD_PORT}:8080 &

echo Starting hive port-forward
kubectl -n "$METERING_NAMESPACE" \
    port-forward "svc/hive-server" ${METERING_HIVE_PORT_FORWARD_PORT}:10000 &

if [ "$METERING_PROMETHEUS_PORT_FORWARD" == "true" ]; then
    echo Starting Prometheus port-forward
    kubectl -n "$METERING_PROMETHEUS_NAMESPACE" \
        port-forward "svc/${METERING_PROMETHEUS_SVC}" \
        "${METERING_PROMETHEUS_PORT_FORWARD_PORT}":"${METERING_PROMETHEUS_SVC_PORT}" &
***REMOVED***
    echo Skipping Prometheus port-forward
***REMOVED***

sleep 6

ARGS=("$@")

if [ "$METERING_PROMETHEUS_SCHEME" == "https" ]; then
    ARGS+=(--prometheus-skip-tls-verify)
***REMOVED***

echo Starting reporting-operator
set -x

"$REPORTING_OPERATOR_BIN_OUT" \
    start \
    --namespace "$METERING_NAMESPACE" \
    --presto-host "$METERING_PRESTO_HOST" \
    --hive-host "$METERING_HIVE_HOST" \
    --prometheus-host "${METERING_PROMETHEUS_SCHEME}://${METERING_PROMETHEUS_HOST}" \
    "${ARGS[@]}" &

wait
