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
        "$OLM_MANIFESTS_DIR/metering.catalogsourcecon***REMOVED***g.yaml" \
        > "$TMPDIR/metering.catalogsourcecon***REMOVED***g.yaml"

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
***REMOVED***

msg "Installing Custom Resource De***REMOVED***nitions"
kube-install "$CRD_DIR"

msg "Installing Metering Catalog Source Con***REMOVED***g"
kubectl apply -f \
    "$OLM_MANIFESTS_DIR/metering.catalogsourcecon***REMOVED***g.yaml"

msg "Installing Metering Operator Group"
kube-install \
    "$OLM_MANIFESTS_DIR/metering.operatorgroup.yaml"

msg "Installing Metering Subscription"
kube-install \
    "$OLM_MANIFESTS_DIR/metering.subscription.yaml"

msg "Installing Metering Resource"
kube-install \
    "$METERING_CR_FILE"

