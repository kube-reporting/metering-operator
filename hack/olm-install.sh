#!/bin/bash -e

ROOT_DIR=$(dirname "${BASH_SOURCE}")/..
source "${ROOT_DIR}/hack/common.sh"

kubectl create namespace "${METERING_NAMESPACE}" || true

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

msg "Installing Custom Resource Definitions"
kube-install "$CRD_DIR"

msg "Installing Metering Catalog Source Config"
kubectl apply -f \
    "$OLM_MANIFESTS_DIR/metering.catalogsourceconfig.yaml"

msg "Installing Metering Operator Group"
kube-install \
    "$OLM_MANIFESTS_DIR/metering.operatorgroup.yaml"

msg "Installing Metering Subscription"
kube-install \
    "$OLM_MANIFESTS_DIR/metering.subscription.yaml"

msg "Installing Metering Resource"
kube-install \
    "$METERING_CR_FILE"

