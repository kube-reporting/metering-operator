#!/bin/bash
set -e

export METERING_NAMESPACE=${METERING_NAMESPACE:-metering-ci}

ROOT_DIR=$(dirname "${BASH_SOURCE}")/..
source "${ROOT_DIR}/hack/common.sh"

: "${UNINSTALL_METERING_BEFORE_INSTALL:=true}"
: "${INSTALL_METERING:=true}"
: "${INSTALL_METHOD:=${DEPLOY_PLATFORM}-direct}"
: "${METERING_CREATE_PULL_SECRET:=false}"
: "${METERING_PULL_SECRET_NAME:=metering-pull-secret}"
: "${INSTALL_METERING:=true}"
: "${DEPLOY_METERING_OPERATOR_LOCAL:=false}"

if [ "$METERING_CREATE_PULL_SECRET" == "true" ]; then
    : "${DOCKER_USERNAME:?}"
    : "${DOCKER_PASSWORD:?}"
fi

if [ "$UNINSTALL_METERING_BEFORE_INSTALL" == "true" ]; then
    echo "Uninstalling metering"
    kubectl delete ns "$METERING_NAMESPACE" || true
else
    echo "Skipping uninstall"
fi

while true; do
    echo "Checking namespace status"
    NS="$(kubectl get ns "$METERING_NAMESPACE" -o json --ignore-not-found)"
    if [ "$NS" == "" ]; then
        echo "Namespace ${METERING_NAMESPACE} does not exist"
        break
    fi
    PHASE="$(echo "$NS" | "$FAQ_BIN" -f json -o json -M -c -r '.status.phase')"
    if [ "$PHASE" == "Active" ]; then
        echo "Namespace is active"
        break
    elif [ "$PHASE" == "Terminating" ]; then
        echo "Waiting for namespace "$METERING_NAMESPACE" termination to complete before continuing"
    else
        echo "Namespace phase is $PHASE, unsure how to handle, exiting"
        exit 2
    fi
    sleep 2
done

echo "Creating namespace $METERING_NAMESPACE"
kubectl create ns "$METERING_NAMESPACE" || true

if [ "$METERING_CREATE_PULL_SECRET" == "true" ]; then
    echo "\$METERING_CREATE_PULL_SECRET is true, creating pull-secret $METERING_PULL_SECRET_NAME"
    kubectl -n "$METERING_NAMESPACE" \
        create secret docker-registry "$METERING_PULL_SECRET_NAME" \
        --docker-server=quay.io \
        --docker-username="$DOCKER_USERNAME" \
        --docker-password="$DOCKER_PASSWORD" \
        --docker-email=example@example.com || true
fi

(( BASE_DEPLOY_EXPECTED_POD_COUNT=7 ))

if [ "$DEPLOY_METERING_OPERATOR_LOCAL" == "true" ]; then
    echo "Deploying metering-operator-locally"
    rm -f /tmp/metering-operator.log
    nohup "$ROOT_DIR/hack/run-metering-operator-local.sh" > "$METERING_OPERATOR_LOG_FILE" &
    echo $! > "$METERING_OPERATOR_PID_FILE"
    (( BASE_DEPLOY_EXPECTED_POD_COUNT-- ))
elif [ "$SKIP_METERING_OPERATOR_DEPLOYMENT" == "true" ]; then
    echo "Skipped metering-operator deployment"
    (( BASE_DEPLOY_EXPECTED_POD_COUNT-- ))
else
    if [ "$INSTALL_METERING" == "true" ]; then
        echo "Installing metering"
        install_metering "${INSTALL_METHOD}"
    else
        echo "Skipping install"
        exit 0
    fi

    echo "Waiting for metering-operator pods to be ready"
    until [ "$(kubectl -n "$METERING_NAMESPACE" get pods -l app=metering-operator -o json | "$FAQ_BIN" -f json -o json -M -c -r '.items | map(try .status.containerStatuses[].ready catch false) | all')" == "true" ]; do
        echo 'waiting for metering-operator pods to be ready'
        sleep 5
    done
    echo "metering helm-operator is ready"
fi

if [ "$DEPLOY_REPORTING_OPERATOR_LOCAL" == "true" ]; then
    (( BASE_DEPLOY_EXPECTED_POD_COUNT-- ))
fi
DEPLOY_EXPECTED_POD_COUNT="${DEPLOY_EXPECTED_POD_COUNT:-$BASE_DEPLOY_EXPECTED_POD_COUNT}"

# now wait for the pods to reach our expected count
echo "checking for pod statuses"
until [ "$(kubectl -n "$METERING_NAMESPACE" get pods -o json | "$FAQ_BIN" -f json -o json -M -c -r '.items | length')" == "$DEPLOY_EXPECTED_POD_COUNT" ]; do
    echo 'waiting for metering pods to be created'
    kubectl -n "$METERING_NAMESPACE" get pods --no-headers -o wide
    sleep 10
done
echo "all of the metering pods have been started"

until [ "$(kubectl -n "$METERING_NAMESPACE" get pods  -o json | "$FAQ_BIN" -f json -o json -M -c -r '.items | map(try .status.containerStatuses[].ready catch false) | all')" == "true" ]; do
    echo 'waiting for all pods to be ready'
    kubectl -n "$METERING_NAMESPACE" get pods --no-headers -o wide
    sleep 10
done
echo "metering pods are all ready"

if [ "$DEPLOY_REPORTING_OPERATOR_LOCAL" == "true" ]; then
    echo "Getting reporting-operator service account"
    REPORTING_OPERATOR_PROMETHEUS_TOKEN=""
    while [ -z "$REPORTING_OPERATOR_PROMETHEUS_TOKEN" ]; do
        # semi-colon matters
        REPORTING_OPERATOR_PROMETHEUS_TOKEN="$(oc -n "$METERING_NAMESPACE" serviceaccounts get-token reporting-operator)" || true
        echo "Waiting for reporting-operator service account"
        sleep 5
    done

    rm -f /tmp/reporting-operator.log
    echo "Deploying report-operator-locally"
    nohup "$ROOT_DIR/hack/run-reporting-operator-local.sh" \
        --namespace "$METERING_NAMESPACE" \
        --prometheus-bearer-token "$REPORTING_OPERATOR_PROMETHEUS_TOKEN" "${REPORTING_OPERATOR_ARGS:-}" > "$REPORTING_OPERATOR_LOG_FILE" &
    echo $! > "$REPORTING_OPERATOR_PID_FILE"

    until curl -s --fail "http://127.0.0.1:8080/healthy" > /dev/null; do
        echo "waiting for local reporting-operator to become healthy"
        sleep 5
    done
    echo "reporting-operator is healthy"
fi

