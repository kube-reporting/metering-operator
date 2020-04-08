#!/bin/bash

ROOT_DIR=$(dirname "${BASH_SOURCE}")/..
source "${ROOT_DIR}/hack/common.sh"

: "${REPORTING_OPERATOR_BIN_OUT:=$ROOT_DIR/bin/reporting-operator-local}"
: "${METERING_NAMESPACE:?}"

: "${METERING_USE_SERVICE_ACCOUNT_AS_PROM_TOKEN:=true}"

: "${METERING_PROMETHEUS_NAMESPACE:=openshift-monitoring}"
: "${METERING_PROMETHEUS_SVC:=thanos-querier}"
: "${METERING_PROMETHEUS_SVC_PORT:=9091}"
: "${METERING_PROMETHEUS_SCHEME:=https}"
: "${METERING_PROMETHEUS_PORT_FORWARD:=true}"

: "${METERING_PRESTO_PORT_FORWARD_PORT:=9991}"
: "${METERING_HIVE_PORT_FORWARD_PORT:=9992}"
: "${METERING_PROMETHEUS_PORT_FORWARD_PORT:=9993}"

: "${METERING_PRESTO_HOST:="127.0.0.1:${METERING_PRESTO_PORT_FORWARD_PORT}"}"
: "${METERING_HIVE_HOST:="127.0.0.1:${METERING_HIVE_PORT_FORWARD_PORT}"}"
: "${METERING_PROMETHEUS_HOST:="127.0.0.1:${METERING_PROMETHEUS_PORT_FORWARD_PORT}"}"

: "${METERING_PRESTO_USE_TLS:=true}"
: "${METERING_HIVE_USE_TLS:=true}"

TMPDIR="$(mktemp -d)"

function cleanup() {
    set +e +o pipefail
    exit_status=$?

    echo "Performing cleanup"

    echo "Stopping background jobs"
    # kill any background jobs
    local pids=$(jobs -pr)
    [ -n "$pids" ] && kill $pids
    # Wait for any jobs
    wait 2>/dev/null

    # delete tempfiles
    rm -rf "$TMPDIR"

    echo "Exiting $0"
    exit "$exit_status"
}

set -e -o pipefail
trap cleanup EXIT

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
else
    echo Skipping Prometheus port-forward
fi

if [ "$METERING_PRESTO_USE_TLS" == "true" ]; then
    maxTries=50
    tries=0
    echo "Getting reporting-operator presto server TLS secrets"
    until kubectl -n "$METERING_NAMESPACE" get secrets reporting-operator-presto-server-tls -o json > "$TMPDIR/reporting-operator-presto-server-tls.json"; do
        if [ "$tries" -gt "$maxTries" ]; then
            echo "Timed out waiting for secret reporting-operator-presto-server-tls"
            exit 1
        fi
        tries+=1
        echo 'Waiting for secret reporting-operator-presto-server-tls'
        sleep 5
    done
    echo "Getting reporting-operator presto client TLS secrets"
    until kubectl -n "$METERING_NAMESPACE" get secrets reporting-operator-presto-client-tls -o json > "$TMPDIR/reporting-operator-presto-client-tls.json"; do
        if [ "$tries" -gt "$maxTries" ]; then
            echo "Timed out waiting for secret reporting-operator-presto-client-tls"
            exit 1
        fi
        tries+=1
        echo 'Waiting for secret reporting-operator-presto-client-tls'
        sleep 5
    done

    export REPORTING_OPERATOR_PRESTO_USE_TLS=true
    export REPORTING_OPERATOR_PRESTO_USE_AUTH=true
    export REPORTING_OPERATOR_PRESTO_TLS_INSECURE_SKIP_VERIFY=true

    export REPORTING_OPERATOR_PRESTO_CA_FILE="$TMPDIR/reporting-operator-presto-server-ca.crt"
    export REPORTING_OPERATOR_PRESTO_CLIENT_CERT_FILE="$TMPDIR/reporting-operator-presto-client-tls.crt"
    export REPORTING_OPERATOR_PRESTO_CLIENT_KEY_FILE="$TMPDIR/reporting-operator-presto-client-tls.key"

    jq -Mcr '.data["ca.crt"] | @base64d' "$TMPDIR/reporting-operator-presto-server-tls.json" > "$REPORTING_OPERATOR_PRESTO_CA_FILE"
    jq -Mcr '.data["tls.crt"] | @base64d' "$TMPDIR/reporting-operator-presto-client-tls.json" > "$REPORTING_OPERATOR_PRESTO_CLIENT_CERT_FILE"
    jq -Mcr '.data["tls.key"] | @base64d' "$TMPDIR/reporting-operator-presto-client-tls.json" > "$REPORTING_OPERATOR_PRESTO_CLIENT_KEY_FILE"
