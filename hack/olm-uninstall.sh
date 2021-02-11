#!/bin/bash -e

ROOT_DIR=$(dirname "${BASH_SOURCE[0]}")/..
# shellcheck disable=SC1090
source "${ROOT_DIR}/hack/common.sh"

set +e

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
fi

SUBSCRIPTION_NAME="$(faq -f yaml '.metadata.name' "$OLM_MANIFESTS_DIR/metering.subscription.yaml")"
CSV_NAME="$(kubectl -n $METERING_NAMESPACE get subscriptions $SUBSCRIPTION_NAME -o yaml | faq -f yaml '.status.currentCSV')"

msg "Removing Metering Resource"
kubectl delete -f \
    "$METERING_CR_FILE"

msg "Removing Metering Subscription"
kubectl delete -f \
    "$OLM_MANIFESTS_DIR/metering.subscription.yaml"

msg "Removing Metering Operator Group"
kubectl delete -f \
    "$OLM_MANIFESTS_DIR/metering.operatorgroup.yaml"

msg "Removing Metering Catalog Source"
kubectl delete -f \
    "$OLM_MANIFESTS_DIR/metering.catalogsource.yaml"

msg "Removing Metering ConfigMap"
kubectl -n $METERING_NAMESPACE delete configmap metering-ocp

msg "Removing Metering Catalog Source Version"
kubectl -n $METERING_NAMESPACE delete csv $CSV_NAME
