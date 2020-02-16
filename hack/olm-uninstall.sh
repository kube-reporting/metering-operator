#!/bin/bash -e

ROOT_DIR=$(dirname "${BASH_SOURCE}")/..
source "${ROOT_DIR}/hack/common.sh"

set +e

if [ "$METERING_NAMESPACE" != "openshift-metering" ]; then
    TMPDIR="$(mktemp -d)"
    # shellcheck disable=SC2064
    trap "rm -rf $TMPDIR" EXIT SIGINT

    "$FAQ_BIN" -f yaml -o yaml -M -c -r \
        --kwargs "namespace=$METERING_NAMESPACE" \
        '.spec.targetNamespace=$namespace' \
        "$OLM_MANIFESTS_DIR/metering.catalogsourceconfig.yaml" \
        > "$TMPDIR/metering.catalogsourceconfig.yaml"

    "$FAQ_BIN" -f yaml -o yaml -M -c -r \
        --kwargs "namespace=$METERING_NAMESPACE" \
        '.spec.targetNamespaces[0]=$namespace | .metadata.name=$namespace + "-" + .metadata.name' \
        "$OLM_MANIFESTS_DIR/metering.operatorgroup.yaml" \
        > "$TMPDIR/metering.operatorgroup.yaml"

    "$FAQ_BIN" -f yaml -o yaml -M -c -r \
        --kwargs "namespace=$METERING_NAMESPACE" \
        '.spec.sourceNamespace=$namespace' \
        "$OLM_MANIFESTS_DIR/metering.subscription.yaml" \
        > "$TMPDIR/metering.subscription.yaml"

        export OLM_MANIFESTS_DIR="$TMPDIR"
fi

SUBSCRIPTION_NAME="$(faq -f yaml '.metadata.name' "$OLM_MANIFESTS_DIR/metering.subscription.yaml")"
CSV_NAME="$(kubectl -n $METERING_NAMESPACE get subscriptions $SUBSCRIPTION_NAME -o yaml | faq -f yaml '.status.currentCSV')"

msg "Removing Metering Resource"
kube-remove \
    "$METERING_CR_FILE"

msg "Removing Metering Subscription"
kube-remove \
    "$OLM_MANIFESTS_DIR/metering.subscription.yaml"

msg "Removing Metering Operator Group"
kube-remove \
    "$OLM_MANIFESTS_DIR/metering.operatorgroup.yaml"

msg "Removing Metering Catalog Source Config"
kubectl delete -f \
    "$OLM_MANIFESTS_DIR/metering.catalogsourceconfig.yaml"

msg "Removing Metering Catalog Source Version"
kubectl -n $METERING_NAMESPACE delete csv $CSV_NAME