fi

if [ "$METERING_HIVE_USE_TLS" == "true" ]; then
    maxTries=50
    tries=0
    echo "Getting reporting-operator Hive server TLS secrets"
    until kubectl -n "$METERING_NAMESPACE" get secrets reporting-operator-hive-server-tls -o json > "$TMPDIR/reporting-operator-hive-server-tls.json"; do
        if [ "$tries" -gt "$maxTries" ]; then
            echo "Timed out waiting for secret reporting-operator-hive-server-tls"
            exit 1
        fi
        tries+=1
        echo 'Waiting for secret reporting-operator-hive-server-tls'
        sleep 5
    done
    echo "Getting reporting-operator Hive client TLS secrets"
    until kubectl -n "$METERING_NAMESPACE" get secrets reporting-operator-hive-client-tls -o json > "$TMPDIR/reporting-operator-hive-client-tls.json"; do
        if [ "$tries" -gt "$maxTries" ]; then
            echo "Timed out waiting for secret reporting-operator-hive-client-tls"
            exit 1
        fi
        tries+=1
        echo 'Waiting for secret reporting-operator-hive-client-tls'
        sleep 5
    done

    export REPORTING_OPERATOR_HIVE_USE_TLS=true
    export REPORTING_OPERATOR_HIVE_USE_AUTH=true
    export REPORTING_OPERATOR_HIVE_TLS_INSECURE_SKIP_VERIFY=true

    export REPORTING_OPERATOR_HIVE_CA_FILE="$TMPDIR/reporting-operator-hive-server-ca.crt"
    export REPORTING_OPERATOR_HIVE_CLIENT_CERT_FILE="$TMPDIR/reporting-operator-hive-client-tls.crt"
    export REPORTING_OPERATOR_HIVE_CLIENT_KEY_FILE="$TMPDIR/reporting-operator-hive-client-tls.key"

    jq -Mcr '.data["ca.crt"] | @base64d' "$TMPDIR/reporting-operator-hive-server-tls.json" > "$REPORTING_OPERATOR_HIVE_CA_FILE"
    jq -Mcr '.data["tls.crt"] | @base64d' "$TMPDIR/reporting-operator-hive-client-tls.json" > "$REPORTING_OPERATOR_HIVE_CLIENT_CERT_FILE"
    jq -Mcr '.data["tls.key"] | @base64d' "$TMPDIR/reporting-operator-hive-client-tls.json" > "$REPORTING_OPERATOR_HIVE_CLIENT_KEY_FILE"
fi

sleep 6

ARGS=()

if [ "$METERING_PROMETHEUS_SCHEME" == "https" ]; then
    ARGS+=(--prometheus-skip-tls-verify)
fi

if [ "$METERING_USE_SERVICE_ACCOUNT_AS_PROM_TOKEN" == "true" ]; then
    REPORTING_OPERATOR_PROMETHEUS_BEARER_TOKEN="$(oc serviceaccounts -n "$METERING_NAMESPACE" get-token reporting-operator)"
    export REPORTING_OPERATOR_PROMETHEUS_BEARER_TOKEN
fi

ARGS+=( "$@" )

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
