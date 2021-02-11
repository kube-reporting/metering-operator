#!/bin/bash -e

ROOT_DIR=$(dirname "${BASH_SOURCE[0]}")/..
# shellcheck disable=SC1090
source "${ROOT_DIR}/hack/common.sh"

kubectl create namespace "${METERING_NAMESPACE}" || true

if [ "$METERING_NAMESPACE" != "openshift-metering" ]; then
    TMPDIR="$(mktemp -d)"
    # shellcheck disable=SC2064
    trap "rm -rf $TMPDIR" EXIT SIGINT

    "$FAQ_BIN" -f yaml -o yaml -M -c -r \
        --kwargs "namespace=$METERING_NAMESPACE" \
        '.metadata.namespace=$namespace' \
        "$OLM_MANIFESTS_DIR/metering.catalogsource.yaml" \
        > "$TMPDIR/metering.catalogsource.yaml"

    "$FAQ_BIN" -f yaml -o yaml -M -c -r \
        --kwargs "namespace=$METERING_NAMESPACE" \
        '.spec.targetNamespaces[0]=$namespace | .metadata.namespace=$namespace | .metadata.name=$namespace + "-" + .metadata.name' \
        "$OLM_MANIFESTS_DIR/metering.operatorgroup.yaml" \
        > "$TMPDIR/metering.operatorgroup.yaml"

    "$FAQ_BIN" -f yaml -o yaml -M -c -r \
        --kwargs "namespace=$METERING_NAMESPACE" \
        '.spec.sourceNamespace=$namespace | .metadata.namespace=$namespace' \
        "$OLM_MANIFESTS_DIR/metering.subscription.yaml" \
        > "$TMPDIR/metering.subscription.yaml"

    export OLM_MANIFESTS_DIR="$TMPDIR"
    export NAMESPACE=$METERING_NAMESPACE
fi

msg "Creating the Metering ConfigMap"
"$ROOT_DIR/hack/create-upgrade-configmap.sh"

msg "Installing Metering Catalog Source"
kubectl apply -f \
    "$OLM_MANIFESTS_DIR/metering.catalogsource.yaml"

msg "Installing Metering Operator Group"
kubectl apply -f \
    "$OLM_MANIFESTS_DIR/metering.operatorgroup.yaml"

msg "Installing Metering Subscription"
kubectl apply -f \
    "$OLM_MANIFESTS_DIR/metering.subscription.yaml"

msg "Installing Metering Resource"
kubectl apply -f \
    "$METERING_CR_FILE"
