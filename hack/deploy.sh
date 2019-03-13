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

if [ "$METERING_CREATE_PULL_SECRET" == "true" ]; then
    : "${DOCKER_USERNAME:?}"
    : "${DOCKER_PASSWORD:?}"
***REMOVED***

if [ "$UNINSTALL_METERING_BEFORE_INSTALL" == "true" ]; then
    echo "Uninstalling metering"
    kubectl delete ns "$METERING_NAMESPACE" || true
***REMOVED***
    echo "Skipping uninstall"
***REMOVED***

while true; do
    echo "Checking namespace status"
    NS="$(kubectl get ns "$METERING_NAMESPACE" -o json --ignore-not-found)"
    if [ "$NS" == "" ]; then
        echo "Namespace ${METERING_NAMESPACE} does not exist"
        break
    ***REMOVED***
    PHASE="$(echo "$NS" | "$FAQ_BIN" -f json -o json -M -c -r '.status.phase')"
    if [ "$PHASE" == "Active" ]; then
        echo "Namespace is active"
        break
    elif [ "$PHASE" == "Terminating" ]; then
        echo "Waiting for namespace "$METERING_NAMESPACE" termination to complete before continuing"
    ***REMOVED***
        echo "Namespace phase is $PHASE, unsure how to handle, exiting"
        exit 2
    ***REMOVED***
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
***REMOVED***

if [ "$INSTALL_METERING" == "true" ]; then
    echo "Installing metering"
    install_metering "${INSTALL_METHOD}"
***REMOVED***
    echo "Skipping install"
***REMOVED***

echo "Waiting for metering-operator pods to be ready"
until [ "$(kubectl -n "$METERING_NAMESPACE" get pods -l app=metering-operator -o json | "$FAQ_BIN" -f json -o json -M -c -r '.items | map(try .status.containerStatuses[].ready catch false) | all')" == "true" ]; do
    echo 'waiting for metering-operator pods to be ready'
    sleep 5
done
echo "metering helm-operator is ready"

echo
echo "Waiting a for pods to be recreated"

EXPECTED_POD_COUNT=7

# now wait for the pods to reach our expected count
echo "checking for pod statuses"
until [ "$(kubectl -n "$METERING_NAMESPACE" get pods -o json | "$FAQ_BIN" -f json -o json -M -c -r '.items | length')" == "$EXPECTED_POD_COUNT" ]; do
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
